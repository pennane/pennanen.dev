package igdb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
)

type Client struct {
	ClientID     string
	ClientSecret string
	HTTP         *http.Client
	TokenURL     string
	APIURL       string

	tokenMu      sync.Mutex
	token        string
	tokenExpires time.Time

	requestAllowed <-chan time.Time
}

const (
	apiAllowsFourRequestsPerSecond = 260 * time.Millisecond
	refreshTokenBeforeExpiry       = time.Hour
	MaxPerQuery                    = 500
)

func NewClient(clientID, clientSecret string) *Client {
	return &Client{
		ClientID:       clientID,
		ClientSecret:   clientSecret,
		HTTP:           &http.Client{Timeout: 30 * time.Second},
		TokenURL:       "https://id.twitch.tv/oauth2/token",
		APIURL:         "https://api.igdb.com/v4",
		requestAllowed: time.Tick(apiAllowsFourRequestsPerSecond),
	}
}

const gameFields = `fields name, first_release_date, game_type, cover.image_id, artworks.image_id, screenshots.image_id, ` +
	`genres.name, platforms.name, platforms.abbreviation, involved_companies.developer, involved_companies.company.name;`

var separateGames = []int{mainGame, remake, remaster, expandedGame, port, standaloneExpansion, episode, fork}

const (
	mainGame            = 0
	standaloneExpansion = 4
	episode             = 6
	remake              = 8
	remaster            = 9
	expandedGame        = 10
	port                = 11
	fork                = 12
)

const SearchLimit = 20

func (c *Client) Search(ctx context.Context, query string) ([]json.RawMessage, error) {
	body := fmt.Sprintf(`search %s; %s where game_type = (%s) & version_parent = null; limit %d;`,
		quote(query), gameFields, joinInts(separateGames), SearchLimit)
	var games []json.RawMessage
	return games, c.query(ctx, "games", body, &games)
}

func (c *Client) Games(ctx context.Context, ids []int64) (map[int64]json.RawMessage, error) {
	if len(ids) == 0 {
		return map[int64]json.RawMessage{}, nil
	}
	body := fmt.Sprintf(`%s where id = (%s); limit %d;`, gameFields, joinInts64(ids), MaxPerQuery)
	var games []json.RawMessage
	if err := c.query(ctx, "games", body, &games); err != nil {
		return nil, err
	}
	byID := map[int64]json.RawMessage{}
	for _, raw := range games {
		var g verdict.IGDBGame
		if err := json.Unmarshal(raw, &g); err == nil && g.Matched() {
			byID[g.ID] = raw
		}
	}
	return byID, nil
}

const steamSource = 1

func (c *Client) GamesForSteamApps(ctx context.Context, appIDs []string) (map[string]int64, error) {
	if len(appIDs) == 0 {
		return map[string]int64{}, nil
	}
	quoted := make([]string, len(appIDs))
	for i, id := range appIDs {
		quoted[i] = quote(id)
	}
	body := fmt.Sprintf(`fields uid, game; where external_game_source = %d & uid = (%s); limit %d;`,
		steamSource, strings.Join(quoted, ","), MaxPerQuery)
	var rows []struct {
		UID  string `json:"uid"`
		Game int64  `json:"game"`
	}
	if err := c.query(ctx, "external_games", body, &rows); err != nil {
		return nil, err
	}
	gameOf := map[string]int64{}
	for _, row := range rows {
		if row.Game != 0 {
			gameOf[row.UID] = row.Game
		}
	}
	return gameOf, nil
}

var errUnauthorized = errors.New("igdb refused the token")

func (c *Client) query(ctx context.Context, endpoint, body string, into any) error {
	err := c.queryOnce(ctx, endpoint, body, into)
	if errors.Is(err, errUnauthorized) {
		c.forgetToken()
		err = c.queryOnce(ctx, endpoint, body, into)
	}
	return err
}

func (c *Client) queryOnce(ctx context.Context, endpoint, body string, into any) error {
	token, err := c.accessToken(ctx)
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.requestAllowed:
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.APIURL+"/"+endpoint, strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Client-ID", c.ClientID)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	res, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("igdb %s: %w", endpoint, err)
	}
	defer res.Body.Close()
	payload, _ := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	switch {
	case res.StatusCode == http.StatusUnauthorized:
		return errUnauthorized
	case res.StatusCode != http.StatusOK:
		return fmt.Errorf("igdb %s answered %d: %s", endpoint, res.StatusCode, bytes.TrimSpace(payload))
	}
	return json.Unmarshal(payload, into)
}

func (c *Client) accessToken(ctx context.Context) (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	if c.token != "" && time.Now().Before(c.tokenExpires) {
		return c.token, nil
	}
	form := url.Values{
		"client_id":     {c.ClientID},
		"client_secret": {c.ClientSecret},
		"grant_type":    {"client_credentials"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := c.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("twitch token: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("twitch token answered %d", res.StatusCode)
	}
	var body struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return "", err
	}
	c.token = body.AccessToken
	c.tokenExpires = time.Now().Add(time.Duration(body.ExpiresIn)*time.Second - refreshTokenBeforeExpiry)
	return c.token, nil
}

func (c *Client) forgetToken() {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	c.token = ""
}

func quote(s string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s) + `"`
}

func joinInts(xs []int) string {
	parts := make([]string, len(xs))
	for i, x := range xs {
		parts[i] = strconv.Itoa(x)
	}
	return strings.Join(parts, ",")
}

func joinInts64(xs []int64) string {
	parts := make([]string, len(xs))
	for i, x := range xs {
		parts[i] = strconv.FormatInt(x, 10)
	}
	return strings.Join(parts, ",")
}
