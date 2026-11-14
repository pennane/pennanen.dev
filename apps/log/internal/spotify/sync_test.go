package spotify_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pennane/pennanen.dev/apps/log/internal/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/spotify"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/syncs"
)

type fakeSpotify struct {
	topArtists map[string]string
	saved      string
	albumsNull bool
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
		fmt.Fprint(w, `{"items":[{"played_at":"2026-11-14T20:00:00.000Z","track":{"id":"T","name":"Track","album":{"id":"A","name":"Alpha","artists":[{"id":"X","name":"Xena"}]},"artists":[{"id":"X","name":"Xena"}]}}],"next":null}`)
	case "/me/albums":
		fmt.Fprintf(w, `{"items":%s,"next":null}`, f.saved)
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
	}
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

	for _, id := range []int64{x, y, a, b} {
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
	plays := h.observations(t, a, "play")
	if len(plays) != 1 || plays[0].ObservedAt != "2026-11-14T20:00:00.000Z" {
		t.Fatalf("plays = %+v", plays)
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
	if report.Counts[spotify.CountPlays] != 1 || report.Counts[spotify.CountSaves] != 1 || report.Counts[spotify.CountSnapshots] != 4 {
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
