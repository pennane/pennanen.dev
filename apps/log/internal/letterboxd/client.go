package letterboxd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	Username string
	HTTP     *http.Client
	BaseURL  string
}

func NewClient(username string) *Client {
	return &Client{Username: username, HTTP: &http.Client{Timeout: 30 * time.Second}, BaseURL: "https://letterboxd.com"}
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path, nil)
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
