package take_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/platform/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/take"
)

const (
	t1 = "2026-01-01T00:00:00Z"
	t2 = "2026-02-01T00:00:00Z"
	t3 = "2026-03-01T00:00:00Z"
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

func addItem(t *testing.T, d *sql.DB, id int64, kindName string) {
	t.Helper()
	exec(t, d, `INSERT INTO item (id, kind, slug) VALUES (?, ?, ?)`, id, kindName, kindName+"-"+string(rune('a'+id)))
}

func logged(t *testing.T, ctx context.Context, d *sql.DB, id int64, kindName string) {
	t.Helper()
	addItem(t, d, id, kindName)
	if _, err := take.Log(ctx, d, id, "ap", "public"); err != nil {
		t.Fatal(err)
	}
}

func TestLoggingIsOneWay(t *testing.T) {
	ctx, d := openLog(t)
	addItem(t, d, 1, "movie")
	first, err := take.Log(ctx, d, 1, "letterboxd", "public")
	if err != nil || first != take.NewlyLogged {
		t.Fatalf("first log = %s, %v", first, err)
	}
	exec(t, d, `UPDATE entry SET blurb = 'kept' WHERE item_id = 1`)
	again, err := take.Log(ctx, d, 1, "ap", "private")
	if err != nil || again != take.AlreadyLogged {
		t.Fatalf("second log = %s, %v", again, err)
	}
	var loggedBy, visibility, blurb string
	if err := d.QueryRowContext(ctx, `SELECT logged_by, visibility, blurb FROM entry WHERE item_id = 1`).Scan(&loggedBy, &visibility, &blurb); err != nil {
		t.Fatal(err)
	}
	if loggedBy != "letterboxd" || visibility != "public" || blurb != "kept" {
		t.Fatalf("entry = %s %s %s", loggedBy, visibility, blurb)
	}
}

func TestAStatusChoiceIsReplacedNotStacked(t *testing.T) {
	ctx, d := openLog(t)
	logged(t, ctx, d, 1, "movie")
	if err := take.ChooseStatus(ctx, d, 1, "want", t1); err != nil {
		t.Fatal(err)
	}
	if err := take.ChooseStatus(ctx, d, 1, "watched", t2); err != nil {
		t.Fatal(err)
	}
	choices, err := take.StatusChoices(ctx, d, []int64{1})
	if err != nil {
		t.Fatal(err)
	}
	if got := choices[1]; got != (take.StatusChoice{Status: "watched", ChosenAt: t2}) {
		t.Fatalf("choice = %+v", got)
	}
	var rows int
	if err := d.QueryRowContext(ctx, `SELECT count(*) FROM status_choice`).Scan(&rows); err != nil || rows != 1 {
		t.Fatalf("rows = %d, %v", rows, err)
	}
	if err := take.WithdrawStatusChoice(ctx, d, 1); err != nil {
		t.Fatal(err)
	}
	choices, err = take.StatusChoices(ctx, d, []int64{1})
	if err != nil || len(choices) != 0 {
		t.Fatalf("after withdraw = %+v, %v", choices, err)
	}
}

func TestATagChoiceRemembersRemoval(t *testing.T) {
	ctx, d := openLog(t)
	logged(t, ctx, d, 1, "movie")
	if err := take.ChooseTag(ctx, d, 1, "another-shot", take.TagAdded, t1); err != nil {
		t.Fatal(err)
	}
	if err := take.ChooseTag(ctx, d, 1, "another-shot", take.TagRemoved, t2); err != nil {
		t.Fatal(err)
	}
	var presence, chosenAt string
	var rows int
	if err := d.QueryRowContext(ctx, `SELECT presence, chosen_at, count(*) FROM tag_choice WHERE item_id = 1`).Scan(&presence, &chosenAt, &rows); err != nil {
		t.Fatal(err)
	}
	if presence != string(take.TagRemoved) || chosenAt != t2 || rows != 1 {
		t.Fatalf("tag choice = %s %s x%d", presence, chosenAt, rows)
	}
}

func TestDismissalIsATakeApartFromTheEntry(t *testing.T) {
	ctx, d := openLog(t)
	logged(t, ctx, d, 1, "game")
	addItem(t, d, 2, "game")
	if err := take.Dismiss(ctx, d, 1, t1); err != nil {
		t.Fatal(err)
	}
	if err := take.Dismiss(ctx, d, 2, t2); err != nil {
		t.Fatal(err)
	}
	dismissals, err := take.Dismissals(ctx, d, []int64{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	if dismissals[1] != t1 || dismissals[2] != t2 {
		t.Fatalf("dismissals = %v", dismissals)
	}
	var entries int
	if err := d.QueryRowContext(ctx, `SELECT count(*) FROM entry`).Scan(&entries); err != nil || entries != 1 {
		t.Fatalf("entries = %d, %v", entries, err)
	}
	if err := take.Undismiss(ctx, d, 1); err != nil {
		t.Fatal(err)
	}
	dismissals, err = take.Dismissals(ctx, d, []int64{1, 2})
	if err != nil || len(dismissals) != 1 || dismissals[2] != t2 {
		t.Fatalf("after undismiss = %v, %v", dismissals, err)
	}
}

func TestReadsCoverMoreIDsThanOneChunk(t *testing.T) {
	ctx, d := openLog(t)
	ids := make([]int64, 0, 1200)
	for id := int64(1); id <= 1200; id++ {
		exec(t, d, `INSERT INTO item (id, kind, slug) VALUES (?, 'game', 'g' || ?)`, id, id)
		if err := take.Dismiss(ctx, d, id, t1); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	dismissals, err := take.Dismissals(ctx, d, ids)
	if err != nil || len(dismissals) != len(ids) {
		t.Fatalf("dismissals = %d, %v", len(dismissals), err)
	}
}
