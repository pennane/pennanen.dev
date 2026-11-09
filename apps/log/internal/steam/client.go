package steam

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
	"time"
)

type Client struct {
	APIKey        string
	SteamID       string
	HTTP          *http.Client
	WebAPIBaseURL string
	StoreBaseURL  string
}

func NewClient(apiKey, steamID string) *Client {
	return &Client{
		APIKey:        apiKey,
		SteamID:       steamID,
		HTTP:          &http.Client{Timeout: 30 * time.Second},
		WebAPIBaseURL: "https://api.steampowered.com",
		StoreBaseURL:  "https://store.steampowered.com",
	}
}

type OwnedGame struct {
	AppID           int64
	Name            string
	PlaytimeForever int64
	Raw             json.RawMessage
}

func (c *Client) OwnedGames(ctx context.Context) ([]OwnedGame, error) {
	q := url.Values{
		"key":                       {c.APIKey},
		"steamid":                   {c.SteamID},
		"include_appinfo":           {"1"},
		"include_played_free_games": {"1"},
		"format":                    {"json"},
	}
	var body struct {
		Response struct {
			Games []json.RawMessage `json:"games"`
		} `json:"response"`
	}
	if err := c.getJSON(ctx, c.WebAPIBaseURL+"/IPlayerService/GetOwnedGames/v1/?"+q.Encode(), &body, http.StatusOK); err != nil {
		return nil, fmt.Errorf("owned games: %w", err)
	}
	games := make([]OwnedGame, 0, len(body.Response.Games))
	for _, raw := range body.Response.Games {
		var g struct {
			AppID           int64  `json:"appid"`
			Name            string `json:"name"`
			PlaytimeForever int64  `json:"playtime_forever"`
		}
		if err := json.Unmarshal(raw, &g); err != nil {
			return nil, err
		}
		games = append(games, OwnedGame{AppID: g.AppID, Name: g.Name, PlaytimeForever: g.PlaytimeForever, Raw: raw})
	}
	return games, nil
}

var UnavailableStoreDetails = json.RawMessage(`{}`)

func (c *Client) StoreDetails(ctx context.Context, appID int64) (json.RawMessage, error) {
	id := strconv.FormatInt(appID, 10)
	q := url.Values{"appids": {id}, "l": {"english"}, "cc": {"us"}}
	var body map[string]struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}
	if err := c.getJSON(ctx, c.StoreBaseURL+"/api/appdetails?"+q.Encode(), &body, http.StatusOK); err != nil {
		return nil, fmt.Errorf("store details %d: %w", appID, err)
	}
	details, listed := body[id]
	if !listed || !details.Success || len(details.Data) == 0 {
		return UnavailableStoreDetails, nil
	}
	return details.Data, nil
}

func (c *Client) PlayerAchievements(ctx context.Context, appID int64) (json.RawMessage, error) {
	q := url.Values{"key": {c.APIKey}, "steamid": {c.SteamID}, "appid": {strconv.FormatInt(appID, 10)}}
	var body struct {
		PlayerStats json.RawMessage `json:"playerstats"`
	}
	gameWithoutStatsAnswersBadRequest := http.StatusBadRequest
	if err := c.getJSON(ctx, c.WebAPIBaseURL+"/ISteamUserStats/GetPlayerAchievements/v1/?"+q.Encode(), &body, http.StatusOK, gameWithoutStatsAnswersBadRequest); err != nil {
		return nil, fmt.Errorf("achievements %d: %w", appID, err)
	}
	if len(body.PlayerStats) == 0 {
		return nil, fmt.Errorf("achievements %d: no playerstats in response", appID)
	}
	return body.PlayerStats, nil
}

func (c *Client) getJSON(ctx context.Context, endpoint string, into any, acceptedStatuses ...int) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return withoutURL(err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 32<<20))
	if err != nil {
		return err
	}
	if !slices.Contains(acceptedStatuses, res.StatusCode) {
		return fmt.Errorf("status %d", res.StatusCode)
	}
	return json.Unmarshal(body, into)
}

func withoutURL(err error) error {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return fmt.Errorf("%s request failed: %w", urlErr.Op, urlErr.Err)
	}
	return err
}
