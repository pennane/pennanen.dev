package igdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
)

type MatchState string

const (
	NeverMatched MatchState = "never-matched"
	Matching     MatchState = "matching"
	Matched      MatchState = "matched"
	MatchFailed  MatchState = "failed"
)

type Report struct {
	State      MatchState `json:"state"`
	StartedAt  string     `json:"startedAt,omitempty"`
	FinishedAt string     `json:"finishedAt,omitempty"`
	Checked    int        `json:"checked"`
	Matched    int        `json:"matched"`
	Problems   []string   `json:"problems"`
}

var NoMatch = json.RawMessage(`{"match":"none"}`)

type Matcher struct {
	Store    *store.Store
	Client   *Client
	OnChange func()

	mu   sync.Mutex
	last Report
}

func NewMatcher(s *store.Store, c *Client, onChange func()) *Matcher {
	return &Matcher{Store: s, Client: c, OnChange: onChange, last: Report{State: NeverMatched, Problems: []string{}}}
}

var ErrAlreadyMatching = errors.New("already matching")

func (m *Matcher) Last() Report {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.last
}

func (m *Matcher) Start(ctx context.Context) error {
	if !m.begin() {
		return ErrAlreadyMatching
	}
	go m.run(ctx)
	return nil
}

func (m *Matcher) Run(ctx context.Context) Report {
	if m.begin() {
		m.run(ctx)
	}
	return m.Last()
}

func (m *Matcher) begin() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.last.State == Matching {
		return false
	}
	m.last = Report{State: Matching, StartedAt: now(), Problems: []string{}}
	return true
}

func (m *Matcher) run(ctx context.Context) {
	report := m.Last()
	err := m.matchSteamGames(ctx, &report)
	report.FinishedAt = now()
	report.State = Matched
	if err != nil {
		report.State = MatchFailed
		report.Problems = append(report.Problems, err.Error())
	}
	slog.Info("igdb match finished", "state", report.State, "checked", report.Checked, "matched", report.Matched, "problems", len(report.Problems))
	m.mu.Lock()
	m.last = report
	m.mu.Unlock()
	if m.OnChange != nil && report.Matched > 0 {
		m.OnChange()
	}
}

func (m *Matcher) matchSteamGames(ctx context.Context, report *Report) error {
	itemOfApp, err := m.Store.IdentitiesLackingSource(ctx, kind.SourceSteam, kind.SourceIGDB)
	if err != nil {
		return err
	}
	appIDs := slices.Sorted(maps.Keys(itemOfApp))
	for chunk := range slices.Chunk(appIDs, MaxPerQuery) {
		if err := m.matchChunk(ctx, chunk, itemOfApp, report); err != nil {
			return err
		}
		m.mu.Lock()
		m.last = *report
		m.mu.Unlock()
	}
	return nil
}

func (m *Matcher) matchChunk(ctx context.Context, appIDs []string, itemOfApp map[string]int64, report *Report) error {
	report.Checked += len(appIDs)
	gameOfApp, err := m.Client.GamesForSteamApps(ctx, appIDs)
	if err != nil {
		return err
	}
	games, err := m.Client.Games(ctx, slices.Collect(maps.Values(gameOfApp)))
	if err != nil {
		return err
	}
	for _, appID := range appIDs {
		gameID, found := gameOfApp[appID]
		raw, fetched := games[gameID]
		if !found || !fetched {
			continue
		}
		if err := m.Store.PutClaim(ctx, itemOfApp[appID], kind.SourceIGDB, fmt.Sprint(gameID), raw); err != nil {
			report.Problems = append(report.Problems, fmt.Sprintf("steam %s → igdb %d: %v", appID, gameID, err))
			continue
		}
		report.Matched++
	}
	return nil
}

func now() string { return time.Now().UTC().Format(time.RFC3339) }
