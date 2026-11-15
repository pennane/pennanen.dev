package rule_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/platform/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/rule"
	"github.com/pennane/pennanen.dev/apps/log/internal/take"
)

func openLog(t *testing.T) (context.Context, *sql.DB) {
	t.Helper()
	ctx := context.Background()
	d, err := db.Open(ctx, filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	return ctx, d
}

func exec(t *testing.T, d *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := d.ExecContext(context.Background(), query, args...); err != nil {
		t.Fatal(err)
	}
}

func addItem(t *testing.T, d *sql.DB, id int64) {
	t.Helper()
	exec(t, d, `INSERT INTO item (id, kind, slug) VALUES (?, 'movie', ?)`, id, string(rune('a'+id)))
}

func proposer(source string, proposals ...rule.Proposal) rule.Proposer {
	return rule.Proposer{
		Source: source,
		Propose: func(context.Context, record.Queryer) ([]rule.Proposal, error) {
			return proposals, nil
		},
	}
}

func statusProposal(itemID int64, ruleName, value, at string) rule.Proposal {
	return rule.Proposal{ItemID: itemID, Rule: ruleName, Subject: rule.SubjectStatus, Value: value, Strength: rule.Applying, ChangedAt: at}
}

func countProposals(t *testing.T, d *sql.DB, source string) int {
	t.Helper()
	var n int
	if err := d.QueryRowContext(context.Background(), `SELECT count(*) FROM proposal WHERE source = ?`, source).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func status(t *testing.T, d *sql.DB, id int64) string {
	t.Helper()
	var s string
	if err := d.QueryRowContext(context.Background(), `SELECT status FROM entry_status WHERE item_id = ?`, id).Scan(&s); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestRunReplacesOnlyItsSourcesProposals(t *testing.T) {
	ctx, d := openLog(t)
	addItem(t, d, 1)
	addItem(t, d, 2)
	a := proposer("a", statusProposal(1, "ra", "watched", "2026-01-01"), statusProposal(2, "ra", "watched", "2026-01-01"))
	b := proposer("b", statusProposal(1, "rb", "want", "2026-01-02"))
	if _, err := rule.RunAll(ctx, d, []rule.Proposer{a, b}, "public"); err != nil {
		t.Fatal(err)
	}
	fewer := proposer("a", statusProposal(1, "ra", "watched", "2026-01-01"))
	if _, err := rule.Run(ctx, d, fewer, "public"); err != nil {
		t.Fatal(err)
	}
	if got := countProposals(t, d, "a"); got != 1 {
		t.Fatalf("a proposals = %d", got)
	}
	if got := countProposals(t, d, "b"); got != 1 {
		t.Fatalf("b proposals = %d", got)
	}
}

func TestAWithdrawnProposalHandsTheStatusBack(t *testing.T) {
	ctx, d := openLog(t)
	addItem(t, d, 1)
	exec(t, d, `INSERT INTO kind_status (kind, status) VALUES ('movie', 'watched')`)
	if _, err := take.Log(ctx, d, 1, "ap", "public"); err != nil {
		t.Fatal(err)
	}
	if _, err := rule.Run(ctx, d, proposer("a", statusProposal(1, "ra", "watched", "2026-01-01")), "public"); err != nil {
		t.Fatal(err)
	}
	if got := status(t, d, 1); got != "watched" {
		t.Fatalf("status = %s", got)
	}
	if _, err := rule.Run(ctx, d, proposer("a"), "public"); err != nil {
		t.Fatal(err)
	}
	if got := status(t, d, 1); got != "untriaged" {
		t.Fatalf("status = %s", got)
	}
}

func TestEntryProposalsLogAndAreNotStored(t *testing.T) {
	ctx, d := openLog(t)
	addItem(t, d, 1)
	entry := rule.Proposal{ItemID: 1, Rule: "steam-owned", Subject: rule.SubjectEntry}
	report, err := rule.Run(ctx, d, proposer("steam", entry), "private")
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Logged) != 1 || report.Logged[0] != 1 || report.Proposals != 0 {
		t.Fatalf("report = %+v", report)
	}
	var loggedBy, visibility string
	if err := d.QueryRowContext(ctx, `SELECT logged_by, visibility FROM entry WHERE item_id = 1`).Scan(&loggedBy, &visibility); err != nil {
		t.Fatal(err)
	}
	if loggedBy != "steam-owned" || visibility != "private" {
		t.Fatalf("entry = %s %s", loggedBy, visibility)
	}
	if got := countProposals(t, d, "steam"); got != 0 {
		t.Fatalf("stored = %d", got)
	}
	again, err := rule.Run(ctx, d, proposer("steam", entry), "private")
	if err != nil || len(again.Logged) != 0 {
		t.Fatalf("again = %+v, %v", again, err)
	}
}

func TestADismissedItemIsNeverLoggedAgain(t *testing.T) {
	ctx, d := openLog(t)
	addItem(t, d, 1)
	if err := take.Dismiss(ctx, d, 1, "2026-01-01T00:00:00Z"); err != nil {
		t.Fatal(err)
	}
	entry := rule.Proposal{ItemID: 1, Rule: "steam-owned", Subject: rule.SubjectEntry}
	report, err := rule.Run(ctx, d, proposer("steam", entry), "public")
	if err != nil || len(report.Logged) != 0 {
		t.Fatalf("report = %+v, %v", report, err)
	}
	var n int
	if err := d.QueryRowContext(ctx, `SELECT count(*) FROM entry`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("entries = %d, %v", n, err)
	}
}

func TestDatesAreStoredAsMidnight(t *testing.T) {
	ctx, d := openLog(t)
	addItem(t, d, 1)
	proposals := []rule.Proposal{
		statusProposal(1, "date", "watched", "2026-03-05"),
		statusProposal(1, "stamp", "want", "2026-03-05T10:11:12Z"),
		statusProposal(1, "none", "listened", ""),
	}
	if _, err := rule.Run(ctx, d, proposer("a", proposals...), "public"); err != nil {
		t.Fatal(err)
	}
	stored := map[string]sql.NullString{}
	rows, err := d.QueryContext(ctx, `SELECT rule, changed_at FROM proposal`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		var at sql.NullString
		if err := rows.Scan(&name, &at); err != nil {
			t.Fatal(err)
		}
		stored[name] = at
	}
	if stored["date"].String != "2026-03-05T00:00:00Z" {
		t.Fatalf("date = %v", stored["date"])
	}
	if stored["stamp"].String != "2026-03-05T10:11:12Z" {
		t.Fatalf("stamp = %v", stored["stamp"])
	}
	if stored["none"].Valid {
		t.Fatalf("none = %v", stored["none"])
	}
	if rule.Normalized("2026-03-05") != "2026-03-05T00:00:00Z" {
		t.Fatal("Normalized")
	}
}

func TestADuplicateProposalFailsTheRunAndKeepsThePreviousOnes(t *testing.T) {
	ctx, d := openLog(t)
	addItem(t, d, 1)
	if _, err := rule.Run(ctx, d, proposer("a", statusProposal(1, "ra", "watched", "2026-01-01")), "public"); err != nil {
		t.Fatal(err)
	}
	twice := statusProposal(1, "rb", "want", "2026-01-02")
	if _, err := rule.Run(ctx, d, proposer("a", twice, twice), "public"); err == nil {
		t.Fatal("duplicate proposals did not fail the run")
	}
	var kept string
	if err := d.QueryRowContext(ctx, `SELECT rule FROM proposal WHERE source = 'a'`).Scan(&kept); err != nil || kept != "ra" {
		t.Fatalf("kept = %q, %v", kept, err)
	}
}
