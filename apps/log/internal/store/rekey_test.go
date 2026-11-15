package store_test

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
	"github.com/pennane/pennanen.dev/apps/log/internal/world"
)

const reloaded = `{"id":604,"title":"The Matrix Reloaded","release_date":"2003-05-15"}`

func TestCorrectingTheKeyFreesItAndKeepsApsEdits(t *testing.T) {
	corrections := map[string]func(s *store.Store, item int64) error{
		"identify": func(s *store.Store, item int64) error {
			return s.Identify(context.Background(), item, store.Identity{Kind: kind.Movie, Source: kind.SourceTMDB, ExternalID: "604", Data: json.RawMessage(reloaded)},
				[]string{kind.SourceTMDB, kind.SourceTMDBSeries})
		},
		"a pick": func(s *store.Store, item int64) error {
			return s.AnswerMatch(context.Background(), item, kind.SourceTMDB, "604", json.RawMessage(reloaded))
		},
	}
	for name, correct := range corrections {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			s, d := openStoreWithDB(t)
			corrected, err := s.CreateItemFromSource(ctx, kind.Movie, "The Matrix", kind.SourceTMDB, "603", json.RawMessage(matrix))
			if err != nil {
				t.Fatal(err)
			}
			if err := s.ChangeOverride(ctx, corrected, map[string]any{"title": "Reloaded, by ap"}); err != nil {
				t.Fatal(err)
			}
			if err := s.PutClaim(ctx, corrected, kind.SourceLetterboxdTake, "", json.RawMessage(`{"review":"twice as long"}`)); err != nil {
				t.Fatal(err)
			}
			other, err := s.CreateItem(ctx, kind.Movie, "Animatrix", nil)
			if err != nil {
				t.Fatal(err)
			}
			with := store.Relation{FromID: corrected, Type: kind.LinkWith, ToID: other}
			if err := s.AddRelation(ctx, with); err != nil {
				t.Fatal(err)
			}
			if _, err := s.Observe(ctx, store.Observation{ItemID: corrected, Source: kind.SourceLetterboxd, Signal: "watched", Key: "watched-once", ObservedAt: "2026-11-01", Data: json.RawMessage(`{}`)}); err != nil {
				t.Fatal(err)
			}

			if err := correct(s, corrected); err != nil {
				t.Fatal(err)
			}

			if owner, err := s.FindByIdentity(ctx, kind.SourceTMDB, "603"); !errors.Is(err, store.ErrNotFound) {
				t.Errorf("tmdb:603 still belongs to %d, %v", owner, err)
			}
			if owner, err := s.FindByIdentity(ctx, kind.SourceTMDB, "604"); err != nil || owner != corrected {
				t.Errorf("tmdb:604 belongs to %d, %v; want %d", owner, err, corrected)
			}
			claims := claimsStored(t, d)
			for _, moved := range []string{"ap:fields/tmdb:603", "letterboxd-take:tmdb:603", "ap:match/tmdb/tmdb:603"} {
				if _, kept := claims[moved]; kept {
					t.Errorf("%s is still keyed by the freed identity", moved)
				}
			}
			for _, kept := range []string{"ap:fields/tmdb:604", "letterboxd-take:tmdb:604", "ap:match/tmdb/tmdb:604"} {
				if _, found := claims[kept]; !found {
					t.Errorf("no %s", kept)
				}
			}
			v, err := s.GetItem(ctx, corrected)
			if err != nil {
				t.Fatal(err)
			}
			if v.Verdicts.Values["title"] != "Reloaded, by ap" {
				t.Errorf("the corrected item lost ap's edits: %v", v.Verdicts.Values)
			}
			if _, related := relationsOf(t, s, corrected)[with]; !related {
				t.Errorf("ap's link edit was lost: %v", relationsOf(t, s, corrected))
			}

			if watched, err := s.Observations(ctx, corrected, kind.SourceLetterboxd, "watched"); err != nil || len(watched) != 1 {
				t.Errorf("observations on the corrected item = %v, %v", watched, err)
			}

			real, err := s.CreateItemFromSource(ctx, kind.Movie, "The Matrix", kind.SourceTMDB, "603", json.RawMessage(matrix))
			if err != nil {
				t.Fatalf("the real tmdb:603 can't be added: %v", err)
			}
			if real == corrected {
				t.Fatal("the real tmdb:603 landed on the corrected item")
			}
			found, err := s.FindByIdentity(ctx, kind.SourceTMDB, "603")
			if err != nil || found != real {
				t.Fatalf("tmdb:603 belongs to %d, %v; want %d", found, err, real)
			}
			if err := s.PutClaim(ctx, found, kind.SourceLetterboxd, "2bg8", json.RawMessage(`{"id":"2bg8","title":"The Matrix","tmdbId":603}`)); err != nil {
				t.Fatal(err)
			}
			if owner, err := s.FindByIdentity(ctx, kind.SourceLetterboxd, "2bg8"); err != nil || owner != real {
				t.Errorf("the letterboxd film belongs to %d, %v; want %d", owner, err, real)
			}
			fresh, err := s.GetItem(ctx, real)
			if err != nil {
				t.Fatal(err)
			}
			if fresh.Verdicts.Values["title"] != "The Matrix" {
				t.Errorf("ap's edits followed the freed identity: %v", fresh.Verdicts.Values)
			}
			if got := relationsOf(t, s, real); len(got) != 0 {
				t.Errorf("ap's link edit followed the freed identity: %v", got)
			}
			if watched, err := s.Observations(ctx, real, kind.SourceLetterboxd, "watched"); err != nil || len(watched) != 0 {
				t.Errorf("observations followed the freed identity: %v, %v", watched, err)
			}
		})
	}
}

func TestNoMatchOnTheKeyRekeysOntoTheNextIdentity(t *testing.T) {
	ctx := context.Background()
	s, d := openStoreWithDB(t)
	game, err := s.CreateItemFromSource(ctx, kind.Game, "Wrong", kind.SourceIGDB, "7", json.RawMessage(`{"id":7,"name":"Wrong"}`))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.PutClaim(ctx, game, kind.SourceSteam, "220", json.RawMessage(`{"appid":220,"name":"Half-Life 2"}`)); err != nil {
		t.Fatal(err)
	}
	if err := s.ChangeOverride(ctx, game, map[string]any{"title": "HL2"}); err != nil {
		t.Fatal(err)
	}

	if err := s.AnswerNoMatch(ctx, game, kind.SourceIGDB); err != nil {
		t.Fatal(err)
	}

	if owner, err := s.FindByIdentity(ctx, kind.SourceIGDB, "7"); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("igdb:7 still belongs to %d, %v", owner, err)
	}
	claims := claimsStored(t, d)
	for _, kept := range []string{"ap:fields/steam:220", "ap:match/igdb/steam:220"} {
		if _, found := claims[kept]; !found {
			t.Errorf("no %s", kept)
		}
	}
	if about := claimData(t, d, "matched:steam:220")["about"]; !reflect.DeepEqual(about, []any{"steam:220", "steam:220"}) {
		t.Errorf("matched about = %v", about)
	}
	v, err := s.GetItem(ctx, game)
	if err != nil {
		t.Fatal(err)
	}
	if v.Verdicts.Values["title"] != "HL2" {
		t.Errorf("verdicts = %v", v.Verdicts.Values)
	}
	other, err := s.CreateItemFromSource(ctx, kind.Game, "Wrong", kind.SourceIGDB, "7", json.RawMessage(`{"id":7,"name":"Wrong"}`))
	if err != nil || other == game {
		t.Errorf("igdb:7 can't be added on its own: %d, %v", other, err)
	}
}

func TestACorrectedKeyAnotherSourceStillNamesStaysHeld(t *testing.T) {
	ctx := context.Background()
	s, d := openStoreWithDB(t)
	film, err := s.CreateItemFromSource(ctx, kind.Movie, "The Matrix", kind.SourceTMDB, "603", json.RawMessage(matrix))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.PutClaim(ctx, film, kind.SourceLetterboxd, "2bg8", json.RawMessage(`{"id":"2bg8","title":"The Matrix","tmdbId":603}`)); err != nil {
		t.Fatal(err)
	}

	if err := s.AnswerMatch(ctx, film, kind.SourceTMDB, "604", json.RawMessage(reloaded)); err != nil {
		t.Fatal(err)
	}

	if owner, err := s.FindByIdentity(ctx, kind.SourceTMDB, "603"); err != nil || owner != film {
		t.Errorf("tmdb:603 belongs to %d, %v; want %d while letterboxd names it", owner, err, film)
	}
	if _, kept := claimsStored(t, d)["ap:match/tmdb/tmdb:603"]; !kept {
		t.Error("the answer moved off the key that is still held")
	}
}

func TestRekeyingRewritesTheLegacyRecordsNamingTheKey(t *testing.T) {
	ctx := context.Background()
	d, err := db.Open(ctx, filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	records := record.New(d, verdict.ReadRecord, verdict.ReadersVersion)
	w := world.New(d, records)
	s := store.New(d, records, w)
	film, err := s.CreateItemFromSource(ctx, kind.Movie, "The Matrix", kind.SourceTMDB, "603", json.RawMessage(matrix))
	if err != nil {
		t.Fatal(err)
	}
	other, err := s.CreateItem(ctx, kind.Movie, "Animatrix", nil)
	if err != nil {
		t.Fatal(err)
	}
	otherKey, err := w.KeyOf(ctx, d, other)
	if err != nil {
		t.Fatal(err)
	}
	adopted := map[record.Identity]string{
		{Source: kind.SourceLegacy, ExternalID: "item/1"}: `{"about":["tmdb:603"],"links":[{"from":"tmdb:603","type":"with","to":"` + otherKey.String() + `"}]}`,
		{Source: kind.SourceLegacy, ExternalID: "item/2"}: `{"about":["` + otherKey.String() + `"],"links":[{"from":"` + otherKey.String() + `","type":"by","to":"tmdb:603"}]}`,
	}
	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	dirty := []record.Identity{}
	for id, data := range adopted {
		if _, err := records.PutClaim(ctx, tx, id, json.RawMessage(data)); err != nil {
			t.Fatal(err)
		}
		dirty = append(dirty, id)
	}
	if _, err := w.Settle(ctx, tx, dirty); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	if err := s.Identify(ctx, film, store.Identity{Kind: kind.Movie, Source: kind.SourceTMDB, ExternalID: "604", Data: json.RawMessage(reloaded)}, []string{kind.SourceTMDB}); err != nil {
		t.Fatal(err)
	}

	want := map[string]map[string]any{
		"legacy:item/1": {"about": []any{"tmdb:604"}, "links": []any{map[string]any{"from": "tmdb:604", "type": "with", "to": otherKey.String()}}},
		"legacy:item/2": {"about": []any{otherKey.String()}, "links": []any{map[string]any{"from": otherKey.String(), "type": "by", "to": "tmdb:604"}}},
	}
	for id, data := range want {
		if got := claimData(t, d, id); !reflect.DeepEqual(got, data) {
			t.Errorf("%s = %v, want %v", id, got, data)
		}
	}
	if owner, err := s.FindByIdentity(ctx, kind.SourceTMDB, "603"); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("tmdb:603 still belongs to %d, %v", owner, err)
	}
	if _, related := relationsOf(t, s, film)[store.Relation{FromID: film, Type: kind.LinkWith, ToID: other}]; !related {
		t.Errorf("film relations = %v", relationsOf(t, s, film))
	}
	real, err := s.CreateItemFromSource(ctx, kind.Movie, "The Matrix", kind.SourceTMDB, "603", json.RawMessage(matrix))
	if err != nil {
		t.Fatal(err)
	}
	if got := relationsOf(t, s, real); len(got) != 0 {
		t.Errorf("the adopted links followed the freed identity: %v", got)
	}
}
