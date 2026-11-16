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
		err := s.Store.PutClaimAbout(ctx, claimOf(record.SourceSteamTake, appID), steamGame(appID), r.Take)
		if errors.Is(err, store.ErrNotFound) {
			p.Problem("review of %d: no game with that app id", r.AppID)
			continue
		}
		if err != nil {
			return err
		}
		p.Add(CountReviewsRecorded, 1)
	}
	p.ChangedData()
	return nil
}

func (s *Syncer) syncGame(ctx context.Context, g OwnedGame, p *syncs.Progress, storeRequestAllowed <-chan time.Time) error {
	appID := strconv.FormatInt(g.AppID, 10)
	game := steamGame(appID)
	_, err := s.Store.ClaimOf(ctx, game, record.SourceSteam)
	var previousPlaytime int64
	previousPlaytimeKnown := false
	switch {
	case errors.Is(err, store.ErrNotFound):
		if err = s.Store.PutClaim(ctx, game, g.Raw); err != nil {
			return err
		}
		p.Add(CountGamesAdded, 1)
	case err != nil:
		return err
	default:
		previousPlaytime, previousPlaytimeKnown = s.playtimeClaimedBefore(ctx, game)
		if err := s.Store.PutClaim(ctx, game, g.Raw); err != nil {
			return err
		}
	}

	if s.lacksClaim(ctx, game, record.SourceSteamStore) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-storeRequestAllowed:
		}
		details, err := s.Client.StoreDetails(ctx, g.AppID)
		if err != nil {
			return err
		}
		if err := s.Store.PutClaimAbout(ctx, claimOf(record.SourceSteamStore, appID), game, details); err != nil {
			return err
		}
		p.Add(CountStoreDetailsFetched, 1)
	}

	playtimeChanged := !previousPlaytimeKnown || g.PlaytimeForever != previousPlaytime
	if g.PlaytimeForever > 0 && (playtimeChanged || s.lacksClaim(ctx, game, record.SourceSteamAchievements)) {
		stats, err := s.Client.PlayerAchievements(ctx, g.AppID)
		if err != nil {
			return err
		}
		if err := s.Store.PutClaimAbout(ctx, claimOf(record.SourceSteamAchievements, appID), game, stats); err != nil {
			return err
		}
		p.Add(CountAchievementsFetched, 1)
	}
	return nil
}

func (s *Syncer) playtimeClaimedBefore(ctx context.Context, game record.Identity) (int64, bool) {
	c, err := s.Store.ClaimOf(ctx, game, record.SourceSteam)
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

func (s *Syncer) lacksClaim(ctx context.Context, game record.Identity, source string) bool {
	_, err := s.Store.ClaimOf(ctx, game, source)
	return errors.Is(err, store.ErrNotFound)
}

func steamGame(appID string) record.Identity {
	return claimOf(record.SourceSteam, appID)
}

func claimOf(source, externalID string) record.Identity {
	return record.Identity{Source: source, ExternalID: externalID}
}
