package store_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
)

const playtimeAxis = "playtime"

func tiers() []store.Division {
	return []store.Division{{Key: "s", Name: "S"}, {Key: "a", Name: "A"}, {Key: "f", Name: "F"}}
}

func playtimeReading(sort store.AxisSort) *store.Reading {
	return &store.Reading{
		Sort:   sort,
		Source: kind.SourceSteam,
		Field:  "playtime_minutes",
		Floors: []float64{1000, 500, 1},
	}
}

func withAxis(t *testing.T, s *store.Store, axis store.Axis) {
	t.Helper()
	ctx := context.Background()
	axes, err := s.Axes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetAxes(ctx, append(axes, axis)); err != nil {
		t.Fatal(err)
	}
	game, _ := kind.Get("game")
	settings, err := s.KindSettings(ctx, game)
	if err != nil {
		t.Fatal(err)
	}
	settings.Axes = append(settings.Axes, axis.Key)
	if _, err := s.SetKindSettings(ctx, game, settings); err != nil {
		t.Fatal(err)
	}
}

func TestACrowdAxisIsReadFromClaimsAndNeverSetByHand(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	withAxis(t, s, store.Axis{Key: playtimeAxis, Name: playtimeAxis, Divisions: tiers(), Reading: playtimeReading(store.CrowdAxis)})

	long, _ := s.CreateItemFromSource(ctx, "game", "Long", kind.SourceSteam, "1", json.RawMessage(`{"appid":1,"playtime_forever":600}`))
	never, _ := s.CreateItemFromSource(ctx, "game", "Never", kind.SourceSteam, "2", json.RawMessage(`{"appid":2,"playtime_forever":0}`))

	v, err := s.GetItem(ctx, long)
	if err != nil {
		t.Fatal(err)
	}
	if v.Entry.ReadScores[playtimeAxis] != "a" || v.Entry.Scores[playtimeAxis] != "" {
		t.Errorf("long: read %v own %v", v.Entry.ReadScores, v.Entry.Scores)
	}
	if v.Entry.EveryScore()[playtimeAxis] != "a" {
		t.Errorf("every score: %v", v.Entry.EveryScore())
	}
	if v, _ := s.GetItem(ctx, never); len(v.Entry.ReadScores) != 0 {
		t.Errorf("below every floor still scored: %v", v.Entry.ReadScores)
	}

	err = s.ChangeEntry(ctx, long, store.EntryChange{Scores: map[string]*string{playtimeAxis: ptr("s")}})
	if !errors.Is(err, store.ErrInvalid) {
		t.Errorf("set a crowd score by hand: %v", err)
	}
}

func TestReadingsMustBeWhole(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	axes, _ := s.Axes(ctx)

	backwards := playtimeReading(store.OutsideAxis)
	backwards.Floors[0] = 1
	cases := map[string]store.Axes{
		"floors that rise": append(axes, store.Axis{Key: playtimeAxis, Name: playtimeAxis, Divisions: tiers(), Reading: backwards}),
		"own as a sort":    append(axes, store.Axis{Key: playtimeAxis, Name: playtimeAxis, Divisions: tiers(), Reading: playtimeReading(store.OwnAxis)}),
		"a missing floor": append(axes, store.Axis{Key: playtimeAxis, Name: playtimeAxis, Divisions: append(tiers(), store.Division{Key: "z", Name: "Z"}),
			Reading: playtimeReading(store.CrowdAxis)}),
	}
	for name, candidate := range cases {
		if _, err := s.SetAxes(ctx, candidate); !errors.Is(err, store.ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}

	readPrestige := append(store.Axes{}, axes...)
	readPrestige[0].Divisions = tiers()
	readPrestige[0].Reading = playtimeReading(store.CrowdAxis)
	if _, err := s.SetAxes(ctx, readPrestige); !errors.Is(err, store.ErrInvalid) {
		t.Errorf("prestige read from a source: %v", err)
	}
}

func TestAnAxisScoredByHandIsNotReadFromASource(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	withAxis(t, s, store.Axis{Key: playtimeAxis, Name: playtimeAxis, Divisions: tiers()})
	id, _ := s.CreateItem(ctx, "game", "Hades", nil)
	if err := s.ChangeEntry(ctx, id, store.EntryChange{Status: ptr("played"), Scores: map[string]*string{playtimeAxis: ptr("s")}}); err != nil {
		t.Fatal(err)
	}
	axes, _ := s.Axes(ctx)
	for i := range axes {
		if axes[i].Key == playtimeAxis {
			axes[i].Reading = playtimeReading(store.CrowdAxis)
		}
	}
	if _, err := s.SetAxes(ctx, axes); !errors.Is(err, store.ErrConflict) {
		t.Errorf("hand scores would vanish: %v", err)
	}
}
