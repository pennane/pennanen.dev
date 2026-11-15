package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/spotify"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
)

func TestAnAlbumPageListsItsTracksInOrder(t *testing.T) {
	var st *store.Store
	srv := spotifyServer(t, func(s *store.Store) *spotify.Client { st = s; return nil })
	ctx := context.Background()
	album, err := st.CreateItemFromSource(ctx, kind.Album, "Beta", kind.SourceSpotifyAlbum, "B", map[string]any{"id": "B", "name": "Beta"})
	if err != nil {
		t.Fatal(err)
	}
	listened := kind.Listened
	if err := st.ChangeEntry(ctx, album, store.EntryChange{Status: &listened}); err != nil {
		t.Fatal(err)
	}
	second, _ := st.KnowItemFromSource(ctx, kind.Track, "Second", kind.SourceSpotifyTrack, "I2", map[string]any{"id": "t2", "name": "Second"})
	first, _ := st.KnowItemFromSource(ctx, kind.Track, "First", kind.SourceSpotifyTrack, "I1", map[string]any{"id": "t1", "name": "First"})
	if err := st.ReplaceParts(ctx, album, []int64{first, second}); err != nil {
		t.Fatal(err)
	}
	status, body := call(t, srv, "GET", "/public/items/album/beta", "", "")
	var item struct {
		Parts struct {
			Title string `json:"title"`
			Items []struct {
				Kind   string         `json:"kind"`
				Fields map[string]any `json:"fields"`
			} `json:"items"`
		} `json:"parts"`
	}
	if err := json.Unmarshal([]byte(body), &item); status != http.StatusOK || err != nil {
		t.Fatalf("%d %s %v", status, body, err)
	}
	parts := item.Parts.Items
	if item.Parts.Title != "tracks" || len(parts) != 2 || parts[0].Fields["title"] != "First" || parts[1].Fields["title"] != "Second" || parts[0].Kind != kind.Track {
		t.Errorf("parts %+v", item.Parts)
	}
}
