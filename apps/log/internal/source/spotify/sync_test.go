package spotify_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pennane/pennanen.dev/apps/log/internal/platform/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
	"github.com/pennane/pennanen.dev/apps/log/internal/rule"
	"github.com/pennane/pennanen.dev/apps/log/internal/source/spotify"
	"github.com/pennane/pennanen.dev/apps/log/internal/source/syncs"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/take"
	"github.com/pennane/pennanen.dev/apps/log/internal/world"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

type fakeSpotify struct {
	topArtists       map[string]string
	saved            string
	albumsNull       bool
	savedTotal       *int
	artistsGone      bool
	plays            string
	shortTermTracks  string
	catalogueFetches atomic.Int64
	isrcs            map[string]string
	albumTracks      map[string]string
	trackFetches     atomic.Int64
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
			items = f.shortTermTracks
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
		f.catalogueFetches.Add(1)
		if f.albumsNull {
			fmt.Fprint(w, `{"albums":[null]}`)
			return
		}
		items := []string{}
		for _, id := range strings.Split(r.URL.Query().Get("ids"), ",") {
			items = append(items, f.album(id))
		}
		fmt.Fprintf(w, `{"albums":[%s]}`, strings.Join(items, ","))
	case "/tracks":
		f.trackFetches.Add(1)
		items := []string{}
		for _, id := range strings.Split(r.URL.Query().Get("ids"), ",") {
			items = append(items, fmt.Sprintf(`{"id":%q,"name":"Song %s","duration_ms":200400,"external_ids":{"isrc":%q}}`, id, id, f.isrcs[id]))
		}
		fmt.Fprintf(w, `{"tracks":[%s]}`, strings.Join(items, ","))
	case "/artists":
		f.catalogueFetches.Add(1)
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

func (f *fakeSpotify) album(id string) string {
	album := map[string]any{}
	var saved []struct {
		Album map[string]any `json:"album"`
	}
	if json.Unmarshal([]byte(f.saved), &saved) == nil {
		for _, s := range saved {
			if s.Album["id"] == id {
				album = s.Album
			}
		}
	}
	if len(album) == 0 {
		json.Unmarshal([]byte(fmt.Sprintf(`{"id":%q,"name":"Album %s","label":"Label %s","popularity":3,"artists":[{"id":"X","name":"Xena"}]}`, id, id, id)), &album)
	}
	if listed, known := f.albumTracks[id]; known {
		album["tracks"] = json.RawMessage(listed)
	}
	encoded, _ := json.Marshal(album)
	return string(encoded)
}

func newFake() *fakeSpotify {
	return &fakeSpotify{
		topArtists: map[string]string{
			"short_term":  "[" + fullX + "]",
			"medium_term": "[]",
			"long_term":   "[" + fullX + "," + fullY + "]",
		},
		saved:           savedAlbum,
		plays:           playedAt("2026-11-14T20:00:00.000Z"),
		shortTermTracks: topTrackFrom(`{"id":"A","name":"Alpha","artists":[{"id":"X","name":"Xena"}]}`),
	}
}

func topTrackFrom(album string) string {
	return fmt.Sprintf(`[{"id":"T","name":"Track","album":%s,"artists":[{"id":"X","name":"Xena"}]}]`, album)
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
	s, _ := openStoreProposing(t, spotify.Proposer)
	return s
}

func openStoreProposing(t *testing.T, proposer rule.Proposer) (*store.Store, *sql.DB) {
	t.Helper()
	d, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	if err := take.WriteKindStatuses(context.Background(), d, kind.All()); err != nil {
		t.Fatal(err)
	}
	records := record.New(d, reading.ReadRecord, reading.ReadersVersion)
	return store.New(d, records, world.New(d, records), []rule.Proposer{proposer}), d
}

type harness struct {
	store   *store.Store
	db      *sql.DB
	syncer  *spotify.Syncer
	day     string
	hourUTC int
}

func newHarness(t *testing.T, fake *fakeSpotify, connection spotify.Connection) *harness {
	t.Helper()
	return newHarnessProposing(t, fake, connection, spotify.Proposer)
}

func newHarnessProposing(t *testing.T, fake *fakeSpotify, connection spotify.Connection, proposer rule.Proposer) *harness {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(fake.handler))
	t.Cleanup(server.Close)
	s, d := openStoreProposing(t, proposer)
	if connection == spotify.Connected {
		if err := s.PutSecret(context.Background(), spotify.RefreshTokenSecret, "r1"); err != nil {
			t.Fatal(err)
		}
	}
	client := spotify.NewClient("id", "secret", s)
	client.AccountsURL = server.URL
	client.APIURL = server.URL
	client.HTTP = server.Client()
	h := &harness{store: s, db: d, syncer: spotify.NewSyncer(s, client), day: "2026-11-15", hourUTC: 3}
	h.syncer.Now = func() time.Time {
		day, err := time.Parse(time.DateOnly, h.day)
		if err != nil {
			t.Fatal(err)
		}
		return day.Add(time.Duration(h.hourUTC) * time.Hour)
	}
	return h
}

func (h *harness) run(t *testing.T) syncs.Report {
	t.Helper()
	sync, err := syncs.New(context.Background(), spotify.SyncName, record.SourceSpotify, h.syncer.Work, h.store, nil)
	if err != nil {
		t.Fatal(err)
	}
	return sync.Run(context.Background())
}

func (h *harness) runPlays(t *testing.T) syncs.Report {
	t.Helper()
	sync, err := syncs.New(context.Background(), spotify.PlaysSyncName, record.SourceSpotify, h.syncer.WorkPlays, h.store, nil)
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
	x := h.item(t, record.SourceSpotifyArtist, "X")
	y := h.item(t, record.SourceSpotifyArtist, "Y")
	a := h.item(t, record.SourceSpotifyAlbum, "A")
	b := h.item(t, record.SourceSpotifyAlbum, "B")

	if view, err := h.store.GetItem(ctx, a); err != nil || view.Entry != nil {
		t.Fatalf("unsaved album entry = %v, err = %v", view.Entry, err)
	}
	if status, _ := h.status(t, b); status != kind.Want {
		t.Fatalf("saved album = %q, want want", status)
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

	claim, err := h.store.Claim(ctx, a, record.SourceSpotifyAlbum)
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
	want := store.Relation{FromID: albumID, Type: kind.LinkBy, ToID: artistID}
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
	a := h.item(t, record.SourceSpotifyAlbum, "A")
	x := h.item(t, record.SourceSpotifyArtist, "X")
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
	b := h.item(t, record.SourceSpotifyAlbum, "B")

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
	b := h.item(t, record.SourceSpotifyAlbum, "B")

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
	b := h.item(t, record.SourceSpotifyAlbum, "B")
	claim, err := h.store.Claim(context.Background(), b, record.SourceSpotifyAlbum)
	if err != nil || !strings.Contains(string(claim.Data), `"label":"Lb"`) {
		t.Fatalf("claim = %s, %v", claim.Data, err)
	}
}

func TestAnAlbumSavedAgainAfterAnUnsaveIsSavedAgain(t *testing.T) {
	ctx := context.Background()
	fake := newFake()
	h := newHarness(t, fake, spotify.Connected)
	h.run(t)
	b := h.item(t, record.SourceSpotifyAlbum, "B")

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
	b := h.item(t, record.SourceSpotifyAlbum, "B")

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
	x := h.item(t, record.SourceSpotifyArtist, "X")
	y := h.item(t, record.SourceSpotifyArtist, "Y")
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
	return view.Entry.Status, view.Entry.StatusDecidedBy
}

func TestArtistsAreLoggedFromTopLists(t *testing.T) {
	fake := newFake()
	h := newHarness(t, fake, spotify.Connected)
	h.run(t)
	x := h.item(t, record.SourceSpotifyArtist, "X")
	y := h.item(t, record.SourceSpotifyArtist, "Y")
	if status, setBy := h.status(t, x); status != kind.Listening || setBy != spotify.RuleArtist {
		t.Fatalf("X = %s by %s", status, setBy)
	}
	if status, setBy := h.status(t, y); status != kind.Listened || setBy != spotify.RuleArtist {
		t.Fatalf("Y = %s by %s", status, setBy)
	}

	fake.topArtists["short_term"] = "[]"
	fake.topArtists["long_term"] = "[" + fullX + "," + fullY + "]"
	h.day = "2026-11-16"
	h.run(t)
	t.Run("X after an empty four weeks is listened", func(t *testing.T) {
		if status, _ := h.status(t, x); status != kind.Listened {
			t.Fatalf("X after an empty four weeks = %s, want listened", status)
		}
	})
	if status, _ := h.status(t, y); status != kind.Listened {
		t.Fatalf("Y stays logged as %s", status)
	}
}

func TestAnArtistReturnsToListeningWhenItIsBackInTheNewestFourWeeks(t *testing.T) {
	fake := newFake()
	h := newHarness(t, fake, spotify.Connected)
	h.run(t)
	x := h.item(t, record.SourceSpotifyArtist, "X")

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
	ctx := context.Background()
	fake := newFake()
	h := newHarness(t, fake, spotify.Connected)
	h.run(t)
	x := h.item(t, record.SourceSpotifyArtist, "X")
	dismissed := store.Dismissed
	if err := h.store.ChangeEntry(ctx, x, store.EntryChange{Dismissal: &dismissed}); err != nil {
		t.Fatal(err)
	}

	h.day = "2026-11-16"
	h.run(t)
	if v, _ := h.store.GetItem(ctx, x); v.DismissedAt == "" {
		t.Fatalf("X was undismissed by a sync")
	}
	if listed, _ := h.store.ListItems(ctx, store.ItemFilter{Kind: kind.Artist}); slices.ContainsFunc(listed, func(v store.ItemView) bool { return v.ID == x }) {
		t.Fatalf("a dismissed artist is listed")
	}

	kept := store.Kept
	listened := kind.Listened
	if err := h.store.ChangeEntry(ctx, x, store.EntryChange{Dismissal: &kept, Status: &listened}); err != nil {
		t.Fatal(err)
	}
	untriaged := kind.Untriaged
	if err := h.store.ChangeEntry(ctx, x, store.EntryChange{Status: &untriaged}); err != nil {
		t.Fatal(err)
	}
	h.day = "2026-11-17"
	h.run(t)
	if status, setBy := h.status(t, x); status != kind.Listening || setBy != spotify.RuleArtist {
		t.Fatalf("X after ap handed it back = %s by %s, want listening by the lists", status, setBy)
	}
}

func TestARerunTheSameDayRecomputesTheSameStatuses(t *testing.T) {
	h := newHarness(t, newFake(), spotify.Connected)
	h.run(t)
	entries := func() map[string]store.Entry {
		t.Helper()
		got := map[string]store.Entry{}
		for _, id := range []struct{ source, externalID string }{
			{record.SourceSpotifyArtist, "X"}, {record.SourceSpotifyArtist, "Y"}, {record.SourceSpotifyAlbum, "B"},
		} {
			v, err := h.store.GetItem(context.Background(), h.item(t, id.source, id.externalID))
			if err != nil || v.Entry == nil {
				t.Fatalf("%s %s: %+v %v", id.source, id.externalID, v.Entry, err)
			}
			got[id.externalID] = *v.Entry
		}
		return got
	}
	before := entries()
	if report := h.run(t); report.State != syncs.Succeeded {
		t.Fatalf("report = %+v", report)
	}
	after := entries()
	for id, was := range before {
		now := after[id]
		if now.Status != was.Status || now.StatusDecidedBy != was.StatusDecidedBy || !reflect.DeepEqual(now.StatusProposals, was.StatusProposals) {
			t.Errorf("%s: %+v, then %+v", id, was, now)
		}
	}
	logged, err := h.store.ListItems(context.Background(), store.ItemFilter{Kind: kind.Artist})
	if err != nil || len(logged) != 2 {
		t.Fatalf("artists = %d, %v", len(logged), err)
	}
}

func TestAnArtistThatLeftTheShortTermListIsListened(t *testing.T) {
	for _, c := range []struct {
		name      string
		shortTerm string
	}{
		{"an empty list", "[]"},
		{"another artist in its place", "[" + fullY + "]"},
	} {
		t.Run(c.name, func(t *testing.T) {
			fake := newFake()
			h := newHarness(t, fake, spotify.Connected)
			h.run(t)
			x := h.item(t, record.SourceSpotifyArtist, "X")
			if status, _ := h.status(t, x); status != kind.Listening {
				t.Fatalf("X on the list = %s", status)
			}

			fake.topArtists["short_term"] = c.shortTerm
			h.day = "2026-11-16"
			h.run(t)
			v, err := h.store.GetItem(context.Background(), x)
			if err != nil {
				t.Fatal(err)
			}
			want := []store.StatusProposal{{Rule: spotify.RuleArtist, Value: kind.Listened, ChangedAt: "2026-11-16T03:00:00Z"}}
			if v.Entry.Status != kind.Listened || v.Entry.StatusDecidedBy != spotify.RuleArtist || !reflect.DeepEqual(v.Entry.StatusProposals, want) {
				t.Fatalf("X off the list = %+v", v.Entry)
			}
		})
	}
}

func TestApsListenedArtistFollowsTheListsWhenTheyChange(t *testing.T) {
	ctx := context.Background()
	fake := newFake()
	h := newHarness(t, fake, spotify.Connected)
	h.run(t)
	x := h.item(t, record.SourceSpotifyArtist, "X")
	listened := kind.Listened
	if err := h.store.ChangeEntry(ctx, x, store.EntryChange{Status: &listened}); err != nil {
		t.Fatal(err)
	}

	h.day = "2099-01-01"
	h.run(t)
	if status, by := h.status(t, x); status != kind.Listened || by != record.SourceAP {
		t.Fatalf("lists unchanged = %s by %s, want ap's listened", status, by)
	}

	fake.topArtists["short_term"] = "[]"
	h.day = "2099-01-02"
	h.run(t)
	if status, by := h.status(t, x); status != kind.Listened || by != spotify.RuleArtist {
		t.Fatalf("off the list = %s by %s, want listened by the lists", status, by)
	}

	fake.topArtists["short_term"] = "[" + fullX + "]"
	h.day = "2099-01-03"
	h.run(t)
	if status, by := h.status(t, x); status != kind.Listening || by != spotify.RuleArtist {
		t.Fatalf("back on the list = %s by %s, want listening by the lists", status, by)
	}
}

func TestUnsavingAnAlbumDoesNotWithdrawIt(t *testing.T) {
	fake := newFake()
	h := newHarness(t, fake, spotify.Connected)
	h.run(t)
	b := h.item(t, record.SourceSpotifyAlbum, "B")

	fake.saved = `[]`
	h.day = "2026-11-16"
	h.run(t)
	if got := h.observations(t, b, record.SignalUnsaved); len(got) != 1 {
		t.Fatalf("unsaved = %+v", got)
	}
	v, err := h.store.GetItem(context.Background(), b)
	if err != nil {
		t.Fatal(err)
	}
	if v.Entry == nil || v.Entry.LoggedBy != spotify.RuleAlbum || v.Entry.Status != kind.Want || v.Entry.StatusDecidedBy != spotify.RuleAlbum {
		t.Fatalf("unsaved album = %+v", v.Entry)
	}
}

func TestAnArtistWhoseCatalogueRefreshFailedHasNoPopularityOrFollowers(t *testing.T) {
	fake := newFake()
	fake.topArtists = map[string]string{"short_term": "[]", "medium_term": "[]", "long_term": "[]"}
	fake.artistsGone = true
	h := newHarness(t, fake, spotify.Connected)
	h.run(t)
	x := h.item(t, record.SourceSpotifyArtist, "X")
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
	a := h.item(t, record.SourceSpotifyAlbum, "A")

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
	a := h.item(t, record.SourceSpotifyAlbum, "A")
	if got := h.observations(t, a, "play"); len(got) != 0 {
		t.Fatalf("plays = %+v", got)
	}
}

func TestThePlaysSyncAndTheNightlySyncCanOverlap(t *testing.T) {
	fake := newFake()
	h := newHarness(t, fake, spotify.Connected)
	nightly, err := syncs.New(context.Background(), spotify.SyncName, record.SourceSpotify, h.syncer.Work, h.store, nil)
	if err != nil {
		t.Fatal(err)
	}
	hourly, err := syncs.New(context.Background(), spotify.PlaysSyncName, record.SourceSpotify, h.syncer.WorkPlays, h.store, nil)
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
	a := h.item(t, record.SourceSpotifyAlbum, "A")
	if got := h.observations(t, a, "play"); len(got) != 1 {
		t.Fatalf("plays = %+v, want the one play once", got)
	}
}

const overlapRounds = 5

func TestTheHourlySyncFetchesNothingForAlbumsItAlreadyHasInFull(t *testing.T) {
	fake := newFake()
	h := newHarness(t, fake, spotify.Connected)
	h.run(t)
	if report := h.runPlays(t); report.State != syncs.Succeeded {
		t.Fatalf("report = %+v", report)
	}
	fetchesBefore := fake.catalogueFetches.Load()
	a := h.item(t, record.SourceSpotifyAlbum, "A")
	before, err := h.store.Claim(context.Background(), a, record.SourceSpotifyAlbum)
	if err != nil {
		t.Fatal(err)
	}

	h.runPlays(t)
	if got := fake.catalogueFetches.Load(); got != fetchesBefore {
		t.Fatalf("catalogue fetches went from %d to %d", fetchesBefore, got)
	}
	after, err := h.store.Claim(context.Background(), a, record.SourceSpotifyAlbum)
	if err != nil || string(after.Data) != string(before.Data) || after.FetchedAt != before.FetchedAt {
		t.Fatalf("claim changed: %+v -> %+v, %v", before, after, err)
	}
}

func TestThePlaysSyncSucceedsWithNothingToDoWhenNotConnected(t *testing.T) {
	h := newHarness(t, newFake(), spotify.NotConnected)
	if report := h.runPlays(t); report.State != syncs.Succeeded || len(report.Problems) != 0 {
		t.Fatalf("report = %+v", report)
	}
}

func TestAnArtistApTaggedBeforeAnySyncStillGetsItsComputedStatus(t *testing.T) {
	h := newHarness(t, newFake(), spotify.Connected)
	h.run(t)
	z, err := h.store.KnowItemFromSource(context.Background(), kind.Artist, "Zed", record.SourceSpotifyArtist, "Z", json.RawMessage(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	if err := h.store.ChangeEntry(context.Background(), z, store.EntryChange{AddTags: []string{"favorite"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := h.store.Observe(context.Background(), store.Observation{ItemID: z, Source: spotify.ObservationSource, Signal: record.SignalTopArtist, Key: "top-artist:long_term:2026-11-15:Z", ObservedAt: "2026-11-15T03:00:00Z", Data: json.RawMessage(`{}`)}); err != nil {
		t.Fatal(err)
	}
	h.run(t)
	if status, setBy := h.status(t, z); status != kind.Listened || setBy != spotify.RuleArtist {
		t.Fatalf("Z = %s by %s", status, setBy)
	}
}

func TestAHandPickedListeningOnAnUnlistedArtistStands(t *testing.T) {
	h := newHarness(t, newFake(), spotify.Connected)
	h.run(t)
	z, err := h.store.KnowItemFromSource(context.Background(), kind.Artist, "Zed", record.SourceSpotifyArtist, "Z", json.RawMessage(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	listening := kind.Listening
	if err := h.store.ChangeEntry(context.Background(), z, store.EntryChange{Status: &listening}); err != nil {
		t.Fatal(err)
	}
	h.run(t)
	if status, setBy := h.status(t, z); status != kind.Listening || setBy != record.SourceAP {
		t.Fatalf("Z = %s by %s, want ap's listening", status, setBy)
	}
}

func (h *harness) relations(t *testing.T, albumID int64) map[store.Relation]struct{} {
	t.Helper()
	view, err := h.store.GetItem(context.Background(), albumID)
	if err != nil {
		t.Fatal(err)
	}
	from := map[store.Relation]struct{}{}
	for _, r := range view.Relations {
		if r.FromID == albumID {
			from[r] = struct{}{}
		}
	}
	return from
}

const (
	collabAlbum      = `{"added_at":"2026-11-10T09:00:00Z","album":{"id":"C","name":"Collab","artists":[{"id":"P","name":"Pia"},{"id":"Q","name":"Quin"}]}}`
	compilationAlbum = `{"added_at":"2026-11-10T09:00:00Z","album":{"id":"V","name":"Hits","artists":[{"id":"` + kind.VariousArtists + `","name":"Various Artists"},{"id":"R","name":"Rei"}]}}`
)

func TestKnowAlbumRelatesTheMainArtistOnly(t *testing.T) {
	fake := newFake()
	fake.saved = "[" + collabAlbum + "," + compilationAlbum + "]"
	h := newHarness(t, fake, spotify.Connected)
	if report := h.run(t); len(report.Problems) != 0 {
		t.Fatalf("report = %+v", report)
	}
	collab := h.item(t, record.SourceSpotifyAlbum, "C")
	p := h.item(t, record.SourceSpotifyArtist, "P")
	q := h.item(t, record.SourceSpotifyArtist, "Q")
	want := map[store.Relation]struct{}{
		{FromID: collab, Type: kind.LinkBy, ToID: p}:   {},
		{FromID: collab, Type: kind.LinkWith, ToID: q}: {},
	}
	if got := h.relations(t, collab); !reflect.DeepEqual(got, want) {
		t.Fatalf("collab relations = %v, want %v", got, want)
	}

	compilation := h.item(t, record.SourceSpotifyAlbum, "V")
	various := h.item(t, record.SourceSpotifyArtist, kind.VariousArtists)
	r := h.item(t, record.SourceSpotifyArtist, "R")
	want = map[store.Relation]struct{}{
		{FromID: compilation, Type: kind.LinkWith, ToID: various}: {},
		{FromID: compilation, Type: kind.LinkWith, ToID: r}:       {},
	}
	if got := h.relations(t, compilation); !reflect.DeepEqual(got, want) {
		t.Fatalf("compilation relations = %v, want %v", got, want)
	}
}

const (
	zedAlbum      = `{"id":"S","name":"Sigma","label":"Ls","artists":[{"id":"Z","name":"Zed"}]}`
	savedZedAlbum = `[{"added_at":"2026-11-10T09:00:00Z","album":` + zedAlbum + `}]`
	alphaAlbum    = `{"id":"A","name":"Alpha","label":"La","artists":[{"id":"X","name":"Xena"}]}`
)

func TestASavedAlbumIsWantedUntilHeard(t *testing.T) {
	fake := newFake()
	fake.saved = savedZedAlbum
	h := newHarness(t, fake, spotify.Connected)
	if report := h.run(t); len(report.Problems) != 0 {
		t.Fatalf("report = %+v", report)
	}
	s := h.item(t, record.SourceSpotifyAlbum, "S")
	z := h.item(t, record.SourceSpotifyArtist, "Z")
	if status, setBy := h.status(t, s); status != kind.Want || setBy != spotify.RuleAlbum {
		t.Fatalf("day 1 album = %s by %s, want want by spotify", status, setBy)
	}
	if status, setBy := h.status(t, z); status != kind.Listened || setBy != spotify.RuleArtist {
		t.Fatalf("day 1 main artist = %s by %s, want listened by spotify", status, setBy)
	}

	fake.shortTermTracks = topTrackFrom(zedAlbum)
	h.day = "2026-11-16"
	h.run(t)
	if status, _ := h.status(t, s); status != kind.Listened {
		t.Fatalf("day 2 album = %s, want listened", status)
	}

	fake.saved = `[]`
	fake.shortTermTracks = `[]`
	h.day = "2026-11-17"
	h.run(t)
	if got := h.observations(t, s, record.SignalUnsaved); len(got) != 1 {
		t.Fatalf("unsaved = %+v", got)
	}
	if status, setBy := h.status(t, s); status != kind.Listened || setBy != spotify.RuleAlbum {
		t.Fatalf("day 3 album = %s by %s, want listened by spotify", status, setBy)
	}
}

func TestApsAlbumStatusSticks(t *testing.T) {
	fake := newFake()
	fake.saved = `[{"added_at":"2026-11-10T09:00:00Z","album":` + alphaAlbum + `}]`
	h := newHarness(t, fake, spotify.Connected)
	h.runPlays(t)
	h.run(t)
	a := h.item(t, record.SourceSpotifyAlbum, "A")
	if status, _ := h.status(t, a); status != kind.Listened {
		t.Fatalf("heard album = %s, want listened", status)
	}
	want := kind.Want
	if err := h.store.ChangeEntry(context.Background(), a, store.EntryChange{Status: &want}); err != nil {
		t.Fatal(err)
	}
	h.day = "2026-11-16"
	h.run(t)
	if status, setBy := h.status(t, a); status != kind.Want || setBy != record.SourceAP {
		t.Fatalf("album = %s by %s, want want by ap", status, setBy)
	}
}

func TestTheArtistPassLeavesAlbumsAlone(t *testing.T) {
	h := newHarness(t, newFake(), spotify.Connected)
	for _, day := range []string{"2026-11-15", "2026-11-16"} {
		h.day = day
		if report := h.run(t); report.State != syncs.Succeeded || len(report.Problems) != 0 {
			t.Fatalf("%s report = %+v", day, report)
		}
		b := h.item(t, record.SourceSpotifyAlbum, "B")
		if status, setBy := h.status(t, b); status != kind.Want || setBy != spotify.RuleAlbum {
			t.Fatalf("%s unheard album = %s by %s, want want by spotify", day, status, setBy)
		}
	}
}

func TestACollabLogsOnlyItsMainArtist(t *testing.T) {
	fake := newFake()
	fake.saved = "[" + collabAlbum + "," + compilationAlbum + "]"
	h := newHarness(t, fake, spotify.Connected)
	if report := h.run(t); len(report.Problems) != 0 {
		t.Fatalf("report = %+v", report)
	}
	if status, _ := h.status(t, h.item(t, record.SourceSpotifyArtist, "P")); status != kind.Listened {
		t.Fatalf("main artist = %q, want listened", status)
	}
	for _, onlyKnown := range []string{"Q", "R", kind.VariousArtists} {
		if status, _ := h.status(t, h.item(t, record.SourceSpotifyArtist, onlyKnown)); status != "" {
			t.Fatalf("%s = %q, want only known", onlyKnown, status)
		}
	}
	compilation := h.item(t, record.SourceSpotifyAlbum, "V")
	view, err := h.store.GetItem(context.Background(), compilation)
	if err != nil {
		t.Fatal(err)
	}
	if genres, present := view.Verdicts.Values["genres"]; present {
		t.Fatalf("compilation genres = %v, want none", genres)
	}
}

func savedWithCredits(credits ...string) string {
	artists := []string{}
	for _, id := range credits {
		artists = append(artists, fmt.Sprintf(`{"id":%q,"name":"Artist %s"}`, id, id))
	}
	return `[{"added_at":"2026-11-10T09:00:00Z","album":{"id":"K","name":"Kappa","artists":[` + strings.Join(artists, ",") + `]}}]`
}

func TestAnAlbumsCreditsFollowSpotify(t *testing.T) {
	fake := newFake()
	fake.saved = savedWithCredits("CA", "CB")
	h := newHarness(t, fake, spotify.Connected)
	h.run(t)
	fake.saved = savedWithCredits("CB", "CA")
	h.day = "2026-11-16"
	if report := h.run(t); len(report.Problems) != 0 {
		t.Fatalf("report = %+v", report)
	}
	album := h.item(t, record.SourceSpotifyAlbum, "K")
	want := map[store.Relation]struct{}{
		{FromID: album, Type: kind.LinkBy, ToID: h.item(t, record.SourceSpotifyArtist, "CB")}:   {},
		{FromID: album, Type: kind.LinkWith, ToID: h.item(t, record.SourceSpotifyArtist, "CA")}: {},
	}
	if got := h.relations(t, album); !reflect.DeepEqual(got, want) {
		t.Fatalf("relations = %v, want %v", got, want)
	}
}

func TestAnAlbumMetWithoutCreditsKeepsItsRelations(t *testing.T) {
	fake := newFake()
	fake.saved = savedWithCredits("CA")
	h := newHarness(t, fake, spotify.Connected)
	h.run(t)
	fake.saved = `[{"added_at":"2026-11-10T09:00:00Z","album":{"id":"K","name":"Kappa"}}]`
	h.day = "2026-11-16"
	h.run(t)
	album := h.item(t, record.SourceSpotifyAlbum, "K")
	want := map[store.Relation]struct{}{
		{FromID: album, Type: kind.LinkBy, ToID: h.item(t, record.SourceSpotifyArtist, "CA")}: {},
	}
	if got := h.relations(t, album); !reflect.DeepEqual(got, want) {
		t.Fatalf("relations = %v, want %v", got, want)
	}
}

func playOf(album, at string) string {
	return fmt.Sprintf(`[{"played_at":%q,"track":{"id":"ST","name":"Sigma Track","album":%s,"artists":[{"id":"Z","name":"Zed"}]}}]`, at, album)
}

func TestAPlayMovesAWantedAlbumToListenedWithinTheHour(t *testing.T) {
	fake := newFake()
	fake.saved = savedZedAlbum
	h := newHarness(t, fake, spotify.Connected)
	h.run(t)
	s := h.item(t, record.SourceSpotifyAlbum, "S")
	if status, _ := h.status(t, s); status != kind.Want {
		t.Fatalf("album = %s, want want", status)
	}
	fake.plays = playOf(zedAlbum, "2026-11-15T10:00:00.000Z")
	report := h.runPlays(t)
	if report.State != syncs.Succeeded || len(report.Problems) != 0 {
		t.Fatalf("report = %+v", report)
	}
	if status, setBy := h.status(t, s); status != kind.Listened || setBy != spotify.RuleAlbum {
		t.Fatalf("album after a play = %s by %s, want listened by spotify", status, setBy)
	}
}

func (h *harness) listening(t *testing.T, itemID int64) any {
	t.Helper()
	view, err := h.store.GetItem(context.Background(), itemID)
	if err != nil {
		t.Fatal(err)
	}
	return view.Verdicts.Values[kind.FieldListening]
}

func TestArtistsListenedToMoreWeighMore(t *testing.T) {
	h := newHarness(t, newFake(), spotify.Connected)
	if report := h.run(t); len(report.Problems) != 0 {
		t.Fatalf("report = %+v", report)
	}
	x := h.listening(t, h.item(t, record.SourceSpotifyArtist, "X"))
	y := h.listening(t, h.item(t, record.SourceSpotifyArtist, "Y"))
	if x != int64(100) || y != int64(0) {
		t.Fatalf("listening X = %v, Y = %v, want 100 and 0", x, y)
	}
}

func TestAWeightIsAPercentileThatHoldsAcrossDays(t *testing.T) {
	fake := newFake()
	h := newHarness(t, fake, spotify.Connected)
	h.run(t)
	h.day = "2026-11-16"
	h.run(t)
	if x := h.listening(t, h.item(t, record.SourceSpotifyArtist, "X")); x != int64(100) {
		t.Fatalf("listening X on day two = %v, want 100", x)
	}
}

func TestASavedAlbumListsItsTracksAsRecordingsInOrder(t *testing.T) {
	fake := newFake()
	fake.saved = `[
		{"added_at":"2026-11-10T09:00:00Z","album":{"id":"B","name":"Beta","popularity":1,"artists":[{"id":"Y","name":"Yara"}],
			"tracks":{"total":2,"items":[{"id":"b2","disc_number":1,"track_number":2},{"id":"b1","disc_number":1,"track_number":1}]}}},
		{"added_at":"2026-11-10T09:00:00Z","album":{"id":"S","name":"Single","popularity":1,"artists":[{"id":"Y","name":"Yara"}],
			"tracks":{"total":1,"items":[{"id":"s1","disc_number":1,"track_number":1}]}}}]`
	fake.isrcs = map[string]string{"b1": "ISRC1", "s1": "ISRC1"}
	h := newHarness(t, fake, spotify.Connected)
	if report := h.run(t); report.State != syncs.Succeeded {
		t.Fatalf("report %+v", report)
	}
	ctx := context.Background()
	recording := h.item(t, record.SourceSpotifyTrack, "ISRC1")
	withoutISRC := h.item(t, record.SourceSpotifyTrack, "spotify:b2")
	if parts, err := h.store.PartsOf(ctx, h.item(t, record.SourceSpotifyAlbum, "B")); err != nil || !slices.Equal(parts, []int64{recording, withoutISRC}) {
		t.Errorf("album parts %v %v", parts, err)
	}
	if parts, err := h.store.PartsOf(ctx, h.item(t, record.SourceSpotifyAlbum, "S")); err != nil || !slices.Equal(parts, []int64{recording}) {
		t.Errorf("single parts %v %v", parts, err)
	}
	track, err := h.store.GetItem(ctx, withoutISRC)
	if err != nil || track.Verdicts.Values["duration"] != int64(200) || track.Entry != nil {
		t.Errorf("track %+v %v", track, err)
	}
	fetches := fake.trackFetches.Load()
	h.day = "2026-11-16"
	h.run(t)
	if fake.trackFetches.Load() != fetches {
		t.Errorf("tracklists in step fetched tracks again")
	}
}

func TestAnAlbumMetWithoutItsTracksGetsItsTracklistOnceFetchedInFull(t *testing.T) {
	fake := newFake()
	fake.saved = `[{"added_at":"2026-11-10T09:00:00Z","album":{"id":"A","name":"Alpha","artists":[{"id":"X","name":"Xena"}]}}]`
	fake.albumTracks = map[string]string{"A": `{"total":2,"items":[{"id":"a2","disc_number":1,"track_number":2},{"id":"a1","disc_number":1,"track_number":1}]}`}
	h := newHarness(t, fake, spotify.Connected)
	report := h.run(t)
	if report.State != syncs.Succeeded || report.Counts[spotify.CountTracklists] != 1 {
		t.Fatalf("report %+v", report)
	}
	ctx := context.Background()
	album := h.item(t, record.SourceSpotifyAlbum, "A")
	first := h.item(t, record.SourceSpotifyTrack, "spotify:a1")
	second := h.item(t, record.SourceSpotifyTrack, "spotify:a2")
	if parts, err := h.store.PartsOf(ctx, album); err != nil || !slices.Equal(parts, []int64{first, second}) {
		t.Errorf("album parts %v %v, want %v", parts, err, []int64{first, second})
	}
}

func TestAnAlbumMetWithoutCreditsButWithATracklistIsStillWritten(t *testing.T) {
	fake := newFake()
	fake.saved = savedWithCredits("CA")
	h := newHarness(t, fake, spotify.Connected)
	h.run(t)
	fake.saved = `[{"added_at":"2026-11-10T09:00:00Z","album":{"id":"K","name":"Kappa","tracks":{"total":1,"items":[{"id":"k1","disc_number":1,"track_number":1}]}}}]`
	h.day = "2026-11-16"
	h.run(t)
	album := h.item(t, record.SourceSpotifyAlbum, "K")
	if parts, err := h.store.PartsOf(context.Background(), album); err != nil || !slices.Equal(parts, []int64{h.item(t, record.SourceSpotifyTrack, "spotify:k1")}) {
		t.Errorf("album parts %v %v", parts, err)
	}
}

func TestARerunTheSameDayAfterACrashKeepsTheStreak(t *testing.T) {
	h := newHarness(t, newFake(), spotify.Connected)
	h.run(t)
	x := h.item(t, record.SourceSpotifyArtist, "X")
	h.day = "2026-11-16"
	h.run(t)
	if _, err := h.db.Exec(`DELETE FROM record_observation WHERE signal = ? AND key LIKE '%:2026-11-16'`, record.SignalTopArtistsListed); err != nil {
		t.Fatal(err)
	}

	h.hourUTC = 9
	if report := h.run(t); report.State != syncs.Succeeded || len(report.Problems) != 0 {
		t.Fatalf("rerun: %+v", report)
	}
	v, err := h.store.GetItem(context.Background(), x)
	if err != nil {
		t.Fatal(err)
	}
	want := []store.StatusProposal{{Rule: spotify.RuleArtist, Value: kind.Listening, ChangedAt: "2026-11-15T03:00:00Z"}}
	if v.Entry.Status != kind.Listening || !reflect.DeepEqual(v.Entry.StatusProposals, want) {
		t.Fatalf("X after a rerun the same day = %+v", v.Entry)
	}
	var listingsAfterTheDaysArtists int
	if err := h.db.QueryRow(`SELECT count(*) FROM record_observation WHERE signal = ? AND key LIKE ? AND observed_at > '2026-11-16T03:00:00Z'`, record.SignalTopArtistsListed, record.SignalTopArtistsListed+":short_term:%").Scan(&listingsAfterTheDaysArtists); err != nil || listingsAfterTheDaysArtists != 0 {
		t.Fatalf("listings without that day's artists = %d, %v", listingsAfterTheDaysArtists, err)
	}
}

func TestAFailedProposalMidSyncIsAProblemAndWeighingGoesOn(t *testing.T) {
	failing := rule.Proposer{Source: record.SourceSpotify, Propose: func(context.Context, record.Queryer) ([]rule.Proposal, error) {
		return nil, errors.New("a rule broke")
	}}
	h := newHarnessProposing(t, newFake(), spotify.Connected, failing)
	report := h.run(t)
	if report.State != syncs.Succeeded || !slices.ContainsFunc(report.Problems, func(p string) bool { return strings.Contains(p, "a rule broke") }) {
		t.Fatalf("report = %+v", report)
	}
	x := h.item(t, record.SourceSpotifyArtist, "X")
	listening := kind.Listening
	if err := h.store.ChangeEntry(context.Background(), x, store.EntryChange{Status: &listening}); err != nil {
		t.Fatal(err)
	}
	h.run(t)
	if weight := h.listening(t, x); weight == nil {
		t.Fatalf("X was not weighed after a failed proposal")
	}
}
