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
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
)

const token = "secret"

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	d, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	srv := httptest.NewServer((&api.Server{Store: store.New(d), AdminToken: token}).Handler())
	t.Cleanup(srv.Close)
	return srv
}

func call(t *testing.T, srv *httptest.Server, method, path, auth, body string) (int, string) {
	t.Helper()
	req, _ := http.NewRequest(method, srv.URL+path, strings.NewReader(body))
	if auth != "" {
		req.Header.Set("Authorization", "Bearer "+auth)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	return res.StatusCode, string(b)
}

func TestAdminRequiresToken(t *testing.T) {
	srv := newServer(t)
	if status, _ := call(t, srv, "GET", "/admin/kinds", "", ""); status != http.StatusUnauthorized {
		t.Errorf("no token: %d", status)
	}
	if status, _ := call(t, srv, "GET", "/admin/kinds", "wrong", ""); status != http.StatusUnauthorized {
		t.Errorf("wrong token: %d", status)
	}
	if status, body := call(t, srv, "GET", "/admin/kinds", token, ""); status != http.StatusOK || !strings.Contains(body, `"game"`) {
		t.Errorf("kinds: %d %s", status, body)
	}
}

func TestPublicShowsOnlyWhatIsPublic(t *testing.T) {
	srv := newServer(t)
	status, body := call(t, srv, "POST", "/admin/items", token, `{"kind":"game","title":"Outer Wilds"}`)
	if status != http.StatusCreated {
		t.Fatalf("create: %d %s", status, body)
	}
	var created struct{ ID int64 }
	json.Unmarshal([]byte(body), &created)
	id := itoa(created.ID)

	if status, _ := call(t, srv, "GET", "/public/items/game/outer-wilds", "", ""); status != http.StatusNotFound {
		t.Errorf("private item visible: %d", status)
	}

	call(t, srv, "PATCH", "/admin/items/"+id+"/entry", token,
		`{"status":"finished","visibility":"public","notes":"cried","blurb":"go play it","scores":{"enjoyment":98},"addTags":["favorite/all-time"]}`)
	call(t, srv, "POST", "/admin/items/"+id+"/moments", token, `{"at":"2024-01-02","note":"public moment"}`)
	call(t, srv, "POST", "/admin/items/"+id+"/moments", token, `{"at":"2024-01-03","note":"private moment","visibility":"private"}`)

	status, body = call(t, srv, "GET", "/public/items/game/outer-wilds", "", "")
	if status != http.StatusOK {
		t.Fatalf("public: %d %s", status, body)
	}
	for _, leaked := range []string{"cried", "private moment", "claims", "notes"} {
		if strings.Contains(body, leaked) {
			t.Errorf("public item leaks %q: %s", leaked, body)
		}
	}
	for _, shown := range []string{"go play it", "public moment", `"enjoyment":98`, "favorite/all-time"} {
		if !strings.Contains(body, shown) {
			t.Errorf("public item misses %q: %s", shown, body)
		}
	}
	if status, body := call(t, srv, "PATCH", "/admin/items/"+id+"/entry", token, `{"status":"seen"}`); status != http.StatusUnprocessableEntity {
		t.Errorf("game accepted thing status: %d %s", status, body)
	}
}

func itoa(n int64) string { b, _ := json.Marshal(n); return string(b) }
