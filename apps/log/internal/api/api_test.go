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
		`{"status":"finished","visibility":"public","blurb":"go play it","scores":{"enjoyment":98},"addTags":["favorite/all-time"]}`)
	call(t, srv, "POST", "/admin/items/"+id+"/moments", token, `{"at":"2024-01-02","note":"public moment"}`)
	call(t, srv, "POST", "/admin/items/"+id+"/moments", token, `{"at":"2024-01-03","note":"private moment","visibility":"private"}`)

	status, body = call(t, srv, "GET", "/public/items/game/outer-wilds", "", "")
	if status != http.StatusOK {
		t.Fatalf("public: %d %s", status, body)
	}
	for _, leaked := range []string{"private moment", "claims"} {
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

func TestDismissedStaysOffPublicPages(t *testing.T) {
	srv := newServer(t)
	_, body := call(t, srv, "POST", "/admin/items", token, `{"kind":"game","title":"Wallpaper Engine"}`)
	var created struct{ ID int64 }
	json.Unmarshal([]byte(body), &created)

	if status, body := call(t, srv, "PATCH", "/admin/items/"+itoa(created.ID)+"/entry", token, `{"status":"dismissed","visibility":"public"}`); status != http.StatusOK {
		t.Fatalf("dismiss: %d %s", status, body)
	}
	if status, _ := call(t, srv, "GET", "/public/items/game/wallpaper-engine", "", ""); status != http.StatusNotFound {
		t.Errorf("dismissed item visible: %d", status)
	}
	if _, body := call(t, srv, "GET", "/public/items", "", ""); strings.Contains(body, "wallpaper-engine") {
		t.Errorf("dismissed item listed: %s", body)
	}
}

func createGame(t *testing.T, srv *httptest.Server, title, entry string) string {
	t.Helper()
	status, body := call(t, srv, "POST", "/admin/items", token, `{"kind":"game","title":"`+title+`"}`)
	if status != http.StatusCreated {
		t.Fatalf("create %s: %d %s", title, status, body)
	}
	var created struct{ ID int64 }
	json.Unmarshal([]byte(body), &created)
	id := itoa(created.ID)
	if status, body := call(t, srv, "PATCH", "/admin/items/"+id+"/entry", token, entry); status != http.StatusOK {
		t.Fatalf("entry %s: %d %s", title, status, body)
	}
	return id
}

func TestRankings(t *testing.T) {
	srv := newServer(t)
	hades := createGame(t, srv, "Hades", `{"status":"finished","visibility":"public"}`)
	wilds := createGame(t, srv, "Outer Wilds", `{"status":"finished","visibility":"private","blurb":"secret blurb"}`)
	wallpaper := createGame(t, srv, "Wallpaper Engine", `{"status":"dismissed"}`)

	status, body := call(t, srv, "POST", "/admin/rankings", token, `{"name":"Games I want to single out"}`)
	if status != http.StatusCreated || !strings.Contains(body, `"slug":"games-i-want-to-single-out"`) {
		t.Fatalf("create ranking: %d %s", status, body)
	}
	var ranking struct{ ID int64 }
	json.Unmarshal([]byte(body), &ranking)
	orderPath := "/admin/rankings/" + itoa(ranking.ID) + "/order"

	rejected := map[string]string{
		"dismissed item":      `{"layout":"numbered","placements":[{"itemId":` + wallpaper + `}]}`,
		"unknown item":        `{"layout":"numbered","placements":[{"itemId":9999}]}`,
		"placed twice":        `{"layout":"numbered","placements":[{"itemId":` + hades + `},{"itemId":` + hades + `}]}`,
		"numbered with tiers": `{"layout":"numbered","tiers":[{"name":"S","placements":[]}]}`,
		"tiered loose items":  `{"layout":"tiered","placements":[{"itemId":` + hades + `}]}`,
		"unnamed tier":        `{"layout":"tiered","tiers":[{"name":" ","placements":[]}]}`,
		"tier named twice":    `{"layout":"tiered","tiers":[{"name":"S","placements":[]},{"name":"S","placements":[]}]}`,
		"unknown layout":      `{"layout":"spiral"}`,
	}
	for name, order := range rejected {
		if status, body := call(t, srv, "PUT", orderPath, token, order); status != http.StatusUnprocessableEntity {
			t.Errorf("%s accepted: %d %s", name, status, body)
		}
	}

	tiered := `{"layout":"tiered","tiers":[{"name":"S","placements":[{"itemId":` + wilds + `,"note":"cried"},{"itemId":` + hades + `}]},{"name":"A","placements":[]}]}`
	if status, body := call(t, srv, "PUT", orderPath, token, tiered); status != http.StatusOK || !strings.Contains(body, `"title":"Outer Wilds"`) {
		t.Fatalf("tiered order: %d %s", status, body)
	}

	status, body = call(t, srv, "GET", "/public/rankings/games-i-want-to-single-out", "", "")
	if status != http.StatusOK {
		t.Fatalf("public ranking: %d %s", status, body)
	}
	var public struct {
		Order struct {
			Layout string
			Tiers  []struct {
				Name       string
				Placements []struct {
					Title string
					Note  string
					Page  *struct{ Kind, Slug string }
				}
			}
		}
	}
	json.Unmarshal([]byte(body), &public)
	if public.Order.Layout != "tiered" || len(public.Order.Tiers) != 2 || public.Order.Tiers[1].Name != "A" || len(public.Order.Tiers[1].Placements) != 0 {
		t.Fatalf("tiers: %s", body)
	}
	top := public.Order.Tiers[0].Placements
	if len(top) != 2 || top[0].Title != "Outer Wilds" || top[0].Note != "cried" || top[1].Title != "Hades" {
		t.Fatalf("S tier: %s", body)
	}
	if top[0].Page != nil {
		t.Errorf("private entry linked: %s", body)
	}
	if top[1].Page == nil || top[1].Page.Slug != "hades" {
		t.Errorf("public entry not linked: %s", body)
	}
	if strings.Contains(body, "secret blurb") {
		t.Errorf("ranking leaks a private blurb: %s", body)
	}

	numbered := `{"layout":"numbered","placements":[{"itemId":` + hades + `},{"itemId":` + wilds + `}]}`
	if status, body := call(t, srv, "PUT", orderPath, token, numbered); status != http.StatusOK || strings.Contains(body, `"tiers"`) {
		t.Fatalf("numbered order: %d %s", status, body)
	}

	call(t, srv, "DELETE", "/admin/items/"+hades, token, "")
	_, body = call(t, srv, "GET", "/public/rankings/games-i-want-to-single-out", "", "")
	if strings.Contains(body, "Hades") || !strings.Contains(body, "Outer Wilds") {
		t.Errorf("deleted item still ranked: %s", body)
	}

	if _, body := call(t, srv, "GET", "/public/rankings", "", ""); !strings.Contains(body, "games-i-want-to-single-out") {
		t.Errorf("ranking not listed: %s", body)
	}
	if status, _ := call(t, srv, "DELETE", "/admin/rankings/"+itoa(ranking.ID), token, ""); status != http.StatusNoContent {
		t.Errorf("delete ranking: %d", status)
	}
	if status, _ := call(t, srv, "GET", "/public/rankings/games-i-want-to-single-out", "", ""); status != http.StatusNotFound {
		t.Errorf("deleted ranking still public: %d", status)
	}
}

func TestPlayedGamesArePublicByDefault(t *testing.T) {
	srv := newServer(t)
	createGame(t, srv, "Hades", `{"status":"finished"}`)
	createGame(t, srv, "Forgotten", `{"status":"maybe-played"}`)
	createGame(t, srv, "Backlog", `{"status":"never-played"}`)
	createGame(t, srv, "Hidden", `{"status":"played","visibility":"private"}`)
	call(t, srv, "POST", "/admin/items", token, `{"kind":"game","title":"Unsorted"}`)

	_, body := call(t, srv, "GET", "/public/items?kind=game", "", "")
	for _, slug := range []string{"hades", "forgotten"} {
		if !strings.Contains(body, `"slug":"`+slug+`"`) {
			t.Errorf("%s not listed: %s", slug, body)
		}
	}
	for _, slug := range []string{"backlog", "hidden", "unsorted"} {
		if strings.Contains(body, `"slug":"`+slug+`"`) {
			t.Errorf("%s listed: %s", slug, body)
		}
	}
	if status, _ := call(t, srv, "GET", "/public/items/game/backlog", "", ""); status != http.StatusNotFound {
		t.Errorf("never-played page: %d", status)
	}
	if status, body := call(t, srv, "GET", "/public/kinds/game", "", ""); status != http.StatusOK || body != `{"name":"game","untriaged":1}`+"\n" {
		t.Errorf("kind: %d %q", status, body)
	}
	if status, _ := call(t, srv, "GET", "/public/kinds/nope", "", ""); status != http.StatusNotFound {
		t.Errorf("unknown kind: %d", status)
	}
}

func TestNeedsScoreListsPlayedGamesWithoutScores(t *testing.T) {
	srv := newServer(t)
	createGame(t, srv, "Rated", `{"status":"finished","scores":{"enjoyment":80}}`)
	createGame(t, srv, "Unrated", `{"status":"played"}`)
	createGame(t, srv, "Forgotten", `{"status":"maybe-played"}`)
	_, body := call(t, srv, "GET", "/admin/items?needs=score", token, "")
	var views []struct{ Slug string }
	json.Unmarshal([]byte(body), &views)
	if len(views) != 1 || views[0].Slug != "unrated" {
		t.Errorf("needs score: %s", body)
	}
}

func TestTimeSpentIsApsWord(t *testing.T) {
	srv := newServer(t)
	id := createGame(t, srv, "Hades", `{"status":"finished","timeSpent":"too-much"}`)
	if _, body := call(t, srv, "GET", "/public/items/game/hades", "", ""); !strings.Contains(body, `"timeSpent":"too-much"`) {
		t.Errorf("public: %s", body)
	}
	if status, _ := call(t, srv, "PATCH", "/admin/items/"+id+"/entry", token, `{"timeSpent":"forever"}`); status != http.StatusUnprocessableEntity {
		t.Errorf("unknown choice: %d", status)
	}
	call(t, srv, "PATCH", "/admin/items/"+id+"/entry", token, `{"blurb":"x"}`)
	if _, body := call(t, srv, "GET", "/public/items/game/hades", "", ""); !strings.Contains(body, `"timeSpent":"too-much"`) {
		t.Errorf("kept when not mentioned: %s", body)
	}
	call(t, srv, "PATCH", "/admin/items/"+id+"/entry", token, `{"timeSpent":""}`)
	if _, body := call(t, srv, "GET", "/public/items/game/hades", "", ""); strings.Contains(body, "timeSpent") {
		t.Errorf("unsaid: %s", body)
	}
}

func TestSettingsDecideWhatShows(t *testing.T) {
	srv := newServer(t)
	createGame(t, srv, "Forgotten", `{"status":"maybe-played"}`)
	createGame(t, srv, "Backlog", `{"status":"never-played"}`)
	createGame(t, srv, "Wanted", `{"status":"never-played","visibility":"public"}`)

	settings := `{"statusesShown":["played","never-played"],"axisLabels":{"enjoyment":"liked"},"scoreDisplay":"both","scoreWords":[{"from":80,"word":"loved"},{"from":0,"word":"meh"}]}`
	if status, body := call(t, srv, "PUT", "/admin/settings/kinds/game", token, settings); status != http.StatusOK || !strings.Contains(body, `"scoreWords":[{"from":0,"word":"meh"},{"from":80,"word":"loved"}]`) {
		t.Fatalf("put settings: %d %s", status, body)
	}
	_, body := call(t, srv, "GET", "/public/items?kind=game", "", "")
	if strings.Contains(body, `"slug":"forgotten"`) || !strings.Contains(body, `"slug":"backlog"`) || !strings.Contains(body, `"slug":"wanted"`) {
		t.Errorf("shown after settings: %s", body)
	}
	if status, _ := call(t, srv, "PUT", "/admin/settings/kinds/game", token, `{"statusesShown":["dismissed"],"scoreDisplay":"number"}`); status != http.StatusUnprocessableEntity {
		t.Errorf("dismissed shown: %d", status)
	}
	call(t, srv, "PUT", "/admin/settings/tag-titles", token, `{"Another Shot":"worth another shot"}`)
	if _, body := call(t, srv, "GET", "/public/presentation", "", ""); !strings.Contains(body, `"tagTitles":{"another-shot":"worth another shot"}`) || !strings.Contains(body, `"axisLabels":{"enjoyment":"liked","quality":"quality"}`) {
		t.Errorf("presentation: %s", body)
	}
}

func TestEntriesListOnlyTriagedAndKeepsItSmall(t *testing.T) {
	srv := newServer(t)
	createGame(t, srv, "Hades", `{"status":"finished","blurb":"long public text","scores":{"enjoyment":92},"addTags":["comfort"]}`)
	createGame(t, srv, "Wallpaper Engine", `{"status":"dismissed"}`)
	call(t, srv, "POST", "/admin/items", token, `{"kind":"game","title":"Untouched"}`)

	status, body := call(t, srv, "GET", "/admin/entries?kind=game", token, "")
	if status != http.StatusOK {
		t.Fatalf("entries: %d %s", status, body)
	}
	var entries []struct {
		Title  string
		Scores map[string]int
		Tags   []string
	}
	if err := json.Unmarshal([]byte(body), &entries); err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Title != "Hades" || entries[0].Scores["enjoyment"] != 92 || entries[0].Tags[0] != "comfort" {
		t.Errorf("entries: %s", body)
	}
	if strings.Contains(body, "long public text") || strings.Contains(body, "verdicts") {
		t.Errorf("entries carry more than the list shows: %s", body)
	}
}
