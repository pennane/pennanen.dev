package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/syncs"
)

const SyncName = "spotify"

type Syncer struct {
	Store  *store.Store
	Client *Client
	Now    func() time.Time
}

func NewSyncer(s *store.Store, c *Client) *Syncer {
	return &Syncer{Store: s, Client: c, Now: time.Now}
}

type snapshot struct {
	recorder *Recorder
	p        *syncs.Progress
	start    string
	day      string
}

type track struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Album json.RawMessage `json:"album"`
}

func (s *Syncer) Work(ctx context.Context, p *syncs.Progress) error {
	now := s.Now().UTC()
	snap := snapshot{
		recorder: NewRecorder(s.Store),
		p:        p,
		start:    now.Format(time.RFC3339),
		day:      now.Format(time.DateOnly),
	}
	for _, timeRange := range TimeRanges {
		artists, err := s.Client.TopArtists(ctx, timeRange)
		if err != nil {
			return err
		}
		p.ChangedData()
		snap.topArtists(ctx, timeRange, artists)
	}
	for _, timeRange := range TimeRanges {
		tracks, err := s.Client.TopTracks(ctx, timeRange)
		if err != nil {
			return err
		}
		p.ChangedData()
		snap.topTracks(ctx, timeRange, tracks)
	}
	plays, err := s.Client.RecentlyPlayed(ctx)
	if err != nil {
		return err
	}
	p.ChangedData()
	snap.plays(ctx, plays)
	saved, err := s.Client.SavedAlbums(ctx)
	if err != nil {
		return err
	}
	snap.saves(ctx, saved)
	if err := s.unsaves(ctx, snap, saved); err != nil {
		return err
	}
	snap.recorder.RefreshClaims(ctx, p, s.Client)
	return ctx.Err()
}

func (n snapshot) topArtists(ctx context.Context, timeRange TimeRange, artists []json.RawMessage) {
	for i, raw := range artists {
		itemID, err := n.recorder.KnowArtist(ctx, n.p, raw, Full)
		if err != nil {
			n.p.Problem("top artist: %v", err)
			continue
		}
		var ref reference
		if err := json.Unmarshal(raw, &ref); err != nil {
			n.p.Problem("top artist: %v", err)
			continue
		}
		key := fmt.Sprintf("%s:%s:%s:%s", SignalTopArtist, timeRange, n.day, ref.ID)
		data := map[string]any{"range": timeRange, "rank": i + 1, "day": n.day}
		if err := n.recorder.Observe(ctx, n.p, itemID, SignalTopArtist, key, n.start, CountSnapshots, data); err != nil {
			n.p.Problem("%s: %v", ref.Name, err)
		}
	}
}

func (n snapshot) topTracks(ctx context.Context, timeRange TimeRange, tracks []json.RawMessage) {
	for i, raw := range tracks {
		var t track
		if err := json.Unmarshal(raw, &t); err != nil {
			n.p.Problem("top track: %v", err)
			continue
		}
		itemID, err := n.recorder.KnowAlbum(ctx, n.p, t.Album, Simplified)
		if err != nil {
			n.p.Problem("top track %s: %v", t.Name, err)
			continue
		}
		key := fmt.Sprintf("%s:%s:%s:%s", SignalTopTrack, timeRange, n.day, t.ID)
		data := map[string]any{"range": timeRange, "rank": i + 1, "day": n.day, "trackId": t.ID, "trackName": t.Name}
		if err := n.recorder.Observe(ctx, n.p, itemID, SignalTopTrack, key, n.start, CountSnapshots, data); err != nil {
			n.p.Problem("top track %s: %v", t.Name, err)
		}
	}
}

func (n snapshot) plays(ctx context.Context, plays []Play) {
	for _, play := range plays {
		var t track
		if err := json.Unmarshal(play.Track, &t); err != nil {
			n.p.Problem("play: %v", err)
			continue
		}
		itemID, err := n.recorder.KnowAlbum(ctx, n.p, t.Album, Simplified)
		if err != nil {
			n.p.Problem("play %s: %v", t.Name, err)
			continue
		}
		key := fmt.Sprintf("%s:%s:%s", SignalPlay, play.PlayedAt, t.ID)
		data := map[string]any{"trackId": t.ID, "trackName": t.Name, "playedAt": play.PlayedAt}
		if err := n.recorder.Observe(ctx, n.p, itemID, SignalPlay, key, play.PlayedAt, CountPlays, data); err != nil {
			n.p.Problem("play %s: %v", t.Name, err)
		}
	}
}

func (n snapshot) saves(ctx context.Context, saved []SavedAlbum) {
	for _, save := range saved {
		itemID, err := n.recorder.KnowAlbum(ctx, n.p, save.Album, Full)
		if err != nil {
			n.p.Problem("saved album: %v", err)
			continue
		}
		var ref reference
		if err := json.Unmarshal(save.Album, &ref); err != nil {
			n.p.Problem("saved album: %v", err)
			continue
		}
		key := fmt.Sprintf("%s:%s:%s", SignalSaved, ref.ID, save.AddedAt)
		if err := n.recorder.Observe(ctx, n.p, itemID, SignalSaved, key, save.AddedAt, CountSaves, map[string]any{"addedAt": save.AddedAt}); err != nil {
			n.p.Problem("saved %s: %v", ref.Name, err)
		}
	}
}

func (s *Syncer) unsaves(ctx context.Context, n snapshot, saved []SavedAlbum) error {
	stillSaved := map[string]struct{}{}
	for _, save := range saved {
		var ref reference
		if err := json.Unmarshal(save.Album, &ref); err == nil {
			stillSaved[ref.ID] = struct{}{}
		}
	}
	wasSaved, err := s.Store.ItemsLastObservedAs(ctx, ObservationSource, []string{SignalSaved, SignalUnsaved}, SignalSaved)
	if err != nil {
		return err
	}
	for _, itemID := range wasSaved {
		claim, err := s.Store.Claim(ctx, itemID, kind.SourceSpotifyAlbum)
		if errors.Is(err, store.ErrNotFound) {
			continue
		}
		if err != nil {
			return err
		}
		if _, still := stillSaved[claim.ExternalID]; still {
			continue
		}
		key := fmt.Sprintf("%s:%s:%s", SignalUnsaved, claim.ExternalID, n.day)
		if err := n.recorder.Observe(ctx, n.p, itemID, SignalUnsaved, key, n.start, CountUnsaves, map[string]any{"day": n.day}); err != nil {
			n.p.Problem("unsaved %s: %v", claim.ExternalID, err)
		}
	}
	return nil
}
