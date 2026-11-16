package store_test

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

func TestItemsLastObservedAs(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	a, err := s.KnowItemFromSource(ctx, kind.Album, "A", record.SourceSpotifyAlbum, "a", map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.KnowItemFromSource(ctx, kind.Album, "B", record.SourceSpotifyAlbum, "b", map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	observe := func(itemID int64, signal, key, at string) {
		t.Helper()
		if _, err := s.ObserveItemKeyed(ctx, store.Observation{ItemID: itemID, Source: "spotify", Signal: signal, Key: key, ObservedAt: at, Data: json.RawMessage(`{}`)}); err != nil {
			t.Fatal(err)
		}
	}
	observe(a, "saved", "saved:a", "2026-11-01T10:00:00Z")
	observe(a, "unsaved", "unsaved:a", "2026-11-02T10:00:00Z")
	observe(b, "saved", "saved:b", "2026-11-01T10:00:00Z")

	got, err := s.ItemsLastObservedAsItemKeyed(ctx, "spotify", []string{"saved", "unsaved"}, "saved")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, []int64{b}) {
		t.Fatalf("items = %v, want [%d]", got, b)
	}
}
