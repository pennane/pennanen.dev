package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

type TimeRange string

const (
	ShortTerm  TimeRange = "short_term"
	MediumTerm TimeRange = "medium_term"
	LongTerm   TimeRange = "long_term"
)

var TimeRanges = []TimeRange{ShortTerm, MediumTerm, LongTerm}

const (
	pageSize       = 50
	artistsPerCall = 50
	albumsPerCall  = 20
	maxAttempts    = 5
	maxRetryWait   = 60 * time.Second
)

type Client struct {
	AccountsURL string
	APIURL      string
	HTTP        *http.Client

	clientID      string
	clientSecret  string
	secrets       Secrets
	tokenMu       sync.Mutex
	access        string
	accessExpires time.Time
}

type SavedAlbum struct {
	AddedAt string
	Album   json.RawMessage
}

type Play struct {
	PlayedAt string
	Track    json.RawMessage
}

func NewClient(clientID, clientSecret string, secrets Secrets) *Client {
	return &Client{
		AccountsURL:  "https://accounts.spotify.com",
		APIURL:       "https://api.spotify.com/v1",
		HTTP:         &http.Client{Timeout: 30 * time.Second},
		clientID:     clientID,
		clientSecret: clientSecret,
		secrets:      secrets,
	}
}

type page struct {
	Items []json.RawMessage `json:"items"`
	Next  *string           `json:"next"`
	Total *int              `json:"total"`
}

type listing struct {
	items []json.RawMessage
	total *int
}

func (c *Client) TopArtists(ctx context.Context, timeRange TimeRange) ([]json.RawMessage, error) {
	return c.topItems(ctx, "artists", timeRange)
}

func (c *Client) TopTracks(ctx context.Context, timeRange TimeRange) ([]json.RawMessage, error) {
	return c.topItems(ctx, "tracks", timeRange)
}

func (c *Client) topItems(ctx context.Context, what string, timeRange TimeRange) ([]json.RawMessage, error) {
	query := url.Values{"time_range": {string(timeRange)}, "limit": {strconv.Itoa(pageSize)}}
	items, err := c.pages(ctx, c.APIURL+"/me/top/"+what+"?"+query.Encode())
	if err != nil {
		return nil, fmt.Errorf("top %s (%s): %w", what, timeRange, err)
	}
	return items, nil
}

func (c *Client) SavedAlbums(ctx context.Context) ([]SavedAlbum, error) {
	listed, err := c.pageListing(ctx, c.APIURL+"/me/albums?limit="+strconv.Itoa(pageSize))
	if err != nil {
		return nil, fmt.Errorf("saved albums: %w", err)
	}
	saved := make([]SavedAlbum, 0, len(listed.items))
	distinct := map[string]struct{}{}
	for _, raw := range listed.items {
		var item struct {
			AddedAt string          `json:"added_at"`
			Album   json.RawMessage `json:"album"`
		}
		if err := json.Unmarshal(raw, &item); err != nil {
			return nil, fmt.Errorf("saved albums: %w", err)
		}
		var album struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(item.Album, &album); err != nil {
			return nil, fmt.Errorf("saved albums: %w", err)
		}
		distinct[album.ID] = struct{}{}
		saved = append(saved, SavedAlbum{AddedAt: item.AddedAt, Album: item.Album})
	}
	if listed.total == nil {
		return nil, errors.New("saved albums: spotify reported no total")
	}
	if len(distinct) != *listed.total {
		return nil, fmt.Errorf("saved albums: read %d distinct albums but spotify reports %d", len(distinct), *listed.total)
	}
	return saved, nil
}

func (c *Client) RecentlyPlayed(ctx context.Context) ([]Play, error) {
	items, err := c.pages(ctx, c.APIURL+"/me/player/recently-played?limit="+strconv.Itoa(pageSize))
	if err != nil {
		return nil, fmt.Errorf("recently played: %w", err)
	}
	plays := make([]Play, 0, len(items))
	for _, raw := range items {
		var item struct {
			PlayedAt string          `json:"played_at"`
			Track    json.RawMessage `json:"track"`
		}
		if err := json.Unmarshal(raw, &item); err != nil {
			return nil, fmt.Errorf("recently played: %w", err)
		}
		plays = append(plays, Play{PlayedAt: item.PlayedAt, Track: item.Track})
	}
	return plays, nil
}

func (c *Client) Artists(ctx context.Context, ids []string) ([]json.RawMessage, error) {
	return c.batches(ctx, "artists", ids, artistsPerCall)
}

func (c *Client) Albums(ctx context.Context, ids []string) ([]json.RawMessage, error) {
	return c.batches(ctx, "albums", ids, albumsPerCall)
}

func (c *Client) batches(ctx context.Context, what string, ids []string, perCall int) ([]json.RawMessage, error) {
	found := []json.RawMessage{}
	for batch := range slices.Chunk(ids, perCall) {
		body, err := c.get(ctx, c.APIURL+"/"+what+"?ids="+url.QueryEscape(strings.Join(batch, ",")))
		if err != nil {
			return found, fmt.Errorf("%s: %w", what, err)
		}
		var response map[string][]json.RawMessage
		if err := json.Unmarshal(body, &response); err != nil {
			return found, fmt.Errorf("%s: %w", what, err)
		}
		for _, raw := range response[what] {
			if string(raw) != "null" {
				found = append(found, raw)
			}
		}
	}
	return found, nil
}

func (c *Client) pages(ctx context.Context, first string) ([]json.RawMessage, error) {
	listed, err := c.pageListing(ctx, first)
	return listed.items, err
}

func (c *Client) pageListing(ctx context.Context, first string) (listing, error) {
	items := []json.RawMessage{}
	var total *int
	next := first
	for next != "" {
		body, err := c.get(ctx, next)
		if err != nil {
			return listing{}, err
		}
		var p page
		if err := json.Unmarshal(body, &p); err != nil {
			return listing{}, err
		}
		if next == first {
			total = p.Total
		}
		items = append(items, p.Items...)
		next = ""
		if p.Next != nil {
			next = *p.Next
		}
	}
	return listing{items: items, total: total}, nil
}

func (c *Client) get(ctx context.Context, address string) ([]byte, error) {
	reauthorisations := 0
	for attempt := 1; ; attempt++ {
		token, err := c.accessToken(ctx)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		res, err := c.HTTP.Do(req)
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(io.LimitReader(res.Body, 16<<20))
		res.Body.Close()
		if err != nil {
			return nil, err
		}
		switch {
		case res.StatusCode == http.StatusUnauthorized && reauthorisations == 0:
			reauthorisations++
			c.forgetAccess()
		case res.StatusCode == http.StatusUnauthorized:
			return nil, ErrNotConnected
		case res.StatusCode == http.StatusTooManyRequests && attempt < maxAttempts:
			wait := retryAfter(res)
			if wait > maxRetryWait {
				return nil, fmt.Errorf("spotify asks to wait %s, longer than the %s allowed", wait, maxRetryWait)
			}
			if err := sleep(ctx, wait); err != nil {
				return nil, err
			}
		case res.StatusCode != http.StatusOK:
			return nil, fmt.Errorf("spotify: %s: %s", res.Status, strings.TrimSpace(string(body)))
		default:
			return body, nil
		}
	}
}

func retryAfter(res *http.Response) time.Duration {
	seconds, err := strconv.Atoi(res.Header.Get("Retry-After"))
	if err != nil || seconds < 0 {
		return time.Second
	}
	return time.Duration(seconds) * time.Second
}

func sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
