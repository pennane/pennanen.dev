package store_test

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
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

	err := s.Identify(ctx, ripley, store.Identity{Kind: kind.Series, Source: record.SourceTMDBSeries, ExternalID: "94028", Data: json.RawMessage(`{"id":94028,"name":"Ripley"}`)},
		[]string{record.SourceTMDB, record.SourceTMDBSeries})
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
	if found, err := s.FindByIdentity(ctx, record.SourceTMDBSeries, "94028"); err != nil || found != ripley {
		t.Errorf("identity: %d %v", found, err)
	}
}

func TestAnIdentityAnotherItemHasIsRefused(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	added, _ := s.CreateItemFromSource(ctx, kind.Series, "Ripley", record.SourceTMDBSeries, "94028", json.RawMessage(`{"id":94028}`))
	film, _ := s.CreateItem(ctx, kind.Movie, "Ripley", nil)

	err := s.Identify(ctx, film, store.Identity{Kind: kind.Series, Source: record.SourceTMDBSeries, ExternalID: "94028", Data: json.RawMessage(`{"id":94028}`)}, nil)
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

func TestAKindMoveSetsTheChoiceAsideAndMovingBackRestoresIt(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	series, err := s.CreateItemFromSource(ctx, kind.Series, "Ripley", record.SourceTMDBSeries, "94028", json.RawMessage(`{"id":94028,"name":"Ripley"}`))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ChangeEntry(ctx, series, store.EntryChange{Status: ptr(kind.Dropped)}); err != nil {
		t.Fatal(err)
	}
	moveTo := func(identity store.Identity) store.ItemView {
		t.Helper()
		if err := s.Identify(ctx, series, identity, []string{record.SourceTMDB, record.SourceTMDBSeries}); err != nil {
			t.Fatalf("moving to %s: %v", identity.Kind, err)
		}
		v, err := s.GetItem(ctx, series)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}

	film := moveTo(store.Identity{Kind: kind.Movie, Source: record.SourceTMDB, ExternalID: "603", Data: json.RawMessage(matrix)})
	setAside := &store.StatusChoice{Status: kind.Dropped, Standing: store.ChoiceSetAside}
	if film.Kind != kind.Movie || film.Entry.Status != kind.Untriaged || !sameChoice(film.Entry.StatusChoice, setAside) {
		t.Fatalf("as a film: %s %+v %+v", film.Kind, film.Entry, film.Entry.StatusChoice)
	}

	back := moveTo(store.Identity{Kind: kind.Series, Source: record.SourceTMDBSeries, ExternalID: "94028", Data: json.RawMessage(`{"id":94028,"name":"Ripley"}`)})
	decides := &store.StatusChoice{Status: kind.Dropped, Standing: store.ChoiceDecides}
	if back.Kind != kind.Series || back.Entry.Status != kind.Dropped || back.Entry.StatusDecidedBy != record.SourceAP || !sameChoice(back.Entry.StatusChoice, decides) {
		t.Fatalf("a series again: %s %+v %+v", back.Kind, back.Entry, back.Entry.StatusChoice)
	}
}

func sameChoice(got, want *store.StatusChoice) bool {
	return got != nil && got.Status == want.Status && got.Standing == want.Standing
}
