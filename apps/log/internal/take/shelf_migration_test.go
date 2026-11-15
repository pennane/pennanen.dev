package take_test

import (
	"context"
	"testing"
	"time"

	"github.com/pennane/pennanen.dev/apps/log/internal/platform/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
	"github.com/pennane/pennanen.dev/apps/log/internal/rule"
	"github.com/pennane/pennanen.dev/apps/log/internal/source/letterboxd"
	"github.com/pennane/pennanen.dev/apps/log/internal/source/syncs"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/take"
	"github.com/pennane/pennanen.dev/apps/log/internal/world"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

func TestApsOverriddenFilmsKeepTheirChoiceThroughTheFirstSync(t *testing.T) {
	ctx, d := openAt(t, lastBeforeTakes)
	exec(t, d, `
		INSERT INTO item (id, kind, slug) VALUES (1, 'series', 'ripley'), (2, 'movie', 'heat'), (3, 'movie', 'memento'), (4, 'movie', 'ran');
		INSERT INTO record_claim (source, external_id, data, fetched_at, changed_at) VALUES
			('letterboxd', 'rip1', '{"id":"rip1","title":"Ripley","tmdbId":94028,"tmdbType":"tv"}', '2025-01-01T00:00:00Z', '2025-01-01T00:00:00Z'),
			('letterboxd', 'heat1', '{"id":"heat1","title":"Heat"}', '2025-01-01T00:00:00Z', '2025-01-01T00:00:00Z'),
			('letterboxd', 'mem1', '{"id":"mem1","title":"Memento"}', '2025-01-01T00:00:00Z', '2025-01-01T00:00:00Z'),
			('letterboxd', 'ran1', '{"id":"ran1","title":"Ran"}', '2025-01-01T00:00:00Z', '2025-01-01T00:00:00Z'),
			('letterboxd-take', 'letterboxd:mem1', '{"onWatchlistSince":"2025-05-05"}', '2025-05-05T00:00:00Z', '2025-05-05T00:00:00Z');
		INSERT INTO item_identity (id, item_id, source, external_id) VALUES
			(10, 1, 'letterboxd', 'rip1'), (20, 2, 'letterboxd', 'heat1'),
			(30, 3, 'letterboxd', 'mem1'), (31, 3, 'letterboxd-take', 'letterboxd:mem1'),
			(40, 4, 'letterboxd', 'ran1');
		INSERT INTO entry (item_id, status, status_set_by, updated_at) VALUES
			(1, 'dropped', 'ap', '2026-01-10T00:00:00Z'),
			(2, 'want', 'ap', '2026-01-11T00:00:00Z'),
			(3, 'want', 'ap', '2026-01-12T00:00:00Z'),
			(4, 'untriaged', 'ap', '2026-01-13T00:00:00Z');`)
	if err := db.MigrateUpTo(ctx, d, firstWithTakes); err != nil {
		t.Fatal(err)
	}

	if ripley := letterboxdTake(t, d, "letterboxd:rip1"); ripley.data["onWatchedShelfSince"] != "2026-01-10" {
		t.Errorf("ripley take = %+v", ripley.data)
	}
	if heat := letterboxdTake(t, d, "letterboxd:heat1"); heat.data["onWatchedShelfSince"] != "2026-01-11" {
		t.Errorf("heat take = %+v", heat.data)
	}
	if memento := letterboxdTake(t, d, "letterboxd:mem1"); memento.data["onWatchedShelfSince"] != nil {
		t.Errorf("a film ap wants, only on the watchlist, got a shelf date: %+v", memento.data)
	}
	assertRows(t, d, `SELECT 'ran takes', count(*) FROM record_claim WHERE source = 'letterboxd-take' AND external_id = 'letterboxd:ran1'`,
		[][2]string{{"ran takes", "0"}})

	if err := take.WriteKindStatuses(ctx, d, kind.All()); err != nil {
		t.Fatal(err)
	}
	records := record.New(d, reading.ReadRecord, reading.ReadersVersion)
	s := store.New(d, records, world.New(d, records), []rule.Proposer{letterboxd.Proposer})
	if _, err := s.RereadClaims(ctx); err != nil {
		t.Fatal(err)
	}
	r := &letterboxd.Recorder{Store: s, Today: func() string { return time.Now().UTC().Format(time.DateOnly) }}
	firstSync, err := syncs.New(ctx, letterboxd.SyncName, record.SourceLetterboxd, func(ctx context.Context, p *syncs.Progress) error {
		for _, film := range []reading.LetterboxdFilm{
			{ID: "rip1", Title: "Ripley", TMDBID: 94028, TMDBType: reading.SeriesOnTMDB},
			{ID: "heat1", Title: "Heat"},
			{ID: "ran1", Title: "Ran"},
		} {
			if _, err := r.Shelve(ctx, p, film, letterboxd.WatchedShelf); err != nil {
				return err
			}
		}
		_, err := r.Shelve(ctx, p, reading.LetterboxdFilm{ID: "mem1", Title: "Memento"}, letterboxd.WatchlistShelf)
		return err
	}, s, nil)
	if err != nil {
		t.Fatal(err)
	}
	if report := firstSync.Run(ctx); report.State != syncs.Succeeded || len(report.Problems) != 0 {
		t.Fatalf("first sync: %+v", report)
	}
	assertRows(t, d, `SELECT item_id, status || ' by ' || decided_by FROM entry_status ORDER BY item_id`, [][2]string{
		{"1", "dropped by ap"}, {"2", "want by ap"}, {"3", "want by ap"}, {"4", "watched by letterboxd"},
	})
}
