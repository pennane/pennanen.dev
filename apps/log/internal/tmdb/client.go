package tmdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	ReadToken string
	HTTP      *http.Client
	BaseURL   string
}

func NewClient(readToken string) *Client {
	return &Client{
		ReadToken: readToken,
		HTTP:      &http.Client{Timeout: 30 * time.Second},
		BaseURL:   "https://api.themoviedb.org",
	}
}

func (c *Client) Movie(ctx context.Context, id int64) (json.RawMessage, error) {
	var raw json.RawMessage
	q := url.Values{"append_to_response": {"credits"}, "language": {"en-US"}}
	if err := c.getJSON(ctx, "/3/movie/"+strconv.FormatInt(id, 10)+"?"+q.Encode(), &raw); err != nil {
		return nil, fmt.Errorf("movie %d: %w", id, err)
	}
	return raw, nil
}

type Found struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	OriginalTitle string `json:"original_title"`
	ReleaseDate   string `json:"release_date"`
}

func (f Found) Year() int {
	if len(f.ReleaseDate) < 4 {
		return 0
	}
	year, _ := strconv.Atoi(f.ReleaseDate[:4])
	return year
}

func (c *Client) SearchMovies(ctx context.Context, title string, year int) ([]Found, error) {
	q := url.Values{"query": {title}, "include_adult": {"false"}, "language": {"en-US"}}
	if year > 0 {
		q.Set("year", strconv.Itoa(year))
	}
	var body struct {
		Results []Found `json:"results"`
	}
	if err := c.getJSON(ctx, "/3/search/movie?"+q.Encode(), &body); err != nil {
		return nil, fmt.Errorf("search %q: %w", title, err)
	}
	return body.Results, nil
}

func (c *Client) getJSON(ctx context.Context, path string, into any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.ReadToken)
	req.Header.Set("Accept", "application/json")
	res, err := c.HTTP.Do(req)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			return fmt.Errorf("%s request failed: %w", urlErr.Op, urlErr.Err)
		}
		return err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", res.StatusCode)
	}
	return json.Unmarshal(body, into)
}
