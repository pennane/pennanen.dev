package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/source/spotify"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

func TestAnAlbumPageListsItsTracksInOrder(t *testing.T) {
	var st *store.Store
	srv := spotifyServer(t, func(s *store.Store) *spotify.Client { st = s; return nil })
	ctx := context.Background()
	album, err := st.CreateItemFromSource(ctx, kind.Album, "Beta", record.SourceSpotifyAlbum, "B", json.RawMessage(`{"id":"B","name":"Beta","tracks":{"items":[{"id":"t1","disc_number":1,"track_number":1},{"id":"t2","disc_number":1,"track_number":2}]}}`))
	if err != nil {
		t.Fatal(err)
	}
	listened := kind.Listened
	if err := st.ChangeEntry(ctx, album, store.EntryChange{Status: &listened}); err != nil {
		t.Fatal(err)
	}
	second, _ := knowItem(ctx, st, record.Identity{Source: record.SourceSpotifyTrack, ExternalID: "I2"}, map[string]any{"id": "t2", "name": "Second"})
	first, _ := knowItem(ctx, st, record.Identity{Source: record.SourceSpotifyTrack, ExternalID: "I1"}, map[string]any{"id": "t1", "name": "First"})
	if first == 0 || second == 0 {
		t.Fatalf("tracks %d %d", first, second)
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
