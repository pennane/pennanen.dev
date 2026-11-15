package store_test

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
)

func TestItemsLastObservedAs(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	a, err := s.KnowItemFromSource(ctx, kind.Album, "A", kind.SourceSpotifyAlbum, "a", map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.KnowItemFromSource(ctx, kind.Album, "B", kind.SourceSpotifyAlbum, "b", map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	observe := func(itemID int64, signal, key, at string) {
		t.Helper()
		if _, err := s.Observe(ctx, store.Observation{ItemID: itemID, Source: "spotify", Signal: signal, Key: key, ObservedAt: at, Data: json.RawMessage(`{}`)}); err != nil {
			t.Fatal(err)
		}
	}
	observe(a, "saved", "saved:a", "2026-11-01T10:00:00Z")
	observe(a, "unsaved", "unsaved:a", "2026-11-02T10:00:00Z")
	observe(b, "saved", "saved:b", "2026-11-01T10:00:00Z")

	got, err := s.ItemsLastObservedAs(ctx, "spotify", []string{"saved", "unsaved"}, "saved")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, []int64{b}) {
		t.Fatalf("items = %v, want [%d]", got, b)
	}
}

func TestNewestSnapshotIsByDay(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	older, err := s.KnowItemFromSource(ctx, kind.Artist, "Older", kind.SourceSpotifyArtist, "older", map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	newer, err := s.KnowItemFromSource(ctx, kind.Artist, "Newer", kind.SourceSpotifyArtist, "newer", map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	observe := func(itemID int64, key, at string) {
		t.Helper()
		if _, err := s.Observe(ctx, store.Observation{ItemID: itemID, Source: "spotify", Signal: "top-artist", Key: key, ObservedAt: at, Data: json.RawMessage(`{}`)}); err != nil {
			t.Fatal(err)
		}
	}
	observe(older, "top-artist:short_term:2026-11-14:older", "2026-11-14T03:00:00Z")
	observe(newer, "top-artist:long_term:2026-11-15:newer", "2026-11-15T03:00:00Z")

	got, err := s.ItemsInNewestSnapshot(ctx, "spotify", "top-artist", "top-artist:short_term:")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("short_term items = %v, want none: the newest day has no short_term rows", got)
	}

	observe(newer, "top-artist:short_term:2026-11-15:newer", "2026-11-15T03:00:00Z")
	got, err = s.ItemsInNewestSnapshot(ctx, "spotify", "top-artist", "top-artist:short_term:")
	if err != nil || !slices.Equal(got, []int64{newer}) {
		t.Fatalf("short_term items = %v, %v, want [%d]", got, err, newer)
	}

	everObserved, err := s.ItemsObserved(ctx, "spotify", "top-artist")
	if err != nil || !slices.Equal(everObserved, []int64{older, newer}) {
		t.Fatalf("ever observed = %v, %v", everObserved, err)
	}
}
