package store_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
)

const sourceWatching = "watching-site"

func statusOf(t *testing.T, s *store.Store, id int64) (string, string) {
	t.Helper()
	v, err := s.GetItem(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if v.Entry == nil {
		return "", ""
	}
	return v.Entry.Status, v.Entry.StatusSetBy
}

func TestASourceFollowsOnlyStatusesItSetUntilApTakesOver(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	id, err := s.CreateItemFromSource(ctx, "game", "Hades", kind.SourceSteam, "1", json.RawMessage(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	if status, setBy := statusOf(t, s, id); status != kind.Untriaged || setBy != kind.SourceSteam {
		t.Fatalf("made by steam: %s by %s", status, setBy)
	}

	steps := []struct {
		status string
		want   store.StatusOutcome
	}{
		{"played", store.StatusFollowed},
		{"played", store.StatusAlreadySo},
		{"finished", store.StatusFollowed},
		{kind.Untriaged, store.StatusFollowed},
	}
	for _, step := range steps {
		got, err := s.SetStatusFromSource(ctx, id, sourceWatching, step.status)
		if err != nil || got != step.want {
			t.Fatalf("%s: %s, %v; want %s", step.status, got, err, step.want)
		}
	}

	if err := s.ChangeEntry(ctx, id, store.EntryChange{Status: ptr("dropped")}); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.SetStatusFromSource(ctx, id, sourceWatching, "finished"); got != store.StatusKeptAsSet {
		t.Errorf("source overrode ap: %s", got)
	}
	if status, setBy := statusOf(t, s, id); status != "dropped" || setBy != kind.SourceAP {
		t.Errorf("after ap: %s by %s", status, setBy)
	}
}

func TestAnotherSourcesStatusIsKeptUnlessUntriaged(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	id, _ := s.CreateItemFromSource(ctx, "game", "Celeste", kind.SourceSteam, "2", json.RawMessage(`{}`))
	if got, _ := s.SetStatusFromSource(ctx, id, "other-site", "played"); got != store.StatusFollowed {
		t.Fatalf("untriaged from steam: %s", got)
	}
	if got, _ := s.SetStatusFromSource(ctx, id, sourceWatching, "finished"); got != store.StatusKeptAsSet {
		t.Errorf("took another source's status: %s", got)
	}
}

func TestAnUnknownStatusOrApAsASourceIsRefused(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	id, _ := s.CreateItemFromSource(ctx, "game", "Fez", kind.SourceSteam, "3", json.RawMessage(`{}`))
	if _, err := s.SetStatusFromSource(ctx, id, sourceWatching, "watched"); err == nil {
		t.Error("games have no watched")
	}
	if _, err := s.SetStatusFromSource(ctx, id, kind.SourceAP, "played"); err == nil {
		t.Error("ap is not a sync")
	}
}

func knownArtist(t *testing.T, s *store.Store, externalID string) int64 {
	t.Helper()
	id, err := s.KnowItemFromSource(context.Background(), kind.Artist, "Artist "+externalID, kind.SourceSpotifyArtist, externalID, json.RawMessage(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestApHandsAnArtistsStatusBack(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	id := knownArtist(t, s, "x")
	if got, err := s.SetStatusFromSource(ctx, id, kind.SourceSpotify, kind.Listening); err != nil || got != store.StatusLogged {
		t.Fatalf("logging = %v, %v", got, err)
	}

	listened := kind.Listened
	if err := s.ChangeEntry(ctx, id, store.EntryChange{Status: &listened}); err != nil {
		t.Fatal(err)
	}
	if status, setBy := statusOf(t, s, id); status != kind.Listened || setBy != kind.SourceSpotify {
		t.Fatalf("after ap picked listened: %s by %s", status, setBy)
	}
	if got, _ := s.SetStatusFromSource(ctx, id, kind.SourceSpotify, kind.Listening); got != store.StatusFollowed {
		t.Fatalf("spotify after hand back = %v, want followed", got)
	}
}

func TestADismissedArtistStaysApsToDismiss(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	id := knownArtist(t, s, "x")
	if _, err := s.SetStatusFromSource(ctx, id, kind.SourceSpotify, kind.Listening); err != nil {
		t.Fatal(err)
	}
	dismissed := kind.Dismissed
	if err := s.ChangeEntry(ctx, id, store.EntryChange{Status: &dismissed}); err != nil {
		t.Fatal(err)
	}
	if status, setBy := statusOf(t, s, id); status != kind.Dismissed || setBy != kind.SourceAP {
		t.Fatalf("after dismissing: %s by %s", status, setBy)
	}
	if got, _ := s.SetStatusFromSource(ctx, id, kind.SourceSpotify, kind.Listened); got != store.StatusKeptAsSet {
		t.Fatalf("spotify over a dismissal = %v, want kept-as-set", got)
	}
}

func TestComputedStatusesOnlyHandBackForTheirKind(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	id, err := s.CreateItemFromSource(ctx, kind.Movie, "Heat", kind.SourceLetterboxd, "heat", json.RawMessage(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	watched := "watched"
	if err := s.ChangeEntry(ctx, id, store.EntryChange{Status: &watched}); err != nil {
		t.Fatal(err)
	}
	if status, setBy := statusOf(t, s, id); status != "watched" || setBy != kind.SourceAP {
		t.Fatalf("movie: %s by %s", status, setBy)
	}
}
