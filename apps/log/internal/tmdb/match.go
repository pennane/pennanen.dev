package tmdb

import (
	"context"
	"encoding/json"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/syncs"
	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
)

const SyncName = "tmdb-match"

const (
	CountChecked   = "checked"
	CountMatched   = "matched"
	CountUnmatched = "unmatched"
)

type Matcher struct {
	Store           *store.Store
	Client          *Client
	RequestInterval time.Duration
}

const tmdbAllowsAboutFortyRequestsASecond = 50 * time.Millisecond

func NewMatcher(s *store.Store, c *Client) *Matcher {
	return &Matcher{Store: s, Client: c, RequestInterval: tmdbAllowsAboutFortyRequestsASecond}
}

func (m *Matcher) Work(ctx context.Context, p *syncs.Progress) error {
	itemOfFilm, err := m.Store.IdentitiesLackingSource(ctx, kind.SourceLetterboxd, kind.SourceTMDB)
	if err != nil {
		return err
	}
	requests := time.NewTicker(m.RequestInterval)
	defer requests.Stop()
	for _, filmID := range slices.Sorted(maps.Keys(itemOfFilm)) {
		if err := ctx.Err(); err != nil {
			return err
		}
		p.Add(CountChecked, 1)
		if err := m.match(ctx, itemOfFilm[filmID], p, requests.C); err != nil {
			p.Problem("letterboxd %s: %v", filmID, err)
		}
	}
	return nil
}

func (m *Matcher) match(ctx context.Context, itemID int64, p *syncs.Progress, allowed <-chan time.Time) error {
	c, err := m.Store.Claim(ctx, itemID, kind.SourceLetterboxd)
	if err != nil {
		return err
	}
	var film verdict.LetterboxdFilm
	if err := json.Unmarshal(c.Data, &film); err != nil {
		return err
	}
	movieID := film.TMDBID
	if movieID == 0 {
		if err := wait(ctx, allowed); err != nil {
			return err
		}
		found, err := m.Client.SearchMovies(ctx, film.Title, film.Year)
		if err != nil {
			return err
		}
		var confident bool
		if movieID, confident = onlyMatch(film, found); !confident {
			p.Add(CountUnmatched, 1)
			return nil
		}
	}
	if err := wait(ctx, allowed); err != nil {
		return err
	}
	raw, err := m.Client.Movie(ctx, movieID)
	if err != nil {
		return err
	}
	if err := m.Store.PutClaim(ctx, itemID, kind.SourceTMDB, strconv.FormatInt(movieID, 10), raw); err != nil {
		return err
	}
	p.Add(CountMatched, 1)
	p.ChangedData()
	return nil
}

func onlyMatch(film verdict.LetterboxdFilm, found []Found) (int64, bool) {
	sameYear := []Found{}
	for _, f := range found {
		if film.Year == 0 || f.Year() == film.Year {
			sameYear = append(sameYear, f)
		}
	}
	sameTitle := []Found{}
	for _, f := range sameYear {
		if sameWords(f.Title, film.Title) || sameWords(f.OriginalTitle, film.Title) {
			sameTitle = append(sameTitle, f)
		}
	}
	switch {
	case len(sameTitle) == 1:
		return sameTitle[0].ID, true
	case len(sameTitle) == 0 && len(sameYear) == 1:
		return sameYear[0].ID, true
	}
	return 0, false
}

func sameWords(a, b string) bool {
	return strings.EqualFold(strings.Join(strings.Fields(a), " "), strings.Join(strings.Fields(b), " "))
}

func wait(ctx context.Context, allowed <-chan time.Time) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-allowed:
		return nil
	}
}
