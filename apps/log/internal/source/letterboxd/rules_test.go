package letterboxd

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/platform/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/rule"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

func logged(itemID int64) rule.Proposal {
	return rule.Proposal{ItemID: itemID, Rule: RuleLetterboxd, Subject: rule.SubjectEntry, Value: rule.EntryLogged, Strength: rule.Applying}
}

func applying(itemID int64, subject rule.Subject, value, at string) rule.Proposal {
	return rule.Proposal{ItemID: itemID, Rule: RuleLetterboxd, Subject: subject, Value: value, Strength: rule.Applying, ChangedAt: at}
}

func assertProposals(t *testing.T, got, want []rule.Proposal) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("proposals = %+v\nwant %+v", got, want)
	}
}

func TestLetterboxdWatchedFilm(t *testing.T) {
	got := propose([]shelvedTake{{itemID: 1, LastWatched: "2026-02-01", OnWatchedShelfSince: "2026-01-01"}})
	assertProposals(t, got, []rule.Proposal{logged(1), applying(1, rule.SubjectStatus, kind.Watched, "2026-02-01")})
}

func TestLetterboxdShelfWithoutDiary(t *testing.T) {
	got := propose([]shelvedTake{{itemID: 1, OnWatchedShelfSince: "2026-01-01"}})
	assertProposals(t, got, []rule.Proposal{logged(1), applying(1, rule.SubjectStatus, kind.Watched, "2026-01-01")})
}

func TestLetterboxdWatchlistOnly(t *testing.T) {
	got := propose([]shelvedTake{{itemID: 1, OnWatchlistSince: "2025-06-01"}})
	assertProposals(t, got, []rule.Proposal{logged(1), applying(1, rule.SubjectStatus, kind.Want, "2025-06-01")})
}

func TestLetterboxdWatchlistAfterWatchingIsAnotherShot(t *testing.T) {
	got := propose([]shelvedTake{{itemID: 1, LastWatched: "2025-01-01", OnWatchedShelfSince: "2025-03-01", OnWatchlistSince: "2026-01-01"}})
	assertProposals(t, got, []rule.Proposal{
		logged(1),
		applying(1, rule.SubjectStatus, kind.Watched, "2025-03-01"),
		applying(1, rule.SubjectTag, kind.TagAnotherShot, "2026-01-01"),
	})
}

func TestLetterboxdSaysNothingWithoutShelves(t *testing.T) {
	assertProposals(t, propose([]shelvedTake{{itemID: 1}}), nil)
}

func TestLetterboxdOneItemWithTwoTakesSaysOneWord(t *testing.T) {
	got := propose([]shelvedTake{
		{itemID: 1, OnWatchedShelfSince: "2026-01-01"},
		{itemID: 1, LastWatched: "2026-03-01"},
	})
	assertProposals(t, got, []rule.Proposal{logged(1), applying(1, rule.SubjectStatus, kind.Watched, "2026-03-01")})
}

func TestLetterboxdProposerReadsStoredRecords(t *testing.T) {
	ctx := context.Background()
	d, err := db.Open(ctx, filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	seed := `
		INSERT INTO item (id, kind, slug) VALUES (1, 'movie', 'heat'), (2, 'movie', 'alien'), (3, 'movie', 'ran');
		INSERT INTO record_claim (source, external_id, data, fetched_at, changed_at) VALUES
			('letterboxd-take', 'letterboxd:heat1', '{}', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z'),
			('letterboxd-take', 'letterboxd:alien1', '{}', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z'),
			('letterboxd-take', 'letterboxd:ran1', '{}', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z');
		INSERT INTO record_reading (source, external_id, version, reading) VALUES
			('letterboxd-take', 'letterboxd:heat1', 5, '{"rating":4.5,"last_watched":"2025-12-24","on_watched_shelf_since":"2026-01-02","on_watchlist_since":"2026-02-01"}'),
			('letterboxd-take', 'letterboxd:alien1', 5, '{"on_watchlist_since":"2025-05-05"}'),
			('letterboxd-take', 'letterboxd:ran1', 5, '{"rating":3}');
		INSERT INTO item_identity (item_id, source, external_id) VALUES
			(1, 'letterboxd-take', 'letterboxd:heat1'), (2, 'letterboxd-take', 'letterboxd:alien1'), (3, 'letterboxd-take', 'letterboxd:ran1');`
	if _, err := d.Exec(seed); err != nil {
		t.Fatal(err)
	}
	got, err := Proposer.Propose(ctx, d)
	if err != nil {
		t.Fatal(err)
	}
	assertProposals(t, got, []rule.Proposal{
		logged(1),
		applying(1, rule.SubjectStatus, kind.Watched, "2026-01-02"),
		applying(1, rule.SubjectTag, kind.TagAnotherShot, "2026-02-01"),
		logged(2),
		applying(2, rule.SubjectStatus, kind.Want, "2025-05-05"),
	})
	if Proposer.Source != "letterboxd" {
		t.Errorf("source = %s", Proposer.Source)
	}
	if _, err := rule.Run(ctx, d, Proposer, "public"); err != nil {
		t.Errorf("running the proposals = %v", err)
	}
}
