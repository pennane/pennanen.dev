package kind

import (
	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
	"testing"
)

func TestGuessGameStatus(t *testing.T) {
	cases := []struct {
		name     string
		verdicts map[string]any
		want     Hint
	}{
		{"unknown playtime", map[string]any{}, NoHint},
		{"never launched", map[string]any{"playtime_minutes": int64(0)}, HintLikelyNeverPlayed},
		{"achievements unlocked", map[string]any{"playtime_minutes": int64(30), "achievements_unlocked": int64(2)}, HintLikelyPlayed},
		{"short with card drops", map[string]any{"playtime_minutes": int64(140), "steam_categories": []string{SteamTradingCardsCategory}}, HintLikelyIdled},
		{"long without card drops", map[string]any{"playtime_minutes": int64(900)}, HintLikelyPlayed},
		{"short without signals", map[string]any{"playtime_minutes": int64(35)}, NoHint},
		{"software with heavy playtime", map[string]any{"playtime_minutes": int64(40000), "genres": []string{"Casual", "Utilities"}}, HintLikelyNotAGame},
	}
	for _, c := range cases {
		if got := guessGameStatus(c.verdicts); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

func TestEveryKindCanBeDismissed(t *testing.T) {
	for _, k := range All() {
		if !k.HasStatus(Dismissed) {
			t.Errorf("%s cannot be dismissed", k.Name)
		}
	}
}

func TestEveryPictureHasAShapeAndOnlyPicturesDo(t *testing.T) {
	for _, k := range All() {
		for _, f := range k.Fields {
			if (f.Type == Picture) != (f.Shape != NoShape) {
				t.Errorf("%s.%s is %s with shape %q", k.Name, f.Name, f.Type, f.Shape)
			}
		}
	}
}

func TestEveryTakeSourceCanBeRead(t *testing.T) {
	for _, k := range All() {
		for _, take := range k.Takes {
			if !reading.Reads(take.Source) {
				t.Errorf("%s takes from %s, which nothing reads", k.Name, take.Source)
			}
		}
	}
}
