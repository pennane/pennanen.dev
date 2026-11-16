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

func (r *Recorder) Shelve(ctx context.Context, p *syncs.Progress, film reading.LetterboxdFilm, shelf Shelf) (record.Identity, error) {
	p.Add(CountFilmsSeen, 1)
	known, err := r.remember(ctx, film)
	if errors.Is(err, store.ErrNotFound) && film.ID != "" {
		film = r.identified(ctx, p, film)
		known, err = r.remember(ctx, film)
	}
	if errors.Is(err, store.ErrNotFound) && film.ID != "" {
		if err = r.Store.PutClaim(ctx, filmIdentity(film.ID), film); err == nil {
			p.Add(CountFilmsAdded, 1)
		}
		known = film
	}
	if err != nil {
		return record.Identity{}, err
	}
	id, err := identityOf(known)
	if err != nil {
		return record.Identity{}, err
	}
	if shelf == WatchlistShelf {
		return id, r.arriveOnWatchlist(ctx, id)
	}
	return id, r.arriveOnWatchedShelf(ctx, p, id)
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

func (r *Recorder) arriveOnWatchlist(ctx context.Context, film record.Identity) error {
	return r.changeTake(ctx, film, func(t *reading.LetterboxdTake) {
		if t.OnWatchlistSince == "" {
			t.OnWatchlistSince = r.today()
		}
	})
}

func (r *Recorder) arriveOnWatchedShelf(ctx context.Context, p *syncs.Progress, film record.Identity) error {
	return r.leaveWatchlistWhen(ctx, p, film, func(t *reading.LetterboxdTake) WatchlistMove {
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

func (r *Recorder) LeaveWatchlist(ctx context.Context, p *syncs.Progress, film record.Identity) error {
	return r.leaveWatchlistWhen(ctx, p, film, func(*reading.LetterboxdTake) WatchlistMove { return Left })
}

type WatchlistMove int

const (
	Stays WatchlistMove = iota
	Left
)

func (r *Recorder) leaveWatchlistWhen(ctx context.Context, p *syncs.Progress, film record.Identity, move func(t *reading.LetterboxdTake) WatchlistMove) error {
	left := Stays
	if err := r.changeTake(ctx, film, func(t *reading.LetterboxdTake) {
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

func (r *Recorder) PutListTake(ctx context.Context, film record.Identity, listed ListedFilm) error {
	return r.changeTake(ctx, film, func(t *reading.LetterboxdTake) {
		t.Rating = listed.Rating
		t.Reaction = listed.Reaction
	})
}

func (r *Recorder) RecordWatch(ctx context.Context, p *syncs.Progress, entry DiaryEntry, idOfSlug map[string]string) error {
	if entry.Film.ID == "" {
		entry.Film.ID = idOfSlug[entry.Film.Slug]
	}
	var film record.Identity
	known, err := r.remember(ctx, entry.Film)
	if errors.Is(err, store.ErrNotFound) {
		entry.Film, err = r.named(ctx, entry.Film)
		if err != nil {
			return fmt.Errorf("watched %s (%d) on %s, but the film isn't in the log yet and its page can't name it: %w", entry.Film.Title, entry.Film.Year, entry.WatchedOn, err)
		}
		film, err = r.Shelve(ctx, p, entry.Film, WatchedShelf)
	} else if err == nil {
		film, err = identityOf(known)
	}
	if err != nil {
		return err
	}
	if err := r.observe(ctx, p, film, entry); err != nil {
		return err
	}
	return r.leaveWatchlistWhen(ctx, p, film, func(t *reading.LetterboxdTake) WatchlistMove {
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

func (r *Recorder) observe(ctx context.Context, p *syncs.Progress, film record.Identity, entry DiaryEntry) error {
	data, err := json.Marshal(watchObserved{Rewatch: entry.Rewatch, Rating: entry.Rating, Reaction: entry.Reaction, Review: entry.Review})
	if err != nil {
		return err
	}
	outcome, err := r.Store.Observe(ctx, record.Observation{
		About:      film,
		Source:     record.SourceLetterboxd,
		Signal:     record.SignalWatched,
		Key:        fmt.Sprintf("%s/watched/%s", film, entry.WatchedOn),
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

func filmIdentity(filmID string) record.Identity {
	return record.Identity{Source: record.SourceLetterboxd, ExternalID: filmID}
}

func takeIdentity(film record.Identity) record.Identity {
	return record.Identity{Source: record.SourceLetterboxdTake, ExternalID: film.String()}
}

func identityOf(film reading.LetterboxdFilm) (record.Identity, error) {
	if film.ID == "" {
		return record.Identity{}, fmt.Errorf("%s has no letterboxd id", film.Title)
	}
	return filmIdentity(film.ID), nil
}

func (r *Recorder) holders(film reading.LetterboxdFilm) []record.Identity {
	var ids []record.Identity
	if film.ID != "" {
		ids = append(ids, filmIdentity(film.ID))
	}
	if link, known := film.TMDB(); known {
		ids = append(ids, record.Identity{Source: link.Type.Source(), ExternalID: link.ExternalID()})
	}
	return ids
}

func (r *Recorder) remember(ctx context.Context, film reading.LetterboxdFilm) (reading.LetterboxdFilm, error) {
	for _, holder := range r.holders(film) {
		merged, err := r.rememberOn(ctx, holder, film)
		if !errors.Is(err, store.ErrNotFound) {
			return merged, err
		}
	}
	return film, store.ErrNotFound
}

func (r *Recorder) rememberOn(ctx context.Context, holder record.Identity, film reading.LetterboxdFilm) (reading.LetterboxdFilm, error) {
	var known reading.LetterboxdFilm
	c, err := r.Store.ClaimOf(ctx, holder, record.SourceLetterboxd)
	switch {
	case err == nil:
		if err := json.Unmarshal(c.Data, &known); err != nil {
			return film, err
		}
	case !errors.Is(err, store.ErrNotFound):
		return film, err
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
		return merged, err
	}
	if err == nil && merged == known {
		return merged, nil
	}
	return merged, r.Store.PutClaimAbout(ctx, filmIdentity(merged.ID), holder, merged)
}

func (r *Recorder) changeTake(ctx context.Context, film record.Identity, change func(t *reading.LetterboxdTake)) error {
	var take reading.LetterboxdTake
	c, err := r.Store.ClaimOf(ctx, film, record.SourceLetterboxdTake)
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
	return r.Store.PutClaimAbout(ctx, takeIdentity(film), film, json.RawMessage(after))
}
