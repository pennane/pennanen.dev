package api_test

import (
	"bytes"
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
	"github.com/pennane/pennanen.dev/apps/log/internal/images"
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
	pictures := &images.Store{Dir: t.TempDir()}
	srv := httptest.NewServer((&api.Server{Store: store.New(d), AdminToken: token, Images: pictures}).Handler())
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
		`{"status":"finished","visibility":"public","blurb":"go play it","prose":"**2024-01-02** the ending","scores":{"prestige":"s"},"addTags":["favorite/all-time"]}`)

	status, body = call(t, srv, "GET", "/public/items/game/outer-wilds", "", "")
	if status != http.StatusOK {
		t.Fatalf("public: %d %s", status, body)
	}
	for _, leaked := range []string{"claims"} {
		if strings.Contains(body, leaked) {
			t.Errorf("public item leaks %q: %s", leaked, body)
		}
	}
	for _, shown := range []string{"go play it", "the ending", `"prestige":"s"`, "favorite/all-time"} {
		if !strings.Contains(body, shown) {
			t.Errorf("public item misses %q: %s", shown, body)
		}
	}
	if status, body := call(t, srv, "PATCH", "/admin/items/"+id+"/entry", token, `{"status":"seen"}`); status != http.StatusUnprocessableEntity {
		t.Errorf("game accepted thing status: %d %s", status, body)
	}
}

func TestPublicListCountsProseWordsWithoutTheProse(t *testing.T) {
	srv := newServer(t)
	createGame(t, srv, "Outer Wilds", `{"status":"finished","prose":"the **ending** made me  cry\n\nagain"}`)
	createGame(t, srv, "Hades", `{"status":"finished"}`)

	_, body := call(t, srv, "GET", "/public/items?kind=game", "", "")
	var items []struct {
		Slug       string
		Prose      string
		ProseWords int
	}
	if err := json.Unmarshal([]byte(body), &items); err != nil {
		t.Fatalf("list: %v %s", err, body)
	}
	words := map[string]int{}
	for _, item := range items {
		if item.Prose != "" {
			t.Errorf("list carries prose for %s", item.Slug)
		}
		words[item.Slug] = item.ProseWords
	}
	if words["outer-wilds"] != 6 || words["hades"] != 0 {
		t.Errorf("prose words = %v", words)
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
	createGame(t, srv, "Rated", `{"status":"finished","scores":{"prestige":"a"}}`)
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

	settings := `{"statusesShown":["played","never-played"],"axes":["prestige"]}`
	if status, body := call(t, srv, "PUT", "/admin/settings/kinds/game", token, settings); status != http.StatusOK || !strings.Contains(body, `"axes":["prestige"]`) {
		t.Fatalf("put settings: %d %s", status, body)
	}
	_, body := call(t, srv, "GET", "/public/items?kind=game", "", "")
	if strings.Contains(body, `"slug":"forgotten"`) || !strings.Contains(body, `"slug":"backlog"`) || !strings.Contains(body, `"slug":"wanted"`) {
		t.Errorf("shown after settings: %s", body)
	}
	if status, _ := call(t, srv, "PUT", "/admin/settings/kinds/game", token, `{"statusesShown":["dismissed"],"axes":["prestige"]}`); status != http.StatusUnprocessableEntity {
		t.Errorf("dismissed shown: %d", status)
	}
	if status, _ := call(t, srv, "PUT", "/admin/settings/kinds/game", token, `{"statusesShown":["played"],"axes":["quality"]}`); status != http.StatusUnprocessableEntity {
		t.Errorf("game dropped prestige: %d", status)
	}
	axes := `[{"key":"prestige","name":"nautinto","divisions":[{"name":"All Time"},{"key":"b","name":"b"}]},{"name":"Replay Value","divisions":[{"name":"yes"}]}]`
	if status, body := call(t, srv, "PUT", "/admin/settings/axes", token, axes); status != http.StatusOK || !strings.Contains(body, `{"key":"replay-value","name":"Replay Value","divisions":[{"key":"yes","name":"yes"}]}`) {
		t.Fatalf("put axes: %d %s", status, body)
	}
	if status, _ := call(t, srv, "PUT", "/admin/settings/axes", token, `[{"key":"quality","name":"quality","divisions":[{"name":"s"}]}]`); status != http.StatusUnprocessableEntity {
		t.Errorf("removed prestige: %d", status)
	}
	call(t, srv, "PUT", "/admin/settings/tag-titles", token, `{"Another Shot":"worth another shot"}`)
	if _, body := call(t, srv, "GET", "/public/presentation", "", ""); !strings.Contains(body, `"tagTitles":{"another-shot":"worth another shot"}`) || !strings.Contains(body, `"name":"nautinto"`) {
		t.Errorf("presentation: %s", body)
	}
}

func TestEntriesListOnlyTriagedAndKeepsItSmall(t *testing.T) {
	srv := newServer(t)
	createGame(t, srv, "Hades", `{"status":"finished","blurb":"long public text","scores":{"prestige":"s"},"addTags":["comfort"]}`)
	createGame(t, srv, "Wallpaper Engine", `{"status":"dismissed"}`)
	call(t, srv, "POST", "/admin/items", token, `{"kind":"game","title":"Untouched"}`)

	status, body := call(t, srv, "GET", "/admin/entries?kind=game", token, "")
	if status != http.StatusOK {
		t.Fatalf("entries: %d %s", status, body)
	}
	var entries []struct {
		Title  string
		Scores map[string]string
		Rank   int
		Tags   []string
	}
	if err := json.Unmarshal([]byte(body), &entries); err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Title != "Hades" || entries[0].Scores["prestige"] != "s" || entries[0].Rank != 1 || entries[0].Tags[0] != "comfort" {
		t.Errorf("entries: %s", body)
	}
	if strings.Contains(body, "long public text") || strings.Contains(body, "verdicts") {
		t.Errorf("entries carry more than the list shows: %s", body)
	}
}

func TestPublicRankCountsOnlyPublicEntries(t *testing.T) {
	srv := newServer(t)
	createGame(t, srv, "Secret Favourite", `{"status":"finished","visibility":"private","scores":{"prestige":"s"}}`)
	createGame(t, srv, "Hades", `{"status":"finished","scores":{"prestige":"a"}}`)
	celeste := createGame(t, srv, "Celeste", `{"status":"finished","scores":{"prestige":"a"}}`)

	_, body := call(t, srv, "GET", "/public/items/game/celeste", "", "")
	if !strings.Contains(body, `"rank":{"place":2,"of":2}`) {
		t.Errorf("celeste: %s", body)
	}
	if status, body := call(t, srv, "PUT", "/admin/items/"+celeste+"/place", token, `{"division":"a","index":0}`); status != http.StatusNoContent {
		t.Fatalf("place: %d %s", status, body)
	}
	_, body = call(t, srv, "GET", "/public/items?kind=game", "", "")
	if !strings.Contains(body, `"slug":"celeste"`) || strings.Contains(body, "secret-favourite") {
		t.Fatalf("public items: %s", body)
	}
	_, body = call(t, srv, "GET", "/public/items/game/celeste", "", "")
	if !strings.Contains(body, `"rank":{"place":1,"of":2}`) {
		t.Errorf("celeste after move: %s", body)
	}
}

func uploadImage(t *testing.T, srv *httptest.Server, ticket string, body []byte) (int, string) {
	t.Helper()
	req, _ := http.NewRequest("POST", srv.URL+"/images", bytes.NewReader(body))
	req.Header.Set("Authorization", "Ticket "+ticket)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	return res.StatusCode, string(b)
}

func TestImagesUploadWithATicketAndServePublicly(t *testing.T) {
	srv := newServer(t)
	gif := []byte("GIF89a\x01\x00\x01\x00\x00\x00\x00;")

	if status, _ := uploadImage(t, srv, "123.forged", gif); status != http.StatusUnauthorized {
		t.Errorf("forged ticket: %d", status)
	}
	status, body := call(t, srv, "POST", "/admin/images/ticket", token, "")
	if status != http.StatusOK {
		t.Fatalf("ticket: %d %s", status, body)
	}
	var ticket api.UploadTicket
	json.Unmarshal([]byte(body), &ticket)

	if status, _ := uploadImage(t, srv, ticket.Ticket, []byte("not an image")); status != http.StatusUnprocessableEntity {
		t.Errorf("text upload: %d", status)
	}
	status, body = uploadImage(t, srv, ticket.Ticket, gif)
	if status != http.StatusCreated {
		t.Fatalf("upload: %d %s", status, body)
	}
	var uploaded api.UploadedImage
	json.Unmarshal([]byte(body), &uploaded)
	status, body = call(t, srv, "GET", "/images/"+uploaded.Name, "", "")
	if status != http.StatusOK || body != string(gif) {
		t.Errorf("serve %s: %d", uploaded.Name, status)
	}
}
