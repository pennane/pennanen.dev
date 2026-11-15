package take_test

import (
	"context"
	"database/sql"
	"slices"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/take"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

const tagIdled = "idled"

type proposal struct {
	itemID    int64
	rule      string
	subject   string
	value     string
	strength  string
	changedAt any
}

func propose(t *testing.T, d *sql.DB, p proposal) {
	t.Helper()
	exec(t, d, `INSERT INTO proposal (item_id, source, rule, subject, value, strength, changed_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		p.itemID, p.rule, p.rule, p.subject, p.value, p.strength, p.changedAt)
}

func kindsWritten(t *testing.T, ctx context.Context, d *sql.DB, names ...string) {
	t.Helper()
	kinds := make([]kind.Kind, 0, len(names))
	for _, name := range names {
		k, known := kind.Get(name)
		if !known {
			t.Fatalf("no kind %s", name)
		}
		kinds = append(kinds, k)
	}
	if err := take.WriteKindStatuses(ctx, d, kinds); err != nil {
		t.Fatal(err)
	}
}

type decided struct {
	status, by string
}

func statusOf(t *testing.T, ctx context.Context, d *sql.DB, itemID int64) decided {
	t.Helper()
	var got decided
	if err := d.QueryRowContext(ctx, `SELECT status, decided_by FROM entry_status WHERE item_id = ?`, itemID).Scan(&got.status, &got.by); err != nil {
		t.Fatal(err)
	}
	return got
}

func tagsOf(t *testing.T, ctx context.Context, d *sql.DB, itemID int64) []decided {
	t.Helper()
	rows, err := d.QueryContext(ctx, `SELECT tag, decided_by FROM entry_tag WHERE item_id = ? ORDER BY tag`, itemID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	tags := []decided{}
	for rows.Next() {
		var tag decided
		if err := rows.Scan(&tag.status, &tag.by); err != nil {
			t.Fatal(err)
		}
		tags = append(tags, tag)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return tags
}

func TestTheNewestWordDecides(t *testing.T) {
	ctx, d := openLog(t)
	kindsWritten(t, ctx, d, kind.Movie)
	logged(t, ctx, d, 1, kind.Movie)
	if err := take.ChooseStatus(ctx, d, 1, kind.Want, t2); err != nil {
		t.Fatal(err)
	}
	propose(t, d, proposal{1, "letterboxd", "status", kind.Watched, "applying", t3})
	if got := statusOf(t, ctx, d, 1); got != (decided{kind.Watched, "letterboxd"}) {
		t.Fatalf("status = %+v", got)
	}
}

func TestAnOlderProposalLosesToAChoice(t *testing.T) {
	ctx, d := openLog(t)
	kindsWritten(t, ctx, d, kind.Movie)
	logged(t, ctx, d, 1, kind.Movie)
	propose(t, d, proposal{1, "letterboxd", "status", kind.Watched, "applying", t1})
	if err := take.ChooseStatus(ctx, d, 1, kind.Want, t2); err != nil {
		t.Fatal(err)
	}
	if got := statusOf(t, ctx, d, 1); got != (decided{kind.Want, "ap"}) {
		t.Fatalf("status = %+v", got)
	}
}

func TestApWinsATie(t *testing.T) {
	ctx, d := openLog(t)
	kindsWritten(t, ctx, d, kind.Movie)
	logged(t, ctx, d, 1, kind.Movie)
	propose(t, d, proposal{1, "letterboxd", "status", kind.Watched, "applying", t2})
	if err := take.ChooseStatus(ctx, d, 1, kind.Want, t2); err != nil {
		t.Fatal(err)
	}
	if got := statusOf(t, ctx, d, 1); got != (decided{kind.Want, "ap"}) {
		t.Fatalf("status = %+v", got)
	}
}

func TestAnUnknownTimeIsOlderThanAnyChoice(t *testing.T) {
	ctx, d := openLog(t)
	kindsWritten(t, ctx, d, kind.Movie)
	logged(t, ctx, d, 1, kind.Movie)
	logged(t, ctx, d, 2, kind.Movie)
	propose(t, d, proposal{1, "letterboxd", "status", kind.Watched, "applying", nil})
	propose(t, d, proposal{2, "letterboxd", "status", kind.Watched, "applying", nil})
	if err := take.ChooseStatus(ctx, d, 1, kind.Want, t1); err != nil {
		t.Fatal(err)
	}
	if got := statusOf(t, ctx, d, 1); got != (decided{kind.Want, "ap"}) {
		t.Fatalf("status with a choice = %+v", got)
	}
	if got := statusOf(t, ctx, d, 2); got != (decided{kind.Watched, "letterboxd"}) {
		t.Fatalf("status without a choice = %+v", got)
	}
}

func TestSuggestingProposalsNeverDecide(t *testing.T) {
	ctx, d := openLog(t)
	kindsWritten(t, ctx, d, kind.Game)
	logged(t, ctx, d, 1, kind.Game)
	propose(t, d, proposal{1, "likely-idled", "status", "never-played", "suggesting", t3})
	propose(t, d, proposal{1, "likely-idled", "tag", tagIdled, "suggesting", t3})
	if got := statusOf(t, ctx, d, 1); got != (decided{kind.Untriaged, take.Nobody}) {
		t.Fatalf("status = %+v", got)
	}
	if got := tagsOf(t, ctx, d, 1); len(got) != 0 {
		t.Fatalf("tags = %+v", got)
	}
}

func TestAChoiceTheKindLacksIsSetAside(t *testing.T) {
	ctx, d := openLog(t)
	kindsWritten(t, ctx, d, kind.Artist)
	logged(t, ctx, d, 1, kind.Movie)
	if err := take.ChooseStatus(ctx, d, 1, kind.Watched, t1); err != nil {
		t.Fatal(err)
	}
	exec(t, d, `UPDATE item SET kind = ? WHERE id = 1`, kind.Artist)
	if got := statusOf(t, ctx, d, 1); got != (decided{kind.Untriaged, take.Nobody}) {
		t.Fatalf("status as artist = %+v", got)
	}
	exec(t, d, `UPDATE item SET kind = ? WHERE id = 1`, kind.Movie)
	kindsWritten(t, ctx, d, kind.Artist, kind.Movie)
	if got := statusOf(t, ctx, d, 1); got != (decided{kind.Watched, "ap"}) {
		t.Fatalf("status back as movie = %+v", got)
	}
}

func TestNothingSaidIsUntriaged(t *testing.T) {
	ctx, d := openLog(t)
	kindsWritten(t, ctx, d, kind.Artist)
	logged(t, ctx, d, 1, kind.Artist)
	var decidedAt sql.NullString
	if err := d.QueryRowContext(ctx, `SELECT decided_at FROM entry_status WHERE item_id = 1`).Scan(&decidedAt); err != nil {
		t.Fatal(err)
	}
	if got := statusOf(t, ctx, d, 1); got != (decided{kind.Untriaged, take.Nobody}) || decidedAt.Valid {
		t.Fatalf("status = %+v at %v", got, decidedAt)
	}
}

func TestEveryKindHasUntriaged(t *testing.T) {
	for _, k := range kind.All() {
		if !k.HasStatus(kind.Untriaged) {
			t.Errorf("%s has no %s", k.Name, kind.Untriaged)
		}
	}
}

func TestKindStatusesLeaveDismissalOut(t *testing.T) {
	ctx, d := openLog(t)
	if err := take.WriteKindStatuses(ctx, d, kind.All()); err != nil {
		t.Fatal(err)
	}
	var dismissedRows int
	if err := d.QueryRowContext(ctx, `SELECT count(*) FROM kind_status WHERE status = ?`, kind.Dismissed).Scan(&dismissedRows); err != nil || dismissedRows != 0 {
		t.Fatalf("dismissed rows = %d, %v", dismissedRows, err)
	}
	rows, err := d.QueryContext(ctx, `SELECT status FROM kind_status WHERE kind = ? ORDER BY status`, kind.Artist)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	statuses := []string{}
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatal(err)
		}
		statuses = append(statuses, s)
	}
	if want := []string{kind.Listened, kind.Listening, kind.Untriaged}; !slices.Equal(statuses, want) {
		t.Fatalf("artist statuses = %v", statuses)
	}
}

func TestTagsFollowTheNewestWord(t *testing.T) {
	ctx, d := openLog(t)
	kindsWritten(t, ctx, d, kind.Movie)
	logged(t, ctx, d, 1, kind.Movie)
	propose(t, d, proposal{1, "letterboxd", "tag", kind.TagAnotherShot, "applying", t1})
	if err := take.ChooseTag(ctx, d, 1, kind.TagAnotherShot, take.TagRemoved, t2); err != nil {
		t.Fatal(err)
	}
	if got := tagsOf(t, ctx, d, 1); len(got) != 0 {
		t.Fatalf("tags after removal = %+v", got)
	}
	exec(t, d, `UPDATE proposal SET changed_at = ? WHERE item_id = 1`, t3)
	if got := tagsOf(t, ctx, d, 1); !slices.Equal(got, []decided{{kind.TagAnotherShot, "letterboxd"}}) {
		t.Fatalf("tags after a newer proposal = %+v", got)
	}
}

func TestADismissedEntryIsNotShown(t *testing.T) {
	ctx, d := openLog(t)
	logged(t, ctx, d, 1, kind.Game)
	logged(t, ctx, d, 2, kind.Game)
	if err := take.Dismiss(ctx, d, 1, t1); err != nil {
		t.Fatal(err)
	}
	rows, err := d.QueryContext(ctx, `SELECT item_id FROM shown_entry ORDER BY item_id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	shown := []int64{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		shown = append(shown, id)
	}
	if err := rows.Err(); err != nil || !slices.Equal(shown, []int64{2}) {
		t.Fatalf("shown = %v, %v", shown, err)
	}
}
