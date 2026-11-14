package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pennane/pennanen.dev/apps/log/internal/store"
)

const RefreshTokenSecret = "spotify-refresh-token"

const scopes = "user-top-read user-library-read user-read-recently-played"

const tokenExpiryMargin = 30 * time.Second

var ErrCodeRefused = errors.New("spotify refused the authorization code")

var ErrNotConnected = errors.New("spotify is not connected; connect it in admin settings")

type Connection string

const (
	Connected    Connection = "connected"
	NotConnected Connection = "not-connected"
)

type Secrets interface {
	Secret(ctx context.Context, name string) (string, error)
	PutSecret(ctx context.Context, name, value string) error
	DeleteSecret(ctx context.Context, name string) error
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type tokenError struct {
	Error string `json:"error"`
}

func (c *Client) AuthorizeURL(redirectURI, state string) string {
	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", c.clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("state", state)
	q.Set("scope", scopes)
	return c.AccountsURL + "/authorize?" + q.Encode()
}

func (c *Client) Connect(ctx context.Context, code, redirectURI string) error {
	token, err := c.requestToken(ctx, url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {redirectURI},
	})
	if err != nil {
		return err
	}
	if token.RefreshToken == "" {
		return errors.New("spotify returned no refresh token")
	}
	if err := c.secrets.PutSecret(ctx, RefreshTokenSecret, token.RefreshToken); err != nil {
		return err
	}
	c.remember(token)
	return nil
}

func (c *Client) Connection(ctx context.Context) (Connection, error) {
	_, err := c.secrets.Secret(ctx, RefreshTokenSecret)
	switch {
	case errors.Is(err, store.ErrNotFound):
		return NotConnected, nil
	case err != nil:
		return "", err
	}
	return Connected, nil
}

func (c *Client) accessToken(ctx context.Context) (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	if c.access != "" && time.Now().Before(c.accessExpires) {
		return c.access, nil
	}
	refresh, err := c.secrets.Secret(ctx, RefreshTokenSecret)
	if errors.Is(err, store.ErrNotFound) {
		return "", ErrNotConnected
	}
	if err != nil {
		return "", err
	}
	token, err := c.requestToken(ctx, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refresh},
	})
	if errors.Is(err, ErrNotConnected) {
		if err := c.secrets.DeleteSecret(ctx, RefreshTokenSecret); err != nil {
			return "", err
		}
		return "", ErrNotConnected
	}
	if err != nil {
		return "", err
	}
	if token.RefreshToken != "" && token.RefreshToken != refresh {
		if err := c.secrets.PutSecret(ctx, RefreshTokenSecret, token.RefreshToken); err != nil {
			return "", err
		}
	}
	c.access = token.AccessToken
	c.accessExpires = time.Now().Add(time.Duration(token.ExpiresIn)*time.Second - tokenExpiryMargin)
	return c.access, nil
}

func (c *Client) forgetAccess() {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	c.access = ""
}

func (c *Client) remember(token tokenResponse) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	c.access = token.AccessToken
	c.accessExpires = time.Now().Add(time.Duration(token.ExpiresIn)*time.Second - tokenExpiryMargin)
}

func (c *Client) requestToken(ctx context.Context, form url.Values) (tokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.AccountsURL+"/api/token", strings.NewReader(form.Encode()))
	if err != nil {
		return tokenResponse{}, err
	}
	req.SetBasicAuth(c.clientID, c.clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := c.HTTP.Do(req)
	if err != nil {
		return tokenResponse{}, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return tokenResponse{}, err
	}
	if res.StatusCode != http.StatusOK {
		var failure tokenError
		if json.Unmarshal(body, &failure) == nil && failure.Error == "invalid_grant" && form.Get("grant_type") == "refresh_token" {
			return tokenResponse{}, ErrNotConnected
		}
		detail := fmt.Errorf("spotify token: %s: %s", res.Status, strings.TrimSpace(string(body)))
		if form.Get("grant_type") == "authorization_code" && res.StatusCode >= 400 && res.StatusCode < 500 {
			return tokenResponse{}, fmt.Errorf("%w: %w", ErrCodeRefused, detail)
		}
		return tokenResponse{}, detail
	}
	var token tokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return tokenResponse{}, err
	}
	return token, nil
}
