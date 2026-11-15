package letterboxd_test

import (
	"context"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/pennane/pennanen.dev/apps/log/internal/platform/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
	"github.com/pennane/pennanen.dev/apps/log/internal/source/letterboxd"
	"github.com/pennane/pennanen.dev/apps/log/internal/source/syncs"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

type shelves struct {
	t     *testing.T
	store *store.Store
	day   string
	r     *letterboxd.Recorder
}

func openShelves(t *testing.T, links letterboxd.FilmPages) *shelves {
	t.Helper()
	d, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	sh := &shelves{t: t, store: openProposing(t, d), day: "2026-11-01"}
	sh.r = &letterboxd.Recorder{Store: sh.store, Links: links, Today: func() string { return sh.day }}
	return sh
}

func (sh *shelves) sync(work func(ctx context.Context, p *syncs.Progress) error) syncs.Report {
	sh.t.Helper()
	ctx := context.Background()
	run, err := syncs.New(ctx, "letterboxd-test", record.SourceLetterboxd, work, sh.store, nil)
	if err != nil {
		sh.t.Fatal(err)
	}
	report := run.Run(ctx)
	if report.State != syncs.Succeeded || len(report.Problems) != 0 {
		sh.t.Fatalf("sync: %+v", report)
	}
	return report
}

func (sh *shelves) shelve(film reading.LetterboxdFilm, shelf letterboxd.Shelf) syncs.Report {
	sh.t.Helper()
	return sh.sync(func(ctx context.Context, p *syncs.Progress) error {
		_, err := sh.r.Shelve(ctx, p, film, shelf)
		return err
	})
}

func (sh *shelves) watch(film reading.LetterboxdFilm, on string) syncs.Report {
	sh.t.Helper()
	return sh.sync(func(ctx context.Context, p *syncs.Progress) error {
		return sh.r.RecordWatch(ctx, p, letterboxd.DiaryEntry{Film: film, WatchedOn: on}, nil)
	})
}

func (sh *shelves) item(film reading.LetterboxdFilm) store.ItemView {
	sh.t.Helper()
	return item(sh.t, sh.store, film.ID)
}

var (
	heatFilm    = reading.LetterboxdFilm{ID: "bbb2", Slug: "heat", Title: "Heat", Year: 1995}
	mementoFilm = reading.LetterboxdFilm{ID: "ccc3", Slug: "memento", Title: "Memento", Year: 2000}
)

func TestAFilmOnTheWatchedShelfIsWatchedWithoutADiary(t *testing.T) {
	sh := openShelves(t, nil)
	sh.day = "2026-11-14"
	sh.shelve(heatFilm, letterboxd.WatchedShelf)

	v := sh.item(heatFilm)
	if v.Entry == nil || v.Entry.Status != kind.Watched || v.Entry.StatusDecidedBy != letterboxd.RuleLetterboxd || v.Entry.LoggedBy != letterboxd.RuleLetterboxd {
		t.Fatalf("heat on the watched shelf: %+v", v.Entry)
	}
	if since := v.Takes[0].Values["on_watched_shelf_since"]; since != "2026-11-14" {
		t.Errorf("on the watched shelf since %v, want the day it was first seen there", since)
	}

	sh.day = "2026-11-20"
	sh.shelve(heatFilm, letterboxd.WatchedShelf)
	if since := sh.item(heatFilm).Takes[0].Values["on_watched_shelf_since"]; since != "2026-11-14" {
		t.Errorf("seeing it there again moved the day to %v", since)
	}
}

func TestWatchingAWatchlistedFilmLeavesTheWatchlist(t *testing.T) {
	sh := openShelves(t, nil)
	sh.shelve(mementoFilm, letterboxd.WatchlistShelf)
	sh.shelve(heatFilm, letterboxd.WatchlistShelf)
	if e := sh.item(mementoFilm).Entry; e == nil || e.Status != kind.Want {
		t.Fatalf("memento on the watchlist: %+v", e)
	}

	sh.day = "2026-11-14"
	sh.shelve(mementoFilm, letterboxd.WatchedShelf)
	sh.watch(heatFilm, "2026-11-10")

	for _, film := range []reading.LetterboxdFilm{mementoFilm, heatFilm} {
		v := sh.item(film)
		if v.Entry.Status != kind.Watched || len(v.Entry.Tags) != 0 {
			t.Errorf("%s watched off the watchlist: %+v", film.Title, v.Entry)
		}
		if since, on := v.Takes[0].Values["on_watchlist_since"]; on {
			t.Errorf("%s is still on the watchlist since %v", film.Title, since)
		}
	}
}

func TestAWatchedFilmBackOnTheWatchlistGetsAnotherShot(t *testing.T) {
	sh := openShelves(t, nil)
	sh.shelve(heatFilm, letterboxd.WatchedShelf)
	sh.day = "2026-11-14"
	sh.shelve(heatFilm, letterboxd.WatchlistShelf)
	sh.day = "2026-11-15"
	sh.shelve(heatFilm, letterboxd.WatchedShelf)

	e := sh.item(heatFilm).Entry
	if e.Status != kind.Watched || !slices.Contains(e.Tags, kind.TagAnotherShot) || e.TagsFromRules[kind.TagAnotherShot] != letterboxd.RuleLetterboxd {
		t.Fatalf("heat back on the watchlist: %+v", e)
	}
}

func TestLeavingTheWatchlistWithdrawsAnotherShot(t *testing.T) {
	sh := openShelves(t, nil)
	sh.shelve(heatFilm, letterboxd.WatchedShelf)
	sh.day = "2026-11-14"
	sh.shelve(heatFilm, letterboxd.WatchlistShelf)
	heatID := sh.item(heatFilm).ID

	report := sh.sync(func(ctx context.Context, p *syncs.Progress) error {
		return sh.r.LeaveWatchlist(ctx, p, heatID)
	})
	e := sh.item(heatFilm).Entry
	if e.Status != kind.Watched || slices.Contains(e.Tags, kind.TagAnotherShot) {
		t.Fatalf("heat off the watchlist: %+v", e)
	}
	if report.Counts[letterboxd.CountLeftWatchlist] != 1 {
		t.Errorf("left the watchlist = %+v", report.Counts)
	}
}

func TestADiaryWatchForAFilmNewToTheLogIsWatchedOnTheDiaryDate(t *testing.T) {
	sh := openShelves(t, nil)
	sh.day = "2026-11-16"
	sh.watch(heatFilm, "2024-03-05")

	e := sh.item(heatFilm).Entry
	want := []store.StatusProposal{{Rule: letterboxd.RuleLetterboxd, Value: kind.Watched, ChangedAt: "2024-03-05T00:00:00Z"}}
	if e.Status != kind.Watched || !reflect.DeepEqual(e.StatusProposals, want) {
		t.Fatalf("heat from the diary: %+v", e)
	}
}

func daysBeforeToday(n int) string {
	return time.Now().UTC().AddDate(0, 0, -n).Format(time.DateOnly)
}

func TestApsDroppedFilmStaysDroppedWhileLetterboxdRepeatsItself(t *testing.T) {
	ctx := context.Background()
	ripleyFilm := reading.LetterboxdFilm{ID: "rip1", Slug: "ripley-2024", Title: "Ripley", Year: 2024}
	sh := openShelves(t, pagesLinking{"rip1": {Type: reading.SeriesOnTMDB, ID: 94028}})
	sh.day = daysBeforeToday(10)
	watchedOn := daysBeforeToday(9)
	sh.shelve(ripleyFilm, letterboxd.WatchedShelf)
	sh.watch(ripleyFilm, watchedOn)
	ripley := sh.item(ripleyFilm)
	if ripley.Kind != kind.Series || ripley.Entry.Status != kind.Watched {
		t.Fatalf("ripley: %s %+v", ripley.Kind, ripley.Entry)
	}

	dropped := kind.Dropped
	if err := sh.store.ChangeEntry(ctx, ripley.ID, store.EntryChange{Status: &dropped}); err != nil {
		t.Fatal(err)
	}
	for _, day := range []string{daysBeforeToday(1), daysBeforeToday(0)} {
		sh.day = day
		sh.shelve(ripleyFilm, letterboxd.WatchedShelf)
		sh.watch(ripleyFilm, watchedOn)
		if e := sh.item(ripleyFilm).Entry; e.Status != kind.Dropped || e.StatusDecidedBy != record.SourceAP {
			t.Fatalf("%s: ripley = %s by %s, want dropped by ap", day, e.Status, e.StatusDecidedBy)
		}
	}
}
