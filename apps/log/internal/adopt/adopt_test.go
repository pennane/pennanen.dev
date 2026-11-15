package adopt_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"math"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/adopt"
	"github.com/pennane/pennanen.dev/apps/log/internal/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
	"github.com/pennane/pennanen.dev/apps/log/internal/world"
)

const (
	lastBeforeLayers = 18
	everyMigration   = math.MaxInt
)

type fixture struct {
	t       *testing.T
	ctx     context.Context
	db      *sql.DB
	records *record.Store
	world   *world.World
}

func legacy(t *testing.T, rows string) *fixture {
	t.Helper()
	ctx := context.Background()
	d, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "log.db")+"?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatal(err)
	}
	d.SetMaxOpenConns(1)
	t.Cleanup(func() { d.Close() })
	if err := db.MigrateUpTo(ctx, d, lastBeforeLayers); err != nil {
		t.Fatal(err)
	}
	if _, err := d.ExecContext(ctx, rows); err != nil {
		t.Fatal(err)
	}
	if err := db.MigrateUpTo(ctx, d, everyMigration); err != nil {
		t.Fatal(err)
	}
	records := record.New(d, verdict.ReadRecord, verdict.ReadersVersion)
	return &fixture{t: t, ctx: ctx, db: d, records: records, world: world.New(d, records)}
}

func (f *fixture) adopt() adopt.Adoption {
	f.t.Helper()
	adoption, err := adopt.Run(f.ctx, f.db, f.records, f.world)
	if err != nil {
		f.t.Fatal(err)
	}
	return adoption
}

func (f *fixture) count(query string, args ...any) int {
	f.t.Helper()
	var n int
	if err := f.db.QueryRowContext(f.ctx, query, args...).Scan(&n); err != nil {
		f.t.Fatal(err)
	}
	return n
}

func (f *fixture) strings(query string, args ...any) []string {
	f.t.Helper()
	rows, err := f.db.QueryContext(f.ctx, query, args...)
	if err != nil {
		f.t.Fatal(err)
	}
	defer rows.Close()
	got := []string{}
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			f.t.Fatal(err)
		}
		got = append(got, s)
	}
	if err := rows.Err(); err != nil {
		f.t.Fatal(err)
	}
	return got
}

func (f *fixture) data(id string) map[string]any {
	f.t.Helper()
	source, externalID, _ := strings.Cut(id, ":")
	var raw string
	if err := f.db.QueryRowContext(f.ctx, `SELECT data FROM record_claim WHERE source = ? AND external_id = ?`, source, externalID).Scan(&raw); err != nil {
		f.t.Fatalf("record %s: %v", id, err)
	}
	var values map[string]any
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		f.t.Fatal(err)
	}
	return values
}

func (f *fixture) itemOf(id string) int64 {
	f.t.Helper()
	parsed, err := record.ParseIdentity(id)
	if err != nil {
		f.t.Fatal(err)
	}
	itemID, err := f.world.ItemOf(f.ctx, f.db, parsed)
	if err != nil {
		f.t.Fatalf("item of %s: %v", id, err)
	}
	return itemID
}

func (f *fixture) keyOf(itemID int64) string {
	f.t.Helper()
	key, err := f.world.KeyOf(f.ctx, f.db, itemID)
	if err != nil {
		f.t.Fatal(err)
	}
	return key.String()
}

func (f *fixture) tableExists(name string) bool {
	f.t.Helper()
	return f.count(`SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = ?`, name) > 0
}

func sameJSON(t *testing.T, got map[string]any, want string) {
	t.Helper()
	var wanted map[string]any
	if err := json.Unmarshal([]byte(want), &wanted); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, wanted) {
		encoded, _ := json.Marshal(got)
		t.Errorf("data = %s, want %s", encoded, want)
	}
}

func TestEveryClaimBecomesARecordOnTheSameItem(t *testing.T) {
	f := legacy(t, `
		INSERT INTO item (id, kind, slug) VALUES (7, 'game', 'half-life-2');
		INSERT INTO entry (item_id, status) VALUES (7, 'finished');
		INSERT INTO claim (item_id, source, external_id, data, fetched_at) VALUES
			(7, 'steam', '220', '{"appid":220,"name":"Half-Life 2","playtime_forever":900}', '2025-01-01T00:00:00Z'),
			(7, 'steam-store', '220', '{"name":"Half-Life 2"}', '2025-02-01T00:00:00Z'),
			(7, 'igdb', '1', '{"id":1,"name":"Half-Life 2"}', '2025-03-01T00:00:00Z');`)
	adoption := f.adopt()
	if adoption.Outcome != adopt.Adopted {
		t.Errorf("outcome = %s", adoption.Outcome)
	}
	got := f.strings(`SELECT source || ':' || external_id || ' ' || fetched_at || ' ' || changed_at || ' ' || data FROM record_claim
		WHERE source IN ('steam', 'steam-store', 'igdb') ORDER BY source`)
	want := []string{
		`igdb:1 2025-03-01T00:00:00Z 2025-03-01T00:00:00Z {"id":1,"name":"Half-Life 2"}`,
		`steam:220 2025-01-01T00:00:00Z 2025-01-01T00:00:00Z {"appid":220,"name":"Half-Life 2","playtime_forever":900}`,
		`steam-store:220 2025-02-01T00:00:00Z 2025-02-01T00:00:00Z {"name":"Half-Life 2"}`,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("records = %q, want %q", got, want)
	}
	for _, id := range []string{"steam:220", "steam-store:220", "igdb:1"} {
		if itemID := f.itemOf(id); itemID != 7 {
			t.Errorf("%s is on item %d, want 7", id, itemID)
		}
	}
	if key := f.keyOf(7); key != "igdb:1" {
		t.Errorf("key = %s, want igdb:1", key)
	}
	if n := f.count(`SELECT count(*) FROM claim WHERE item_id = 7 AND source IN ('steam', 'steam-store', 'igdb')`); n != 3 {
		t.Errorf("claim view shows %d claims on item 7", n)
	}
}

func TestItemsKeepIdsSlugsEntriesScoresAndTags(t *testing.T) {
	f := legacy(t, `
		INSERT INTO item (id, kind, slug, created_at) VALUES
			(3, 'game', 'portal', '2024-01-01T00:00:00Z'),
			(9, 'artist', 'pendulum', '2024-02-01T00:00:00Z');
		INSERT INTO entry (item_id, status, visibility, blurb, prose, status_set_by, time_spent) VALUES
			(3, 'finished', 'public', 'short', 'long', 'ap', 'just-right'),
			(9, 'listening', 'private', '', '', 'spotify-artist', NULL);
		INSERT INTO score (item_id, axis, division, position) VALUES (3, 'prestige', 's', 0);
		INSERT INTO tag (item_id, tag, added_by) VALUES (3, 'favorite', 'ap'), (9, 'comfort', 'ap');
		INSERT INTO claim (item_id, source, external_id, data, fetched_at) VALUES
			(3, 'steam', '400', '{"appid":400,"name":"Portal"}', '2025-01-01T00:00:00Z'),
			(9, 'spotify-artist', 'P1', '{"id":"P1","name":"Pendulum"}', '2025-01-01T00:00:00Z');`)
	f.adopt()
	items := f.strings(`SELECT id || ' ' || kind || ' ' || slug || ' ' || created_at FROM item ORDER BY id`)
	wantItems := []string{"3 game portal 2024-01-01T00:00:00Z", "9 artist pendulum 2024-02-01T00:00:00Z"}
	if !reflect.DeepEqual(items, wantItems) {
		t.Errorf("items = %q, want %q", items, wantItems)
	}
	entries := f.strings(`SELECT item_id || ' ' || status || ' ' || visibility || ' ' || blurb || ' ' || prose || ' ' || status_set_by || ' ' || coalesce(time_spent, '-') FROM entry ORDER BY item_id`)
	wantEntries := []string{"3 finished public short long ap just-right", "9 listening private   spotify-artist -"}
	if !reflect.DeepEqual(entries, wantEntries) {
		t.Errorf("entries = %q, want %q", entries, wantEntries)
	}
	if scores := f.strings(`SELECT item_id || ' ' || axis || ' ' || division || ' ' || position FROM score`); !reflect.DeepEqual(scores, []string{"3 prestige s 0"}) {
		t.Errorf("scores = %q", scores)
	}
	if tags := f.strings(`SELECT item_id || ' ' || tag FROM tag ORDER BY item_id`); !reflect.DeepEqual(tags, []string{"3 favorite", "9 comfort"}) {
		t.Errorf("tags = %q", tags)
	}
	if itemID := f.itemOf("steam:400"); itemID != 3 {
		t.Errorf("steam:400 on item %d", itemID)
	}
	if itemID := f.itemOf("spotify-artist:P1"); itemID != 9 {
		t.Errorf("spotify-artist:P1 on item %d", itemID)
	}
}

func TestAHandMadeItemBecomesAnApItemRecord(t *testing.T) {
	f := legacy(t, `
		INSERT INTO item (id, kind, slug) VALUES (5, 'movie', 'home-video');
		INSERT INTO entry (item_id, status) VALUES (5, 'watched');
		INSERT INTO claim (item_id, source, external_id, data, fetched_at) VALUES
			(5, 'ap', NULL, '{"title":"Home video","released":"1999"}', '2025-01-01T00:00:00Z');`)
	f.adopt()
	sameJSON(t, f.data("ap:item/5"), `{"kind":"movie","fields":{"title":"Home video","released":"1999"}}`)
	if itemID := f.itemOf("ap:item/5"); itemID != 5 {
		t.Errorf("ap:item/5 on item %d", itemID)
	}
	if key := f.keyOf(5); key != "ap:item/5" {
		t.Errorf("key = %s", key)
	}
	if n := f.count(`SELECT count(*) FROM item_verdict WHERE item_id = 5 AND json_extract(own, '$.values.title') = 'Home video'`); n != 1 {
		t.Errorf("item 5 has no verdict titled Home video")
	}
	if n := f.count(`SELECT count(*) FROM entry WHERE item_id = 5`); n != 1 {
		t.Errorf("entry of item 5 lost")
	}
}

func TestAnOverrideBecomesAnApFieldsRecord(t *testing.T) {
	f := legacy(t, `
		INSERT INTO item (id, kind, slug) VALUES (7, 'game', 'half-life-2');
		INSERT INTO claim (item_id, source, external_id, data, fetched_at) VALUES
			(7, 'steam', '220', '{"appid":220,"name":"Half-Life 2"}', '2025-01-01T00:00:00Z'),
			(7, 'ap', NULL, '{"title":"HL2"}', '2025-02-01T00:00:00Z');`)
	f.adopt()
	sameJSON(t, f.data("ap:fields/steam:220"), `{"fields":{"title":"HL2"}}`)
	if itemID := f.itemOf("ap:fields/steam:220"); itemID != 7 {
		t.Errorf("override on item %d", itemID)
	}
	if n := f.count(`SELECT count(*) FROM record_claim WHERE source = 'ap' AND external_id = 'fields/steam:220' AND changed_at = '2025-02-01T00:00:00Z'`); n != 1 {
		t.Errorf("override did not keep its time")
	}
	if n := f.count(`SELECT count(*) FROM item_verdict WHERE item_id = 7 AND json_extract(own, '$.values.title') = 'HL2'`); n != 1 {
		t.Errorf("override does not win the title")
	}
}

func TestNoMatchBecomesAnApMatchAnswer(t *testing.T) {
	f := legacy(t, `
		INSERT INTO item (id, kind, slug) VALUES (7, 'game', 'half-life-2');
		INSERT INTO claim (item_id, source, external_id, data, fetched_at) VALUES
			(7, 'steam', '220', '{"appid":220,"name":"Half-Life 2"}', '2025-01-01T00:00:00Z'),
			(7, 'igdb', NULL, '{"match":"none"}', '2025-02-01T00:00:00Z');`)
	f.adopt()
	sameJSON(t, f.data("ap:match/igdb/steam:220"), `{"to":null}`)
	if itemID := f.itemOf("ap:match/igdb/steam:220"); itemID != 7 {
		t.Errorf("match answer on item %d", itemID)
	}
	if n := f.count(`SELECT count(*) FROM record_claim WHERE source = 'igdb'`); n != 0 {
		t.Errorf("no match left %d igdb records", n)
	}
}

func TestClaimsWithoutAnExternalIdTakeTheItemsKey(t *testing.T) {
	f := legacy(t, `
		INSERT INTO item (id, kind, slug) VALUES (4, 'movie', 'heat');
		INSERT INTO claim (item_id, source, external_id, data, fetched_at) VALUES
			(4, 'letterboxd', 'abc', '{"id":"abc","title":"Heat"}', '2025-01-01T00:00:00Z'),
			(4, 'letterboxd-take', NULL, '{"rating":4.5}', '2025-02-01T00:00:00Z');`)
	f.adopt()
	sameJSON(t, f.data("letterboxd-take:letterboxd:abc"), `{"rating":4.5}`)
	if itemID := f.itemOf("letterboxd-take:letterboxd:abc"); itemID != 4 {
		t.Errorf("take on item %d", itemID)
	}
}

func TestJoinsNoDeriveExplainsBecomeALegacyRecord(t *testing.T) {
	f := legacy(t, `
		INSERT INTO item (id, kind, slug) VALUES (7, 'game', 'half-life-2'), (8, 'game', 'portal');
		INSERT INTO claim (item_id, source, external_id, data, fetched_at) VALUES
			(7, 'steam', '220', '{"appid":220,"name":"Half-Life 2"}', '2025-01-01T00:00:00Z'),
			(7, 'igdb', '1', '{"id":1,"name":"Half-Life 2"}', '2025-02-01T00:00:00Z'),
			(8, 'steam', '400', '{"appid":400,"name":"Portal"}', '2025-01-01T00:00:00Z'),
			(8, 'steam-store', '400', '{"name":"Portal"}', '2025-01-01T00:00:00Z');`)
	f.adopt()
	about, _ := f.data("legacy:item/7")["about"].([]any)
	if !reflect.DeepEqual(about, []any{"igdb:1", "steam:220"}) {
		t.Errorf("legacy:item/7 about %v", about)
	}
	if n := f.count(`SELECT count(*) FROM record_claim WHERE source = 'legacy' AND external_id = 'item/8'`); n != 0 {
		t.Errorf("item 8's derives already join it, yet it got a legacy record")
	}
	for _, table := range []string{"link", "item_verdict", "item_overlap", "identity_waiting", "item_identity"} {
		if _, err := f.db.ExecContext(f.ctx, `DELETE FROM `+table); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := f.world.Rebuild(f.ctx); err != nil {
		t.Fatal(err)
	}
	if f.itemOf("steam:220") != f.itemOf("igdb:1") {
		t.Errorf("a rebuild from scratch split steam:220 and igdb:1")
	}
}

func TestObservationsMoveToTheItemsKey(t *testing.T) {
	f := legacy(t, `
		INSERT INTO item (id, kind, slug) VALUES (9, 'artist', 'pendulum');
		INSERT INTO claim (item_id, source, external_id, data, fetched_at) VALUES
			(9, 'spotify-artist', 'P1', '{"id":"P1","name":"Pendulum"}', '2025-01-01T00:00:00Z');
		INSERT INTO observation (item_id, source, signal, key, observed_at, data) VALUES
			(9, 'spotify', 'top-artist', 'top-artist:short_term:2026-03-14:P1', '2026-03-14T00:00:00Z', '{"rank":3,"of":50}');`)
	f.adopt()
	got := f.strings(`SELECT about_source || ':' || about_external_id || ' ' || source || ' ' || signal || ' ' || key || ' ' || observed_at || ' ' || data FROM record_observation`)
	want := []string{`spotify-artist:P1 spotify top-artist top-artist:short_term:2026-03-14:P1 2026-03-14T00:00:00Z {"rank":3,"of":50}`}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("observations = %q, want %q", got, want)
	}
	if n := f.count(`SELECT count(*) FROM observation WHERE item_id = 9`); n != 1 {
		t.Errorf("observation view shows %d on item 9", n)
	}
}

func TestRelationsAndPartsAreKept(t *testing.T) {
	f := legacy(t, `
		INSERT INTO item (id, kind, slug) VALUES
			(1, 'album', 'hold-your-colour'), (2, 'artist', 'pendulum'), (3, 'track', 'slam'),
			(4, 'album', 'immersion'), (5, 'track', 'watercolour');
		INSERT INTO claim (item_id, source, external_id, data, fetched_at) VALUES
			(1, 'spotify-album', 'AL', '{"id":"AL","name":"Hold Your Colour"}', '2025-01-01T00:00:00Z'),
			(2, 'spotify-artist', 'P1', '{"id":"P1","name":"Pendulum"}', '2025-01-01T00:00:00Z'),
			(3, 'spotify-track', 'T1', '{"id":"T1","name":"Slam"}', '2025-01-01T00:00:00Z'),
			(4, 'spotify-album', 'IM', '{"id":"IM","name":"Immersion","artists":[{"id":"P1","name":"Pendulum"}],"tracks":{"items":[{"id":"T2","disc_number":1,"track_number":1}]}}', '2025-01-01T00:00:00Z'),
			(5, 'spotify-track', 'T2', '{"id":"T2","name":"Watercolour"}', '2025-01-01T00:00:00Z');
		INSERT INTO relation (from_id, type, to_id) VALUES (1, 'by', 2), (4, 'by', 2);
		INSERT INTO part (whole_id, position, part_id) VALUES (1, 4, 3), (4, 1, 5);`)
	f.adopt()
	if relations := f.strings(`SELECT from_id || ' ' || type || ' ' || to_id FROM relation ORDER BY from_id`); !reflect.DeepEqual(relations, []string{"1 by 2", "4 by 2"}) {
		t.Errorf("relations = %q", relations)
	}
	if parts := f.strings(`SELECT whole_id || ' ' || position || ' ' || part_id FROM part ORDER BY whole_id`); !reflect.DeepEqual(parts, []string{"1 4 3", "4 1 5"}) {
		t.Errorf("parts = %q", parts)
	}
	links, _ := f.data("legacy:item/1")["links"].([]any)
	if len(links) != 1 {
		t.Errorf("legacy:item/1 links = %v, want the by relation", links)
	}
	partLinks, _ := f.data("legacy:item/3")["links"].([]any)
	if len(partLinks) != 1 {
		t.Errorf("legacy:item/3 links = %v, want the part", partLinks)
	}
	if n := f.count(`SELECT count(*) FROM record_claim WHERE source = 'legacy' AND external_id IN ('item/4', 'item/5')`); n != 0 {
		t.Errorf("relations the album claim derives got %d legacy records", n)
	}
}

func TestAdoptingTwiceDoesNothing(t *testing.T) {
	f := legacy(t, `
		INSERT INTO item (id, kind, slug) VALUES (7, 'game', 'half-life-2');
		INSERT INTO claim (item_id, source, external_id, data, fetched_at) VALUES
			(7, 'steam', '220', '{"appid":220,"name":"Half-Life 2"}', '2025-01-01T00:00:00Z');`)
	if first := f.adopt(); first.Outcome != adopt.Adopted {
		t.Errorf("first outcome = %s", first.Outcome)
	}
	for _, gone := range []string{"legacy_claim", "legacy_observation", "legacy_relation", "legacy_part", "moment_archive", "moment_ref_archive", "number_score_archive"} {
		if f.tableExists(gone) {
			t.Errorf("%s survived adoption", gone)
		}
	}
	records := f.count(`SELECT count(*) FROM record_claim`)
	if second := f.adopt(); second.Outcome != adopt.NothingToAdopt {
		t.Errorf("second outcome = %s", second.Outcome)
	}
	if again := f.count(`SELECT count(*) FROM record_claim`); again != records {
		t.Errorf("records went from %d to %d", records, again)
	}
}

func TestAFailedAdoptLeavesLegacyTablesInPlace(t *testing.T) {
	f := legacy(t, `
		INSERT INTO item (id, kind, slug) VALUES (7, 'game', 'half-life-2');
		INSERT INTO claim (item_id, source, external_id, data, fetched_at) VALUES
			(7, 'steam', '220', '{"appid":220,"name":"Half-Life 2"}', '2025-01-01T00:00:00Z');`)
	if _, err := f.db.ExecContext(f.ctx, `PRAGMA foreign_keys = OFF`); err != nil {
		t.Fatal(err)
	}
	if _, err := f.db.ExecContext(f.ctx, `INSERT INTO legacy_claim (item_id, source, external_id, data, fetched_at) VALUES (99, 'steam', '400', '{"appid":400}', '2025-01-01T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if _, err := f.db.ExecContext(f.ctx, `PRAGMA foreign_keys = ON`); err != nil {
		t.Fatal(err)
	}
	_, err := adopt.Run(f.ctx, f.db, f.records, f.world)
	if err == nil || !strings.Contains(err.Error(), "adopted 1 of 2 legacy claims") {
		t.Fatalf("err = %v, want one of two legacy claims adopted", err)
	}
	if n := f.count(`SELECT count(*) FROM legacy_claim`); n != 2 {
		t.Errorf("legacy_claim has %d rows, want 2", n)
	}
	if n := f.count(`SELECT count(*) FROM record_claim`); n != 0 {
		t.Errorf("record_claim has %d rows, want 0", n)
	}
	if n := f.count(`SELECT count(*) FROM item_identity`); n != 0 {
		t.Errorf("item_identity has %d rows, want 0", n)
	}
}

func TestAnItemWithoutAnyExternalIdKeepsItsIdKindAndEntry(t *testing.T) {
	f := legacy(t, `
		INSERT INTO item (id, kind, slug) VALUES (6, 'movie', 'lost-tape');
		INSERT INTO entry (item_id, status) VALUES (6, 'watched');
		INSERT INTO claim (item_id, source, external_id, data, fetched_at) VALUES
			(6, 'tmdb', NULL, '{"match":"none"}', '2025-01-01T00:00:00Z'),
			(6, 'letterboxd-take', NULL, '{"rating":3}', '2025-02-01T00:00:00Z');`)
	f.adopt()
	if key := f.keyOf(6); key != "ap:item/6" {
		t.Errorf("key = %s", key)
	}
	sameJSON(t, f.data("ap:item/6"), `{"kind":"movie","fields":{}}`)
	for _, id := range []string{"ap:match/tmdb/ap:item/6", "letterboxd-take:ap:item/6"} {
		if itemID := f.itemOf(id); itemID != 6 {
			t.Errorf("%s on item %d", id, itemID)
		}
	}
	if kinds := f.strings(`SELECT kind || ' ' || slug FROM item WHERE id = 6`); !reflect.DeepEqual(kinds, []string{"movie lost-tape"}) {
		t.Errorf("item 6 = %q", kinds)
	}
	if n := f.count(`SELECT count(*) FROM entry WHERE item_id = 6 AND status = 'watched'`); n != 1 {
		t.Errorf("entry of item 6 lost")
	}
}
