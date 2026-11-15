package tmdb

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
	"github.com/pennane/pennanen.dev/apps/log/internal/source/letterboxd"
	"github.com/pennane/pennanen.dev/apps/log/internal/source/syncs"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

const SyncName = "tmdb-match"

const (
	CountChecked   = "checked"
	CountMatched   = "matched"
	CountUnmatched = "unmatched"
)

type LetterboxdPages interface {
	TMDBLinkOf(ctx context.Context, film reading.LetterboxdFilm) (reading.TMDBLink, error)
}

type Matcher struct {
	Store           *store.Store
	Client          *Client
	Pages           LetterboxdPages
	RequestInterval time.Duration
}

const tmdbAllowsAboutFortyRequestsASecond = 50 * time.Millisecond

func NewMatcher(s *store.Store, c *Client, pages LetterboxdPages) *Matcher {
	return &Matcher{Store: s, Client: c, Pages: pages, RequestInterval: tmdbAllowsAboutFortyRequestsASecond}
}

func (m *Matcher) Work(ctx context.Context, p *syncs.Progress) error {
	itemOfFilm, err := m.Store.IdentitiesLackingSources(ctx, record.SourceLetterboxd, record.SourceTMDB, record.SourceTMDBSeries)
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
	c, err := m.Store.Claim(ctx, itemID, record.SourceLetterboxd)
	if err != nil {
		return err
	}
	var film reading.LetterboxdFilm
	if err := json.Unmarshal(c.Data, &film); err != nil {
		return err
	}
	link, known := film.TMDB()
	if !known {
		if link, known = m.linkedFromPage(ctx, film, p); !known {
			if link, known, err = m.searched(ctx, film, allowed); err != nil || !known {
				p.Add(CountUnmatched, 1)
				return err
			}
		}
	}
	if err := wait(ctx, allowed); err != nil {
		return err
	}
	if err := Identify(ctx, m.Store, m.Client, itemID, link); err != nil {
		return err
	}
	p.Add(CountMatched, 1)
	p.ChangedData()
	return nil
}

func (m *Matcher) linkedFromPage(ctx context.Context, film reading.LetterboxdFilm, p *syncs.Progress) (reading.TMDBLink, bool) {
	if m.Pages == nil {
		return reading.TMDBLink{}, false
	}
	link, err := m.Pages.TMDBLinkOf(ctx, film)
	if err != nil {
		if !errors.Is(err, letterboxd.ErrNoTMDBLink) {
			p.Problem("%s: %v", film.Title, err)
		}
		return reading.TMDBLink{}, false
	}
	return link, true
}

func (m *Matcher) searched(ctx context.Context, film reading.LetterboxdFilm, allowed <-chan time.Time) (reading.TMDBLink, bool, error) {
	if err := wait(ctx, allowed); err != nil {
		return reading.TMDBLink{}, false, err
	}
	found, err := m.Client.SearchMovies(ctx, film.Title, film.Year)
	if err != nil {
		return reading.TMDBLink{}, false, err
	}
	movieID, confident := onlyMatch(film, found)
	return reading.TMDBLink{Type: reading.MovieOnTMDB, ID: movieID}, confident, nil
}

func Identify(ctx context.Context, s *store.Store, c *Client, itemID int64, link reading.TMDBLink) error {
	raw, err := c.Entry(ctx, link)
	if err != nil {
		return err
	}
	return s.Identify(ctx, itemID, store.Identity{
		Kind:       kind.OfTMDB(link.Type),
		Source:     link.Type.Source(),
		ExternalID: link.ExternalID(),
		Data:       raw,
	}, tmdbSources)
}

var tmdbSources = []string{record.SourceTMDB, record.SourceTMDBSeries}

func onlyMatch(film reading.LetterboxdFilm, found []Found) (int64, bool) {
	fits := []Found{}
	for _, f := range found {
		sameYear := film.Year == 0 || f.Year() == film.Year
		if sameYear && (sameWords(f.Title, film.Title) || sameWords(f.OriginalTitle, film.Title)) {
			fits = append(fits, f)
		}
	}
	if len(fits) != 1 {
		return 0, false
	}
	return fits[0].ID, true
}

func sameWords(a, b string) bool {
	words := func(s string) string {
		return strings.Join(strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r)
		}), " ")
	}
	return words(a) == words(b)
}

func wait(ctx context.Context, allowed <-chan time.Time) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-allowed:
		return nil
	}
}
