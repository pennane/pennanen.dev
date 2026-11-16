package igdb

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/source/syncs"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
)

const SyncName = "igdb-match"

const (
	CountChecked = "checked"
	CountMatched = "matched"
)

type Matcher struct {
	Store  *store.Store
	Client *Client
}

func NewMatcher(s *store.Store, c *Client) *Matcher {
	return &Matcher{Store: s, Client: c}
}

func (m *Matcher) Work(ctx context.Context, p *syncs.Progress) error {
	itemOfApp, err := m.Store.IdentitiesLackingSourcesItemKeyed(ctx, record.SourceSteam, record.SourceIGDB)
	if err != nil {
		return err
	}
	appIDs := slices.Sorted(maps.Keys(itemOfApp))
	for chunk := range slices.Chunk(appIDs, MaxPerQuery) {
		if err := m.matchChunk(ctx, chunk, itemOfApp, p); err != nil {
			return err
		}
	}
	return nil
}

func (m *Matcher) matchChunk(ctx context.Context, appIDs []string, itemOfApp map[string]int64, p *syncs.Progress) error {
	p.Add(CountChecked, len(appIDs))
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
		if err := m.Store.PutClaimItemKeyed(ctx, itemOfApp[appID], record.SourceIGDB, fmt.Sprint(gameID), raw); err != nil {
			p.Problem("steam %s → igdb %d: %v", appID, gameID, err)
			continue
		}
		p.Add(CountMatched, 1)
		p.ChangedData()
	}
	return nil
}
