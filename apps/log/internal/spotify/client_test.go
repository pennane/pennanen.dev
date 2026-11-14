package spotify_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/spotify"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
)

type memorySecrets struct {
	mu     sync.Mutex
	values map[string]string
}

func newSecrets(initial map[string]string) *memorySecrets {
	values := map[string]string{}
	for k, v := range initial {
		values[k] = v
	}
	return &memorySecrets{values: values}
}

func (m *memorySecrets) Secret(ctx context.Context, name string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	value, ok := m.values[name]
	if !ok {
		return "", store.ErrNotFound
	}
	return value, nil
}

func (m *memorySecrets) PutSecret(ctx context.Context, name, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.values[name] = value
	return nil
}

func clientAgainst(server *httptest.Server, secrets spotify.Secrets) *spotify.Client {
	c := spotify.NewClient("id", "secret", secrets)
	c.AccountsURL = server.URL
	c.APIURL = server.URL
	c.HTTP = server.Client()
	return c
}

func tokenResponse(w http.ResponseWriter, extra string) {
	fmt.Fprintf(w, `{"access_token":"access","token_type":"Bearer","expires_in":3600%s}`, extra)
}

func TestAuthorizeURLAsksForTheScopes(t *testing.T) {
	c := spotify.NewClient("the-id", "secret", newSecrets(nil))
	parsed, err := url.Parse(c.AuthorizeURL("https://x.test/callback", "state-1"))
	if err != nil {
		t.Fatal(err)
	}
	q := parsed.Query()
	want := map[string]string{
		"response_type": "code",
		"client_id":     "the-id",
		"redirect_uri":  "https://x.test/callback",
		"state":         "state-1",
		"scope":         "user-top-read user-library-read user-read-recently-played",
	}
	for key, value := range want {
		if q.Get(key) != value {
			t.Errorf("%s = %q, want %q", key, q.Get(key), value)
		}
	}
}

func TestConnectStoresTheRefreshToken(t *testing.T) {
	var form url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		form = r.PostForm
		tokenResponse(w, `,"refresh_token":"r1"`)
	}))
	defer server.Close()
	secrets := newSecrets(nil)
	c := clientAgainst(server, secrets)
	if err := c.Connect(context.Background(), "the-code", "https://x.test/callback"); err != nil {
		t.Fatal(err)
	}
	if form.Get("grant_type") != "authorization_code" || form.Get("code") != "the-code" {
		t.Fatalf("form = %v", form)
	}
	if got, _ := secrets.Secret(context.Background(), spotify.RefreshTokenSecret); got != "r1" {
		t.Fatalf("stored refresh token = %q", got)
	}
	connection, err := c.Connection(context.Background())
	if err != nil || connection != spotify.Connected {
		t.Fatalf("connection = %q, %v", connection, err)
	}
}

func TestConnectionIsNotConnectedWithoutAToken(t *testing.T) {
	c := spotify.NewClient("id", "secret", newSecrets(nil))
	connection, err := c.Connection(context.Background())
	if err != nil || connection != spotify.NotConnected {
		t.Fatalf("connection = %q, %v", connection, err)
	}
}

func TestARotatedRefreshTokenReplacesTheStoredOne(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/token") {
			tokenResponse(w, `,"refresh_token":"r2"`)
			return
		}
		fmt.Fprint(w, `{"items":[],"next":null}`)
	}))
	defer server.Close()
	secrets := newSecrets(map[string]string{spotify.RefreshTokenSecret: "r1"})
	c := clientAgainst(server, secrets)
	if _, err := c.TopArtists(context.Background(), spotify.ShortTerm); err != nil {
		t.Fatal(err)
	}
	if got, _ := secrets.Secret(context.Background(), spotify.RefreshTokenSecret); got != "r2" {
		t.Fatalf("stored refresh token = %q, want r2", got)
	}
}

func TestNoOrRevokedTokenIsNotConnected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"invalid_grant"}`)
	}))
	defer server.Close()

	none := clientAgainst(server, newSecrets(nil))
	if _, err := none.TopArtists(context.Background(), spotify.ShortTerm); !errors.Is(err, spotify.ErrNotConnected) {
		t.Fatalf("no secret err = %v", err)
	}
	revoked := clientAgainst(server, newSecrets(map[string]string{spotify.RefreshTokenSecret: "gone"}))
	if _, err := revoked.TopArtists(context.Background(), spotify.ShortTerm); !errors.Is(err, spotify.ErrNotConnected) {
		t.Fatalf("revoked err = %v", err)
	}
}

func TestTopArtistsFollowsNext(t *testing.T) {
	var server *httptest.Server
	var ranges []string
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/token"):
			tokenResponse(w, "")
		case r.URL.Query().Get("offset") == "2":
			ranges = append(ranges, r.URL.Query().Get("time_range"))
			fmt.Fprint(w, `{"items":[{"id":"c"}],"next":null}`)
		default:
			ranges = append(ranges, r.URL.Query().Get("time_range"))
			fmt.Fprintf(w, `{"items":[{"id":"a"},{"id":"b"}],"next":"%s/me/top/artists?time_range=short_term&limit=50&offset=2"}`, server.URL)
		}
	}))
	defer server.Close()
	c := clientAgainst(server, newSecrets(map[string]string{spotify.RefreshTokenSecret: "r1"}))
	artists, err := c.TopArtists(context.Background(), spotify.ShortTerm)
	if err != nil {
		t.Fatal(err)
	}
	if len(artists) != 3 {
		t.Fatalf("artists = %d", len(artists))
	}
	for _, got := range ranges {
		if got != "short_term" {
			t.Fatalf("time_range = %q", got)
		}
	}
}

func TestTooManyRequestsWaitsAndRetries(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/token") {
			tokenResponse(w, "")
			return
		}
		calls++
		if calls == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		fmt.Fprint(w, `{"items":[{"id":"a"}],"next":null}`)
	}))
	defer server.Close()
	c := clientAgainst(server, newSecrets(map[string]string{spotify.RefreshTokenSecret: "r1"}))
	artists, err := c.TopArtists(context.Background(), spotify.LongTerm)
	if err != nil || len(artists) != 1 || calls != 2 {
		t.Fatalf("artists = %d, calls = %d, err = %v", len(artists), calls, err)
	}
}

func TestAlbumsAreFetchedTwentyAtATime(t *testing.T) {
	var batches []int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/token") {
			tokenResponse(w, "")
			return
		}
		ids := strings.Split(r.URL.Query().Get("ids"), ",")
		batches = append(batches, len(ids))
		items := make([]string, len(ids))
		for i, id := range ids {
			items[i] = fmt.Sprintf(`{"id":%q}`, id)
		}
		fmt.Fprintf(w, `{"albums":[%s]}`, strings.Join(items, ","))
	}))
	defer server.Close()
	ids := make([]string, 45)
	for i := range ids {
		ids[i] = fmt.Sprintf("a%d", i)
	}
	c := clientAgainst(server, newSecrets(map[string]string{spotify.RefreshTokenSecret: "r1"}))
	albums, err := c.Albums(context.Background(), ids)
	if err != nil || len(albums) != 45 {
		t.Fatalf("albums = %d, %v", len(albums), err)
	}
	if fmt.Sprint(batches) != "[20 20 5]" {
		t.Fatalf("batches = %v", batches)
	}
}
