package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"slices"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/syncs"
	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
)

const (
	CountTracks          = "tracks"
	CountTracklists      = "tracklists"
	recordingWithoutISRC = "spotify:"
)

type albumTracklist struct {
	Name   string `json:"name"`
	Tracks struct {
		Items []verdict.TracklistEntry `json:"items"`
		Total int                      `json:"total"`
	} `json:"tracks"`
}

type tracklistToFill struct {
	albumID  int64
	trackIDs []string
}

func RecordingKey(t verdict.SpotifyTrack) string {
	if t.ExternalIDs.ISRC != "" {
		return t.ExternalIDs.ISRC
	}
	return recordingWithoutISRC + t.ID
}

func fillTracklists(ctx context.Context, st *store.Store, c *Client, p *syncs.Progress) error {
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
		p.Problem("%s: %v", kind.SourceSpotifyTrack, err)
	}
	trackItems := map[string]int64{}
	for _, raw := range fetched {
		spotifyID, itemID, err := knowTrack(ctx, st, p, raw)
		if err != nil {
			p.Problem("track %s: %v", spotifyID, err)
			continue
		}
		trackItems[spotifyID] = itemID
	}
	for _, list := range toFill {
		partIDs := make([]int64, 0, len(list.trackIDs))
		for _, trackID := range list.trackIDs {
			if itemID, known := trackItems[trackID]; known {
				partIDs = append(partIDs, itemID)
			}
		}
		if len(partIDs) != len(list.trackIDs) {
			p.Problem("album %d: %d of its %d tracks came back", list.albumID, len(partIDs), len(list.trackIDs))
			continue
		}
		if err := st.ReplaceParts(ctx, list.albumID, partIDs); err != nil {
			p.Problem("album %d: %v", list.albumID, err)
			continue
		}
		p.Add(CountTracklists, 1)
		p.ChangedData()
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
		claim, err := st.Claim(ctx, albumID, kind.SourceSpotifyAlbum)
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
		trackIDs := verdict.InTracklistOrder(album.Tracks.Items)
		if album.Tracks.Total > len(trackIDs) {
			p.Problem("%s lists %d of its %d tracks", album.Name, len(trackIDs), album.Tracks.Total)
			continue
		}
		parts, err := st.PartsOf(ctx, albumID)
		if err != nil {
			return nil, err
		}
		if len(parts) != len(trackIDs) {
			toFill = append(toFill, tracklistToFill{albumID: albumID, trackIDs: trackIDs})
		}
	}
	return toFill, nil
}

func knowTrack(ctx context.Context, st *store.Store, p *syncs.Progress, raw json.RawMessage) (string, int64, error) {
	var t verdict.SpotifyTrack
	if err := json.Unmarshal(raw, &t); err != nil {
		return "", 0, err
	}
	if t.ID == "" {
		return "", 0, errors.New("a spotify track without an id")
	}
	key := RecordingKey(t)
	itemID, err := st.FindByIdentity(ctx, kind.SourceSpotifyTrack, key)
	if err == nil {
		return t.ID, itemID, st.PutClaim(ctx, itemID, kind.SourceSpotifyTrack, key, raw)
	}
	if !errors.Is(err, store.ErrNotFound) {
		return t.ID, 0, err
	}
	itemID, err = st.KnowItemFromSource(ctx, kind.Track, t.Name, kind.SourceSpotifyTrack, key, raw)
	if errors.Is(err, store.ErrConflict) {
		itemID, err = st.FindByIdentity(ctx, kind.SourceSpotifyTrack, key)
	} else if err == nil {
		p.Add(CountTracks, 1)
	}
	return t.ID, itemID, err
}
