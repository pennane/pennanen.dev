package db

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

const lastNumberScoreVersion = 6

func TestNumberScoresBecomeDivisionsAndRanks(t *testing.T) {
	ctx := context.Background()
	d, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "log.db")+"?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatal(err)
	}
	d.SetMaxOpenConns(singleWriter)
	t.Cleanup(func() { d.Close() })
	if err := migrateUpTo(ctx, d, lastNumberScoreVersion); err != nil {
		t.Fatal(err)
	}
	if _, err := d.ExecContext(ctx, `
		INSERT INTO item (id, kind, slug) VALUES
			(1, 'game', 'hades'), (2, 'game', 'celeste'), (3, 'game', 'outer-wilds'),
			(4, 'game', 'wallpaper'), (5, 'game', 'fez'), (6, 'thing', 'walk');
		INSERT INTO entry (item_id, status) VALUES
			(1, 'finished'), (2, 'finished'), (3, 'finished'), (4, 'dismissed'), (5, 'played'), (6, 'seen');
		INSERT INTO claim (item_id, source, external_id, data) VALUES
			(1, 'steam', '1', '{"playtime_forever": 3000}'),
			(2, 'steam', '2', '{"playtime_forever": 9000}');
		INSERT INTO score (item_id, axis, value) VALUES
			(1, 'enjoyment', 92), (2, 'enjoyment', 92), (3, 'enjoyment', 95),
			(4, 'enjoyment', 80), (5, 'enjoyment', 60), (5, 'quality', 81), (6, 'enjoyment', 79);
		INSERT INTO moment (id, item_id, at) VALUES (1, 1, '2024-01-01');
		INSERT INTO moment_score (moment_id, axis, value) VALUES (1, 'enjoyment', 97);
		INSERT INTO ranking (id, slug, name) VALUES (1, 'best', 'best');`); err != nil {
		t.Fatal(err)
	}
	if err := migrate(ctx, d); err != nil {
		t.Fatal(err)
	}

	type spot struct {
		division string
		position sql.NullInt64
	}
	want := map[[2]string]spot{
		{"outer-wilds", "enjoyment"}: {"loved", sql.NullInt64{Int64: 0, Valid: true}},
		{"celeste", "enjoyment"}:     {"loved", sql.NullInt64{Int64: 1, Valid: true}},
		{"hades", "enjoyment"}:       {"loved", sql.NullInt64{Int64: 2, Valid: true}},
		{"fez", "enjoyment"}:         {"meh", sql.NullInt64{Int64: 0, Valid: true}},
		{"fez", "quality"}:           {"great", sql.NullInt64{}},
		{"walk", "enjoyment"}:        {"liked", sql.NullInt64{Int64: 0, Valid: true}},
	}
	rows, err := d.QueryContext(ctx, `SELECT i.slug, s.axis, s.division, s.position FROM score s JOIN item i ON i.id = s.item_id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	got := map[[2]string]spot{}
	for rows.Next() {
		var slug, axis string
		var s spot
		if err := rows.Scan(&slug, &axis, &s.division, &s.position); err != nil {
			t.Fatal(err)
		}
		got[[2]string{slug, axis}] = s
	}
	if len(got) != len(want) {
		t.Errorf("scores = %v, want %v", got, want)
	}
	for key, w := range want {
		if got[key] != w {
			t.Errorf("%v = %+v, want %+v", key, got[key], w)
		}
	}
	var archived int
	d.QueryRowContext(ctx, `SELECT count(*) FROM number_score_archive`).Scan(&archived)
	if archived != 7 {
		t.Errorf("archived %d old scores, want 7", archived)
	}
	for _, gone := range []string{"ranking", "ranking_tier", "placement", "moment_score"} {
		var tables int
		d.QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = ?`, gone).Scan(&tables)
		if tables != 0 {
			t.Errorf("table %s survived", gone)
		}
	}
}
