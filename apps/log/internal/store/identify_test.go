package store_test

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
)

func ranksIn(t *testing.T, s *store.Store, kindName string) map[int64]int {
	t.Helper()
	views, err := s.ListItems(context.Background(), store.ItemFilter{Kind: kindName})
	if err != nil {
		t.Fatal(err)
	}
	got := map[int64]int{}
	for _, v := range views {
		if v.Entry != nil && v.Entry.Rank > 0 {
			got[v.ID] = v.Entry.Rank
		}
	}
	return got
}

func TestAFilmFoundToBeASeriesMovesWithItsEntry(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	watched := func(kindName, title, division string) int64 {
		t.Helper()
		id, err := s.CreateItem(ctx, kindName, title, nil)
		if err != nil {
			t.Fatal(err)
		}
		change := store.EntryChange{Status: ptr(kind.Watched), Scores: map[string]*string{kind.Prestige: &division, kind.Quality: ptr("b")}, AddTags: []string{"another-shot"}}
		if err := s.ChangeEntry(ctx, id, change); err != nil {
			t.Fatal(err)
		}
		return id
	}
	first := watched(kind.Movie, "First Film", "a")
	ripley := watched(kind.Movie, "Ripley", "a")
	last := watched(kind.Movie, "Last Film", "a")
	ranked := watched(kind.Series, "Ranked Series", "a")
	namesake := watched(kind.Series, "Ripley", "s")

	err := s.Identify(ctx, ripley, store.Identity{Kind: kind.Series, Source: kind.SourceTMDBSeries, ExternalID: "94028", Data: json.RawMessage(`{"id":94028,"name":"Ripley"}`)},
		[]string{kind.SourceTMDB, kind.SourceTMDBSeries})
	if err != nil {
		t.Fatal(err)
	}

	v, err := s.GetItem(ctx, ripley)
	if err != nil {
		t.Fatal(err)
	}
	if v.Kind != kind.Series || v.Slug != "ripley-2" {
		t.Errorf("moved item: %s/%s", v.Kind, v.Slug)
	}
	if v.Entry.Status != kind.Watched || v.Entry.Scores[kind.Prestige] != "a" || v.Entry.Scores[kind.Quality] != "b" || len(v.Entry.Tags) != 1 {
		t.Errorf("entry didn't move along: %+v", v.Entry)
	}
	if got, want := ranksIn(t, s, kind.Movie), map[int64]int{first: 1, last: 2}; !maps.Equal(got, want) {
		t.Errorf("films left behind = %v, want %v", got, want)
	}
	if got, want := ranksIn(t, s, kind.Series), map[int64]int{namesake: 1, ranked: 2, ripley: 3}; !maps.Equal(got, want) {
		t.Errorf("series = %v, want %v", got, want)
	}
	if found, err := s.FindByIdentity(ctx, kind.SourceTMDBSeries, "94028"); err != nil || found != ripley {
		t.Errorf("identity: %d %v", found, err)
	}
}

func TestAnIdentityAnotherItemHasIsRefused(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	added, _ := s.CreateItemFromSource(ctx, kind.Series, "Ripley", kind.SourceTMDBSeries, "94028", json.RawMessage(`{"id":94028}`))
	film, _ := s.CreateItem(ctx, kind.Movie, "Ripley", nil)

	err := s.Identify(ctx, film, store.Identity{Kind: kind.Series, Source: kind.SourceTMDBSeries, ExternalID: "94028", Data: json.RawMessage(`{"id":94028}`)}, nil)
	if !errors.Is(err, store.ErrConflict) {
		t.Fatalf("err = %v", err)
	}
	if v, _ := s.GetItem(ctx, film); v.Kind != kind.Movie {
		t.Errorf("the film moved anyway: %s", v.Kind)
	}
	if v, _ := s.GetItem(ctx, added); v.Kind != kind.Series {
		t.Errorf("the series changed: %s", v.Kind)
	}
}

func TestASeriesWhoseStatusAFilmCannotHaveIsNotMovedToFilms(t *testing.T) {
	ctx := context.Background()
	s, d := openStoreWithDB(t)
	series, err := s.CreateItemFromSource(ctx, kind.Series, "Ripley", kind.SourceTMDBSeries, "94028", json.RawMessage(`{"id":94028,"name":"Ripley"}`))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ChangeEntry(ctx, series, store.EntryChange{Status: ptr(kind.Dropped)}); err != nil {
		t.Fatal(err)
	}

	err = s.Identify(ctx, series, store.Identity{Kind: kind.Movie, Source: kind.SourceTMDB, ExternalID: "603", Data: json.RawMessage(matrix)},
		[]string{kind.SourceTMDB, kind.SourceTMDBSeries})
	if !errors.Is(err, store.ErrInvalid) {
		t.Fatalf("err = %v, want ErrInvalid", err)
	}

	v, err := s.GetItem(ctx, series)
	if err != nil {
		t.Fatal(err)
	}
	if v.Kind != kind.Series || v.Entry.Status != kind.Dropped {
		t.Errorf("item = %s with status %s", v.Kind, v.Entry.Status)
	}
	if _, kept := claimsStored(t, d)["tmdb-tv:94028"]; !kept {
		t.Error("the tmdb-tv claim was removed")
	}
	if owner, err := s.FindByIdentity(ctx, kind.SourceTMDBSeries, "94028"); err != nil || owner != series {
		t.Errorf("tmdb-tv:94028 belongs to %d, %v", owner, err)
	}
}
