package letterboxd

import (
	"context"

	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/syncs"
)

const SyncName = "letterboxd"

type Syncer struct {
	Recorder *Recorder
	Client   *Client
}

func NewSyncer(s *store.Store, c *Client) *Syncer {
	return &Syncer{Recorder: &Recorder{Store: s, Links: c}, Client: c}
}

func (s *Syncer) Work(ctx context.Context, p *syncs.Progress) error {
	watchlist, err := s.Client.RecentWatchlist(ctx)
	if err != nil {
		return err
	}
	films, err := s.Client.RecentFilms(ctx)
	if err != nil {
		return err
	}
	diary, err := s.Client.RecentDiary(ctx)
	if err != nil {
		return err
	}
	p.ChangedData()
	for _, listed := range watchlist {
		if _, err := s.Recorder.Shelve(ctx, p, listed.Film, WatchlistShelf); err != nil {
			p.Problem("%s: %v", listed.Film.Title, err)
		}
	}
	idOfSlug := map[string]string{}
	for _, listed := range films {
		idOfSlug[listed.Film.Slug] = listed.Film.ID
		s.shelveWatched(ctx, p, listed)
	}
	for _, entry := range oldestFirst(diary) {
		if err := s.Recorder.RecordWatch(ctx, p, entry, idOfSlug); err != nil {
			p.Problem("%v", err)
		}
	}
	return ctx.Err()
}

func (s *Syncer) shelveWatched(ctx context.Context, p *syncs.Progress, listed ListedFilm) (int64, error) {
	itemID, err := s.Recorder.Shelve(ctx, p, listed.Film, WatchedShelf)
	if err == nil {
		err = s.Recorder.PutListTake(ctx, itemID, listed)
	}
	if err != nil {
		p.Problem("%s: %v", listed.Film.Title, err)
	}
	return itemID, err
}

func (s *Syncer) Import(ctx context.Context, p *syncs.Progress, export Export) error {
	p.ChangedData()
	onWatchlist := map[int64]struct{}{}
	for _, film := range export.Watchlist {
		itemID, err := s.Recorder.Shelve(ctx, p, film, WatchlistShelf)
		if err != nil {
			p.Problem("%s: %v", film.Title, err)
			continue
		}
		onWatchlist[itemID] = struct{}{}
	}
	for _, listed := range export.Watched {
		itemID, err := s.shelveWatched(ctx, p, listed)
		if _, stillWanted := onWatchlist[itemID]; err == nil && !stillWanted {
			if err := s.Recorder.LeaveWatchlist(ctx, p, itemID); err != nil {
				p.Problem("%s: %v", listed.Film.Title, err)
			}
		}
	}
	for _, entry := range oldestFirst(export.Diary) {
		if err := s.Recorder.RecordWatch(ctx, p, entry, map[string]string{}); err != nil {
			p.Problem("%v", err)
		}
	}
	return ctx.Err()
}
