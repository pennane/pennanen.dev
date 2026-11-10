package store_test

import (
	"context"
	"errors"
	"maps"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
)

func rankedGames(t *testing.T, s *store.Store, titles ...string) map[string]int64 {
	t.Helper()
	ids := map[string]int64{}
	for _, title := range titles {
		id, err := s.CreateItem(context.Background(), "game", title, nil)
		if err != nil {
			t.Fatal(err)
		}
		ids[title] = id
	}
	return ids
}

func ranks(t *testing.T, s *store.Store) map[int64]int {
	t.Helper()
	views, err := s.ListItems(context.Background(), store.ItemFilter{Kind: "game"})
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

func enjoyed(division string) store.EntryChange {
	return store.EntryChange{Status: ptr("finished"), Scores: map[string]*string{kind.Enjoyment: &division}}
}

func TestRanksRunThroughDivisionsInOrder(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	g := rankedGames(t, s, "Meh One", "Loved One", "Loved Two", "Liked One")
	for title, division := range map[string]string{"Meh One": "meh", "Liked One": "liked"} {
		if err := s.ChangeEntry(ctx, g[title], enjoyed(division)); err != nil {
			t.Fatal(err)
		}
	}
	s.ChangeEntry(ctx, g["Loved One"], enjoyed("loved"))
	s.ChangeEntry(ctx, g["Loved Two"], enjoyed("loved"))

	want := map[int64]int{g["Loved One"]: 1, g["Loved Two"]: 2, g["Liked One"]: 3, g["Meh One"]: 4}
	if got := ranks(t, s); !maps.Equal(got, want) {
		t.Errorf("ranks = %v, want %v", got, want)
	}

	if err := s.Place(ctx, g["Meh One"], "loved", 0); err != nil {
		t.Fatal(err)
	}
	if err := s.Place(ctx, g["Loved One"], "loved", store.AtBottom); err != nil {
		t.Fatal(err)
	}
	want = map[int64]int{g["Meh One"]: 1, g["Loved Two"]: 2, g["Loved One"]: 3, g["Liked One"]: 4}
	if got := ranks(t, s); !maps.Equal(got, want) {
		t.Errorf("after moves ranks = %v, want %v", got, want)
	}

	s.ChangeEntry(ctx, g["Loved Two"], enjoyed("loved"))
	if got := ranks(t, s); !maps.Equal(got, want) {
		t.Errorf("same division moved it: %v", got)
	}

	if err := s.ChangeEntry(ctx, g["Meh One"], store.EntryChange{Status: ptr(kind.Dismissed)}); err != nil {
		t.Fatal(err)
	}
	want = map[int64]int{g["Loved Two"]: 1, g["Loved One"]: 2, g["Liked One"]: 3}
	if got := ranks(t, s); !maps.Equal(got, want) {
		t.Errorf("after dismissal ranks = %v, want %v", got, want)
	}
}

func TestDivisionsMustExistAndStayWhileUsed(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	g := rankedGames(t, s, "Hades")
	if err := s.ChangeEntry(ctx, g["Hades"], enjoyed("sublime")); !errors.Is(err, store.ErrInvalid) {
		t.Errorf("unknown division: %v", err)
	}
	if err := s.Place(ctx, g["Hades"], "sublime", 0); !errors.Is(err, store.ErrInvalid) {
		t.Errorf("placed into unknown division: %v", err)
	}
	if err := s.ChangeEntry(ctx, g["Hades"], enjoyed("loved")); err != nil {
		t.Fatal(err)
	}
	game, _ := kind.Get("game")
	settings, _ := s.KindSettings(ctx, game)
	settings.Divisions[kind.Enjoyment] = settings.Divisions[kind.Enjoyment][1:]
	if _, err := s.SetKindSettings(ctx, game, settings); !errors.Is(err, store.ErrConflict) {
		t.Errorf("removed a division in use: %v", err)
	}
	settings.Divisions[kind.Enjoyment] = append([]store.Division{{Key: "loved", Name: "all-time"}}, settings.Divisions[kind.Enjoyment]...)
	if _, err := s.SetKindSettings(ctx, game, settings); err != nil {
		t.Errorf("renaming a division: %v", err)
	}
}
