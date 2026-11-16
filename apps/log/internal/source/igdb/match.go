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
	apps, err := m.Store.IdentitiesLackingSources(ctx, record.SourceSteam, record.SourceIGDB)
	if err != nil {
		return err
	}
	for chunk := range slices.Chunk(apps, MaxPerQuery) {
		if err := m.matchChunk(ctx, chunk, p); err != nil {
			return err
		}
	}
	return nil
}

func (m *Matcher) matchChunk(ctx context.Context, apps []record.Identity, p *syncs.Progress) error {
	appIDs := make([]string, len(apps))
	for i, app := range apps {
		appIDs[i] = app.ExternalID
	}
	p.Add(CountChecked, len(appIDs))
	gameOfApp, err := m.Client.GamesForSteamApps(ctx, appIDs)
	if err != nil {
		return err
	}
	games, err := m.Client.Games(ctx, slices.Collect(maps.Values(gameOfApp)))
	if err != nil {
		return err
	}
	for _, app := range apps {
		appID := app.ExternalID
		gameID, found := gameOfApp[appID]
		raw, fetched := games[gameID]
		if !found || !fetched {
			continue
		}
		if err := m.Store.PutClaimAbout(ctx, record.Identity{Source: record.SourceIGDB, ExternalID: fmt.Sprint(gameID)}, app, raw); err != nil {
			p.Problem("steam %s → igdb %d: %v", appID, gameID, err)
			continue
		}
		p.Add(CountMatched, 1)
		p.ChangedData()
	}
	return nil
}
