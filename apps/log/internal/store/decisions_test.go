package store_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"maps"
	"path/filepath"
	"slices"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/platform/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
	"github.com/pennane/pennanen.dev/apps/log/internal/rule"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/take"
	"github.com/pennane/pennanen.dev/apps/log/internal/world"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

const (
	longerAgo = "2019-01-01T00:00:00Z"
	longAgo   = "2020-01-01T00:00:00Z"
	farAhead  = "2999-01-01T00:00:00Z"
	ruleLists = "lists"
)

func openStoreProposing(t *testing.T, proposers ...rule.Proposer) (*store.Store, *sql.DB) {
	t.Helper()
	ctx := context.Background()
	d, err := db.Open(ctx, filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	if err := take.WriteKindStatuses(ctx, d, kind.All()); err != nil {
		t.Fatal(err)
	}
	records := record.New(d, reading.ReadRecord, reading.ReadersVersion)
	return store.New(d, records, world.New(d, records), proposers), d
}

func saying(proposals func() []rule.Proposal) rule.Proposer {
	return rule.Proposer{Source: ruleLists, Propose: func(context.Context, record.Queryer) ([]rule.Proposal, error) {
		return proposals(), nil
	}}
}

func applyingStatus(itemID int64, value, at string) rule.Proposal {
	return rule.Proposal{ItemID: itemID, Rule: ruleLists, Subject: rule.SubjectStatus, Value: value, Strength: rule.Applying, ChangedAt: at}
}

func entryOf(t *testing.T, s *store.Store, id int64) store.Entry {
	t.Helper()
	v, err := s.GetItem(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if v.Entry == nil {
		t.Fatalf("item %d is not logged", id)
	}
	return *v.Entry
}

func handMadeGame(t *testing.T, s *store.Store, title string) int64 {
	t.Helper()
	id, err := s.CreateItem(context.Background(), kind.Game, title, nil)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func chosenAt(t *testing.T, d *sql.DB, id int64) string {
	t.Helper()
	var at string
	if err := d.QueryRow(`SELECT chosen_at FROM status_choice WHERE item_id = ?`, id).Scan(&at); err != nil {
		t.Fatal(err)
	}
	return at
}

func TestApChoosingAStatusStoresAChoice(t *testing.T) {
	ctx := context.Background()
	s, d := openStoreProposing(t)
	id := handMadeGame(t, s, "Hades")
	if e := entryOf(t, s, id); e.Status != kind.Untriaged || e.StatusDecidedBy != take.Nobody || e.StatusChoice != nil || e.LoggedBy != record.SourceAP {
		t.Fatalf("fresh entry: %+v", e)
	}

	if err := s.ChangeEntry(ctx, id, store.EntryChange{Status: ptr("played")}); err != nil {
		t.Fatal(err)
	}
	e := entryOf(t, s, id)
	if e.Status != "played" || e.StatusDecidedBy != record.SourceAP {
		t.Fatalf("after choosing: %s by %s", e.Status, e.StatusDecidedBy)
	}
	if e.StatusChoice == nil || e.StatusChoice.Status != "played" || e.StatusChoice.ChosenAt == "" || e.StatusChoice.Standing != store.ChoiceDecides {
		t.Fatalf("choice = %+v", e.StatusChoice)
	}

	if _, err := d.Exec(`UPDATE status_choice SET chosen_at = ? WHERE item_id = ?`, longAgo, id); err != nil {
		t.Fatal(err)
	}
	if err := s.ChangeEntry(ctx, id, store.EntryChange{Blurb: ptr("a blurb moves nothing")}); err != nil {
		t.Fatal(err)
	}
	if err := s.ChangeEntry(ctx, id, store.EntryChange{Status: ptr("played")}); err != nil {
		t.Fatal(err)
	}
	if at := chosenAt(t, d, id); at != longAgo {
		t.Errorf("saying the same again moved the choice to %s", at)
	}
	if err := s.ChangeEntry(ctx, id, store.EntryChange{Status: ptr("finished")}); err != nil {
		t.Fatal(err)
	}
	if at := chosenAt(t, d, id); at == longAgo {
		t.Errorf("a new status kept the old time")
	}
	if err := s.ChangeEntry(ctx, id, store.EntryChange{Status: ptr("watched")}); !errors.Is(err, store.ErrInvalid) {
		t.Errorf("a game took a film status: %v", err)
	}
	if err := s.ChangeEntry(ctx, id, store.EntryChange{Status: ptr(kind.Dismissed)}); !errors.Is(err, store.ErrInvalid) {
		t.Errorf("dismissed is a dismissal, not a status: %v", err)
	}
}

func TestChoosingUntriagedHandsTheStatusBackToTheRules(t *testing.T) {
	ctx := context.Background()
	var id int64
	s, _ := openStoreProposing(t, saying(func() []rule.Proposal {
		return []rule.Proposal{applyingStatus(id, "played", longAgo)}
	}))
	id = handMadeGame(t, s, "Celeste")
	if err := s.ProposeAll(ctx); err != nil {
		t.Fatal(err)
	}
	if e := entryOf(t, s, id); e.Status != "played" || e.StatusDecidedBy != ruleLists {
		t.Fatalf("from the rule: %s by %s", e.Status, e.StatusDecidedBy)
	}
	if err := s.ChangeEntry(ctx, id, store.EntryChange{Status: ptr("finished")}); err != nil {
		t.Fatal(err)
	}
	if e := entryOf(t, s, id); e.Status != "finished" || e.StatusDecidedBy != record.SourceAP {
		t.Fatalf("ap's newer word: %s by %s", e.Status, e.StatusDecidedBy)
	}
	if err := s.ChangeEntry(ctx, id, store.EntryChange{Status: ptr(kind.Untriaged)}); err != nil {
		t.Fatal(err)
	}
	if e := entryOf(t, s, id); e.Status != "played" || e.StatusDecidedBy != ruleLists || e.StatusChoice != nil {
		t.Errorf("after untriaged: %s by %s, choice %+v", e.Status, e.StatusDecidedBy, e.StatusChoice)
	}
}

func TestDismissingKeepsTheEntryButHidesIt(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	id := handMadeGame(t, s, "Wallpaper Engine")
	if err := s.ChangeEntry(ctx, id, store.EntryChange{Status: ptr("played"), Scores: map[string]*string{kind.Prestige: ptr("b")}}); err != nil {
		t.Fatal(err)
	}
	dismissed := store.Dismissed
	if err := s.ChangeEntry(ctx, id, store.EntryChange{Dismissal: &dismissed}); err != nil {
		t.Fatal(err)
	}
	v, err := s.GetItem(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if v.Entry == nil || v.DismissedAt == "" || len(v.Entry.Scores) != 0 || v.Entry.Status != "played" {
		t.Fatalf("dismissed item = %+v, entry %+v", v, v.Entry)
	}
	listed := func() []int64 {
		views, err := s.ListItems(ctx, store.ItemFilter{Kind: kind.Game})
		if err != nil {
			t.Fatal(err)
		}
		ids := []int64{}
		for _, v := range views {
			ids = append(ids, v.ID)
		}
		return ids
	}
	if got := listed(); len(got) != 0 {
		t.Errorf("listed a dismissed item: %v", got)
	}
	if got, err := s.LoggedItemsNotDismissed(ctx, kind.Game); err != nil || len(got) != 0 {
		t.Errorf("logged and not dismissed = %v, %v", got, err)
	}

	kept := store.Kept
	if err := s.ChangeEntry(ctx, id, store.EntryChange{Dismissal: &kept}); err != nil {
		t.Fatal(err)
	}
	if got := listed(); !slices.Equal(got, []int64{id}) {
		t.Errorf("kept again but listed %v", got)
	}
	if v, _ := s.GetItem(ctx, id); v.DismissedAt != "" {
		t.Errorf("kept item still dismissed at %s", v.DismissedAt)
	}
	unknown := store.DismissalChange("forgotten")
	if err := s.ChangeEntry(ctx, id, store.EntryChange{Dismissal: &unknown}); !errors.Is(err, store.ErrInvalid) {
		t.Errorf("unknown dismissal change: %v", err)
	}
}

func TestTagsFromRulesAreReportedApart(t *testing.T) {
	ctx := context.Background()
	var id int64
	s, d := openStoreProposing(t, saying(func() []rule.Proposal {
		return []rule.Proposal{{ItemID: id, Rule: ruleLists, Subject: rule.SubjectTag, Value: kind.TagAnotherShot, Strength: rule.Applying, ChangedAt: longAgo}}
	}))
	id = handMadeGame(t, s, "Hades")
	if err := s.ProposeAll(ctx); err != nil {
		t.Fatal(err)
	}
	if err := s.ChangeEntry(ctx, id, store.EntryChange{AddTags: []string{"Comfort"}}); err != nil {
		t.Fatal(err)
	}
	e := entryOf(t, s, id)
	if !slices.Equal(e.Tags, []string{kind.TagAnotherShot, "comfort"}) {
		t.Errorf("tags = %v", e.Tags)
	}
	if want := map[string]string{kind.TagAnotherShot: ruleLists}; !maps.Equal(e.TagsFromRules, want) {
		t.Errorf("tags from rules = %v, want %v", e.TagsFromRules, want)
	}
	tagged, err := s.ListItems(ctx, store.ItemFilter{Tag: kind.TagAnotherShot})
	if err != nil || len(tagged) != 1 {
		t.Errorf("tag filter = %d, %v", len(tagged), err)
	}

	if err := s.ChangeEntry(ctx, id, store.EntryChange{RemoveTags: []string{kind.TagAnotherShot}}); err != nil {
		t.Fatal(err)
	}
	e = entryOf(t, s, id)
	if !slices.Equal(e.Tags, []string{"comfort"}) || len(e.TagsFromRules) != 0 {
		t.Errorf("after ap removed the rule's tag: %v, %v", e.Tags, e.TagsFromRules)
	}

	if _, err := d.Exec(`UPDATE tag_choice SET chosen_at = ? WHERE item_id = ? AND tag = ?`, longerAgo, id, kind.TagAnotherShot); err != nil {
		t.Fatal(err)
	}
	if e := entryOf(t, s, id); !slices.Contains(e.Tags, kind.TagAnotherShot) {
		t.Fatalf("a newer rule did not put the tag back: %v", e.Tags)
	}
	if err := s.ChangeEntry(ctx, id, store.EntryChange{RemoveTags: []string{kind.TagAnotherShot}}); err != nil {
		t.Fatal(err)
	}
	if e := entryOf(t, s, id); slices.Contains(e.Tags, kind.TagAnotherShot) {
		t.Errorf("removing again left the rule's newer tag: %v", e.Tags)
	}

	if err := s.ChangeEntry(ctx, id, store.EntryChange{RemoveTags: []string{"never-had"}}); err != nil {
		t.Fatal(err)
	}
	var stored int
	if err := d.QueryRow(`SELECT count(*) FROM tag_choice WHERE item_id = ? AND tag = 'never-had'`, id).Scan(&stored); err != nil || stored != 0 {
		t.Errorf("removing a tag the entry never had stored %d choices, %v", stored, err)
	}
}

func TestTheEditorSeesWhenTheListsDisagree(t *testing.T) {
	ctx := context.Background()
	var artist int64
	s, d := openStoreProposing(t, saying(func() []rule.Proposal {
		return []rule.Proposal{applyingStatus(artist, kind.Listening, longAgo)}
	}))
	artist = knowArtist(t, s, "x", "Pendulum")
	if err := s.ChangeEntry(ctx, artist, store.EntryChange{Status: ptr(kind.Listened)}); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Exec(`UPDATE status_choice SET chosen_at = ? WHERE item_id = ?`, longerAgo, artist); err != nil {
		t.Fatal(err)
	}
	if err := s.ProposeAll(ctx); err != nil {
		t.Fatal(err)
	}
	e := entryOf(t, s, artist)
	if e.Status != kind.Listening || e.StatusDecidedBy != ruleLists {
		t.Errorf("status = %s by %s", e.Status, e.StatusDecidedBy)
	}
	if e.StatusChoice == nil || e.StatusChoice.Status != kind.Listened || e.StatusChoice.Standing != store.ChoiceOutvoted {
		t.Errorf("choice = %+v", e.StatusChoice)
	}
	want := []store.StatusProposal{{Rule: ruleLists, Value: kind.Listening, ChangedAt: longAgo}}
	if !slices.Equal(e.StatusProposals, want) {
		t.Errorf("proposals = %+v, want %+v", e.StatusProposals, want)
	}
	if err := s.ChangeEntry(ctx, artist, store.EntryChange{Status: ptr(kind.Listened)}); err != nil {
		t.Fatal(err)
	}
	if e := entryOf(t, s, artist); e.Status != kind.Listened || e.StatusChoice == nil || e.StatusChoice.Standing != store.ChoiceDecides {
		t.Errorf("re-picking an outvoted choice: %s, %+v", e.Status, e.StatusChoice)
	}

	game := handMadeGame(t, s, "Moved Here")
	if _, err := d.Exec(`INSERT INTO status_choice (item_id, status, chosen_at) VALUES (?, ?, ?)`, game, kind.Watched, longAgo); err != nil {
		t.Fatal(err)
	}
	if e := entryOf(t, s, game); e.Status != kind.Untriaged || e.StatusChoice == nil || e.StatusChoice.Standing != store.ChoiceSetAside {
		t.Errorf("a choice the kind lacks: %s, %+v", e.Status, e.StatusChoice)
	}
}

func TestUntriagedCountReadsDecidedStatus(t *testing.T) {
	ctx := context.Background()
	var decided int64
	s, _ := openStoreProposing(t, saying(func() []rule.Proposal {
		return []rule.Proposal{applyingStatus(decided, "played", longAgo)}
	}))
	handMadeGame(t, s, "Untouched")
	decided = handMadeGame(t, s, "Played By The Rule")
	gone := handMadeGame(t, s, "Untriaged But Dismissed")
	dismissed := store.Dismissed
	if err := s.ChangeEntry(ctx, gone, store.EntryChange{Dismissal: &dismissed}); err != nil {
		t.Fatal(err)
	}
	if err := s.ProposeAll(ctx); err != nil {
		t.Fatal(err)
	}
	if count, err := s.CountUntriaged(ctx, kind.Game); err != nil || count != 1 {
		t.Errorf("untriaged = %d, %v", count, err)
	}
	played, err := s.ListItems(ctx, store.ItemFilter{Status: "played"})
	if err != nil || len(played) != 1 || played[0].ID != decided {
		t.Errorf("status filter = %+v, %v", played, err)
	}
}

func TestSuggestionsAreListedByRule(t *testing.T) {
	ctx := context.Background()
	var hinted int64
	s, _ := openStoreProposing(t, saying(func() []rule.Proposal {
		return []rule.Proposal{{ItemID: hinted, Rule: "likely-idled", Subject: rule.SubjectTag, Value: "idled", Strength: rule.Suggesting}}
	}))
	handMadeGame(t, s, "Unhinted")
	hinted = handMadeGame(t, s, "Card Farm")
	if err := s.ProposeAll(ctx); err != nil {
		t.Fatal(err)
	}
	views, err := s.ListItems(ctx, store.ItemFilter{Suggested: "likely-idled"})
	if err != nil || len(views) != 1 || views[0].ID != hinted {
		t.Fatalf("suggested = %+v, %v", views, err)
	}
	want := []store.Suggestion{{Rule: "likely-idled", Subject: rule.SubjectTag, Value: "idled"}}
	if !slices.Equal(views[0].Suggestions, want) {
		t.Errorf("suggestions = %+v", views[0].Suggestions)
	}
	if e := entryOf(t, s, hinted); len(e.Tags) != 0 {
		t.Errorf("a suggestion tagged the entry: %v", e.Tags)
	}
}

func TestMatchingAnItemReproposes(t *testing.T) {
	ctx := context.Background()
	matchedOnIGDB := rule.Proposer{Source: ruleLists, Propose: func(ctx context.Context, q record.Queryer) ([]rule.Proposal, error) {
		rows, err := q.QueryContext(ctx, `SELECT item_id FROM claim WHERE source = ?`, record.SourceIGDB)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var proposals []rule.Proposal
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				return nil, err
			}
			proposals = append(proposals, applyingStatus(id, "played", longAgo))
		}
		return proposals, rows.Err()
	}}
	s, _ := openStoreProposing(t, matchedOnIGDB)
	id := handMadeGame(t, s, "Tetris")
	if err := s.AnswerMatch(ctx, id, record.SourceIGDB, "1942", json.RawMessage(`{"id":1942,"name":"Tetris"}`)); err != nil {
		t.Fatal(err)
	}
	if e := entryOf(t, s, id); e.Status != "played" || e.StatusDecidedBy != ruleLists {
		t.Errorf("after matching: %s by %s", e.Status, e.StatusDecidedBy)
	}
}

func TestProposeRunsOnlyThatSource(t *testing.T) {
	ctx := context.Background()
	var id int64
	other := rule.Proposer{Source: "other", Propose: func(context.Context, record.Queryer) ([]rule.Proposal, error) {
		return []rule.Proposal{{ItemID: id, Rule: "other", Subject: rule.SubjectStatus, Value: "finished", Strength: rule.Applying, ChangedAt: farAhead}}, nil
	}}
	s, _ := openStoreProposing(t, saying(func() []rule.Proposal {
		return []rule.Proposal{applyingStatus(id, "played", longAgo)}
	}), other)
	id = handMadeGame(t, s, "Fez")
	if err := s.Propose(ctx, ruleLists); err != nil {
		t.Fatal(err)
	}
	if e := entryOf(t, s, id); e.Status != "played" {
		t.Errorf("status = %s, want only the lists' word", e.Status)
	}
	if err := s.Propose(ctx, "nobody-knows"); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("unknown source: %v", err)
	}
}

func TestTheDismissalFilterListsOnlyDismissedItems(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	kept := handMadeGame(t, s, "Hades")
	gone := handMadeGame(t, s, "Wallpaper Engine")
	dismissed := store.Dismissed
	if err := s.ChangeEntry(ctx, gone, store.EntryChange{Dismissal: &dismissed}); err != nil {
		t.Fatal(err)
	}
	views, err := s.ListItems(ctx, store.ItemFilter{Kind: kind.Game, Dismissal: store.Dismissed})
	if err != nil || len(views) != 1 || views[0].ID != gone || views[0].DismissedAt == "" {
		t.Fatalf("dismissed = %+v, %v", views, err)
	}
	shown, err := s.ListItems(ctx, store.ItemFilter{Kind: kind.Game})
	if err != nil || len(shown) != 1 || shown[0].ID != kept {
		t.Fatalf("shown = %+v, %v", shown, err)
	}
	kept2 := store.Kept
	if _, err := s.ListItems(ctx, store.ItemFilter{Dismissal: kept2}); !errors.Is(err, store.ErrInvalid) {
		t.Errorf("kept filter error = %v", err)
	}
}
