package store_test

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
)

func TestItemsLastObservedAs(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	a, err := knowItem(ctx, s, record.Identity{Source: record.SourceSpotifyAlbum, ExternalID: "a"}, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	b, err := knowItem(ctx, s, record.Identity{Source: record.SourceSpotifyAlbum, ExternalID: "b"}, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	observe := func(itemID int64, signal, key, at string) {
		t.Helper()
		if _, err := observeOnItem(ctx, s, store.Observation{ItemID: itemID, Source: "spotify", Signal: signal, Key: key, ObservedAt: at, Data: json.RawMessage(`{}`)}); err != nil {
			t.Fatal(err)
		}
	}
	observe(a, "saved", "saved:a", "2026-11-01T10:00:00Z")
	observe(a, "unsaved", "unsaved:a", "2026-11-02T10:00:00Z")
	observe(b, "saved", "saved:b", "2026-11-01T10:00:00Z")

	got, err := s.IdentitiesLastObservedAs(ctx, "spotify", []string{"saved", "unsaved"}, "saved")
	if err != nil {
		t.Fatal(err)
	}
	if want := []record.Identity{{Source: record.SourceSpotifyAlbum, ExternalID: "b"}}; !slices.Equal(got, want) {
		t.Fatalf("identities = %v, want %v", got, want)
	}
}
