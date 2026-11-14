package letterboxd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/syncs"
	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
)

const (
	CountFilmsSeen        = "films-seen"
	CountFilmsAdded       = "films-added"
	CountStatusesFollowed = "statuses-followed"
	CountWatchesObserved  = "watches-observed"
	CountTagsFollowed     = "tags-followed"

	SignalWatched = "watched"
	movieKind     = "movie"
)

var statusOn = map[Shelf]string{
	WatchedShelf:   kind.Watched,
	WatchlistShelf: kind.Want,
}

type Recorder struct {
	Store *store.Store
	Today func() string
}

func (r *Recorder) today() string {
	if r.Today != nil {
		return r.Today()
	}
	return time.Now().UTC().Format(time.DateOnly)
}

func (r *Recorder) Shelve(ctx context.Context, p *syncs.Progress, film verdict.LetterboxdFilm, shelf Shelf) (int64, error) {
	p.Add(CountFilmsSeen, 1)
	itemID, err := r.itemOf(ctx, film)
	switch {
	case errors.Is(err, store.ErrNotFound) && film.ID != "":
		if itemID, err = r.Store.CreateItemFromSource(ctx, movieKind, film.Title, kind.SourceLetterboxd, film.ID, film); err != nil {
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
		return itemID, r.arriveOnWatchlist(ctx, p, itemID)
	}
	return itemID, r.follow(ctx, p, itemID, shelf)
}

func (r *Recorder) arriveOnWatchlist(ctx context.Context, p *syncs.Progress, itemID int64) error {
	move := NoMove
	if err := r.changeTake(ctx, itemID, func(t *verdict.LetterboxdTake) {
		if t.OnWatchlistSince == "" {
			t.OnWatchlistSince = r.today()
			move = Arrived
		}
	}); err != nil {
		return err
	}
	watched, err := r.isWatched(ctx, itemID)
	switch {
	case err != nil:
		return err
	case watched == NotWatchedYet:
		return r.follow(ctx, p, itemID, WatchlistShelf)
	case move == Arrived:
		return r.followTag(ctx, p, itemID, store.TagPresent)
	}
	return nil
}

func (r *Recorder) LeaveWatchlist(ctx context.Context, p *syncs.Progress, itemID int64) error {
	if err := r.changeTake(ctx, itemID, func(t *verdict.LetterboxdTake) { t.OnWatchlistSince = "" }); err != nil {
		return err
	}
	return r.followTag(ctx, p, itemID, store.TagAbsent)
}

func (r *Recorder) followTag(ctx context.Context, p *syncs.Progress, itemID int64, presence store.TagPresence) error {
	outcome, err := r.Store.SetTagFromSource(ctx, itemID, kind.SourceLetterboxd, kind.TagAnotherShot, presence)
	if outcome == store.TagAdded || outcome == store.TagRemoved {
		p.Add(CountTagsFollowed, 1)
	}
	return err
}

type WatchlistMove int

const (
	NoMove WatchlistMove = iota
	Arrived
	Left
)

type Watched int

const (
	NotWatchedYet Watched = iota
	AlreadyWatched
)

func (r *Recorder) isWatched(ctx context.Context, itemID int64) (Watched, error) {
	v, err := r.Store.GetItem(ctx, itemID)
	if err != nil {
		return NotWatchedYet, err
	}
	if v.Entry != nil && v.Entry.Status == kind.Watched {
		return AlreadyWatched, nil
	}
	return NotWatchedYet, nil
}

func (r *Recorder) PutListTake(ctx context.Context, itemID int64, listed ListedFilm) error {
	return r.changeTake(ctx, itemID, func(t *verdict.LetterboxdTake) {
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
		return fmt.Errorf("watched %s (%d) on %s, but the film isn't in the log yet", entry.Film.Title, entry.Film.Year, entry.WatchedOn)
	}
	if err != nil {
		return err
	}
	if entry.Film.ID != "" {
		if err := r.remember(ctx, itemID, entry.Film); err != nil {
			return err
		}
	}
	if err := r.follow(ctx, p, itemID, WatchedShelf); err != nil {
		return err
	}
	if err := r.observe(ctx, p, itemID, entry); err != nil {
		return err
	}
	move := NoMove
	if err := r.changeTake(ctx, itemID, func(t *verdict.LetterboxdTake) {
		if entry.WatchedOn > t.LastWatched {
			t.LastWatched = entry.WatchedOn
		}
		if entry.Review != "" {
			t.Review = entry.Review
		}
		if t.OnWatchlistSince != "" && entry.WatchedOn >= t.OnWatchlistSince {
			move = Left
		}
	}); err != nil {
		return err
	}
	if move == Left {
		return r.LeaveWatchlist(ctx, p, itemID)
	}
	return nil
}

type watchObserved struct {
	Rewatch  Rewatch                    `json:"rewatch"`
	Rating   *float64                   `json:"rating,omitempty"`
	Reaction verdict.LetterboxdReaction `json:"reaction,omitempty"`
	Review   string                     `json:"review,omitempty"`
}

func (r *Recorder) observe(ctx context.Context, p *syncs.Progress, itemID int64, entry DiaryEntry) error {
	data, err := json.Marshal(watchObserved{Rewatch: entry.Rewatch, Rating: entry.Rating, Reaction: entry.Reaction, Review: entry.Review})
	if err != nil {
		return err
	}
	outcome, err := r.Store.Observe(ctx, store.Observation{
		ItemID:     itemID,
		Source:     kind.SourceLetterboxd,
		Signal:     SignalWatched,
		Key:        fmt.Sprintf("item/%d/watched/%s", itemID, entry.WatchedOn),
		ObservedAt: entry.WatchedOn,
		Data:       data,
	})
	if outcome == store.Observed {
		p.Add(CountWatchesObserved, 1)
	}
	return err
}

func (r *Recorder) itemOf(ctx context.Context, film verdict.LetterboxdFilm) (int64, error) {
	if film.ID != "" {
		itemID, err := r.Store.FindByIdentity(ctx, kind.SourceLetterboxd, film.ID)
		if !errors.Is(err, store.ErrNotFound) {
			return itemID, err
		}
	}
	if film.TMDBID > 0 {
		return r.Store.FindByIdentity(ctx, kind.SourceTMDB, strconv.FormatInt(film.TMDBID, 10))
	}
	return 0, store.ErrNotFound
}

func (r *Recorder) remember(ctx context.Context, itemID int64, film verdict.LetterboxdFilm) error {
	var known verdict.LetterboxdFilm
	if c, err := r.Store.Claim(ctx, itemID, kind.SourceLetterboxd); err == nil {
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
		merged.TMDBID = known.TMDBID
	}
	if merged.ID == "" {
		return nil
	}
	if merged == known {
		return nil
	}
	return r.Store.PutClaim(ctx, itemID, kind.SourceLetterboxd, merged.ID, merged)
}

func (r *Recorder) follow(ctx context.Context, p *syncs.Progress, itemID int64, shelf Shelf) error {
	outcome, err := r.Store.SetStatusFromSource(ctx, itemID, kind.SourceLetterboxd, statusOn[shelf])
	if outcome == store.StatusFollowed || outcome == store.StatusLogged {
		p.Add(CountStatusesFollowed, 1)
	}
	if err == nil && shelf == WatchedShelf && outcome == store.StatusFollowed {
		return r.LeaveWatchlist(ctx, p, itemID)
	}
	return err
}

func (r *Recorder) changeTake(ctx context.Context, itemID int64, change func(t *verdict.LetterboxdTake)) error {
	var take verdict.LetterboxdTake
	c, err := r.Store.Claim(ctx, itemID, kind.SourceLetterboxdTake)
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
	return r.Store.PutClaim(ctx, itemID, kind.SourceLetterboxdTake, "", json.RawMessage(after))
}
