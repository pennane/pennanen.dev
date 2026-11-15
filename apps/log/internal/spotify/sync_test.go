package spotify_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pennane/pennanen.dev/apps/log/internal/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/spotify"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/syncs"
)

type fakeSpotify struct {
	topArtists  map[string]string
	saved       string
	albumsNull  bool
	savedTotal  *int
	artistsGone bool
	plays       string
}

func (f *fakeSpotify) handler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/token":
		fmt.Fprint(w, `{"access_token":"access","expires_in":3600}`)
	case "/me/top/artists":
		fmt.Fprintf(w, `{"items":%s,"next":null}`, f.topArtists[r.URL.Query().Get("time_range")])
	case "/me/top/tracks":
		items := `[]`
		if r.URL.Query().Get("time_range") == "short_term" {
			items = `[{"id":"T","name":"Track","album":{"id":"A","name":"Alpha","artists":[{"id":"X","name":"Xena"}]},"artists":[{"id":"X","name":"Xena"}]}]`
		}
		fmt.Fprintf(w, `{"items":%s,"next":null}`, items)
	case "/me/player/recently-played":
		fmt.Fprintf(w, `{"items":%s,"next":null}`, f.plays)
	case "/me/albums":
		var listed []json.RawMessage
		json.Unmarshal([]byte(f.saved), &listed)
		total := len(listed)
		if f.savedTotal != nil {
			total = *f.savedTotal
		}
		fmt.Fprintf(w, `{"items":%s,"next":null,"total":%d}`, f.saved, total)
	case "/albums":
		if f.albumsNull {
			fmt.Fprint(w, `{"albums":[null]}`)
			return
		}
		items := []string{}
		for _, id := range strings.Split(r.URL.Query().Get("ids"), ",") {
			items = append(items, fmt.Sprintf(`{"id":%q,"name":"Album %s","label":"Label %s","artists":[{"id":"X","name":"Xena"}]}`, id, id, id))
		}
		fmt.Fprintf(w, `{"albums":[%s]}`, strings.Join(items, ","))
	case "/artists":
		if f.artistsGone {
			http.NotFound(w, r)
			return
		}
		items := []string{}
		for _, id := range strings.Split(r.URL.Query().Get("ids"), ",") {
			items = append(items, fmt.Sprintf(`{"id":%q,"name":"Artist %s","genres":["g"],"popularity":5}`, id, id))
		}
		fmt.Fprintf(w, `{"artists":[%s]}`, strings.Join(items, ","))
	default:
		http.NotFound(w, r)
	}
}

const (
	fullX      = `{"id":"X","name":"Xena","genres":["a"],"popularity":9}`
	fullY      = `{"id":"Y","name":"Yara","genres":["b"],"popularity":8}`
	savedAlbum = `[{"added_at":"2026-11-10T09:00:00Z","album":{"id":"B","name":"Beta","label":"Lb","artists":[{"id":"Y","name":"Yara"}]}}]`
)

func newFake() *fakeSpotify {
	return &fakeSpotify{
		topArtists: map[string]string{
			"short_term":  "[" + fullX + "]",
			"medium_term": "[]",
			"long_term":   "[" + fullX + "," + fullY + "]",
		},
		saved: savedAlbum,
		plays: playedAt("2026-11-14T20:00:00.000Z"),
	}
}

func playedAt(times ...string) string {
	items := []string{}
	for _, at := range times {
		items = append(items, fmt.Sprintf(`{"played_at":%q,"track":{"id":"T","name":"Track","album":{"id":"A","name":"Alpha","artists":[{"id":"X","name":"Xena"}]},"artists":[{"id":"X","name":"Xena"}]}}`, at))
	}
	return "[" + strings.Join(items, ",") + "]"
}

func openStore(t *testing.T) *store.Store {
	t.Helper()
	d, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	return store.New(d)
}

type harness struct {
	store  *store.Store
	syncer *spotify.Syncer
	day    string
}

func newHarness(t *testing.T, fake *fakeSpotify, connection spotify.Connection) *harness {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(fake.handler))
	t.Cleanup(server.Close)
	s := openStore(t)
	if connection == spotify.Connected {
		if err := s.PutSecret(context.Background(), spotify.RefreshTokenSecret, "r1"); err != nil {
			t.Fatal(err)
		}
	}
	client := spotify.NewClient("id", "secret", s)
	client.AccountsURL = server.URL
	client.APIURL = server.URL
	client.HTTP = server.Client()
	h := &harness{store: s, syncer: spotify.NewSyncer(s, client), day: "2026-11-15"}
	h.syncer.Now = func() time.Time {
		day, err := time.Parse(time.DateOnly, h.day)
		if err != nil {
			t.Fatal(err)
		}
		return day.Add(3 * time.Hour)
	}
	return h
}

func (h *harness) run(t *testing.T) syncs.Report {
	t.Helper()
	sync, err := syncs.New(context.Background(), spotify.SyncName, h.syncer.Work, h.store, nil)
	if err != nil {
		t.Fatal(err)
	}
	return sync.Run(context.Background())
}

func (h *harness) runPlays(t *testing.T) syncs.Report {
	t.Helper()
	sync, err := syncs.New(context.Background(), spotify.PlaysSyncName, h.syncer.WorkPlays, h.store, nil)
	if err != nil {
		t.Fatal(err)
	}
	return sync.Run(context.Background())
}

func (h *harness) item(t *testing.T, source, externalID string) int64 {
	t.Helper()
	id, err := h.store.FindByIdentity(context.Background(), source, externalID)
	if err != nil {
		t.Fatalf("%s %s: %v", source, externalID, err)
	}
	return id
}

func (h *harness) observations(t *testing.T, itemID int64, signal string) []store.Observation {
	t.Helper()
	o, err := h.store.Observations(context.Background(), itemID, spotify.ObservationSource, signal)
	if err != nil {
		t.Fatal(err)
	}
	return o
}

func TestASyncRecordsTodaysSnapshot(t *testing.T) {
	ctx := context.Background()
	h := newHarness(t, newFake(), spotify.Connected)
	report := h.run(t)
	if report.State != syncs.Succeeded || len(report.Problems) != 0 {
		t.Fatalf("report = %+v", report)
	}
	x := h.item(t, kind.SourceSpotifyArtist, "X")
	y := h.item(t, kind.SourceSpotifyArtist, "Y")
	a := h.item(t, kind.SourceSpotifyAlbum, "A")
	b := h.item(t, kind.SourceSpotifyAlbum, "B")

	for _, id := range []int64{a, b} {
		view, err := h.store.GetItem(ctx, id)
		if err != nil || view.Entry != nil {
			t.Fatalf("item %d entry = %v, err = %v", id, view.Entry, err)
		}
	}
	if got := len(h.observations(t, x, "top-artist")) + len(h.observations(t, y, "top-artist")); got != 3 {
		t.Fatalf("top-artist observations = %d", got)
	}
	topTracks := h.observations(t, a, "top-track")
	if len(topTracks) != 1 || topTracks[0].Key != "top-track:short_term:2026-11-15:T" {
		t.Fatalf("top-track = %+v", topTracks)
	}
	var topTrack struct {
		Range     string `json:"range"`
		Rank      int    `json:"rank"`
		Day       string `json:"day"`
		TrackID   string `json:"trackId"`
		TrackName string `json:"trackName"`
	}
	if err := json.Unmarshal(topTracks[0].Data, &topTrack); err != nil || topTrack.Rank != 1 || topTrack.Range != "short_term" || topTrack.TrackName != "Track" {
		t.Fatalf("top-track data = %s, %v", topTracks[0].Data, err)
	}
	saved := h.observations(t, b, "saved")
	if len(saved) != 1 || saved[0].ObservedAt != "2026-11-10T09:00:00Z" {
		t.Fatalf("saved = %+v", saved)
	}
	assertRelation(t, h, a, x)
	assertRelation(t, h, b, y)

	claim, err := h.store.Claim(ctx, a, kind.SourceSpotifyAlbum)
	if err != nil || !strings.Contains(string(claim.Data), `"label":"Label A"`) {
		t.Fatalf("album claim = %s, %v", claim.Data, err)
	}
	if report.Counts[spotify.CountSaves] != 1 || report.Counts[spotify.CountSnapshots] != 4 {
		t.Fatalf("counts = %v", report.Counts)
	}
}

func assertRelation(t *testing.T, h *harness, albumID, artistID int64) {
	t.Helper()
	view, err := h.store.GetItem(context.Background(), albumID)
	if err != nil {
		t.Fatal(err)
	}
	want := store.Relation{FromID: albumID, Type: kind.RelationBy, ToID: artistID}
	for _, r := range view.Relations {
		if r == want {
			return
		}
	}
	t.Fatalf("relations of %d = %v, want %v", albumID, view.Relations, want)
}

func TestRunningTwiceADayAddsNothing(t *testing.T) {
	h := newHarness(t, newFake(), spotify.Connected)
	h.run(t)
	a := h.item(t, kind.SourceSpotifyAlbum, "A")
	x := h.item(t, kind.SourceSpotifyArtist, "X")
	before := len(h.observations(t, a, "top-track")) + len(h.observations(t, x, "top-artist")) + len(h.observations(t, a, "play"))
	report := h.run(t)
	after := len(h.observations(t, a, "top-track")) + len(h.observations(t, x, "top-artist")) + len(h.observations(t, a, "play"))
	if before != after || report.Counts[spotify.CountSnapshots] != 0 || report.Counts[spotify.CountPlays] != 0 || report.Counts[spotify.CountSaves] != 0 {
		t.Fatalf("before = %d, after = %d, counts = %v", before, after, report.Counts)
	}
}

func TestAnAlbumMissingFromSavedIsUnsavedOnce(t *testing.T) {
	fake := newFake()
	h := newHarness(t, fake, spotify.Connected)
	h.run(t)
	b := h.item(t, kind.SourceSpotifyAlbum, "B")

	fake.saved = `[]`
	h.day = "2026-11-16"
	report := h.run(t)
	if got := h.observations(t, b, "unsaved"); len(got) != 1 || report.Counts[spotify.CountUnsaves] != 1 {
		t.Fatalf("unsaved = %+v, counts = %v", got, report.Counts)
	}
	h.day = "2026-11-17"
	h.run(t)
	if got := h.observations(t, b, "unsaved"); len(got) != 1 {
		t.Fatalf("unsaved after a third day = %+v", got)
	}
}

func TestNotConnectedFailsTheRun(t *testing.T) {
	h := newHarness(t, newFake(), spotify.NotConnected)
	report := h.run(t)
	if report.State != syncs.Failed || !strings.Contains(strings.Join(report.Problems, "\n"), "connect it in admin settings") {
		t.Fatalf("report = %+v", report)
	}
}

func TestAnUnreadableSavedAlbumFailsTheRunBeforeAnyUnsave(t *testing.T) {
	fake := newFake()
	h := newHarness(t, fake, spotify.Connected)
	h.run(t)
	b := h.item(t, kind.SourceSpotifyAlbum, "B")

	fake.saved = `[{"added_at":"2026-11-10T09:00:00Z","album":5}]`
	h.day = "2026-11-16"
	report := h.run(t)
	if report.State != syncs.Failed {
		t.Fatalf("report = %+v", report)
	}
	if got := h.observations(t, b, "unsaved"); len(got) != 0 {
		t.Fatalf("unsaved = %+v", got)
	}
}

func TestAnAlbumTheBatchNeverReturnedIsAProblemNamingIt(t *testing.T) {
	fake := newFake()
	fake.albumsNull = true
	h := newHarness(t, fake, spotify.Connected)
	report := h.run(t)
	if !strings.Contains(strings.Join(report.Problems, "\n"), "spotify-album never came back in full: A") {
		t.Fatalf("problems = %v", report.Problems)
	}
	b := h.item(t, kind.SourceSpotifyAlbum, "B")
	claim, err := h.store.Claim(context.Background(), b, kind.SourceSpotifyAlbum)
	if err != nil || !strings.Contains(string(claim.Data), `"label":"Lb"`) {
		t.Fatalf("claim = %s, %v", claim.Data, err)
	}
}

func TestAnAlbumSavedAgainAfterAnUnsaveIsSavedAgain(t *testing.T) {
	ctx := context.Background()
	fake := newFake()
	h := newHarness(t, fake, spotify.Connected)
	h.run(t)
	b := h.item(t, kind.SourceSpotifyAlbum, "B")

	fake.saved = `[]`
	h.day = "2026-11-16"
	h.run(t)
	fake.saved = savedAlbum
	h.day = "2026-11-17"
	h.run(t)

	latestSaved, err := h.store.ItemsLastObservedAs(ctx, spotify.ObservationSource, []string{"saved", "unsaved"}, "saved")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(latestSaved, b) {
		t.Fatalf("saved = %v, want it to include %d", latestSaved, b)
	}
	if got := h.observations(t, b, "saved"); len(got) != 2 || got[1].Key != "saved:B:2026-11-10T09:00:00Z:2026-11-17" {
		t.Fatalf("saved = %+v", got)
	}
}

func TestASavedTotalThatDoesNotMatchFailsTheRunBeforeAnyUnsave(t *testing.T) {
	fake := newFake()
	h := newHarness(t, fake, spotify.Connected)
	h.run(t)
	b := h.item(t, kind.SourceSpotifyAlbum, "B")

	missing := 1
	fake.saved = `[]`
	fake.savedTotal = &missing
	h.day = "2026-11-16"
	report := h.run(t)
	if report.State != syncs.Failed {
		t.Fatalf("report = %+v", report)
	}
	if got := h.observations(t, b, "unsaved"); len(got) != 0 {
		t.Fatalf("unsaved = %+v", got)
	}
}

func TestRunningAgainTheSameDayWithAnotherOrderRecordsNoSecondSnapshot(t *testing.T) {
	fake := newFake()
	h := newHarness(t, fake, spotify.Connected)
	h.run(t)
	x := h.item(t, kind.SourceSpotifyArtist, "X")
	y := h.item(t, kind.SourceSpotifyArtist, "Y")
	before := len(h.observations(t, x, "top-artist")) + len(h.observations(t, y, "top-artist"))

	fake.topArtists["long_term"] = "[" + fullY + "," + fullX + "]"
	report := h.run(t)
	after := len(h.observations(t, x, "top-artist")) + len(h.observations(t, y, "top-artist"))
	if before != after || report.Counts[spotify.CountSnapshots] != 0 {
		t.Fatalf("before = %d, after = %d, counts = %v", before, after, report.Counts)
	}
}

func (h *harness) status(t *testing.T, itemID int64) (string, string) {
	t.Helper()
	view, err := h.store.GetItem(context.Background(), itemID)
	if err != nil {
		t.Fatal(err)
	}
	if view.Entry == nil {
		return "", ""
	}
	return view.Entry.Status, view.Entry.StatusSetBy
}

func TestArtistsAreLoggedFromTopLists(t *testing.T) {
	fake := newFake()
	h := newHarness(t, fake, spotify.Connected)
	h.run(t)
	x := h.item(t, kind.SourceSpotifyArtist, "X")
	y := h.item(t, kind.SourceSpotifyArtist, "Y")
	if status, setBy := h.status(t, x); status != kind.Listening || setBy != kind.SourceSpotify {
		t.Fatalf("X = %s by %s", status, setBy)
	}
	if status, setBy := h.status(t, y); status != kind.Listened || setBy != kind.SourceSpotify {
		t.Fatalf("Y = %s by %s", status, setBy)
	}

	fake.topArtists["short_term"] = "[]"
	fake.topArtists["long_term"] = "[" + fullX + "," + fullY + "]"
	h.day = "2026-11-16"
	h.run(t)
	if status, _ := h.status(t, x); status != kind.Listened {
		t.Fatalf("X after an empty four weeks = %s, want listened", status)
	}
	if status, _ := h.status(t, y); status != kind.Listened {
		t.Fatalf("Y stays logged as %s", status)
	}
}

func TestAnArtistReturnsToListeningWhenItIsBackInTheNewestFourWeeks(t *testing.T) {
	fake := newFake()
	h := newHarness(t, fake, spotify.Connected)
	h.run(t)
	x := h.item(t, kind.SourceSpotifyArtist, "X")

	fake.topArtists["short_term"] = "[]"
	h.day = "2026-11-16"
	h.run(t)
	fake.topArtists["short_term"] = "[" + fullX + "]"
	h.day = "2026-11-17"
	h.run(t)
	if status, _ := h.status(t, x); status != kind.Listening {
		t.Fatalf("X = %s, want listening", status)
	}
}

func TestADismissedArtistStaysDismissed(t *testing.T) {
	fake := newFake()
	h := newHarness(t, fake, spotify.Connected)
	h.run(t)
	x := h.item(t, kind.SourceSpotifyArtist, "X")
	dismissed := kind.Dismissed
	if err := h.store.ChangeEntry(context.Background(), x, store.EntryChange{Status: &dismissed}); err != nil {
		t.Fatal(err)
	}

	h.day = "2026-11-16"
	h.run(t)
	if status, setBy := h.status(t, x); status != kind.Dismissed || setBy != kind.SourceAP {
		t.Fatalf("X = %s by %s", status, setBy)
	}

	listened := kind.Listened
	if err := h.store.ChangeEntry(context.Background(), x, store.EntryChange{Status: &listened}); err != nil {
		t.Fatal(err)
	}
	h.day = "2026-11-17"
	h.run(t)
	if status, setBy := h.status(t, x); status != kind.Listening || setBy != kind.SourceSpotify {
		t.Fatalf("X after ap handed it back = %s by %s, want listening by spotify", status, setBy)
	}
}

func TestARerunTheSameDayRecomputesTheSameStatuses(t *testing.T) {
	h := newHarness(t, newFake(), spotify.Connected)
	h.run(t)
	report := h.run(t)
	if report.State != syncs.Succeeded || report.Counts[string(store.StatusAlreadySo)] != 2 || report.Counts[string(store.StatusLogged)] != 0 || report.Counts[string(store.StatusFollowed)] != 0 {
		t.Fatalf("report = %+v", report)
	}
	logged, err := h.store.ListItems(context.Background(), store.ItemFilter{Kind: kind.Artist})
	if err != nil || len(logged) != 2 {
		t.Fatalf("artists = %d, %v", len(logged), err)
	}
}

func TestAnArtistWhoseCatalogueRefreshFailedHasNoPopularityOrFollowers(t *testing.T) {
	fake := newFake()
	fake.topArtists = map[string]string{"short_term": "[]", "medium_term": "[]", "long_term": "[]"}
	fake.artistsGone = true
	h := newHarness(t, fake, spotify.Connected)
	h.run(t)
	x := h.item(t, kind.SourceSpotifyArtist, "X")
	view, err := h.store.GetItem(context.Background(), x)
	if err != nil {
		t.Fatal(err)
	}
	verdicts, err := json.Marshal(view.Verdicts)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(verdicts), "popularity") || strings.Contains(string(verdicts), "followers") {
		t.Fatalf("verdicts = %s", verdicts)
	}
	if !strings.Contains(string(verdicts), "Xena") {
		t.Fatalf("verdicts = %s, want the simplified name", verdicts)
	}
}

func TestPlaysAreRecordedByTheirOwnSync(t *testing.T) {
	fake := newFake()
	fake.plays = playedAt("2026-11-14T20:00:00.000Z", "2026-11-14T20:05:00.000Z")
	h := newHarness(t, fake, spotify.Connected)
	if report := h.runPlays(t); report.State != syncs.Succeeded || report.Counts[spotify.CountPlays] != 2 {
		t.Fatalf("report = %+v", report)
	}
	a := h.item(t, kind.SourceSpotifyAlbum, "A")

	fake.plays = playedAt("2026-11-14T20:05:00.000Z", "2026-11-14T21:00:00.000Z")
	report := h.runPlays(t)
	if report.State != syncs.Succeeded || report.Counts[spotify.CountPlays] != 1 {
		t.Fatalf("second report = %+v", report)
	}
	if got := h.observations(t, a, "play"); len(got) != 3 {
		t.Fatalf("plays = %+v, want one observation per play", got)
	}
}

func TestTheNightlySyncLeavesPlaysAlone(t *testing.T) {
	h := newHarness(t, newFake(), spotify.Connected)
	report := h.run(t)
	if report.Counts[spotify.CountPlays] != 0 {
		t.Fatalf("counts = %v", report.Counts)
	}
	a := h.item(t, kind.SourceSpotifyAlbum, "A")
	if got := h.observations(t, a, "play"); len(got) != 0 {
		t.Fatalf("plays = %+v", got)
	}
}

func TestThePlaysSyncAndTheNightlySyncCanOverlap(t *testing.T) {
	fake := newFake()
	h := newHarness(t, fake, spotify.Connected)
	nightly, err := syncs.New(context.Background(), spotify.SyncName, h.syncer.Work, h.store, nil)
	if err != nil {
		t.Fatal(err)
	}
	hourly, err := syncs.New(context.Background(), spotify.PlaysSyncName, h.syncer.WorkPlays, h.store, nil)
	if err != nil {
		t.Fatal(err)
	}
	reports := make(chan syncs.Report, 2*overlapRounds)
	var running sync.WaitGroup
	for _, s := range []*syncs.Sync{nightly, hourly} {
		running.Add(1)
		go func() {
			defer running.Done()
			for range overlapRounds {
				reports <- s.Run(context.Background())
			}
		}()
	}
	running.Wait()
	close(reports)
	for report := range reports {
		if report.State != syncs.Succeeded || len(report.Problems) != 0 {
			t.Fatalf("report = %+v", report)
		}
	}
	a := h.item(t, kind.SourceSpotifyAlbum, "A")
	if got := h.observations(t, a, "play"); len(got) != 1 {
		t.Fatalf("plays = %+v, want the one play once", got)
	}
}

const overlapRounds = 5
