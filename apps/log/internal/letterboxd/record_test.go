package letterboxd_test

import (
	"context"
	"path/filepath"
	"slices"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/letterboxd"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/syncs"
	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
)

func TestAWatchedFilmBackOnTheWatchlistGetsAnotherShotUntilRewatched(t *testing.T) {
	ctx := context.Background()
	d, err := db.Open(ctx, filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	s := store.New(d)
	r := &letterboxd.Recorder{Store: s, Today: func() string { return "2026-11-14" }}
	heat := verdict.LetterboxdFilm{ID: "bbb2", Slug: "heat", Title: "Heat", Year: 1995}
	memento := verdict.LetterboxdFilm{ID: "ccc3", Slug: "memento", Title: "Memento", Year: 2000}

	step := func(name string, work func(p *syncs.Progress) error) syncs.Report {
		t.Helper()
		sync, err := syncs.New(ctx, name, func(ctx context.Context, p *syncs.Progress) error { return work(p) }, s, nil)
		if err != nil {
			t.Fatal(err)
		}
		report := sync.Run(ctx)
		if report.State != syncs.Succeeded {
			t.Fatalf("%s: %+v", name, report)
		}
		return report
	}
	entryOf := func(lid string) store.Entry {
		t.Helper()
		return *item(t, s, lid).Entry
	}

	step("shelve", func(p *syncs.Progress) error {
		if _, err := r.Shelve(ctx, p, heat, letterboxd.WatchedShelf); err != nil {
			return err
		}
		if _, err := r.Shelve(ctx, p, heat, letterboxd.WatchlistShelf); err != nil {
			return err
		}
		_, err := r.Shelve(ctx, p, memento, letterboxd.WatchlistShelf)
		return err
	})
	if e := entryOf("bbb2"); e.Status != kind.Watched || !slices.Contains(e.Tags, kind.TagAnotherShot) || e.TagsAddedBySources[kind.TagAnotherShot] != kind.SourceLetterboxd {
		t.Fatalf("heat back on the watchlist: %+v", e)
	}
	if e := entryOf("ccc3"); e.Status != kind.Want || len(e.Tags) != 0 {
		t.Errorf("an unwatched film on the watchlist is wanted, not another shot: %+v", e)
	}

	step("old diary", func(p *syncs.Progress) error {
		return r.RecordWatch(ctx, p, letterboxd.DiaryEntry{Film: heat, WatchedOn: "2020-01-01"}, nil)
	})
	if e := entryOf("bbb2"); !slices.Contains(e.Tags, kind.TagAnotherShot) {
		t.Errorf("a watch from before it was back on the watchlist took the tag: %+v", e)
	}

	report := step("rewatch", func(p *syncs.Progress) error {
		return r.RecordWatch(ctx, p, letterboxd.DiaryEntry{Film: heat, WatchedOn: "2026-11-14", Rewatch: letterboxd.Rewatched}, nil)
	})
	if e := entryOf("bbb2"); slices.Contains(e.Tags, kind.TagAnotherShot) || report.Counts[letterboxd.CountTagsFollowed] != 1 {
		t.Errorf("rewatched, still another shot: %+v %+v", e, report)
	}

	heatID := item(t, s, "bbb2").ID
	if err := s.ChangeEntry(ctx, heatID, store.EntryChange{AddTags: []string{kind.TagAnotherShot}}); err != nil {
		t.Fatal(err)
	}
	step("back and off again", func(p *syncs.Progress) error {
		if _, err := r.Shelve(ctx, p, heat, letterboxd.WatchlistShelf); err != nil {
			return err
		}
		return r.LeaveWatchlist(ctx, p, heatID)
	})
	if e := entryOf("bbb2"); !slices.Contains(e.Tags, kind.TagAnotherShot) || e.TagsAddedBySources[kind.TagAnotherShot] != "" {
		t.Errorf("ap's own another shot was taken by the sync: %+v", e)
	}
}
