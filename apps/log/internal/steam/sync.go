package steam

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
)

type SyncState string

const (
	NeverSynced SyncState = "never-synced"
	Syncing     SyncState = "syncing"
	Synced      SyncState = "synced"
	SyncFailed  SyncState = "failed"
)

type Report struct {
	State               SyncState `json:"state"`
	StartedAt           string    `json:"startedAt,omitempty"`
	FinishedAt          string    `json:"finishedAt,omitempty"`
	GamesSeen           int       `json:"gamesSeen"`
	GamesAdded          int       `json:"gamesAdded"`
	StoreDetailsFetched int       `json:"storeDetailsFetched"`
	AchievementsFetched int       `json:"achievementsFetched"`
	Problems            []string  `json:"problems"`
}

type Syncer struct {
	Store                *store.Store
	Client               *Client
	StoreRequestInterval time.Duration
	OnChange             func()

	mu   sync.Mutex
	last Report
}

const storeAPIToleratesAboutOneRequestPerInterval = 1500 * time.Millisecond

func NewSyncer(s *store.Store, c *Client, onChange func()) *Syncer {
	return &Syncer{
		Store:                s,
		Client:               c,
		StoreRequestInterval: storeAPIToleratesAboutOneRequestPerInterval,
		OnChange:             onChange,
		last:                 Report{State: NeverSynced, Problems: []string{}},
	}
}

func (s *Syncer) Last() Report {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.last
}

var ErrAlreadySyncing = errors.New("already syncing")

func (s *Syncer) Start(ctx context.Context) error {
	if !s.begin() {
		return ErrAlreadySyncing
	}
	go s.run(ctx)
	return nil
}

func (s *Syncer) Run(ctx context.Context) Report {
	if s.begin() {
		s.run(ctx)
	}
	return s.Last()
}

func (s *Syncer) begin() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.last.State == Syncing {
		return false
	}
	s.last = Report{State: Syncing, StartedAt: now(), Problems: []string{}}
	return true
}

func (s *Syncer) run(ctx context.Context) {
	report := s.Last()
	err := s.syncAll(ctx, &report)
	report.FinishedAt = now()
	report.State = Synced
	if err != nil {
		report.State = SyncFailed
		report.Problems = append(report.Problems, err.Error())
	}
	slog.Info("steam sync finished", "state", report.State, "seen", report.GamesSeen, "added", report.GamesAdded,
		"storeDetails", report.StoreDetailsFetched, "achievements", report.AchievementsFetched, "problems", len(report.Problems))
	s.mu.Lock()
	s.last = report
	s.mu.Unlock()
	if s.OnChange != nil {
		s.OnChange()
	}
}

func (s *Syncer) syncAll(ctx context.Context, report *Report) error {
	games, err := s.Client.OwnedGames(ctx)
	if err != nil {
		return err
	}
	report.GamesSeen = len(games)
	storeRequests := time.NewTicker(s.StoreRequestInterval)
	defer storeRequests.Stop()
	for _, g := range games {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := s.syncGame(ctx, g, report, storeRequests.C); err != nil {
			report.Problems = append(report.Problems, fmt.Sprintf("%d %s: %v", g.AppID, g.Name, err))
		}
		s.mu.Lock()
		s.last = *report
		s.mu.Unlock()
	}
	return nil
}

func (s *Syncer) syncGame(ctx context.Context, g OwnedGame, report *Report, storeRequestAllowed <-chan time.Time) error {
	appID := strconv.FormatInt(g.AppID, 10)
	itemID, err := s.Store.FindByIdentity(ctx, kind.SourceSteam, appID)
	var previousPlaytime int64
	previousPlaytimeKnown := false
	switch {
	case errors.Is(err, store.ErrNotFound):
		if itemID, err = s.Store.CreateItemFromSource(ctx, "game", g.Name, kind.SourceSteam, appID, g.Raw); err != nil {
			return err
		}
		report.GamesAdded++
	case err != nil:
		return err
	default:
		previousPlaytime, previousPlaytimeKnown = s.playtimeClaimedBefore(ctx, itemID)
		if err := s.Store.PutClaim(ctx, itemID, kind.SourceSteam, appID, g.Raw); err != nil {
			return err
		}
	}

	if s.lacksClaim(ctx, itemID, kind.SourceSteamStore) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-storeRequestAllowed:
		}
		details, err := s.Client.StoreDetails(ctx, g.AppID)
		if err != nil {
			return err
		}
		if err := s.Store.PutClaim(ctx, itemID, kind.SourceSteamStore, appID, details); err != nil {
			return err
		}
		report.StoreDetailsFetched++
	}

	playtimeChanged := !previousPlaytimeKnown || g.PlaytimeForever != previousPlaytime
	if g.PlaytimeForever > 0 && (playtimeChanged || s.lacksClaim(ctx, itemID, kind.SourceSteamAchievements)) {
		stats, err := s.Client.PlayerAchievements(ctx, g.AppID)
		if err != nil {
			return err
		}
		if err := s.Store.PutClaim(ctx, itemID, kind.SourceSteamAchievements, appID, stats); err != nil {
			return err
		}
		report.AchievementsFetched++
	}
	return nil
}

func (s *Syncer) playtimeClaimedBefore(ctx context.Context, itemID int64) (int64, bool) {
	c, err := s.Store.Claim(ctx, itemID, kind.SourceSteam)
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

func now() string { return time.Now().UTC().Format(time.RFC3339) }
