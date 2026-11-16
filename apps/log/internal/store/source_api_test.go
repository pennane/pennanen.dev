package store_test

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
)

var (
	steamHalfLife = record.Identity{Source: record.SourceSteam, ExternalID: "220"}
	igdbEpisode   = record.Identity{Source: record.SourceIGDB, ExternalID: "7"}
)

const halfLife = `{"appid":220,"name":"Half-Life 2"}`

func TestPutClaimMakesKeyItem(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)

	if err := s.PutClaim(ctx, steamHalfLife, json.RawMessage(halfLife)); err != nil {
		t.Fatal(err)
	}

	if _, err := s.FindByIdentity(ctx, record.SourceSteam, "220"); err != nil {
		t.Fatalf("no item for %s: %v", steamHalfLife, err)
	}
	if err := s.PutClaim(ctx, record.Identity{Source: record.SourceSteam}, json.RawMessage(halfLife)); !errors.Is(err, store.ErrInvalid) {
		t.Errorf("blank external id: %v, want invalid", err)
	}
}

func TestPutClaimAboutLandsOnTheItem(t *testing.T) {
	ctx := context.Background()
	s, d := openStoreWithDB(t)
	itemID := steamGame(t, s, "220", "Half-Life 2")

	if err := s.PutClaimAbout(ctx, igdbEpisode, steamHalfLife, json.RawMessage(`{"id":7,"name":"Half-Life 2: Episode"}`)); err != nil {
		t.Fatal(err)
	}

	if owner, err := s.FindByIdentity(ctx, record.SourceIGDB, "7"); err != nil || owner != itemID {
		t.Errorf("igdb:7 belongs to %d, %v; want %d", owner, err, itemID)
	}
	if _, matched := claimsStored(t, d)["matched:igdb:7"]; !matched {
		t.Error("no matched record for igdb:7")
	}
}

func TestPutClaimAboutUnknownIsNotFound(t *testing.T) {
	ctx := context.Background()
	s, d := openStoreWithDB(t)
	before := claimsStored(t, d)

	err := s.PutClaimAbout(ctx, igdbEpisode, steamHalfLife, json.RawMessage(`{"id":7,"name":"X"}`))

	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("err = %v, want not found", err)
	}
	if after := claimsStored(t, d); len(after) != len(before) {
		t.Errorf("claims %v became %v", before, after)
	}
	if _, stored := claimsStored(t, d)["igdb:7"]; stored {
		t.Error("claim igdb:7 was written")
	}
}

func TestPutClaimUnchangedKeepsChangedAt(t *testing.T) {
	ctx := context.Background()
	s, d := openStoreWithDB(t)
	changedAt := func() string {
		var at string
		if err := d.QueryRowContext(ctx, `SELECT changed_at FROM record_claim WHERE source = ? AND external_id = ?`, steamHalfLife.Source, steamHalfLife.ExternalID).Scan(&at); err != nil {
			t.Fatal(err)
		}
		return at
	}
	if err := s.PutClaim(ctx, steamHalfLife, json.RawMessage(halfLife)); err != nil {
		t.Fatal(err)
	}
	first := changedAt()

	if err := s.PutClaim(ctx, steamHalfLife, json.RawMessage(halfLife)); err != nil {
		t.Fatal(err)
	}

	if second := changedAt(); second != first {
		t.Errorf("changed_at went from %s to %s", first, second)
	}
}

func TestRemoveClaimReleases(t *testing.T) {
	ctx := context.Background()
	s, d := openStoreWithDB(t)
	steamGame(t, s, "220", "Half-Life 2")
	if err := s.PutClaimAbout(ctx, igdbEpisode, steamHalfLife, json.RawMessage(`{"id":7,"name":"Half-Life 2: Episode"}`)); err != nil {
		t.Fatal(err)
	}

	if err := s.RemoveClaim(ctx, igdbEpisode); err != nil {
		t.Fatal(err)
	}

	if _, err := s.FindByIdentity(ctx, record.SourceIGDB, "7"); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("igdb:7 still held: %v", err)
	}
	for _, stored := range []string{"igdb:7", "matched:igdb:7"} {
		if _, there := claimsStored(t, d)[stored]; there {
			t.Errorf("%s still stored", stored)
		}
	}
	if err := s.RemoveClaim(ctx, igdbEpisode); err != nil {
		t.Errorf("removing an absent claim: %v", err)
	}
}

func TestIdentitiesLastObservedAs(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	a := record.Identity{Source: record.SourceSpotifyAlbum, ExternalID: "a"}
	b := record.Identity{Source: record.SourceSpotifyAlbum, ExternalID: "b"}
	observe := func(about record.Identity, signal, key, at string) {
		t.Helper()
		if _, err := s.Observe(ctx, record.Observation{About: about, Source: "spotify", Signal: signal, Key: key, ObservedAt: at, Data: json.RawMessage(`{}`)}); err != nil {
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
	if !slices.Equal(got, []record.Identity{b}) {
		t.Fatalf("identities = %v, want [%v]", got, b)
	}
}
