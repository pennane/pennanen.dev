package store_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
)

func TestAnOutsideTakeShowsBesideTheEntryAndNeverAsAField(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	id, err := s.CreateItem(ctx, "movie", "Stalker", nil)
	if err != nil {
		t.Fatal(err)
	}
	take := json.RawMessage(`{"rating":4.5,"reaction":"liked","review":"slow, then not"}`)
	if err := s.PutClaim(ctx, id, kind.SourceLetterboxdTake, "stalker", take); err != nil {
		t.Fatal(err)
	}
	v, err := s.GetItem(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if len(v.Takes) != 1 || v.Takes[0].Source != kind.SourceLetterboxdTake || v.Takes[0].Values["rating"] != 4.5 || len(v.Takes[0].Hidden) != 0 {
		t.Fatalf("takes: %+v", v.Takes)
	}
	if _, leaked := v.Verdicts.Values["review"]; leaked {
		t.Errorf("a take became a field: %v", v.Verdicts.Values)
	}
	err = s.SetTakeFieldVisibility(ctx, id, kind.SourceLetterboxdTake, "review", store.TakeFieldShown)
	if !errors.Is(err, store.ErrInvalid) {
		t.Errorf("review is public on letterboxd already: %v", err)
	}
}

func TestHiddenTakeFieldsStayOffPublicPages(t *testing.T) {
	view := store.TakeView{
		Take:   verdict.Take{Source: "vivino-take", Values: map[string]any{"rating": 4.0, "price": 30.0}},
		Hidden: []string{"price"},
	}
	public := view.Public()
	if _, shown := public.Values["price"]; shown || public.Values["rating"] != 4.0 {
		t.Errorf("public: %v", public.Values)
	}
	if _, kept := view.Values["price"]; !kept {
		t.Error("public view changed the admin view")
	}
}

func TestObservationsAreAppendedOnce(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	id, _ := s.CreateItem(ctx, "movie", "Stalker", nil)
	entry := store.Observation{
		ItemID: id, Source: kind.SourceLetterboxd, Signal: "diary", Key: "diary/123",
		ObservedAt: "2026-11-01", Data: json.RawMessage(`{"rating":4.5}`),
	}
	if got, err := s.Observe(ctx, entry); err != nil || got != store.Observed {
		t.Fatalf("first: %s %v", got, err)
	}
	changed := entry
	changed.Data = json.RawMessage(`{"rating":1}`)
	if got, err := s.Observe(ctx, changed); err != nil || got != store.AlreadyKnown {
		t.Fatalf("again: %s %v", got, err)
	}
	rewatch := entry
	rewatch.Key, rewatch.ObservedAt = "diary/456", "2026-11-10"
	if _, err := s.Observe(ctx, rewatch); err != nil {
		t.Fatal(err)
	}
	seen, err := s.Observations(ctx, id, kind.SourceLetterboxd, "diary")
	if err != nil || len(seen) != 2 || string(seen[0].Data) != `{"rating":4.5}` || seen[1].ObservedAt != "2026-11-10" {
		t.Fatalf("observations: %+v %v", seen, err)
	}
}
