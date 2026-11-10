package store_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
)

func openStore(t *testing.T) *store.Store {
	t.Helper()
	d, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	return store.New(d)
}

func ptr[T any](v T) *T { return &v }

func TestOverrideBeatsSourceAndSurvivesResync(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	id, err := s.CreateItemFromSource(ctx, "game", "Hades", kind.SourceSteam, "1145360",
		json.RawMessage(`{"appid":1145360,"name":"Hades","playtime_forever":3000}`))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ChangeOverride(ctx, id, map[string]any{"title": "HADES"}); err != nil {
		t.Fatal(err)
	}
	if err := s.PutClaim(ctx, id, kind.SourceSteam, "1145360",
		json.RawMessage(`{"appid":1145360,"name":"Hades","playtime_forever":3100}`)); err != nil {
		t.Fatal(err)
	}
	v, err := s.GetItem(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if v.Verdicts.Values["title"] != "HADES" || v.Verdicts.SourceOf["title"] != kind.SourceAP {
		t.Errorf("title verdict = %v from %v", v.Verdicts.Values["title"], v.Verdicts.SourceOf["title"])
	}
	if v.Verdicts.Values["playtime_minutes"] != int64(3100) {
		t.Errorf("playtime = %v", v.Verdicts.Values["playtime_minutes"])
	}
	if v.Slug != "hades" || v.Entry == nil || v.Entry.Status != kind.Untriaged {
		t.Errorf("unexpected item %+v", v)
	}
}

func TestIdleHintFromCardsAndNoAchievements(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	id, _ := s.CreateItemFromSource(ctx, "game", "Card Farm", kind.SourceSteam, "1",
		json.RawMessage(`{"appid":1,"name":"Card Farm","playtime_forever":150}`))
	s.PutClaim(ctx, id, kind.SourceSteamStore, "1", json.RawMessage(`{"name":"Card Farm","categories":[{"description":"Steam Trading Cards"}]}`))
	s.PutClaim(ctx, id, kind.SourceSteamAchievements, "1", json.RawMessage(`{"achievements":[{"achieved":0}]}`))
	idled, err := s.ListItems(ctx, store.ItemFilter{Hint: kind.HintLikelyIdled})
	if err != nil || len(idled) != 1 {
		t.Fatalf("idled = %d, %v", len(idled), err)
	}
	if err := s.ChangeEntries(ctx, []int64{id}, store.EntryChange{Status: ptr("never-played")}); err != nil {
		t.Fatal(err)
	}
	v, _ := s.GetItem(ctx, id)
	if v.Hint != kind.NoHint {
		t.Errorf("triaged entry still hinted %q", v.Hint)
	}
}

func TestEntryRulesFollowKind(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	id, err := s.CreateItem(ctx, "thing", "A Walk", map[string]any{"mood": "foggy"})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ChangeEntry(ctx, id, store.EntryChange{Status: ptr("finished")}); err == nil {
		t.Error("thing accepted a game status")
	}
	if err := s.ChangeEntry(ctx, id, store.EntryChange{Scores: map[string]*string{"vibes": ptr("good")}}); err == nil {
		t.Error("accepted an undeclared axis")
	}
	if err := s.ChangeEntry(ctx, id, store.EntryChange{
		Scores:  map[string]*string{kind.Prestige: ptr("s")},
		AddTags: []string{" Favorite / All Time "},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateItem(ctx, "game", "X", map[string]any{"mood": "foggy"}); err == nil {
		t.Error("game accepted an undeclared field")
	}
	favorites, _ := s.ListItems(ctx, store.ItemFilter{Tag: "favorite"})
	if len(favorites) != 1 || favorites[0].Entry.Tags[0] != "favorite/all-time" || favorites[0].Entry.Scores[kind.Prestige] != "s" {
		t.Errorf("favorites = %+v", favorites)
	}
	if favorites[0].Verdicts.Values["mood"] != "foggy" {
		t.Errorf("free field lost: %v", favorites[0].Verdicts.Values)
	}
}

func TestMomentsKeepTheirRefs(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	id, _ := s.CreateItem(ctx, "game", "Outer Wilds", nil)
	other, _ := s.CreateItem(ctx, "thing", "Rainy weekend", nil)
	if _, err := s.AddMoment(ctx, id, store.NewMoment{At: "2024-03-01", Note: "replayed", Refs: []int64{other}}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddMoment(ctx, id, store.NewMoment{At: "2025-01-01", Note: "again"}); err != nil {
		t.Fatal(err)
	}
	v, _ := s.GetItem(ctx, id)
	if len(v.Moments) != 2 || v.Moments[1].Refs[0] != other {
		t.Errorf("entry %+v moments %+v", v.Entry, v.Moments)
	}
}

func TestSlugify(t *testing.T) {
	for in, want := range map[string]string{
		"Château Margaux 2015":    "chateau-margaux-2015",
		"  Half-Life: Alyx ":      "half-life-alyx",
		"!!!":                     "untitled",
		"Hollow Knight: Silksong": "hollow-knight-silksong",
	} {
		if got := store.Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}
}
