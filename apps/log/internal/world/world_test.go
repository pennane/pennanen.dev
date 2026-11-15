package world_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pennane/pennanen.dev/apps/log/internal/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
	"github.com/pennane/pennanen.dev/apps/log/internal/world"
)

type fixture struct {
	t       *testing.T
	ctx     context.Context
	db      *sql.DB
	records *record.Store
	world   *world.World
	clock   time.Time
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	d, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	records := record.New(d, verdict.ReadRecord, verdict.ReadersVersion)
	return &fixture{
		t:       t,
		ctx:     context.Background(),
		db:      d,
		records: records,
		world:   world.New(d, records),
		clock:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func identity(t *testing.T, s string) record.Identity {
	t.Helper()
	id, err := record.ParseIdentity(s)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func (f *fixture) put(s, data string) record.Identity {
	f.t.Helper()
	id := identity(f.t, s)
	f.clock = f.clock.Add(time.Minute)
	if _, err := f.records.PutClaimAt(f.ctx, f.db, id, json.RawMessage(data), f.clock.Format("2006-01-02T15:04:05Z")); err != nil {
		f.t.Fatal(err)
	}
	return id
}

func (f *fixture) settle(ids ...record.Identity) world.Report {
	f.t.Helper()
	report, err := f.world.Settle(f.ctx, f.db, ids)
	if err != nil {
		f.t.Fatal(err)
	}
	return report
}

func (f *fixture) itemOf(id record.Identity) int64 {
	f.t.Helper()
	itemID, err := f.world.ItemOf(f.ctx, f.db, id)
	if err != nil {
		f.t.Fatalf("item of %s: %v", id, err)
	}
	return itemID
}

func (f *fixture) kindAndSlug(itemID int64) (string, string) {
	f.t.Helper()
	var kindName, slug string
	if err := f.db.QueryRowContext(f.ctx, `SELECT kind, slug FROM item WHERE id = ?`, itemID).Scan(&kindName, &slug); err != nil {
		f.t.Fatal(err)
	}
	return kindName, slug
}

type decided struct {
	own      verdict.Verdicts
	verdicts verdict.Verdicts
}

func (f *fixture) decided(itemID int64) decided {
	f.t.Helper()
	var own, verdicts string
	if err := f.db.QueryRowContext(f.ctx, `SELECT own, verdicts FROM item_verdict WHERE item_id = ?`, itemID).Scan(&own, &verdicts); err != nil {
		f.t.Fatalf("verdicts of item %d: %v", itemID, err)
	}
	var d decided
	if err := json.Unmarshal([]byte(own), &d.own); err != nil {
		f.t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(verdicts), &d.verdicts); err != nil {
		f.t.Fatal(err)
	}
	return d
}

func (f *fixture) count(table string) int {
	f.t.Helper()
	var n int
	if err := f.db.QueryRowContext(f.ctx, `SELECT count(*) FROM `+table).Scan(&n); err != nil {
		f.t.Fatal(err)
	}
	return n
}

func (f *fixture) dump(query string) string {
	f.t.Helper()
	rows, err := f.db.QueryContext(f.ctx, query)
	if err != nil {
		f.t.Fatal(err)
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		f.t.Fatal(err)
	}
	var b strings.Builder
	for rows.Next() {
		values := make([]any, len(columns))
		pointers := make([]any, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			f.t.Fatal(err)
		}
		fmt.Fprintln(&b, values...)
	}
	if err := rows.Err(); err != nil {
		f.t.Fatal(err)
	}
	return b.String()
}

func (f *fixture) snapshot() string {
	f.t.Helper()
	return f.dump(`SELECT id, kind, slug FROM item ORDER BY id`) +
		f.dump(`SELECT id, item_id, source, external_id FROM item_identity ORDER BY id`) +
		f.dump(`SELECT item_id, own, verdicts FROM item_verdict ORDER BY item_id`) +
		f.dump(`SELECT * FROM link ORDER BY asserted_source, asserted_external_id, from_source, from_external_id, type, to_source, to_external_id`) +
		f.dump(`SELECT * FROM item_overlap ORDER BY source, external_id, other_item_id`) +
		f.dump(`SELECT * FROM identity_waiting ORDER BY about_source, about_external_id, waiting_source, waiting_external_id`)
}

const (
	hades       = `{"id":1,"name":"Hades"}`
	celeste     = `{"id":2,"name":"Celeste"}`
	halfLife2   = `{"appid":220,"name":"Half-Life 2","playtime_forever":10}`
	artistX     = `{"id":"a","name":"Artist","genres":["x"]}`
	artistY     = `{"id":"a","name":"Artist","genres":["y"]}`
	albumByA    = `{"id":"b","name":"Album","artists":[{"id":"a","name":"Artist"}]}`
	albumTracks = `{"id":"b","name":"Album","artists":[{"id":"a","name":"Artist"}],"tracks":{"items":[{"id":"t1","disc_number":1,"track_number":1},{"id":"t2","disc_number":1,"track_number":2}]}}`
	trackOne    = `{"id":"t1","name":"One","external_ids":{"isrc":"ISRC1"}}`
)

func TestARecordWithANewIdentityMakesAnItemSluggedByItsTitle(t *testing.T) {
	f := newFixture(t)
	igdb := f.put("igdb:1", hades)

	report := f.settle(igdb)

	itemID := f.itemOf(igdb)
	if !slices.Equal(report.Created, []int64{itemID}) {
		t.Fatalf("created = %v, want [%d]", report.Created, itemID)
	}
	if kindName, slug := f.kindAndSlug(itemID); kindName != kind.Game || slug != "hades" {
		t.Fatalf("kind, slug = %s, %s", kindName, slug)
	}
	if f.count("item") != 1 {
		t.Fatalf("items = %d", f.count("item"))
	}
}

func TestRecordsAboutTheSameThingAreOneItem(t *testing.T) {
	f := newFixture(t)
	steam := f.put("steam:220", halfLife2)
	f.settle(steam)
	store := f.put("steam-store:220", `{"name":"Half-Life 2"}`)
	f.settle(store)

	if f.itemOf(steam) != f.itemOf(store) {
		t.Fatalf("steam on %d, steam-store on %d", f.itemOf(steam), f.itemOf(store))
	}
	if f.count("item") != 1 {
		t.Fatalf("items = %d", f.count("item"))
	}
}

func TestALaterIdentityJoinsWithoutChangingIdOrSlug(t *testing.T) {
	f := newFixture(t)
	steam := f.put("steam:220", halfLife2)
	f.settle(steam)
	itemID := f.itemOf(steam)

	matched := f.put("matched:igdb:1", `{"about":["igdb:1","steam:220"]}`)
	igdb := f.put("igdb:1", `{"id":1,"name":"HL2"}`)
	report := f.settle(matched, igdb)

	if f.itemOf(igdb) != itemID || f.itemOf(matched) != itemID {
		t.Fatalf("igdb on %d, matched on %d, want %d", f.itemOf(igdb), f.itemOf(matched), itemID)
	}
	if _, slug := f.kindAndSlug(itemID); slug != "half-life-2" {
		t.Fatalf("slug = %s", slug)
	}
	if len(report.Created) != 0 {
		t.Fatalf("created = %v", report.Created)
	}
	if title := verdict.Title(f.decided(itemID).verdicts); title != "HL2" {
		t.Fatalf("title = %q", title)
	}
}

func TestAnIdentityOwnedByAnotherItemIsAnOverlapNotAMerge(t *testing.T) {
	f := newFixture(t)
	first := f.put("igdb:1", hades)
	second := f.put("igdb:2", celeste)
	f.settle(first)
	f.settle(second)
	firstItem, secondItem := f.itemOf(first), f.itemOf(second)

	legacy := f.put("legacy:item/9", `{"about":["igdb:1","igdb:2"]}`)
	report := f.settle(legacy)

	want := []world.Overlap{{Identity: second, ItemID: firstItem, OtherItemID: secondItem}}
	if !reflect.DeepEqual(report.Overlaps, want) {
		t.Fatalf("overlaps = %+v, want %+v", report.Overlaps, want)
	}
	if f.itemOf(first) != firstItem || f.itemOf(second) != secondItem || f.itemOf(legacy) != firstItem {
		t.Fatalf("identities moved")
	}
	if f.count("item") != 2 || f.count("item_overlap") != 1 {
		t.Fatalf("items = %d, overlaps = %d", f.count("item"), f.count("item_overlap"))
	}
}

func TestTheKindFollowsTheIdentitySource(t *testing.T) {
	f := newFixture(t)
	film := f.put("letterboxd:abc", `{"id":"abc","title":"Heat"}`)
	f.settle(film)
	itemID := f.itemOf(film)
	if kindName, _ := f.kindAndSlug(itemID); kindName != kind.Movie {
		t.Fatalf("kind = %s", kindName)
	}

	series := f.put("tmdb-tv:5", `{"id":5,"name":"Heat"}`)
	f.put("letterboxd:abc", `{"id":"abc","title":"Heat","tmdbId":5,"tmdbType":"tv"}`)
	report := f.settle(series, film)

	want := []world.Move{{ItemID: itemID, From: kind.Movie, To: kind.Series}}
	if !reflect.DeepEqual(report.Moves, want) {
		t.Fatalf("moves = %+v, want %+v", report.Moves, want)
	}
	if f.itemOf(series) != itemID {
		t.Fatalf("series on %d, want %d", f.itemOf(series), itemID)
	}
	if kindName, slug := f.kindAndSlug(itemID); kindName != kind.Series || slug != "heat" {
		t.Fatalf("kind, slug = %s, %s", kindName, slug)
	}
}

func TestTheOverrideWinsThenSourcesInRank(t *testing.T) {
	f := newFixture(t)
	steam := f.put("steam:220", halfLife2)
	override := f.put("ap:fields/steam:220", `{"fields":{"playtime_minutes":20}}`)
	f.settle(steam, override)

	got := f.decided(f.itemOf(steam)).verdicts
	if got.Values["playtime_minutes"] != float64(20) || got.SourceOf["playtime_minutes"] != kind.SourceAP {
		t.Fatalf("playtime = %v from %s", got.Values["playtime_minutes"], got.SourceOf["playtime_minutes"])
	}
	if got.Values["title"] != "Half-Life 2" || got.SourceOf["title"] != kind.SourceSteam {
		t.Fatalf("title = %v from %s", got.Values["title"], got.SourceOf["title"])
	}
}

func TestTwoApRecordsOnOneItemBothCount(t *testing.T) {
	f := newFixture(t)
	steam := f.put("steam:220", halfLife2)
	override := f.put("ap:fields/steam:220", `{"fields":{"title":"Mine"}}`)
	noMatch := f.put("ap:match/igdb/steam:220", `{"to":null}`)
	f.settle(steam, override, noMatch)

	itemID := f.itemOf(steam)
	if f.itemOf(noMatch) != itemID {
		t.Fatalf("no match on %d, want %d", f.itemOf(noMatch), itemID)
	}
	got := f.decided(itemID).verdicts
	if got.Values["title"] != "Mine" || got.SourceOf["title"] != kind.SourceAP {
		t.Fatalf("title = %v from %s", got.Values["title"], got.SourceOf["title"])
	}
}

func TestAnAlbumInheritsGenresFromItsMainArtistOneStep(t *testing.T) {
	f := newFixture(t)
	artist := f.put("spotify-artist:a", artistX)
	album := f.put("spotify-album:b", albumByA)
	f.settle(artist, album)

	got := f.decided(f.itemOf(album))
	if !reflect.DeepEqual(got.verdicts.Values["genres"], []any{"x"}) || got.verdicts.SourceOf["genres"] != kind.InheritedSource(kind.LinkBy) {
		t.Fatalf("genres = %v from %s", got.verdicts.Values["genres"], got.verdicts.SourceOf["genres"])
	}
	if _, owned := got.own.Values["genres"]; owned {
		t.Fatalf("own genres = %v", got.own.Values["genres"])
	}
}

func TestChangingTheArtistRederivesTheAlbumsItInherits(t *testing.T) {
	f := newFixture(t)
	artist := f.put("spotify-artist:a", artistX)
	album := f.put("spotify-album:b", albumByA)
	f.settle(artist, album)

	f.put("spotify-artist:a", artistY)
	f.settle(artist)

	if got := f.decided(f.itemOf(album)).verdicts.Values["genres"]; !reflect.DeepEqual(got, []any{"y"}) {
		t.Fatalf("genres = %v", got)
	}
}

func TestATracklistResolvesWhenItsTracksArrive(t *testing.T) {
	f := newFixture(t)
	album := f.put("spotify-album:b", albumTracks)
	f.settle(album)
	albumID := f.itemOf(album)

	track := f.put("spotify-track:ISRC1", trackOne)
	f.settle(track)
	trackID := f.itemOf(track)

	got := f.dump(fmt.Sprintf(`
		SELECT p.item_id, l.position FROM link l
		JOIN item_identity p ON p.source = l.from_source AND p.external_id = l.from_external_id
		JOIN item_identity w ON w.source = l.to_source AND w.external_id = l.to_external_id
		WHERE l.type = '%s' AND w.item_id = %d ORDER BY l.position`, kind.LinkPartOf, albumID))
	if want := fmt.Sprintln(trackID, 1); got != want {
		t.Fatalf("parts = %q, want %q", got, want)
	}
	if f.count("link WHERE type = 'part-of'") != 2 {
		t.Fatalf("part-of links = %d", f.count("link WHERE type = 'part-of'"))
	}
}

func TestALinkOfAnOrderedTypeWithoutAPositionIsRefused(t *testing.T) {
	f := newFixture(t)
	edit := f.put("ap:link/spotify-track-id:t1/part-of/spotify-album:b",
		`{"edit":"added","from":"spotify-track-id:t1","type":"part-of","to":"spotify-album:b"}`)

	_, err := f.world.Settle(f.ctx, f.db, []record.Identity{edit})

	if err == nil || !strings.Contains(err.Error(), kind.LinkPartOf) {
		t.Fatalf("err = %v", err)
	}
}

func settleALittleOfEverything(f *fixture) []record.Identity {
	ids := []record.Identity{
		f.put("steam:220", halfLife2),
		f.put("matched:igdb:1", `{"about":["igdb:1","steam:220"]}`),
		f.put("igdb:1", `{"id":1,"name":"HL2"}`),
		f.put("ap:fields/steam:220", `{"fields":{"playtime_minutes":20}}`),
		f.put("spotify-artist:a", artistX),
		f.put("spotify-album:b", albumTracks),
		f.put("spotify-track:ISRC1", trackOne),
	}
	f.settle(ids...)
	return ids
}

func TestSettlingTheSameClaimsAgainChangesNothing(t *testing.T) {
	f := newFixture(t)
	ids := settleALittleOfEverything(f)
	before := f.snapshot()

	report := f.settle(ids...)

	if !reflect.DeepEqual(report, world.Report{}) {
		t.Fatalf("report = %+v", report)
	}
	if after := f.snapshot(); after != before {
		t.Fatalf("before:\n%s\nafter:\n%s", before, after)
	}
}

func TestRebuildReturnsTheSameItems(t *testing.T) {
	f := newFixture(t)
	settleALittleOfEverything(f)
	before := f.snapshot()
	if _, err := f.db.ExecContext(f.ctx, `DELETE FROM item_verdict; DELETE FROM link`); err != nil {
		t.Fatal(err)
	}

	if _, err := f.world.Rebuild(f.ctx); err != nil {
		t.Fatal(err)
	}

	if after := f.snapshot(); after != before {
		t.Fatalf("before:\n%s\nafter:\n%s", before, after)
	}
	var version string
	if err := f.db.QueryRowContext(f.ctx, `SELECT value FROM setting WHERE key = 'world-version'`).Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != strconv.Itoa(world.DeriveVersion) {
		t.Fatalf("world-version = %s", version)
	}
}

func TestRebuildIfStaleRunsOnlyOnAVersionChange(t *testing.T) {
	f := newFixture(t)
	settleALittleOfEverything(f)
	if _, err := f.db.ExecContext(f.ctx, `DELETE FROM item_verdict`); err != nil {
		t.Fatal(err)
	}

	if _, err := f.world.RebuildIfStale(f.ctx); err != nil {
		t.Fatal(err)
	}
	if f.count("item_verdict") == 0 {
		t.Fatal("the first call did not rebuild")
	}
	if _, err := f.db.ExecContext(f.ctx, `DELETE FROM item_verdict`); err != nil {
		t.Fatal(err)
	}

	report, err := f.world.RebuildIfStale(f.ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(report, world.Report{}) || f.count("item_verdict") != 0 {
		t.Fatalf("the second call rebuilt: report %+v, %d verdicts", report, f.count("item_verdict"))
	}
}

func TestSeedingPutsIdentitiesOnAnExistingItemKeyFirst(t *testing.T) {
	f := newFixture(t)
	result, err := f.db.ExecContext(f.ctx, `INSERT INTO item (kind, slug) VALUES ('game', 'half-life-2')`)
	if err != nil {
		t.Fatal(err)
	}
	itemID, _ := result.LastInsertId()
	steam, igdb := identity(t, "steam:220"), identity(t, "igdb:1")

	if err := f.world.Seed(f.ctx, f.db, itemID, []record.Identity{steam, igdb}); err != nil {
		t.Fatal(err)
	}

	key, err := f.world.KeyOf(f.ctx, f.db, itemID)
	if err != nil || key != steam || f.itemOf(igdb) != itemID {
		t.Fatalf("key = %v (%v), igdb on %d", key, err, f.itemOf(igdb))
	}
	if _, err := f.world.ItemOf(f.ctx, f.db, identity(t, "igdb:404")); err != world.ErrNotFound {
		t.Fatalf("err = %v", err)
	}
}

func TestASuppressedLinkIsDroppedAndNoLongerInherits(t *testing.T) {
	f := newFixture(t)
	artist := f.put("spotify-artist:a", artistX)
	album := f.put("spotify-album:b", albumByA)
	f.settle(artist, album)

	edit := f.put("ap:link/spotify-album:b/by/spotify-artist:a",
		`{"edit":"suppressed","from":"spotify-album:b","type":"by","to":"spotify-artist:a"}`)
	f.settle(edit)

	if n := f.count("link WHERE type = 'by'"); n != 0 {
		t.Fatalf("by links = %d", n)
	}
	if got, inherited := f.decided(f.itemOf(album)).verdicts.Values["genres"]; inherited {
		t.Fatalf("genres = %v", got)
	}
	if f.count("item") != 2 {
		t.Fatalf("items = %d", f.count("item"))
	}
}

func TestARecordWaitingOnItsTrackJoinsWhenTheTrackArrives(t *testing.T) {
	f := newFixture(t)
	listening := f.put("spotify-listening:spotify-track:ISRC1", `{"listening":5}`)
	f.settle(listening)
	if f.count("item") != 0 {
		t.Fatalf("items = %d", f.count("item"))
	}

	track := f.put("spotify-track:ISRC1", trackOne)
	f.settle(track)

	if f.itemOf(listening) != f.itemOf(track) {
		t.Fatalf("listening on %d, track on %d", f.itemOf(listening), f.itemOf(track))
	}
	if f.count("identity_waiting") != 0 {
		t.Fatalf("still waiting: %d", f.count("identity_waiting"))
	}
}

func TestATakeWaitingOnItsFilmJoinsWhenTheFilmArrives(t *testing.T) {
	f := newFixture(t)
	take := f.put("letterboxd-take:letterboxd:abc", `{"rating":4}`)
	f.settle(take)

	film := f.put("letterboxd:abc", `{"id":"abc","title":"Heat"}`)
	f.settle(film)

	if f.itemOf(take) != f.itemOf(film) {
		t.Fatalf("take on %d, film on %d", f.itemOf(take), f.itemOf(film))
	}
	if f.count("item") != 1 || f.count("identity_waiting") != 0 {
		t.Fatalf("items = %d, waiting = %d", f.count("item"), f.count("identity_waiting"))
	}
}

func TestRebuildWithinACallersTransactionGivesTheSameItems(t *testing.T) {
	f := newFixture(t)
	settleALittleOfEverything(f)
	before := f.snapshot()
	if _, err := f.db.ExecContext(f.ctx, `DELETE FROM item_verdict; DELETE FROM link`); err != nil {
		t.Fatal(err)
	}

	tx, err := f.db.BeginTx(f.ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.world.RebuildWithin(f.ctx, tx); err != nil {
		tx.Rollback()
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	if after := f.snapshot(); after != before {
		t.Fatalf("before:\n%s\nafter:\n%s", before, after)
	}
	var version string
	if err := f.db.QueryRowContext(f.ctx, `SELECT value FROM setting WHERE key = 'world-version'`).Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != strconv.Itoa(world.DeriveVersion) {
		t.Fatalf("world-version = %s", version)
	}
}
