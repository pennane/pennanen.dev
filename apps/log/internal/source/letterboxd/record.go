package letterboxd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
	"github.com/pennane/pennanen.dev/apps/log/internal/source/syncs"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

const (
	CountFilmsSeen       = "films-seen"
	CountFilmsAdded      = "films-added"
	CountWatchesObserved = "watches-observed"
	CountLeftWatchlist   = "left-watchlist"
)

type FilmPages interface {
	TMDBLinkOf(ctx context.Context, film reading.LetterboxdFilm) (reading.TMDBLink, error)
	IDOf(ctx context.Context, film reading.LetterboxdFilm) (string, error)
}

type Recorder struct {
	Store *store.Store
	Links FilmPages
	Today func() string
}

func (r *Recorder) today() string {
	if r.Today != nil {
		return r.Today()
	}
	return time.Now().UTC().Format(time.DateOnly)
}

func (r *Recorder) Shelve(ctx context.Context, p *syncs.Progress, film reading.LetterboxdFilm, shelf Shelf) (int64, error) {
	p.Add(CountFilmsSeen, 1)
	itemID, err := r.itemOf(ctx, film)
	if errors.Is(err, store.ErrNotFound) && film.ID != "" {
		film = r.identified(ctx, p, film)
		itemID, err = r.itemOf(ctx, film)
	}
	switch {
	case errors.Is(err, store.ErrNotFound) && film.ID != "":
		if itemID, err = r.Store.KnowItemFromSource(ctx, kindOf(film), film.Title, record.SourceLetterboxd, film.ID, film); err != nil {
			return 0, err
		}
		p.Add(CountFilmsAdded, 1)
	case err != nil:
		return 0, err
	default:
		if err := r.remember(ctx, itemID, film); err != nil {
			return 0, err
		}
	}
	if shelf == WatchlistShelf {
		return itemID, r.arriveOnWatchlist(ctx, itemID)
	}
	return itemID, r.arriveOnWatchedShelf(ctx, p, itemID)
}

func (r *Recorder) identified(ctx context.Context, p *syncs.Progress, film reading.LetterboxdFilm) reading.LetterboxdFilm {
	if _, known := film.TMDB(); known || r.Links == nil {
		return film
	}
	link, err := r.Links.TMDBLinkOf(ctx, film)
	if err != nil {
		if !errors.Is(err, ErrNoTMDBLink) {
			p.Problem("%s: %v", film.Title, err)
		}
		return film
	}
	return film.Linked(link)
}

func kindOf(film reading.LetterboxdFilm) string {
	if link, known := film.TMDB(); known {
		return kind.OfTMDB(link.Type)
	}
	return kind.Movie
}

func (r *Recorder) arriveOnWatchlist(ctx context.Context, itemID int64) error {
	return r.changeTake(ctx, itemID, func(t *reading.LetterboxdTake) {
		if t.OnWatchlistSince == "" {
			t.OnWatchlistSince = r.today()
		}
	})
}

func (r *Recorder) arriveOnWatchedShelf(ctx context.Context, p *syncs.Progress, itemID int64) error {
	return r.leaveWatchlistWhen(ctx, p, itemID, func(t *reading.LetterboxdTake) WatchlistMove {
		if t.OnWatchedShelfSince != "" {
			return Stays
		}
		t.OnWatchedShelfSince = r.today()
		if t.LastWatched == "" {
			return Left
		}
		return Stays
	})
}

func (r *Recorder) LeaveWatchlist(ctx context.Context, p *syncs.Progress, itemID int64) error {
	return r.leaveWatchlistWhen(ctx, p, itemID, func(*reading.LetterboxdTake) WatchlistMove { return Left })
}

type WatchlistMove int

const (
	Stays WatchlistMove = iota
	Left
)

func (r *Recorder) leaveWatchlistWhen(ctx context.Context, p *syncs.Progress, itemID int64, move func(t *reading.LetterboxdTake) WatchlistMove) error {
	left := Stays
	if err := r.changeTake(ctx, itemID, func(t *reading.LetterboxdTake) {
		if move(t) == Left && t.OnWatchlistSince != "" {
			t.OnWatchlistSince = ""
			left = Left
		}
	}); err != nil {
		return err
	}
	if left == Left {
		p.Add(CountLeftWatchlist, 1)
	}
	return nil
}

func (r *Recorder) PutListTake(ctx context.Context, itemID int64, listed ListedFilm) error {
	return r.changeTake(ctx, itemID, func(t *reading.LetterboxdTake) {
		t.Rating = listed.Rating
		t.Reaction = listed.Reaction
	})
}

func (r *Recorder) RecordWatch(ctx context.Context, p *syncs.Progress, entry DiaryEntry, idOfSlug map[string]string) error {
	if entry.Film.ID == "" {
		entry.Film.ID = idOfSlug[entry.Film.Slug]
	}
	itemID, err := r.itemOf(ctx, entry.Film)
	if errors.Is(err, store.ErrNotFound) {
		entry.Film, err = r.named(ctx, entry.Film)
		if err != nil {
			return fmt.Errorf("watched %s (%d) on %s, but the film isn't in the log yet and its page can't name it: %w", entry.Film.Title, entry.Film.Year, entry.WatchedOn, err)
		}
		itemID, err = r.Shelve(ctx, p, entry.Film, WatchedShelf)
	}
	if err != nil {
		return err
	}
	if entry.Film.ID != "" {
		if err := r.remember(ctx, itemID, entry.Film); err != nil {
			return err
		}
	}
	if err := r.observe(ctx, p, itemID, entry); err != nil {
		return err
	}
	return r.leaveWatchlistWhen(ctx, p, itemID, func(t *reading.LetterboxdTake) WatchlistMove {
		if entry.WatchedOn >= t.LastWatched && entry.Rating != nil {
			t.Rating = entry.Rating
		}
		if entry.WatchedOn > t.LastWatched {
			t.LastWatched = entry.WatchedOn
		}
		if entry.Review != "" {
			t.Review = entry.Review
		}
		if entry.WatchedOn >= t.OnWatchlistSince {
			return Left
		}
		return Stays
	})
}

type watchObserved struct {
	Rewatch  Rewatch                    `json:"rewatch"`
	Rating   *float64                   `json:"rating,omitempty"`
	Reaction reading.LetterboxdReaction `json:"reaction,omitempty"`
	Review   string                     `json:"review,omitempty"`
}

func (r *Recorder) observe(ctx context.Context, p *syncs.Progress, itemID int64, entry DiaryEntry) error {
	data, err := json.Marshal(watchObserved{Rewatch: entry.Rewatch, Rating: entry.Rating, Reaction: entry.Reaction, Review: entry.Review})
	if err != nil {
		return err
	}
	outcome, err := r.Store.ObserveItemKeyed(ctx, store.Observation{
		ItemID:     itemID,
		Source:     record.SourceLetterboxd,
		Signal:     record.SignalWatched,
		Key:        fmt.Sprintf("item/%d/watched/%s", itemID, entry.WatchedOn),
		ObservedAt: entry.WatchedOn,
		Data:       data,
	})
	if outcome == store.Observed {
		p.Add(CountWatchesObserved, 1)
	}
	return err
}

func (r *Recorder) named(ctx context.Context, film reading.LetterboxdFilm) (reading.LetterboxdFilm, error) {
	if film.ID != "" {
		return film, nil
	}
	if r.Links == nil {
		return film, ErrNoShortLink
	}
	id, err := r.Links.IDOf(ctx, film)
	film.ID = id
	return film, err
}

func (r *Recorder) itemOf(ctx context.Context, film reading.LetterboxdFilm) (int64, error) {
	if film.ID != "" {
		itemID, err := r.Store.FindByIdentity(ctx, record.SourceLetterboxd, film.ID)
		if !errors.Is(err, store.ErrNotFound) {
			return itemID, err
		}
	}
	if link, known := film.TMDB(); known {
		return r.Store.FindByIdentity(ctx, link.Type.Source(), link.ExternalID())
	}
	return 0, store.ErrNotFound
}

func (r *Recorder) remember(ctx context.Context, itemID int64, film reading.LetterboxdFilm) error {
	var known reading.LetterboxdFilm
	if c, err := r.Store.ClaimItemKeyed(ctx, itemID, record.SourceLetterboxd); err == nil {
		if err := json.Unmarshal(c.Data, &known); err != nil {
			return err
		}
	}
	merged := film
	if merged.ID == "" {
		merged.ID = known.ID
	}
	if merged.Slug == "" {
		merged.Slug = known.Slug
	}
	if merged.Year == 0 {
		merged.Year = known.Year
	}
	if merged.TMDBID == 0 {
		merged.TMDBID, merged.TMDBType = known.TMDBID, known.TMDBType
	}
	if merged.ID == "" {
		return nil
	}
	if merged == known {
		return nil
	}
	return r.Store.PutClaimItemKeyed(ctx, itemID, record.SourceLetterboxd, merged.ID, merged)
}

func (r *Recorder) changeTake(ctx context.Context, itemID int64, change func(t *reading.LetterboxdTake)) error {
	var take reading.LetterboxdTake
	c, err := r.Store.ClaimItemKeyed(ctx, itemID, record.SourceLetterboxdTake)
	switch {
	case err == nil:
		if err := json.Unmarshal(c.Data, &take); err != nil {
			return err
		}
	case !errors.Is(err, store.ErrNotFound):
		return err
	}
	before, err := json.Marshal(take)
	if err != nil {
		return err
	}
	change(&take)
	after, err := json.Marshal(take)
	if err != nil || string(before) == string(after) {
		return err
	}
	return r.Store.PutClaimItemKeyed(ctx, itemID, record.SourceLetterboxdTake, "", json.RawMessage(after))
}
