package take_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/platform/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/take"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

const (
	lastBeforeTakes = 19
	firstWithTakes  = 20
)

func openAt(t *testing.T, version int) (context.Context, *sql.DB) {
	t.Helper()
	ctx := context.Background()
	d, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "log.db")+"?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatal(err)
	}
	d.SetMaxOpenConns(1)
	t.Cleanup(func() { d.Close() })
	if err := db.MigrateUpTo(ctx, d, version); err != nil {
		t.Fatal(err)
	}
	return ctx, d
}

func TestMigrationKeepsEveryStatusTagAndDismissal(t *testing.T) {
	ctx, d := openAt(t, lastBeforeTakes)
	exec(t, d, `
		INSERT INTO item (id, kind, slug) VALUES
			(1, 'movie', 'heat'), (2, 'movie', 'alien'), (3, 'game', 'wallpaper'),
			(4, 'game', 'hades'), (5, 'artist', 'pendulum'), (6, 'movie', 'ran'),
			(7, 'movie', 'stalker'), (8, 'game', 'soundtrack');
		INSERT INTO item_identity (id, item_id, source, external_id) VALUES
			(10, 1, 'letterboxd', 'heat1'), (11, 1, 'tmdb', '949'),
			(20, 2, 'tmdb', '348'), (21, 2, 'letterboxd', 'alien1'),
			(30, 3, 'steam', '431960'), (40, 4, 'steam', '1145360'),
			(50, 5, 'spotify-artist', 'pend'), (60, 6, 'letterboxd', 'ran1'),
			(70, 7, 'tmdb', '1398'), (71, 7, 'letterboxd', 'stalker1'), (80, 8, 'steam', '999');
		INSERT INTO record_claim (source, external_id, data, fetched_at, changed_at) VALUES
			('letterboxd-take', 'letterboxd:heat1', '{"rating":4.5,"lastWatched":"2025-12-24"}', '2025-12-25T00:00:00Z', '2025-12-24T00:00:00Z'),
			('letterboxd-take', 'letterboxd:stalker1', '{"rating":5,"lastWatched":"2025-11-01"}', '2025-11-02T00:00:00Z', '2025-11-01T00:00:00Z');
		INSERT INTO record_reading (source, external_id, version, reading) VALUES
			('letterboxd-take', 'letterboxd:heat1', 4, '{"rating":4.5,"last_watched":"2025-12-24"}');
		INSERT INTO item_identity (id, item_id, source, external_id) VALUES
			(12, 1, 'letterboxd-take', 'letterboxd:heat1'), (72, 7, 'letterboxd-take', 'letterboxd:stalker1');
		INSERT INTO entry (item_id, status, status_set_by, updated_at) VALUES
			(1, 'watched', 'letterboxd', '2026-01-02T03:04:05Z'),
			(2, 'watched', 'letterboxd', '2026-01-03T00:00:00Z'),
			(3, 'dismissed', 'ap', '2026-01-04T00:00:00Z'),
			(4, 'finished', 'ap', '2026-01-05T00:00:00Z'),
			(5, 'listening', 'spotify', '2026-01-06T00:00:00Z'),
			(6, 'untriaged', 'ap', '2026-01-07T00:00:00Z'),
			(7, 'watched', 'letterboxd', '2026-01-08T00:00:00Z'),
			(8, 'dismissed', 'steam', '2026-01-09T00:00:00Z');
		INSERT INTO tag (item_id, tag, added_by) VALUES
			(1, 'another-shot', 'letterboxd'), (3, 'formative', 'ap'), (4, 'comfort', 'ap'), (4, 'favorite', 'ap');`)
	if err := db.MigrateUpTo(ctx, d, firstWithTakes); err != nil {
		t.Fatal(err)
	}

	assertRows(t, d, `SELECT item_id, status FROM status_choice ORDER BY item_id`, [][2]string{{"4", "finished"}})
	assertRows(t, d, `SELECT item_id, dismissed_at FROM dismissal ORDER BY item_id`, [][2]string{{"3", "2026-01-04T00:00:00Z"}, {"8", "2026-01-09T00:00:00Z"}})
	assertRows(t, d, `SELECT item_id, tag || ' ' || presence FROM tag_choice ORDER BY item_id, tag`,
		[][2]string{{"3", "formative added"}, {"4", "comfort added"}, {"4", "favorite added"}})
	assertRows(t, d, `SELECT item_id, logged_by FROM entry ORDER BY item_id`, [][2]string{
		{"1", "letterboxd"}, {"2", "letterboxd"}, {"3", "ap"}, {"4", "ap"}, {"5", "spotify"}, {"6", "ap"}, {"7", "letterboxd"}, {"8", "steam"},
	})

	var choicesAtMigration, unstampedTags int
	if err := d.QueryRowContext(ctx, `SELECT count(*) FROM status_choice WHERE chosen_at GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]T[0-9][0-9]:[0-9][0-9]:[0-9][0-9]Z'`).Scan(&choicesAtMigration); err != nil || choicesAtMigration != 1 {
		t.Fatalf("stamped choices = %d, %v", choicesAtMigration, err)
	}
	if err := d.QueryRowContext(ctx, `SELECT count(*) FROM tag_choice WHERE chosen_at <> (SELECT chosen_at FROM status_choice)`).Scan(&unstampedTags); err != nil || unstampedTags != 0 {
		t.Fatalf("tag choices at another time = %d, %v", unstampedTags, err)
	}

	heat := letterboxdTake(t, d, "letterboxd:heat1")
	if heat.data["onWatchedShelfSince"] != "2026-01-02" || heat.data["rating"] != 4.5 || heat.data["lastWatched"] != "2025-12-24" {
		t.Fatalf("heat take = %+v", heat.data)
	}
	if heat.fetchedAt != "2025-12-25T00:00:00Z" || heat.changedAt != "2025-12-24T00:00:00Z" {
		t.Fatalf("heat take times = %s %s", heat.fetchedAt, heat.changedAt)
	}
	alien := letterboxdTake(t, d, "tmdb:348")
	if len(alien.data) != 1 || alien.data["onWatchedShelfSince"] != "2026-01-03" {
		t.Fatalf("alien take = %+v", alien.data)
	}
	if alien.fetchedAt != "2026-01-03T00:00:00Z" || alien.changedAt != "2026-01-03T00:00:00Z" {
		t.Fatalf("alien take times = %s %s", alien.fetchedAt, alien.changedAt)
	}
	assertRows(t, d, `SELECT item_id, external_id FROM item_identity WHERE source = 'letterboxd-take' ORDER BY item_id`,
		[][2]string{{"1", "letterboxd:heat1"}, {"2", "tmdb:348"}, {"7", "letterboxd:stalker1"}})
	stalker := letterboxdTake(t, d, "letterboxd:stalker1")
	if stalker.data["onWatchedShelfSince"] != "2026-01-08" || stalker.data["rating"] != 5.0 || stalker.data["lastWatched"] != "2025-11-01" {
		t.Fatalf("stalker take = %+v", stalker.data)
	}
	assertRows(t, d, `SELECT 'letterboxd-take', count(*) FROM record_claim WHERE source = 'letterboxd-take'`,
		[][2]string{{"letterboxd-take", "3"}})
	var staleReadings int
	if err := d.QueryRowContext(ctx, `SELECT count(*) FROM record_reading WHERE source = 'letterboxd-take'`).Scan(&staleReadings); err != nil || staleReadings != 0 {
		t.Fatalf("letterboxd take readings left to reread = %d, %v", staleReadings, err)
	}

	if err := take.WriteKindStatuses(ctx, d, kind.All()); err != nil {
		t.Fatal(err)
	}
	assertRows(t, d, `SELECT s.item_id, s.status || ' by ' || s.decided_by FROM shown_entry e JOIN entry_status s ON s.item_id = e.item_id ORDER BY s.item_id`, [][2]string{
		{"1", "untriaged by nobody"}, {"2", "untriaged by nobody"}, {"4", "finished by ap"}, {"5", "untriaged by nobody"}, {"6", "untriaged by nobody"}, {"7", "untriaged by nobody"},
	})
}

type claim struct {
	data                 map[string]any
	fetchedAt, changedAt string
}

func letterboxdTake(t *testing.T, d *sql.DB, externalID string) claim {
	t.Helper()
	var raw string
	var c claim
	if err := d.QueryRow(`SELECT data, fetched_at, changed_at FROM record_claim WHERE source = 'letterboxd-take' AND external_id = ?`, externalID).Scan(&raw, &c.fetchedAt, &c.changedAt); err != nil {
		t.Fatalf("%s: %v", externalID, err)
	}
	if err := json.Unmarshal([]byte(raw), &c.data); err != nil {
		t.Fatal(err)
	}
	return c
}

func assertRows(t *testing.T, d *sql.DB, query string, want [][2]string) {
	t.Helper()
	rows, err := d.Query(query)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	got := [][2]string{}
	for rows.Next() {
		var row [2]string
		if err := rows.Scan(&row[0], &row[1]); err != nil {
			t.Fatal(err)
		}
		got = append(got, row)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Fatalf("%s = %v, want %v", query, got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("%s = %v, want %v", query, got, want)
		}
	}
}
