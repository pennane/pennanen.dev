package letterboxd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
)

type Client struct {
	Username      string
	HTTP          *http.Client
	BaseURL       string
	ShortLinkBase string
	PageInterval  time.Duration
	pages         sync.Mutex
	lastPage      time.Time
}

const letterboxdRateLimitsBackToBackPages = 5 * time.Second

func NewClient(username string) *Client {
	return &Client{
		Username:      username,
		HTTP:          &http.Client{Timeout: 30 * time.Second},
		BaseURL:       "https://letterboxd.com",
		ShortLinkBase: "https://boxd.it",
		PageInterval:  letterboxdRateLimitsBackToBackPages,
	}
}

func (c *Client) RecentFilms(ctx context.Context) ([]ListedFilm, error) {
	page, err := c.get(ctx, "/"+c.Username+"/films/")
	if err != nil {
		return nil, fmt.Errorf("films: %w", err)
	}
	return parseFilmGrid(string(page))
}

func (c *Client) RecentWatchlist(ctx context.Context) ([]ListedFilm, error) {
	page, err := c.get(ctx, "/"+c.Username+"/watchlist/")
	if err != nil {
		return nil, fmt.Errorf("watchlist: %w", err)
	}
	films, err := parseFilmGrid(string(page))
	if errors.Is(err, ErrNoFilmsSeen) {
		return []ListedFilm{}, nil
	}
	return films, err
}

func (c *Client) RecentDiary(ctx context.Context) ([]DiaryEntry, error) {
	feed, err := c.get(ctx, "/"+c.Username+"/rss/")
	if err != nil {
		return nil, fmt.Errorf("diary feed: %w", err)
	}
	return parseDiary(feed)
}

func (c *Client) get(ctx context.Context, path string) ([]byte, error) {
	return c.getURL(ctx, c.BaseURL+path)
}

func (c *Client) getURL(ctx context.Context, address string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			return nil, fmt.Errorf("%s request failed: %w", urlErr.Op, urlErr.Err)
		}
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", res.StatusCode)
	}
	return body, nil
}

var (
	tmdbLinkOnPage = regexp.MustCompile(`themoviedb\.org/(movie|tv)/(\d+)`)
	ErrNoTMDBLink  = errors.New("the film's letterboxd page links to no tmdb entry")
)

func (c *Client) TMDBLinkOf(ctx context.Context, film verdict.LetterboxdFilm) (verdict.TMDBLink, error) {
	address := c.ShortLinkBase + "/" + film.ID
	if film.Slug != "" {
		address = c.BaseURL + "/film/" + film.Slug + "/"
	}
	if err := c.waitForPageTurn(ctx); err != nil {
		return verdict.TMDBLink{}, err
	}
	page, err := c.getURL(ctx, address)
	if err != nil {
		return verdict.TMDBLink{}, fmt.Errorf("film page: %w", err)
	}
	return tmdbLinkIn(string(page))
}

func (c *Client) waitForPageTurn(ctx context.Context) error {
	c.pages.Lock()
	defer c.pages.Unlock()
	if wait := time.Until(c.lastPage.Add(c.PageInterval)); wait > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
	c.lastPage = time.Now()
	return nil
}

func tmdbLinkIn(page string) (verdict.TMDBLink, error) {
	m := tmdbLinkOnPage.FindStringSubmatch(page)
	if m == nil {
		return verdict.TMDBLink{}, ErrNoTMDBLink
	}
	id, err := strconv.ParseInt(m[2], 10, 64)
	if err != nil {
		return verdict.TMDBLink{}, err
	}
	return verdict.TMDBLink{Type: verdict.TMDBType(m[1]), ID: id}, nil
}
