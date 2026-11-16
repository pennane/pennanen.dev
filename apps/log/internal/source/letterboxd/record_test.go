package letterboxd_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"slices"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/platform/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
	"github.com/pennane/pennanen.dev/apps/log/internal/source/letterboxd"
	"github.com/pennane/pennanen.dev/apps/log/internal/source/syncs"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

func TestAWatchedFilmBackOnTheWatchlistGetsAnotherShotUntilRewatched(t *testing.T) {
	ctx := context.Background()
	d, err := db.Open(ctx, filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	s := openProposing(t, d)
	r := &letterboxd.Recorder{Store: s, Today: func() string { return "2026-11-14" }}
	heat := reading.LetterboxdFilm{ID: "bbb2", Slug: "heat", Title: "Heat", Year: 1995}
	memento := reading.LetterboxdFilm{ID: "ccc3", Slug: "memento", Title: "Memento", Year: 2000}

	step := func(name string, work func(p *syncs.Progress) error) syncs.Report {
		t.Helper()
		sync, err := syncs.New(ctx, name, record.SourceLetterboxd, func(ctx context.Context, p *syncs.Progress) error { return work(p) }, s, nil)
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
	t.Run("heat back on the watchlist is watched with another shot", func(t *testing.T) {
		if e := entryOf("bbb2"); e.Status != kind.Watched || !slices.Contains(e.Tags, kind.TagAnotherShot) || e.TagsFromRules[kind.TagAnotherShot] != letterboxd.RuleLetterboxd {
			t.Fatalf("heat back on the watchlist: %+v", e)
		}
	})
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
	if e := entryOf("bbb2"); slices.Contains(e.Tags, kind.TagAnotherShot) {
		t.Errorf("rewatched, still another shot: %+v %+v", e, report)
	}
	t.Run("the rewatch counts the watchlist it left", func(t *testing.T) {
		if report.Counts[letterboxd.CountLeftWatchlist] != 1 {
			t.Errorf("tags followed = %+v", report)
		}
	})

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
	if e := entryOf("bbb2"); !slices.Contains(e.Tags, kind.TagAnotherShot) || e.TagsFromRules[kind.TagAnotherShot] != "" {
		t.Errorf("ap's own another shot was taken by the sync: %+v", e)
	}
}

type pagesLinking map[string]reading.TMDBLink

func (p pagesLinking) TMDBLinkOf(_ context.Context, film reading.LetterboxdFilm) (reading.TMDBLink, error) {
	if link, linked := p[film.ID]; linked {
		return link, nil
	}
	return reading.TMDBLink{}, letterboxd.ErrNoTMDBLink
}

func (p pagesLinking) IDOf(_ context.Context, _ reading.LetterboxdFilm) (string, error) {
	return "", letterboxd.ErrNoShortLink
}

type pagesNaming map[string]string

func (p pagesNaming) TMDBLinkOf(_ context.Context, _ reading.LetterboxdFilm) (reading.TMDBLink, error) {
	return reading.TMDBLink{}, letterboxd.ErrNoTMDBLink
}

func (p pagesNaming) IDOf(_ context.Context, film reading.LetterboxdFilm) (string, error) {
	if id, named := p[film.Slug]; named {
		return id, nil
	}
	return "", letterboxd.ErrNoShortLink
}

func TestADiaryEntryForAFilmNotInTheLogAddsIt(t *testing.T) {
	ctx := context.Background()
	d, err := db.Open(ctx, filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	s := openProposing(t, d)
	r := &letterboxd.Recorder{Store: s, Links: pagesNaming{"mulholland-drive": "297o"}}
	five := 5.0
	mulholland := letterboxd.DiaryEntry{
		Film:      reading.LetterboxdFilm{Slug: "mulholland-drive", Title: "Mulholland Drive", Year: 2001, TMDBID: 1018},
		WatchedOn: "2026-11-14",
		Rating:    &five,
		Reaction:  reading.Liked,
		Review:    "wow huh",
	}
	unnamed := letterboxd.DiaryEntry{Film: reading.LetterboxdFilm{Slug: "nowhere", Title: "Nowhere", Year: 1997}, WatchedOn: "2026-11-14"}
	sync, _ := syncs.New(ctx, "diary", record.SourceLetterboxd, func(ctx context.Context, p *syncs.Progress) error {
		for _, entry := range []letterboxd.DiaryEntry{mulholland, unnamed} {
			if err := r.RecordWatch(ctx, p, entry, map[string]string{}); err != nil {
				p.Problem("%v", err)
			}
		}
		return nil
	}, s, nil)
	report := sync.Run(ctx)
	if report.Counts[letterboxd.CountFilmsAdded] != 1 || len(report.Problems) != 1 {
		t.Fatalf("%+v", report)
	}
	id, err := s.FindByIdentity(ctx, record.SourceLetterboxd, "297o")
	if err != nil {
		t.Fatal(err)
	}
	v, _ := s.GetItem(ctx, id)
	if v.Kind != kind.Movie || v.Entry.Status != kind.Watched {
		t.Errorf("mulholland drive: %s %+v", v.Kind, v.Entry)
	}
	if take := v.Takes[0].Values; take["rating"] != 5.0 || take["last_watched"] != "2026-11-14" || take["review"] != "wow huh" {
		t.Errorf("take: %v", take)
	}
	if again, err := s.FindByIdentity(ctx, record.SourceTMDB, "1018"); err == nil && again != id {
		t.Errorf("tmdb 1018 is item %d, not %d", again, id)
	}
}

func TestASeriesOnLetterboxdIsLoggedAsASeriesOrJoinsTheOneAdded(t *testing.T) {
	ctx := context.Background()
	d, err := db.Open(ctx, filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	s := openProposing(t, d)
	added, err := s.CreateItemFromSource(ctx, kind.Series, "Ripley", record.SourceTMDBSeries, "94028", json.RawMessage(`{"id":94028,"name":"Ripley"}`))
	if err != nil {
		t.Fatal(err)
	}
	r := &letterboxd.Recorder{Store: s, Links: pagesLinking{
		"rip1": {Type: reading.SeriesOnTMDB, ID: 94028},
		"twp1": {Type: reading.SeriesOnTMDB, ID: 1920},
	}}
	sync, _ := syncs.New(ctx, "shelve", record.SourceLetterboxd, func(ctx context.Context, p *syncs.Progress) error {
		for _, film := range []reading.LetterboxdFilm{
			{ID: "rip1", Slug: "ripley-2024", Title: "Ripley", Year: 2024},
			{ID: "twp1", Slug: "twin-peaks", Title: "Twin Peaks", Year: 1990},
			{ID: "hea1", Slug: "heat", Title: "Heat", Year: 1995},
		} {
			if _, err := r.Shelve(ctx, p, film, letterboxd.WatchedShelf); err != nil {
				return err
			}
		}
		return nil
	}, s, nil)
	report := sync.Run(ctx)
	if report.State != syncs.Succeeded || report.Counts[letterboxd.CountFilmsAdded] != 2 || len(report.Problems) != 0 {
		t.Fatalf("shelve: %+v", report)
	}

	ripley, _ := s.FindByIdentity(ctx, record.SourceLetterboxd, "rip1")
	if ripley != added {
		t.Fatalf("ripley became item %d, not the one ap added (%d)", ripley, added)
	}
	if v, _ := s.GetItem(ctx, ripley); v.Kind != kind.Series || v.Entry == nil {
		t.Errorf("ripley: %s %+v", v.Kind, v.Entry)
	}
	t.Run("ripley shelved watched is watched", func(t *testing.T) {
		if v, _ := s.GetItem(ctx, ripley); v.Entry.Status != kind.Watched {
			t.Errorf("ripley: %+v", v.Entry)
		}
	})
	twinPeaks, _ := s.FindByIdentity(ctx, record.SourceLetterboxd, "twp1")
	v, _ := s.GetItem(ctx, twinPeaks)
	var claim reading.LetterboxdFilm
	c, _ := s.ClaimItemKeyed(ctx, twinPeaks, record.SourceLetterboxd)
	json.Unmarshal(c.Data, &claim)
	if v.Kind != kind.Series || claim.TMDBType != reading.SeriesOnTMDB || claim.TMDBID != 1920 {
		t.Errorf("twin peaks: %s %+v", v.Kind, claim)
	}
	heat, _ := s.FindByIdentity(ctx, record.SourceLetterboxd, "hea1")
	if v, _ := s.GetItem(ctx, heat); v.Kind != kind.Movie {
		t.Errorf("heat: %s", v.Kind)
	}
}

func TestTheNewestDiaryEntrySetsTheRating(t *testing.T) {
	ctx := context.Background()
	d, err := db.Open(ctx, filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	s := openProposing(t, d)
	r := &letterboxd.Recorder{Store: s}
	film := reading.LetterboxdFilm{ID: "lif1", Slug: "the-life-aquatic-with-steve-zissou", Title: "The Life Aquatic with Steve Zissou", Year: 2004}
	stars := func(n float64) *float64 { return &n }
	sync, _ := syncs.New(ctx, "diary", record.SourceLetterboxd, func(ctx context.Context, p *syncs.Progress) error {
		itemID, err := r.Shelve(ctx, p, film, letterboxd.WatchedShelf)
		if err != nil {
			return err
		}
		if err := r.PutListTake(ctx, itemID, letterboxd.ListedFilm{Film: film, Rating: stars(4.5)}); err != nil {
			return err
		}
		for _, entry := range []letterboxd.DiaryEntry{
			{Film: film, WatchedOn: "2026-11-14", Rewatch: letterboxd.Rewatched, Rating: stars(5)},
			{Film: film, WatchedOn: "2026-07-01", Rating: stars(4)},
		} {
			if err := r.RecordWatch(ctx, p, entry, nil); err != nil {
				return err
			}
		}
		return nil
	}, s, nil)
	report := sync.Run(ctx)
	if report.State != syncs.Succeeded {
		t.Fatalf("%+v", report)
	}
	id, _ := s.FindByIdentity(ctx, record.SourceLetterboxd, "lif1")
	v, _ := s.GetItem(ctx, id)
	if take := v.Takes[0].Values; take["rating"] != 5.0 || take["last_watched"] != "2026-11-14" {
		t.Errorf("take: %v", take)
	}
}
