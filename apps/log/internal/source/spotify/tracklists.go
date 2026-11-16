package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"slices"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
	"github.com/pennane/pennanen.dev/apps/log/internal/source/syncs"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

const (
	CountTracks          = "tracks"
	CountTracklists      = "tracklists"
	recordingWithoutISRC = "spotify:"
)

type albumTracklist struct {
	Name   string `json:"name"`
	Tracks struct {
		Items []reading.TracklistEntry `json:"items"`
		Total int                      `json:"total"`
	} `json:"tracks"`
}

type tracklistToFill struct {
	albumID  int64
	trackIDs []string
}

func RecordingKey(t reading.SpotifyTrack) string {
	if t.ExternalIDs.ISRC != "" {
		return t.ExternalIDs.ISRC
	}
	return recordingWithoutISRC + t.ID
}

func fillTracklists(ctx context.Context, st *store.Store, c *Client, p *syncs.Progress) error {
	if err := fetchAlbumsWithoutTracklists(ctx, st, c, p); err != nil {
		return err
	}
	toFill, err := tracklistsOutOfStep(ctx, st, p)
	if err != nil || len(toFill) == 0 {
		return err
	}
	wanted := []string{}
	for _, list := range toFill {
		wanted = append(wanted, list.trackIDs...)
	}
	fetched, err := c.Tracks(ctx, slices.Compact(slices.Sorted(slices.Values(wanted))))
	if err != nil {
		p.Problem("%s: %v", record.SourceSpotifyTrack, err)
	}
	for _, raw := range fetched {
		if spotifyID, _, err := knowTrack(ctx, st, p, raw); err != nil {
			p.Problem("track %s: %v", spotifyID, err)
		}
	}
	for _, list := range toFill {
		unknown, err := unknownTracks(ctx, st, list.trackIDs)
		if err != nil {
			return err
		}
		if len(unknown) > 0 {
			p.Problem("album %d: %d of its %d tracks came back", list.albumID, len(list.trackIDs)-len(unknown), len(list.trackIDs))
			continue
		}
		p.Add(CountTracklists, 1)
		p.ChangedData()
	}
	return ctx.Err()
}

func unknownTracks(ctx context.Context, st *store.Store, trackIDs []string) ([]string, error) {
	unknown := []string{}
	for _, trackID := range trackIDs {
		_, err := st.FindByIdentity(ctx, record.SourceSpotifyTrackID, trackID)
		switch {
		case errors.Is(err, store.ErrNotFound):
			unknown = append(unknown, trackID)
		case err != nil:
			return nil, err
		}
	}
	return unknown, nil
}

func fetchAlbumsWithoutTracklists(ctx context.Context, st *store.Store, c *Client, p *syncs.Progress) error {
	albums, err := st.LoggedItemsNotDismissed(ctx, kind.Album)
	if err != nil {
		return err
	}
	itemOf := map[string]int64{}
	for _, albumID := range albums {
		claim, err := st.ClaimItemKeyed(ctx, albumID, record.SourceSpotifyAlbum)
		if errors.Is(err, store.ErrNotFound) {
			continue
		}
		if err != nil {
			return err
		}
		var album albumTracklist
		if json.Unmarshal(claim.Data, &album) == nil && album.Tracks.Total == 0 && len(album.Tracks.Items) == 0 {
			itemOf[claim.ExternalID] = albumID
		}
	}
	if len(itemOf) == 0 {
		return nil
	}
	fetched, err := c.Albums(ctx, slices.Sorted(maps.Keys(itemOf)))
	if err != nil {
		p.Problem("%s: %v", record.SourceSpotifyAlbum, err)
	}
	for _, raw := range fetched {
		var album struct {
			ID string `json:"id"`
			albumTracklist
		}
		if json.Unmarshal(raw, &album) != nil || len(album.Tracks.Items) == 0 {
			continue
		}
		albumID, asked := itemOf[album.ID]
		if !asked {
			continue
		}
		if err := st.PutClaimItemKeyed(ctx, albumID, record.SourceSpotifyAlbum, album.ID, raw); err != nil {
			p.Problem("%s %s: %v", record.SourceSpotifyAlbum, album.ID, err)
		}
	}
	return ctx.Err()
}

func tracklistsOutOfStep(ctx context.Context, st *store.Store, p *syncs.Progress) ([]tracklistToFill, error) {
	albums, err := st.LoggedItemsNotDismissed(ctx, kind.Album)
	if err != nil {
		return nil, err
	}
	toFill := []tracklistToFill{}
	for _, albumID := range albums {
		claim, err := st.ClaimItemKeyed(ctx, albumID, record.SourceSpotifyAlbum)
		if errors.Is(err, store.ErrNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}
		var album albumTracklist
		if err := json.Unmarshal(claim.Data, &album); err != nil {
			p.Problem("album %d: %v", albumID, err)
			continue
		}
		trackIDs := reading.InTracklistOrder(album.Tracks.Items)
		if album.Tracks.Total > len(trackIDs) {
			p.Problem("%s lists %d of its %d tracks", album.Name, len(trackIDs), album.Tracks.Total)
			continue
		}
		unknown, err := unknownTracks(ctx, st, trackIDs)
		if err != nil {
			return nil, err
		}
		if len(unknown) > 0 {
			toFill = append(toFill, tracklistToFill{albumID: albumID, trackIDs: trackIDs})
		}
	}
	return toFill, nil
}

func knowTrack(ctx context.Context, st *store.Store, p *syncs.Progress, raw json.RawMessage) (string, int64, error) {
	var t reading.SpotifyTrack
	if err := json.Unmarshal(raw, &t); err != nil {
		return "", 0, err
	}
	if t.ID == "" {
		return "", 0, errors.New("a spotify track without an id")
	}
	key := RecordingKey(t)
	itemID, err := st.FindByIdentity(ctx, record.SourceSpotifyTrack, key)
	if err == nil {
		return t.ID, itemID, st.PutClaimItemKeyed(ctx, itemID, record.SourceSpotifyTrack, key, raw)
	}
	if !errors.Is(err, store.ErrNotFound) {
		return t.ID, 0, err
	}
	itemID, err = st.KnowItemFromSource(ctx, kind.Track, t.Name, record.SourceSpotifyTrack, key, raw)
	if errors.Is(err, store.ErrConflict) {
		itemID, err = st.FindByIdentity(ctx, record.SourceSpotifyTrack, key)
	} else if err == nil {
		p.Add(CountTracks, 1)
	}
	return t.ID, itemID, err
}
