package api_test

import (
	"bytes"
	"cmp"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/api"
	"github.com/pennane/pennanen.dev/apps/log/internal/platform/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/platform/images"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/take"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

const token = "secret"

func openStore(t *testing.T, d *sql.DB) *store.Store {
	t.Helper()
	if err := take.WriteKindStatuses(context.Background(), d, kind.All()); err != nil {
		t.Fatal(err)
	}
	return store.Open(d)
}

func newServerWithDB(t *testing.T) (*httptest.Server, *sql.DB) {
	t.Helper()
	d, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	pictures := &images.Store{Dir: t.TempDir()}
	srv := httptest.NewServer((&api.Server{Store: openStore(t, d), AdminToken: token, Images: pictures}).Handler())
	t.Cleanup(srv.Close)
	return srv, d
}

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv, _ := newServerWithDB(t)
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
	if status, body := call(t, srv, "PATCH", "/admin/items/"+id+"/entry", token, `{"status":"watched"}`); status != http.StatusUnprocessableEntity {
		t.Errorf("game accepted a movie status: %d %s", status, body)
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

func TestPublicItemsNeverIncludeDismissed(t *testing.T) {
	srv := newServer(t)
	id := createGame(t, srv, "Wallpaper Engine", `{"status":"finished","visibility":"public","scores":{"prestige":"s"}}`)
	if _, body := call(t, srv, "GET", "/public/items?kind=game", "", ""); !strings.Contains(body, "wallpaper-engine") {
		t.Fatalf("public before dismissing: %s", body)
	}
	if status, body := call(t, srv, "PATCH", "/admin/items/"+id+"/entry", token, `{"status":"dismissed"}`); status != http.StatusUnprocessableEntity {
		t.Errorf("dismissed as a status: %d %s", status, body)
	}
	status, body := call(t, srv, "PATCH", "/admin/items/"+id+"/entry", token, `{"dismissal":"dismissed"}`)
	if status != http.StatusOK || !strings.Contains(body, `"dismissedAt":"`) {
		t.Fatalf("dismiss: %d %s", status, body)
	}
	if status, _ := call(t, srv, "GET", "/public/items/game/wallpaper-engine", "", ""); status != http.StatusNotFound {
		t.Errorf("dismissed item visible: %d", status)
	}
	if _, body := call(t, srv, "GET", "/public/items", "", ""); strings.Contains(body, "wallpaper-engine") {
		t.Errorf("dismissed item listed: %s", body)
	}
	if _, body := call(t, srv, "GET", "/public/presentation", "", ""); strings.Contains(body, `"game":{`) {
		t.Errorf("a kind with only a dismissed item is presented: %s", body)
	}
	if _, body := call(t, srv, "GET", "/admin/items?kind=game", token, ""); strings.Contains(body, "wallpaper-engine") {
		t.Errorf("dismissed item in the admin list: %s", body)
	}
}

func TestAnEntryTellsWhoDecidedIt(t *testing.T) {
	srv := newServer(t)
	id := createGame(t, srv, "Hades", `{"status":"finished","addTags":["comfort"]}`)
	_, body := call(t, srv, "GET", "/admin/items/"+id, token, "")
	for _, shown := range []string{`"status":"finished"`, `"statusDecidedBy":"ap"`, `"statusChoice":{"status":"finished","chosenAt":"`, `"standing":"decides"`, `"statusProposals":[]`, `"tags":["comfort"]`, `"tagsFromRules":{}`, `"loggedBy":"ap"`, `"suggestions":[]`} {
		if !strings.Contains(body, shown) {
			t.Errorf("item misses %s: %s", shown, body)
		}
	}
	for _, gone := range []string{"statusSetBy", "tagsAddedBySources", `"hint"`, "dismissedAt"} {
		if strings.Contains(body, gone) {
			t.Errorf("item still says %s: %s", gone, body)
		}
	}
}

func TestItemsFilterBySuggestingRule(t *testing.T) {
	srv, d := newServerWithDB(t)
	hinted := createGame(t, srv, "Card Farm", `{"blurb":""}`)
	createGame(t, srv, "Hades", `{"blurb":""}`)
	if _, err := d.Exec(`INSERT INTO proposal (item_id, source, rule, subject, value, strength) VALUES (?, 'steam', 'likely-idled', 'tag', 'idled', 'suggesting')`, hinted); err != nil {
		t.Fatal(err)
	}
	_, body := call(t, srv, "GET", "/admin/items?suggested=likely-idled", token, "")
	var items []struct {
		Slug        string
		Suggestions []struct{ Rule, Subject, Value string }
	}
	if err := json.Unmarshal([]byte(body), &items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Slug != "card-farm" || len(items[0].Suggestions) != 1 || items[0].Suggestions[0].Rule != "likely-idled" {
		t.Errorf("suggested items: %s", body)
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
	createGame(t, srv, "Wallpaper Engine", `{"status":"played","dismissal":"dismissed"}`)
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

func TestPublicFieldsComeFromTheKindAndPrivateOnesNeverShow(t *testing.T) {
	srv := newServer(t)
	id := createGame(t, srv, "Hades", `{"status":"finished"}`)
	if status, body := call(t, srv, "PATCH", "/admin/items/"+id+"/fields", token, `{"released":"2020","playtime_minutes":600}`); status != http.StatusOK {
		t.Fatalf("fields: %d %s", status, body)
	}
	if _, body := call(t, srv, "GET", "/public/items/game/hades", "", ""); !strings.Contains(body, `"released":"2020"`) || strings.Contains(body, "playtime") {
		t.Errorf("public item: %s", body)
	}

	_, body := call(t, srv, "GET", "/public/presentation", "", "")
	var presentation struct {
		Kinds  map[string]store.KindSettings
		Fields map[string][]struct{ Key string }
	}
	if err := json.Unmarshal([]byte(body), &presentation); err != nil {
		t.Fatal(err)
	}
	keys := map[string][]string{}
	for kind, fields := range presentation.Fields {
		for _, f := range fields {
			keys[kind] = append(keys[kind], f.Key)
		}
	}
	if strings.Contains(strings.Join(keys["game"], " "), "playtime") || !strings.Contains(strings.Join(keys["movie"], " "), "letterboxd-take.rating") || strings.Contains(strings.Join(keys["movie"], " "), "on_watchlist_since") {
		t.Errorf("public fields: %v", keys)
	}
	game := presentation.Kinds["game"]
	if strings.Join(toStrings(game.FieldUses["genres"]), ",") != "filter,page" || game.Overview.Count != 5 || game.DefaultView != store.AutoView {
		t.Errorf("game defaults: %+v", game)
	}

	movie := `{"title":"movies","statusesShown":["watched"],"axes":["prestige"],"fieldUses":{"genres":[]},"overview":{"sort":"letterboxd-take.rating","count":8}}`
	if status, body := call(t, srv, "PUT", "/admin/settings/kinds/movie", token, movie); status != http.StatusOK || !strings.Contains(body, `"genres":[]`) || !strings.Contains(body, `"released":["filter","column","page"]`) {
		t.Errorf("movie settings: %d %s", status, body)
	}
	for _, refused := range []string{
		`{"title":"games","statusesShown":["played"],"axes":["prestige"],"overview":{"sort":"genres"}}`,
		`{"title":"games","statusesShown":["played"],"axes":["prestige"],"fieldUses":{"playtime_minutes":["column"]}}`,
		`{"title":"games","statusesShown":["played"],"axes":["prestige"],"defaultView":"grid"}`,
	} {
		if status, _ := call(t, srv, "PUT", "/admin/settings/kinds/game", token, refused); status != http.StatusUnprocessableEntity {
			t.Errorf("accepted %s: %d", refused, status)
		}
	}
}

func toStrings[T ~string](xs []T) []string {
	out := make([]string, len(xs))
	for i, x := range xs {
		out[i] = string(x)
	}
	return out
}

func TestKnownItemsStayOutOfEntriesAndPublic(t *testing.T) {
	ctx := context.Background()
	d, err := db.Open(ctx, filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	st := openStore(t, d)
	srv := httptest.NewServer((&api.Server{Store: st, AdminToken: token, Images: &images.Store{Dir: t.TempDir()}}).Handler())
	t.Cleanup(srv.Close)

	if _, err := st.KnowItemFromSource(ctx, kind.Movie, "Heat", record.SourceTMDB, "949", json.RawMessage(`{"id":949,"title":"Heat"}`)); err != nil {
		t.Fatal(err)
	}
	logged, err := st.CreateItem(ctx, kind.Movie, "Ran", nil)
	if err != nil {
		t.Fatal(err)
	}
	status := "watched"
	if err := st.ChangeEntries(ctx, []int64{logged}, store.EntryChange{Status: &status}); err != nil {
		t.Fatal(err)
	}

	titles := func(path string) []string {
		code, body := call(t, srv, "GET", path, token, "")
		if code != http.StatusOK {
			t.Fatalf("%s: %d %s", path, code, body)
		}
		var listed []struct{ Title, Slug string }
		if err := json.Unmarshal([]byte(body), &listed); err != nil {
			t.Fatal(err)
		}
		var names []string
		for _, l := range listed {
			names = append(names, cmp.Or(l.Title, l.Slug))
		}
		return names
	}
	if got := titles("/admin/entries?kind=movie"); len(got) != 1 || got[0] != "Ran" {
		t.Errorf("entries = %v", got)
	}
	code, body := call(t, srv, "GET", "/admin/items?kind=movie", token, "")
	if code != http.StatusOK {
		t.Fatalf("items: %d %s", code, body)
	}
	var items []struct {
		Slug  string
		Entry *json.RawMessage
	}
	if err := json.Unmarshal([]byte(body), &items); err != nil {
		t.Fatal(err)
	}
	entryOf := map[string]bool{}
	for _, i := range items {
		entryOf[i.Slug] = i.Entry != nil
	}
	if len(items) != 1 || !entryOf["ran"] {
		t.Errorf("admin items = %s", body)
	}
	_, body = call(t, srv, "GET", "/public/items?kind=movie", "", "")
	if strings.Contains(body, `"slug":"heat"`) || !strings.Contains(body, `"slug":"ran"`) {
		t.Errorf("public items = %s", body)
	}
}

func TestPublicPresentationListsOnlyKindsWithSomethingPublic(t *testing.T) {
	srv := newServer(t)
	status, body := call(t, srv, "POST", "/admin/items", token, `{"kind":"movie","title":"Ran"}`)
	if status != http.StatusCreated {
		t.Fatalf("create: %d %s", status, body)
	}
	var created struct{ ID int64 }
	json.Unmarshal([]byte(body), &created)
	call(t, srv, "PATCH", "/admin/items/"+itoa(created.ID)+"/entry", token, `{"status":"watched"}`)

	kindsOf := func(path, auth string) map[string]json.RawMessage {
		code, body := call(t, srv, "GET", path, auth, "")
		if code != http.StatusOK {
			t.Fatalf("%s: %d %s", path, code, body)
		}
		var parsed struct{ Kinds map[string]json.RawMessage }
		if err := json.Unmarshal([]byte(body), &parsed); err != nil {
			t.Fatal(err)
		}
		return parsed.Kinds
	}
	public := kindsOf("/public/presentation", "")
	if _, shown := public["movie"]; !shown {
		t.Errorf("public lacks movie: %v", public)
	}
	for _, empty := range []string{"artist", "album", "game"} {
		if _, shown := public[empty]; shown {
			t.Errorf("public lists %s with nothing public", empty)
		}
	}
	admin := kindsOf("/admin/settings", token)
	for _, name := range []string{"movie", "artist", "album", "game"} {
		if _, listed := admin[name]; !listed {
			t.Errorf("admin settings lacks %s", name)
		}
	}
}
