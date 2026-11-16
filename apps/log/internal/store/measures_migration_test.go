package store_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/platform/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/take"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

const (
	lastBeforeMeasures = 20
	firstWithMeasures  = 21
)

func openBeforeMeasures(t *testing.T) (context.Context, *sql.DB) {
	t.Helper()
	ctx := context.Background()
	d, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "log.db")+"?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatal(err)
	}
	d.SetMaxOpenConns(1)
	t.Cleanup(func() { d.Close() })
	if err := db.MigrateUpTo(ctx, d, lastBeforeMeasures); err != nil {
		t.Fatal(err)
	}
	return ctx, d
}

func rowsOf(t *testing.T, d *sql.DB, query string) []string {
	t.Helper()
	rows, err := d.Query(query)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		t.Fatal(err)
	}
	got := []string{}
	for rows.Next() {
		values := make([]sql.NullString, len(columns))
		into := make([]any, len(columns))
		for i := range values {
			into[i] = &values[i]
		}
		if err := rows.Scan(into...); err != nil {
			t.Fatal(err)
		}
		row := make([]string, len(values))
		for i, v := range values {
			row[i] = v.String
		}
		got = append(got, strings.Join(row, " "))
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return got
}

func assertRowsOf(t *testing.T, d *sql.DB, query string, want ...string) {
	t.Helper()
	if got := rowsOf(t, d, query); !slices.Equal(got, want) {
		t.Errorf("%s\n got %q\nwant %q", query, got, want)
	}
}

func TestMeasuresMigrationRekeysTakesAndWatchesAndMovesListening(t *testing.T) {
	ctx, d := openBeforeMeasures(t)
	if _, err := d.ExecContext(ctx, `
		INSERT INTO item (id, kind, slug) VALUES
			(1, 'movie', 'heat'), (2, 'series', 'ripley'), (3, 'movie', 'matrix'), (4, 'movie', 'reloaded'),
			(5, 'movie', 'ran'), (6, 'movie', 'alien'), (7, 'artist', 'xena'), (8, 'artist', 'yara');
		INSERT INTO record_claim (source, external_id, data, fetched_at, changed_at) VALUES
			('letterboxd', 'heat1', '{"id":"heat1","title":"Heat"}', '2025-01-01T00:00:00Z', '2025-01-01T00:00:00Z'),
			('letterboxd-take', 'tmdb:949', '{"rating":4.5}', '2025-12-25T00:00:00Z', '2025-12-24T00:00:00Z'),
			('letterboxd', 'rip1', '{"id":"rip1","title":"Ripley"}', '2025-01-01T00:00:00Z', '2025-01-01T00:00:00Z'),
			('letterboxd-take', 'tmdb-tv:94028', '{"rating":5}', '2025-11-02T00:00:00Z', '2025-11-01T00:00:00Z'),
			('letterboxd', 'mat1', '{"id":"mat1","title":"The Matrix"}', '2025-01-01T00:00:00Z', '2025-01-01T00:00:00Z'),
			('letterboxd-take', 'tmdb:603', '{"rating":3}', '2025-03-02T00:00:00Z', '2025-03-01T00:00:00Z'),
			('letterboxd-take', 'letterboxd:mat1', '{"rating":4}', '2025-04-02T00:00:00Z', '2025-04-01T00:00:00Z'),
			('letterboxd', 'mat2', '{"id":"mat2","title":"Reloaded"}', '2025-01-01T00:00:00Z', '2025-01-01T00:00:00Z'),
			('letterboxd-take', 'tmdb:604', '{"rating":2}', '2025-06-02T00:00:00Z', '2025-06-01T00:00:00Z'),
			('letterboxd-take', 'letterboxd:mat2', '{"rating":1}', '2025-05-02T00:00:00Z', '2025-05-01T00:00:00Z'),
			('letterboxd', 'ran1', '{"id":"ran1","title":"Ran"}', '2025-01-01T00:00:00Z', '2025-01-01T00:00:00Z'),
			('letterboxd-take', 'letterboxd:ran1', '{"rating":5}', '2025-07-02T00:00:00Z', '2025-07-01T00:00:00Z'),
			('letterboxd-take', 'tmdb:348', '{"rating":4}', '2025-08-02T00:00:00Z', '2025-08-01T00:00:00Z'),
			('spotify-artist', 'X', '{"id":"X","name":"Xena"}', '2025-01-01T00:00:00Z', '2025-01-01T00:00:00Z'),
			('spotify-listening', 'spotify-artist:X', '{"listening":55}', '2025-09-02T00:00:00Z', '2025-09-01T00:00:00Z'),
			('spotify-artist', 'Y', '{"id":"Y","name":"Yara"}', '2025-01-01T00:00:00Z', '2025-01-01T00:00:00Z'),
			('spotify-listening', 'spotify-artist:Y', '{"listening":20}', '2025-09-02T00:00:00Z', '2025-09-01T00:00:00Z');
		INSERT INTO record_reading (source, external_id, version, reading) VALUES
			('letterboxd-take', 'tmdb:949', 5, '{"rating":4.5}'),
			('letterboxd-take', 'letterboxd:mat1', 5, '{"rating":4}'),
			('spotify-listening', 'spotify-artist:X', 5, '{"listening":55}');
		INSERT INTO item_identity (id, item_id, source, external_id) VALUES
			(10, 1, 'tmdb', '949'), (11, 1, 'letterboxd', 'heat1'), (12, 1, 'letterboxd-take', 'tmdb:949'),
			(20, 2, 'tmdb-tv', '94028'), (21, 2, 'letterboxd', 'rip1'), (22, 2, 'letterboxd-take', 'tmdb-tv:94028'),
			(30, 3, 'tmdb', '603'), (31, 3, 'letterboxd', 'mat1'), (32, 3, 'letterboxd-take', 'tmdb:603'), (33, 3, 'letterboxd-take', 'letterboxd:mat1'),
			(40, 4, 'tmdb', '604'), (41, 4, 'letterboxd', 'mat2'), (42, 4, 'letterboxd-take', 'tmdb:604'), (43, 4, 'letterboxd-take', 'letterboxd:mat2'),
			(50, 5, 'letterboxd', 'ran1'), (51, 5, 'letterboxd-take', 'letterboxd:ran1'),
			(60, 6, 'tmdb', '348'), (61, 6, 'letterboxd-take', 'tmdb:348'),
			(70, 7, 'spotify-artist', 'X'), (71, 7, 'spotify-listening', 'spotify-artist:X'),
			(80, 8, 'spotify-artist', 'Y'), (81, 8, 'spotify-listening', 'spotify-artist:Y');
		INSERT INTO identity_waiting (about_source, about_external_id, waiting_source, waiting_external_id) VALUES
			('tmdb', '949', 'letterboxd-take', 'tmdb:949');
		INSERT INTO record_observation (about_source, about_external_id, source, signal, key, observed_at, data) VALUES
			('tmdb', '949', 'letterboxd', 'watched', 'item/1/watched/2025-12-24', '2025-12-24', '{"rating":4.5}'),
			('tmdb-tv', '94028', 'letterboxd', 'watched', 'item/2/watched/2025-11-01', '2025-11-01', '{}'),
			('letterboxd', 'ran1', 'letterboxd', 'watched', 'letterboxd:ran1/watched/2025-07-01', '2025-07-01', '{}'),
			('letterboxd', 'ran1', 'letterboxd', 'watched', 'item/5/watched/2025-07-01', '2025-07-01', '{}'),
			('tmdb', '348', 'letterboxd', 'watched', 'item/6/watched/2025-08-01', '2025-08-01', '{}'),
			('spotify-artist', 'X', 'spotify', 'play', 'play:1', '2025-09-01T00:00:00Z', '{}');
		INSERT INTO item_verdict (item_id, own, verdicts) VALUES
			(7, '{"values":{"title":"Xena","listening":55},"sourceOf":{"title":"spotify-artist","listening":"spotify-listening"}}',
			    '{"values":{"title":"Xena","listening":55},"sourceOf":{"title":"spotify-artist","listening":"spotify-listening"}}'),
			(8, '{"values":{"title":"Yara","listening":90},"sourceOf":{"title":"spotify-artist","listening":"ap"}}',
			    '{"values":{"title":"Yara","listening":90},"sourceOf":{"title":"spotify-artist","listening":"ap"}}');`); err != nil {
		t.Fatal(err)
	}
	if _, err := d.ExecContext(ctx, `
		INSERT OR REPLACE INTO setting (key, value) VALUES ('axes', '[{"key":"own","name":"Own","divisions":[{"key":"a","name":"A"}]},{"key":"listening-share","name":"Listening share","divisions":[{"key":"a","name":"A"}],"reading":{"sort":"activity","source":"spotify-listening","field":"listening","floors":[95]}}]');`); err != nil {
		t.Fatal(err)
	}
	if err := db.MigrateUpTo(ctx, d, firstWithMeasures); err != nil {
		t.Fatal(err)
	}

	assertRowsOf(t, d, `SELECT value FROM setting WHERE key = 'axes'`,
		`[{"key":"own","name":"Own","divisions":[{"key":"a","name":"A"}]},{"key":"listening-share","name":"Listening share","divisions":[{"key":"a","name":"A"}],"reading":{"sort":"activity","source":"spotify","field":"listening","floors":[95]}}]`)
	assertRowsOf(t, d, `SELECT i.item_id, c.external_id, c.data, c.fetched_at, c.changed_at FROM record_claim c JOIN item_identity i ON i.source = c.source AND i.external_id = c.external_id WHERE c.source = 'letterboxd-take' ORDER BY i.item_id`,
		`1 letterboxd:heat1 {"rating":4.5} 2025-12-25T00:00:00Z 2025-12-24T00:00:00Z`,
		`2 letterboxd:rip1 {"rating":5} 2025-11-02T00:00:00Z 2025-11-01T00:00:00Z`,
		`3 letterboxd:mat1 {"rating":4} 2025-04-02T00:00:00Z 2025-04-01T00:00:00Z`,
		`4 letterboxd:mat2 {"rating":2} 2025-06-02T00:00:00Z 2025-06-01T00:00:00Z`,
		`5 letterboxd:ran1 {"rating":5} 2025-07-02T00:00:00Z 2025-07-01T00:00:00Z`,
		`6 tmdb:348 {"rating":4} 2025-08-02T00:00:00Z 2025-08-01T00:00:00Z`)
	assertRowsOf(t, d, `SELECT external_id FROM record_claim WHERE source = 'letterboxd-take' ORDER BY external_id`,
		"letterboxd:heat1", "letterboxd:mat1", "letterboxd:mat2", "letterboxd:ran1", "letterboxd:rip1", "tmdb:348")
	assertRowsOf(t, d, `SELECT id, item_id, external_id FROM item_identity WHERE source = 'letterboxd-take' ORDER BY id`,
		"12 1 letterboxd:heat1", "22 2 letterboxd:rip1", "33 3 letterboxd:mat1", "42 4 letterboxd:mat2", "51 5 letterboxd:ran1", "61 6 tmdb:348")
	assertRowsOf(t, d, `SELECT external_id, version, reading FROM record_reading WHERE source = 'letterboxd-take' ORDER BY external_id`,
		`letterboxd:heat1 5 {"rating":4.5}`, `letterboxd:mat1 5 {"rating":4}`)
	assertRowsOf(t, d, `SELECT about_source, about_external_id, waiting_source, waiting_external_id FROM identity_waiting`,
		"tmdb 949 letterboxd-take letterboxd:heat1")

	assertRowsOf(t, d, `SELECT about_source, about_external_id, key FROM record_observation WHERE source = 'letterboxd' ORDER BY id`,
		"letterboxd heat1 letterboxd:heat1/watched/2025-12-24",
		"letterboxd rip1 letterboxd:rip1/watched/2025-11-01",
		"letterboxd ran1 letterboxd:ran1/watched/2025-07-01",
		"letterboxd ran1 item/5/watched/2025-07-01",
		"tmdb 348 item/6/watched/2025-08-01")
	assertRowsOf(t, d, `SELECT item_id, key FROM observation WHERE source = 'spotify'`, "7 play:1")

	assertRowsOf(t, d, `SELECT count(*) FROM record_claim WHERE source = 'spotify-listening'`, "0")
	assertRowsOf(t, d, `SELECT count(*) FROM record_reading WHERE source = 'spotify-listening'`, "0")
	assertRowsOf(t, d, `SELECT count(*) FROM item_identity WHERE source = 'spotify-listening'`, "0")
	assertRowsOf(t, d, `SELECT item_id, source, field, value FROM measure ORDER BY item_id`,
		"7 spotify listening 55", "8 spotify listening 20")
	assertRowsOf(t, d, `SELECT item_id, own, verdicts FROM item_verdict ORDER BY item_id`,
		`7 {"values":{"title":"Xena"},"sourceOf":{"title":"spotify-artist"}} {"values":{"title":"Xena"},"sourceOf":{"title":"spotify-artist"}}`,
		`8 {"values":{"title":"Yara","listening":90},"sourceOf":{"title":"spotify-artist","listening":"ap"}} {"values":{"title":"Yara","listening":90},"sourceOf":{"title":"spotify-artist","listening":"ap"}}`)
	assertRowsOf(t, d, `SELECT count(*) FROM record_claim`, "13")

	if err := take.WriteKindStatuses(ctx, d, kind.All()); err != nil {
		t.Fatal(err)
	}
	s := store.Open(d)
	heat := record.Identity{Source: record.SourceLetterboxd, ExternalID: "heat1"}
	if c, err := s.ClaimOf(ctx, heat, record.SourceLetterboxdTake); err != nil || c.ExternalID != "letterboxd:heat1" {
		t.Errorf("heat's take = %+v, %v", c, err)
	}
	for itemID, want := range map[int64]struct {
		value  int64
		source string
	}{7: {55, record.SourceSpotify}, 8: {90, record.SourceAP}} {
		v, err := s.GetItem(ctx, itemID)
		if err != nil {
			t.Fatal(err)
		}
		if v.Verdicts.Values[kind.FieldListening] != want.value || v.Verdicts.SourceOf[kind.FieldListening] != want.source {
			t.Errorf("item %d verdicts = %+v", itemID, v.Verdicts)
		}
	}
}
