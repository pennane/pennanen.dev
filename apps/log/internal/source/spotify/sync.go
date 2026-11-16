package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/source/syncs"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
)

const (
	SyncName      = "spotify"
	PlaysSyncName = "spotify-plays"
)

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

func (s *Syncer) snapshotNow(p *syncs.Progress) snapshot {
	now := s.Now().UTC()
	return snapshot{
		recorder: NewRecorder(s.Store),
		p:        p,
		start:    now.Format(time.RFC3339),
		day:      now.Format(time.DateOnly),
	}
}

func (s *Syncer) WorkPlays(ctx context.Context, p *syncs.Progress) error {
	snap := s.snapshotNow(p)
	plays, err := s.Client.RecentlyPlayed(ctx)
	if errors.Is(err, ErrNotConnected) {
		return nil
	}
	if err != nil {
		return err
	}
	snap.plays(ctx, plays)
	snap.recorder.RefreshClaims(ctx, p, s.Client)
	return ctx.Err()
}

func (s *Syncer) Work(ctx context.Context, p *syncs.Progress) error {
	snap := s.snapshotNow(p)
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
	saved, err := s.Client.SavedAlbums(ctx)
	if err != nil {
		return err
	}
	stillSaved, err := savedIDs(saved)
	if err != nil {
		return err
	}
	if err := s.saves(ctx, snap, saved); err != nil {
		return err
	}
	if err := s.unsaves(ctx, snap, stillSaved); err != nil {
		return err
	}
	snap.recorder.RefreshClaims(ctx, p, s.Client)
	if err := s.Store.Propose(ctx, record.SourceSpotify); err != nil {
		p.Problem("logging what the rules say before weighing: %v", err)
	}
	if err := computeListening(ctx, s.Store, p); err != nil {
		return err
	}
	if err := fillTracklists(ctx, s.Store, s.Client, p); err != nil {
		return err
	}
	return ctx.Err()
}

func (n snapshot) listAlreadyRecorded(ctx context.Context, signal string, timeRange TimeRange) bool {
	prefix := fmt.Sprintf("%s:%s:%s", signal, timeRange, n.day)
	recorded, err := n.recorder.Store.HasObservationKeyWithPrefix(ctx, ObservationSource, prefix)
	if err != nil {
		n.p.Problem("%s %s: %v", signal, timeRange, err)
		return true
	}
	return recorded
}

func (n snapshot) topArtists(ctx context.Context, timeRange TimeRange, artists []json.RawMessage) {
	skip := n.listAlreadyRecorded(ctx, record.SignalTopArtist, timeRange) || n.listAlreadyRecorded(ctx, record.SignalTopArtistsListed, timeRange)
	for i, raw := range artists {
		about, err := n.recorder.KnowArtist(ctx, n.p, raw, Full)
		if err != nil {
			n.p.Problem("top artist: %v", err)
			continue
		}
		if skip {
			continue
		}
		var ref reference
		if err := json.Unmarshal(raw, &ref); err != nil {
			n.p.Problem("top artist: %v", err)
			continue
		}
		key := fmt.Sprintf("%s:%s:%s:%s", record.SignalTopArtist, timeRange, n.day, ref.ID)
		data := map[string]any{"range": timeRange, "rank": i + 1, "day": n.day}
		if _, err := n.recorder.Observe(ctx, n.p, about, record.SignalTopArtist, key, n.start, CountSnapshots, data); err != nil {
			n.p.Problem("%s: %v", ref.Name, err)
		}
	}
	if !skip {
		n.topArtistsListed(ctx, timeRange, len(artists))
	}
}

func (n snapshot) topArtistsListed(ctx context.Context, timeRange TimeRange, count int) {
	data, err := json.Marshal(map[string]any{"range": timeRange, "day": n.day, "count": count})
	if err != nil {
		n.p.Problem("top artists listed %s: %v", timeRange, err)
		return
	}
	_, err = n.recorder.Store.Observe(ctx, record.Observation{
		About:      record.Identity{Source: ObservationSource, ExternalID: fmt.Sprintf("%s:%s", record.SignalTopArtistsListed, timeRange)},
		Source:     ObservationSource,
		Signal:     record.SignalTopArtistsListed,
		Key:        fmt.Sprintf("%s:%s:%s", record.SignalTopArtistsListed, timeRange, n.day),
		ObservedAt: n.start,
		Data:       data,
	})
	if err != nil {
		n.p.Problem("top artists listed %s: %v", timeRange, err)
	}
}

func (n snapshot) topTracks(ctx context.Context, timeRange TimeRange, tracks []json.RawMessage) {
	skip := n.listAlreadyRecorded(ctx, record.SignalTopTrack, timeRange)
	for i, raw := range tracks {
		var t track
		if err := json.Unmarshal(raw, &t); err != nil {
			n.p.Problem("top track: %v", err)
			continue
		}
		about, err := n.recorder.KnowAlbum(ctx, n.p, t.Album, Simplified)
		if err != nil {
			n.p.Problem("top track %s: %v", t.Name, err)
			continue
		}
		if skip {
			continue
		}
		key := fmt.Sprintf("%s:%s:%s:%s", record.SignalTopTrack, timeRange, n.day, t.ID)
		data := map[string]any{"range": timeRange, "rank": i + 1, "day": n.day, "trackId": t.ID, "trackName": t.Name}
		if _, err := n.recorder.Observe(ctx, n.p, about, record.SignalTopTrack, key, n.start, CountSnapshots, data); err != nil {
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
		about, err := n.recorder.KnowAlbum(ctx, n.p, t.Album, Simplified)
		if err != nil {
			n.p.Problem("play %s: %v", t.Name, err)
			continue
		}
		key := fmt.Sprintf("%s:%s:%s", record.SignalPlay, play.PlayedAt, t.ID)
		data := map[string]any{"trackId": t.ID, "trackName": t.Name, "playedAt": play.PlayedAt}
		outcome, err := n.recorder.Observe(ctx, n.p, about, record.SignalPlay, key, play.PlayedAt, CountPlays, data)
		if err != nil {
			n.p.Problem("play %s: %v", t.Name, err)
			continue
		}
		if outcome == store.Observed {
			n.p.ChangedData()
		}
	}
}

func (s *Syncer) saves(ctx context.Context, n snapshot, saved []SavedAlbum) error {
	unsavedAlbums, err := s.Store.IdentitiesLastObservedAs(ctx, ObservationSource, []string{record.SignalSaved, record.SignalUnsaved}, record.SignalUnsaved)
	if err != nil {
		return err
	}
	wasUnsaved := map[record.Identity]struct{}{}
	for _, album := range unsavedAlbums {
		wasUnsaved[album] = struct{}{}
	}
	n.recordSaves(ctx, saved, wasUnsaved)
	return nil
}

func (n snapshot) recordSaves(ctx context.Context, saved []SavedAlbum, wasUnsaved map[record.Identity]struct{}) {
	for _, save := range saved {
		about, err := n.recorder.KnowAlbum(ctx, n.p, save.Album, Full)
		if err != nil {
			n.p.Problem("saved album: %v", err)
			continue
		}
		var ref reference
		if err := json.Unmarshal(save.Album, &ref); err != nil {
			n.p.Problem("saved album: %v", err)
			continue
		}
		key := fmt.Sprintf("%s:%s:%s", record.SignalSaved, ref.ID, save.AddedAt)
		if _, err := n.recorder.Observe(ctx, n.p, about, record.SignalSaved, key, save.AddedAt, CountSaves, map[string]any{"addedAt": save.AddedAt}); err != nil {
			n.p.Problem("saved %s: %v", ref.Name, err)
		}
		if _, unsaved := wasUnsaved[about]; unsaved {
			resavedKey := fmt.Sprintf("%s:%s:%s:%s", record.SignalSaved, ref.ID, save.AddedAt, n.day)
			if _, err := n.recorder.Observe(ctx, n.p, about, record.SignalSaved, resavedKey, n.start, CountSaves, map[string]any{"addedAt": save.AddedAt}); err != nil {
				n.p.Problem("saved again %s: %v", ref.Name, err)
			}
		}
	}
}

func savedIDs(saved []SavedAlbum) (map[string]struct{}, error) {
	ids := map[string]struct{}{}
	for _, save := range saved {
		var ref reference
		if err := json.Unmarshal(save.Album, &ref); err != nil {
			return nil, fmt.Errorf("saved albums: %w", err)
		}
		if ref.ID == "" {
			return nil, errors.New("saved albums: an album without an id")
		}
		ids[ref.ID] = struct{}{}
	}
	return ids, nil
}

func (s *Syncer) unsaves(ctx context.Context, n snapshot, stillSaved map[string]struct{}) error {
	wasSaved, err := s.Store.IdentitiesLastObservedAs(ctx, ObservationSource, []string{record.SignalSaved, record.SignalUnsaved}, record.SignalSaved)
	if err != nil {
		return err
	}
	for _, album := range wasSaved {
		if _, still := stillSaved[album.ExternalID]; still {
			continue
		}
		key := fmt.Sprintf("%s:%s:%s", record.SignalUnsaved, album.ExternalID, n.day)
		if _, err := n.recorder.Observe(ctx, n.p, album, record.SignalUnsaved, key, n.start, CountUnsaves, map[string]any{"day": n.day}); err != nil {
			n.p.Problem("unsaved %s: %v", album.ExternalID, err)
		}
	}
	return nil
}
