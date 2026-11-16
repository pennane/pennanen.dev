package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/source/syncs"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

const ObservationSource = record.SourceSpotify

const (
	CountArtists   = "artists"
	CountAlbums    = "albums"
	CountSnapshots = "snapshots"
	CountPlays     = "plays"
	CountSaves     = "saves"
	CountUnsaves   = "unsaves"
)

type Detail int

const (
	Simplified Detail = iota
	Full
	Kept
)

type metItem struct {
	itemID     int64
	externalID string
	detail     Detail
	raw        json.RawMessage
}

type reference struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Artists []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"artists"`
}

type Recorder struct {
	Store   *store.Store
	artists map[string]*metItem
	albums  map[string]*metItem
}

func NewRecorder(s *store.Store) *Recorder {
	return &Recorder{Store: s, artists: map[string]*metItem{}, albums: map[string]*metItem{}}
}

func (r *Recorder) KnowArtist(ctx context.Context, p *syncs.Progress, raw json.RawMessage, detail Detail) (int64, error) {
	return r.know(ctx, p, r.artists, kind.Artist, record.SourceSpotifyArtist, CountArtists, raw, detail)
}

func (r *Recorder) KnowAlbum(ctx context.Context, p *syncs.Progress, raw json.RawMessage, detail Detail) (int64, error) {
	itemID, err := r.know(ctx, p, r.albums, kind.Album, record.SourceSpotifyAlbum, CountAlbums, raw, detail)
	if err != nil {
		return 0, err
	}
	var album reference
	if err := json.Unmarshal(raw, &album); err != nil {
		return 0, err
	}
	if len(album.Artists) == 0 {
		r.keepStoredClaimIfItHasCredits(ctx, itemID, album.ID)
		return itemID, nil
	}
	for _, credited := range album.Artists {
		creditedRaw, err := json.Marshal(credited)
		if err != nil {
			return 0, err
		}
		if _, err := r.KnowArtist(ctx, p, creditedRaw, Simplified); err != nil {
			return 0, err
		}
	}
	return itemID, nil
}

func (r *Recorder) keepStoredClaimIfItHasCredits(ctx context.Context, itemID int64, albumID string) {
	seen, met := r.albums[albumID]
	if !met {
		return
	}
	var arriving albumCredits
	if json.Unmarshal(seen.raw, &arriving) != nil || len(arriving.Artists) > 0 {
		return
	}
	stored, err := r.Store.ClaimItemKeyed(ctx, itemID, record.SourceSpotifyAlbum)
	if err != nil {
		return
	}
	var kept albumCredits
	if json.Unmarshal(stored.Data, &kept) == nil && len(kept.Artists) > 0 && len(arriving.Tracks.Items) <= len(kept.Tracks.Items) {
		seen.detail = Kept
	}
}

type albumCredits struct {
	Artists []json.RawMessage `json:"artists"`
	Tracks  struct {
		Items []json.RawMessage `json:"items"`
	} `json:"tracks"`
}

func (r *Recorder) know(ctx context.Context, p *syncs.Progress, met map[string]*metItem, kindName, source, count string, raw json.RawMessage, detail Detail) (int64, error) {
	var ref reference
	if err := json.Unmarshal(raw, &ref); err != nil {
		return 0, err
	}
	if ref.ID == "" {
		return 0, fmt.Errorf("a spotify %s without an id", kindName)
	}
	if seen, ok := met[ref.ID]; ok {
		if detail == Full && seen.detail != Full {
			seen.detail, seen.raw = Full, raw
		}
		return seen.itemID, nil
	}
	known := detail
	itemID, err := r.Store.FindByIdentity(ctx, source, ref.ID)
	if err == nil && detail == Simplified && r.storedClaimIsFull(ctx, itemID, source) {
		known = Kept
	}
	if errors.Is(err, store.ErrNotFound) {
		itemID, err = r.Store.KnowItemFromSource(ctx, kindName, ref.Name, source, ref.ID, raw)
		switch {
		case errors.Is(err, store.ErrConflict):
			itemID, err = r.Store.FindByIdentity(ctx, source, ref.ID)
		case err == nil:
			p.Add(count, 1)
		}
	}
	if err != nil {
		return 0, err
	}
	met[ref.ID] = &metItem{itemID: itemID, externalID: ref.ID, detail: known, raw: raw}
	return itemID, nil
}

func (r *Recorder) Observe(ctx context.Context, p *syncs.Progress, itemID int64, signal, key, observedAt, count string, data any) (store.ObserveOutcome, error) {
	encoded, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	outcome, err := r.Store.ObserveItemKeyed(ctx, store.Observation{
		ItemID:     itemID,
		Source:     ObservationSource,
		Signal:     signal,
		Key:        key,
		ObservedAt: observedAt,
		Data:       encoded,
	})
	if outcome == store.Observed {
		p.Add(count, 1)
	}
	return outcome, err
}

func (r *Recorder) RefreshClaims(ctx context.Context, p *syncs.Progress, c *Client) {
	r.refresh(ctx, p, r.artists, record.SourceSpotifyArtist, c.Artists)
	r.refresh(ctx, p, r.albums, record.SourceSpotifyAlbum, c.Albums)
}

func stillSimplified(met map[string]*metItem) []string {
	left := []string{}
	for id, seen := range met {
		if seen.detail == Simplified {
			left = append(left, id)
		}
	}
	return left
}

func (r *Recorder) refresh(ctx context.Context, p *syncs.Progress, met map[string]*metItem, source string, fetch func(context.Context, []string) ([]json.RawMessage, error)) {
	missing := stillSimplified(met)
	if len(missing) > 0 {
		fetched, err := fetch(ctx, missing)
		if err != nil {
			p.Problem("%s: %v", source, err)
		}
		defer func() {
			if left := stillSimplified(met); len(left) > 0 {
				sort.Strings(left)
				p.Problem("%s never came back in full: %s", source, strings.Join(left, ", "))
			}
		}()
		for _, raw := range fetched {
			var ref reference
			if json.Unmarshal(raw, &ref) == nil && met[ref.ID] != nil {
				met[ref.ID].detail, met[ref.ID].raw = Full, raw
			}
		}
	}
	for id, seen := range met {
		if seen.detail != Full {
			continue
		}
		if err := r.Store.PutClaimItemKeyed(ctx, seen.itemID, source, id, seen.raw); err != nil {
			p.Problem("%s %s: %v", source, id, err)
		}
	}
}

func (r *Recorder) storedClaimIsFull(ctx context.Context, itemID int64, source string) bool {
	claim, err := r.Store.ClaimItemKeyed(ctx, itemID, source)
	if err != nil {
		return false
	}
	var stored struct {
		Popularity *int64 `json:"popularity"`
	}
	return json.Unmarshal(claim.Data, &stored) == nil && stored.Popularity != nil
}
