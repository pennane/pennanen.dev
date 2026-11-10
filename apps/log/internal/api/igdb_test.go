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
	"github.com/pennane/pennanen.dev/apps/log/internal/igdb"
	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
)

const (
	minecraft = `{"id":121,"name":"Minecraft","first_release_date":1321488000,"cover":{"image_id":"co49x5"},` +
		`"artworks":[{"image_id":"ar1"}],"genres":[{"name":"Sandbox"}],"platforms":[{"name":"PC (Microsoft Windows)","abbreviation":"PC"}],` +
		`"involved_companies":[{"developer":true,"company":{"name":"Mojang"}},{"developer":false,"company":{"name":"Microsoft"}}]}`
	hades = `{"id":113112,"name":"Hades","first_release_date":1600300800,"cover":{"image_id":"co39vc"}}`
)

type fakeIGDB struct {
	queries []string
}

func (f *fakeIGDB) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	query := string(body)
	f.queries = append(f.queries, r.URL.Path+" "+query)
	switch {
	case r.URL.Path == "/token":
		io.WriteString(w, `{"access_token":"t","expires_in":5000000}`)
	case r.Header.Get("Authorization") != "Bearer t":
		w.WriteHeader(http.StatusUnauthorized)
	case r.URL.Path == "/v4/external_games":
		io.WriteString(w, `[{"uid":"1145360","game":113112}]`)
	case r.URL.Path == "/v4/games" && strings.HasPrefix(query, "search"):
		io.WriteString(w, `[`+minecraft+`]`)
	case r.URL.Path == "/v4/games" && strings.Contains(query, "121") && strings.Contains(query, "113112"):
		io.WriteString(w, `[`+minecraft+`,`+hades+`]`)
	case r.URL.Path == "/v4/games" && strings.Contains(query, "113112"):
		io.WriteString(w, `[`+hades+`]`)
	case r.URL.Path == "/v4/games" && strings.Contains(query, "121"):
		io.WriteString(w, `[`+minecraft+`]`)
	default:
		io.WriteString(w, `[]`)
	}
}

func newIGDBServer(t *testing.T) (*httptest.Server, *store.Store, *igdb.Matcher, *fakeIGDB) {
	t.Helper()
	d, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	fake := &fakeIGDB{}
	upstream := httptest.NewServer(fake)
	t.Cleanup(upstream.Close)
	client := igdb.NewClient("id", "secret")
	client.TokenURL = upstream.URL + "/token"
	client.APIURL = upstream.URL + "/v4"
	st := store.New(d)
	matcher := igdb.NewMatcher(st, client, nil)
	srv := httptest.NewServer((&api.Server{Store: st, AdminToken: token, IGDB: client, IGDBMatch: matcher}).Handler())
	t.Cleanup(srv.Close)
	return srv, st, matcher, fake
}

func TestQuickAddFromIGDB(t *testing.T) {
	srv, _, _, _ := newIGDBServer(t)

	status, body := call(t, srv, "GET", "/admin/igdb/search?q=minecraft", token, "")
	var results []api.IGDBResult
	json.Unmarshal([]byte(body), &results)
	if status != http.StatusOK || len(results) != 1 || results[0].Name != "Minecraft" || results[0].Year != "2011" || results[0].ItemID != 0 {
		t.Fatalf("search: %d %s", status, body)
	}

	status, body = call(t, srv, "POST", "/admin/items/from-igdb", token, `{"igdbId":121}`)
	if status != http.StatusCreated {
		t.Fatalf("add: %d %s", status, body)
	}
	var added store.ItemView
	json.Unmarshal([]byte(body), &added)
	values := added.Verdicts.Values
	if values["title"] != "Minecraft" || values["released"] != "2011-11-17" ||
		values["poster"] != "https://images.igdb.com/igdb/image/upload/t_cover_big/co49x5.jpg" ||
		values["cover"] != "https://images.igdb.com/igdb/image/upload/t_screenshot_big/ar1.jpg" {
		t.Errorf("verdicts: %v", values)
	}
	if devs, _ := values["developers"].([]any); len(devs) != 1 || devs[0] != "Mojang" {
		t.Errorf("developers: %v", values["developers"])
	}
	if added.Entry == nil || added.Entry.Status != kind.Untriaged {
		t.Errorf("entry: %+v", added.Entry)
	}

	if status, _ := call(t, srv, "POST", "/admin/items/from-igdb", token, `{"igdbId":121}`); status != http.StatusOK {
		t.Errorf("adding twice: %d", status)
	}
	_, body = call(t, srv, "GET", "/admin/igdb/search?q=minecraft", token, "")
	json.Unmarshal([]byte(body), &results)
	if results[0].ItemID != added.ID {
		t.Errorf("search does not know it is in the log: %s", body)
	}
}

func TestSteamGamesMatchByAppID(t *testing.T) {
	srv, st, matcher, _ := newIGDBServer(t)
	ctx := context.Background()
	hadesID, _ := st.CreateItemFromSource(ctx, "game", "Hades", kind.SourceSteam, "1145360", json.RawMessage(`{"appid":1145360,"name":"Hades™"}`))
	obscureID, _ := st.CreateItemFromSource(ctx, "game", "Obscure", kind.SourceSteam, "999", json.RawMessage(`{"appid":999,"name":"Obscure"}`))

	report := matcher.Run(ctx)
	if report.State != igdb.Matched || report.Checked != 2 || report.Matched != 1 {
		t.Fatalf("report: %+v", report)
	}
	v, _ := st.GetItem(ctx, hadesID)
	if v.Verdicts.Values["title"] != "Hades" || v.Verdicts.SourceOf["title"] != kind.SourceIGDB {
		t.Errorf("hades verdicts: %v", v.Verdicts)
	}

	_, body := call(t, srv, "GET", "/admin/items?kind=game&needs=match", token, "")
	if !strings.Contains(body, `"slug":"obscure"`) || strings.Contains(body, `"slug":"hades"`) {
		t.Errorf("unmatched: %s", body)
	}
	if status, body := call(t, srv, "PUT", "/admin/items/"+itoa(obscureID)+"/igdb/none", token, ""); status != http.StatusOK {
		t.Fatalf("no match: %d %s", status, body)
	}
	if _, body := call(t, srv, "GET", "/admin/items?kind=game&needs=match", token, ""); body != "[]\n" {
		t.Errorf("still unmatched after no match: %s", body)
	}
	if report := matcher.Run(ctx); report.Checked != 0 {
		t.Errorf("matcher asked again: %+v", report)
	}
	if status, _ := call(t, srv, "PUT", "/admin/items/"+itoa(obscureID)+"/igdb", token, `{"igdbId":113112}`); status != http.StatusConflict {
		t.Errorf("matching a second item to hades: %d", status)
	}
	if status, _ := call(t, srv, "PUT", "/admin/items/424242/igdb/none", token, ""); status != http.StatusNotFound {
		t.Errorf("missing item: %d", status)
	}
}

func TestIGDBUnconfigured(t *testing.T) {
	srv := newServer(t)
	if status, _ := call(t, srv, "GET", "/admin/igdb/search?q=x", token, ""); status != http.StatusNotImplemented {
		t.Errorf("search: %d", status)
	}
	if status, _ := call(t, srv, "POST", "/admin/sync/igdb", token, ""); status != http.StatusNotImplemented {
		t.Errorf("match: %d", status)
	}
}
