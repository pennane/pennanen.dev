package steam

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/source/syncs"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
)

const SyncName = "steam"

const (
	CountGamesSeen           = "games-seen"
	CountGamesAdded          = "games-added"
	CountStoreDetailsFetched = "store-details-fetched"
	CountAchievementsFetched = "achievements-fetched"
	CountReviewsRecorded     = "reviews-recorded"
)

type Syncer struct {
	Store                *store.Store
	Client               *Client
	StoreRequestInterval time.Duration
}

const storeAPIToleratesAboutOneRequestPerInterval = 1500 * time.Millisecond

func NewSyncer(s *store.Store, c *Client) *Syncer {
	return &Syncer{Store: s, Client: c, StoreRequestInterval: storeAPIToleratesAboutOneRequestPerInterval}
}

func (s *Syncer) Work(ctx context.Context, p *syncs.Progress) error {
	games, err := s.Client.OwnedGames(ctx)
	if err != nil {
		return err
	}
	p.Add(CountGamesSeen, len(games))
	p.ChangedData()
	storeRequests := time.NewTicker(s.StoreRequestInterval)
	defer storeRequests.Stop()
	for _, g := range games {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := s.syncGame(ctx, g, p, storeRequests.C); err != nil {
			p.Problem("%d %s: %v", g.AppID, g.Name, err)
		}
	}
	return s.recordReviews(ctx, p)
}

func (s *Syncer) recordReviews(ctx context.Context, p *syncs.Progress) error {
	reviews, err := s.Client.Reviews(ctx)
	if err != nil {
		p.Problem("%v", err)
		return nil
	}
	for _, r := range reviews {
		appID := strconv.FormatInt(r.AppID, 10)
		itemID, err := s.Store.FindByIdentity(ctx, record.SourceSteam, appID)
		if errors.Is(err, store.ErrNotFound) {
			p.Problem("review of %d: no game with that app id", r.AppID)
			continue
		}
		if err != nil {
			return err
		}
		if err := s.Store.PutClaim(ctx, itemID, record.SourceSteamTake, appID, r.Take); err != nil {
			return err
		}
		p.Add(CountReviewsRecorded, 1)
	}
	p.ChangedData()
	return nil
}

func (s *Syncer) syncGame(ctx context.Context, g OwnedGame, p *syncs.Progress, storeRequestAllowed <-chan time.Time) error {
	appID := strconv.FormatInt(g.AppID, 10)
	itemID, err := s.Store.FindByIdentity(ctx, record.SourceSteam, appID)
	var previousPlaytime int64
	previousPlaytimeKnown := false
	switch {
	case errors.Is(err, store.ErrNotFound):
		if itemID, err = s.Store.CreateItemFromSource(ctx, "game", g.Name, record.SourceSteam, appID, g.Raw); err != nil {
			return err
		}
		p.Add(CountGamesAdded, 1)
	case err != nil:
		return err
	default:
		previousPlaytime, previousPlaytimeKnown = s.playtimeClaimedBefore(ctx, itemID)
		if err := s.Store.PutClaim(ctx, itemID, record.SourceSteam, appID, g.Raw); err != nil {
			return err
		}
	}

	if s.lacksClaim(ctx, itemID, record.SourceSteamStore) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-storeRequestAllowed:
		}
		details, err := s.Client.StoreDetails(ctx, g.AppID)
		if err != nil {
			return err
		}
		if err := s.Store.PutClaim(ctx, itemID, record.SourceSteamStore, appID, details); err != nil {
			return err
		}
		p.Add(CountStoreDetailsFetched, 1)
	}

	playtimeChanged := !previousPlaytimeKnown || g.PlaytimeForever != previousPlaytime
	if g.PlaytimeForever > 0 && (playtimeChanged || s.lacksClaim(ctx, itemID, record.SourceSteamAchievements)) {
		stats, err := s.Client.PlayerAchievements(ctx, g.AppID)
		if err != nil {
			return err
		}
		if err := s.Store.PutClaim(ctx, itemID, record.SourceSteamAchievements, appID, stats); err != nil {
			return err
		}
		p.Add(CountAchievementsFetched, 1)
	}
	return nil
}

func (s *Syncer) playtimeClaimedBefore(ctx context.Context, itemID int64) (int64, bool) {
	c, err := s.Store.Claim(ctx, itemID, record.SourceSteam)
	if err != nil {
		return 0, false
	}
	var previous struct {
		PlaytimeForever int64 `json:"playtime_forever"`
	}
	if json.Unmarshal(c.Data, &previous) != nil {
		return 0, false
	}
	return previous.PlaytimeForever, true
}

func (s *Syncer) lacksClaim(ctx context.Context, itemID int64, source string) bool {
	_, err := s.Store.Claim(ctx, itemID, source)
	return errors.Is(err, store.ErrNotFound)
}
