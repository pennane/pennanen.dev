package api_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/api"
	"github.com/pennane/pennanen.dev/apps/log/internal/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/spotify"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
)

func spotifyServer(t *testing.T, accounts *httptest.Server, configured bool) *httptest.Server {
	t.Helper()
	d, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	st := store.New(d)
	server := &api.Server{Store: st, AdminToken: token}
	if configured {
		server.Spotify = spotify.NewClient("id", "secret", st)
		server.Spotify.AccountsURL = accounts.URL
	}
	srv := httptest.NewServer(server.Handler())
	t.Cleanup(srv.Close)
	return srv
}

func connectionOf(t *testing.T, srv *httptest.Server) string {
	t.Helper()
	status, body := call(t, srv, "GET", "/admin/spotify", token, "")
	if status != http.StatusOK {
		t.Fatalf("connection: %d %s", status, body)
	}
	var got struct {
		Connection string `json:"connection"`
	}
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatal(err)
	}
	return got.Connection
}

func TestSpotifyConnectionStates(t *testing.T) {
	accounts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if r.PostForm.Get("code") != "good" {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, `{"error":"invalid_grant"}`)
			return
		}
		io.WriteString(w, `{"access_token":"a","refresh_token":"r","expires_in":3600}`)
	}))
	t.Cleanup(accounts.Close)

	if got := connectionOf(t, spotifyServer(t, accounts, false)); got != "not-configured" {
		t.Errorf("without a client: %q", got)
	}

	srv := spotifyServer(t, accounts, true)
	if got := connectionOf(t, srv); got != "not-connected" {
		t.Errorf("before connecting: %q", got)
	}
	status, body := call(t, srv, "POST", "/admin/spotify/connect", token, `{"code":"bad","redirectUri":"http://x/cb"}`)
	if status != http.StatusBadRequest || !strings.Contains(body, "invalid_grant") {
		t.Errorf("refused code: %d %s", status, body)
	}
	if got := connectionOf(t, srv); got != "not-connected" {
		t.Errorf("after a refused code: %q", got)
	}
	if status, body := call(t, srv, "POST", "/admin/spotify/connect", token, `{"code":"good","redirectUri":"http://x/cb"}`); status != http.StatusNoContent {
		t.Fatalf("connect: %d %s", status, body)
	}
	if got := connectionOf(t, srv); got != "connected" {
		t.Errorf("after connecting: %q", got)
	}
}

func TestSpotifyAuthorizeNeedsRedirectAndState(t *testing.T) {
	srv := spotifyServer(t, httptest.NewServer(http.NotFoundHandler()), true)

	for _, query := range []string{"", "?redirect_uri=http://x/cb", "?state=s"} {
		if status, _ := call(t, srv, "GET", "/admin/spotify/authorize"+query, token, ""); status != http.StatusBadRequest {
			t.Errorf("%q: %d", query, status)
		}
	}
	status, body := call(t, srv, "GET", "/admin/spotify/authorize?redirect_uri=http://x/cb&state=s", token, "")
	if status != http.StatusOK || !strings.Contains(body, "client_id=id") || !strings.Contains(body, "state=s") {
		t.Errorf("authorize: %d %s", status, body)
	}
}

func TestSpotifyRoutesNeedAdminAndAConfiguredClient(t *testing.T) {
	srv := spotifyServer(t, httptest.NewServer(http.NotFoundHandler()), false)

	if status, _ := call(t, srv, "GET", "/admin/spotify", "", ""); status != http.StatusUnauthorized {
		t.Errorf("without a token: %d", status)
	}
	if status, _ := call(t, srv, "GET", "/admin/spotify/authorize?redirect_uri=a&state=b", token, ""); status != http.StatusNotImplemented {
		t.Errorf("authorize unconfigured: %d", status)
	}
	if status, _ := call(t, srv, "POST", "/admin/spotify/connect", token, `{"code":"c","redirectUri":"r"}`); status != http.StatusNotImplemented {
		t.Errorf("connect unconfigured: %d", status)
	}
}
