package record_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
)

const (
	firstTime = "2026-01-01T00:00:00Z"
	laterTime = "2026-02-01T00:00:00Z"
)

var steamGame = record.Identity{Source: "steam", ExternalID: "10"}

func echoReader(source string, raw json.RawMessage) (json.RawMessage, record.Readability) {
	if string(raw) == `{"unreadable":true}` {
		return nil, record.Unreadable
	}
	return raw, record.Readable
}

func openDB(t *testing.T) *sql.DB {
	t.Helper()
	d, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func newStore(t *testing.T) (*record.Store, *sql.DB) {
	d := openDB(t)
	return record.New(d, echoReader, 1), d
}

func mustPut(t *testing.T, s *record.Store, d *sql.DB, id record.Identity, data, at string) record.Change {
	t.Helper()
	change, err := s.PutClaimAt(context.Background(), d, id, json.RawMessage(data), at)
	if err != nil {
		t.Fatal(err)
	}
	return change
}

func mustReadings(t *testing.T, s *record.Store, d *sql.DB, ids ...record.Identity) []record.Reading {
	t.Helper()
	readings, err := s.Readings(context.Background(), d, ids)
	if err != nil {
		t.Fatal(err)
	}
	return readings
}

func count(t *testing.T, d *sql.DB, query string, args ...any) int {
	t.Helper()
	var n int
	if err := d.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestAClaimIsKeyedByIdentityAndReadWhenPut(t *testing.T) {
	s, d := newStore(t)
	if change := mustPut(t, s, d, steamGame, `{"name":"Hades"}`, firstTime); change != record.Added {
		t.Fatalf("change = %s", change)
	}
	readings := mustReadings(t, s, d, steamGame)
	if len(readings) != 1 || readings[0].Values["name"] != "Hades" || readings[0].Identity != steamGame {
		t.Fatalf("readings = %+v", readings)
	}
}

func TestPuttingTheSameClaimAgainChangesNothing(t *testing.T) {
	s, d := newStore(t)
	mustPut(t, s, d, steamGame, `{"name":"Hades"}`, firstTime)
	if change := mustPut(t, s, d, steamGame, `{"name":"Hades"}`, laterTime); change != record.Unchanged {
		t.Fatalf("change = %s", change)
	}
	reading := mustReadings(t, s, d, steamGame)[0]
	if reading.ChangedAt != firstTime || reading.FetchedAt != laterTime {
		t.Fatalf("changed %s fetched %s", reading.ChangedAt, reading.FetchedAt)
	}
}

func TestADifferentReadingMovesChangedAt(t *testing.T) {
	s, d := newStore(t)
	mustPut(t, s, d, steamGame, `{"name":"Hades"}`, firstTime)
	if change := mustPut(t, s, d, steamGame, `{"name":"Hades II"}`, laterTime); change != record.Changed {
		t.Fatalf("change = %s", change)
	}
	reading := mustReadings(t, s, d, steamGame)[0]
	if reading.ChangedAt != laterTime || reading.Values["name"] != "Hades II" {
		t.Fatalf("reading = %+v", reading)
	}
}

func TestUnreadableClaimsCompareByData(t *testing.T) {
	s, d := newStore(t)
	mustPut(t, s, d, steamGame, `{"unreadable":true}`, firstTime)
	if change := mustPut(t, s, d, steamGame, `{"unreadable":true}`, laterTime); change != record.Unchanged {
		t.Fatalf("change = %s", change)
	}
	if got := mustReadings(t, s, d, steamGame); len(got) != 0 {
		t.Fatalf("readings = %+v", got)
	}
	if n := count(t, d, `SELECT COUNT(*) FROM record_reading`); n != 0 {
		t.Fatalf("stored readings = %d", n)
	}
}

func TestRemovingAClaimLeavesAnObservationOfIt(t *testing.T) {
	s, d := newStore(t)
	ctx := context.Background()
	mustPut(t, s, d, steamGame, `{"name":"Hades"}`, firstTime)
	removal, err := s.RemoveClaim(ctx, d, steamGame)
	if err != nil {
		t.Fatal(err)
	}
	if removal != record.Removed {
		t.Fatalf("removal = %s", removal)
	}
	if n := count(t, d, `SELECT COUNT(*) FROM record_claim`) + count(t, d, `SELECT COUNT(*) FROM record_reading`); n != 0 {
		t.Fatalf("claim rows left = %d", n)
	}
	var source, externalID, signal, data string
	err = d.QueryRow(`SELECT about_source, about_external_id, signal, data FROM record_observation`).Scan(&source, &externalID, &signal, &data)
	if err != nil {
		t.Fatal(err)
	}
	if (record.Identity{Source: source, ExternalID: externalID}) != steamGame || signal != record.SignalClaimRemoved || data != `{"name":"Hades"}` {
		t.Fatalf("observation = %s:%s %s %s", source, externalID, signal, data)
	}
}

func TestRemovingAnAbsentClaimIsWasAbsent(t *testing.T) {
	s, d := newStore(t)
	removal, err := s.RemoveClaim(context.Background(), d, steamGame)
	if err != nil {
		t.Fatal(err)
	}
	if removal != record.WasAbsent {
		t.Fatalf("removal = %s", removal)
	}
	if n := count(t, d, `SELECT COUNT(*) FROM record_observation`); n != 0 {
		t.Fatalf("observations = %d", n)
	}
}

func TestObservationsAreAppendedOnce(t *testing.T) {
	s, d := newStore(t)
	ctx := context.Background()
	observation := record.Observation{
		About:      steamGame,
		Source:     "steam-achievements",
		Signal:     "unlocked",
		Key:        "achievement-1",
		ObservedAt: firstTime,
		Data:       json.RawMessage(`{"n":1}`),
	}
	first, err := s.Observe(ctx, d, observation)
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.Observe(ctx, d, observation)
	if err != nil {
		t.Fatal(err)
	}
	if first != record.Observed || second != record.AlreadyKnown {
		t.Fatalf("outcomes = %s, %s", first, second)
	}
	if n := count(t, d, `SELECT COUNT(*) FROM record_observation`); n != 1 {
		t.Fatalf("observations = %d", n)
	}
}

func TestRereadRefreshesOldReadingsWithoutMovingChangedAt(t *testing.T) {
	d := openDB(t)
	ctx := context.Background()
	oldStore := record.New(d, echoReader, 1)
	mustPut(t, oldStore, d, steamGame, `{"name":"Hades"}`, firstTime)

	upgraded := func(source string, raw json.RawMessage) (json.RawMessage, record.Readability) {
		return json.RawMessage(`{"name":"Hades","upgraded":true}`), record.Readable
	}
	newStore := record.New(d, upgraded, 2)
	reread, err := newStore.Reread(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if reread != 1 {
		t.Fatalf("reread = %d", reread)
	}
	reading := mustReadings(t, newStore, d, steamGame)[0]
	if reading.ChangedAt != firstTime || reading.Values["upgraded"] != true {
		t.Fatalf("reading = %+v", reading)
	}
	again, err := newStore.Reread(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if again != 0 {
		t.Fatalf("second reread = %d", again)
	}
}

func TestRereadWalksPastBatches(t *testing.T) {
	s, d := newStore(t)
	const claims = 450
	for i := 0; i < claims; i++ {
		mustPut(t, s, d, record.Identity{Source: "steam", ExternalID: fmt.Sprint(i)}, `{"name":"x"}`, firstTime)
	}
	if _, err := d.Exec(`DELETE FROM record_reading`); err != nil {
		t.Fatal(err)
	}
	reread, err := s.Reread(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if reread != claims {
		t.Fatalf("reread = %d", reread)
	}
	if n := count(t, d, `SELECT COUNT(*) FROM record_reading`); n != claims {
		t.Fatalf("readings = %d", n)
	}
}

func TestParseIdentitySplitsAtTheFirstColon(t *testing.T) {
	id, err := record.ParseIdentity("ap:fields/igdb:1942")
	if err != nil {
		t.Fatal(err)
	}
	if id != (record.Identity{Source: "ap", ExternalID: "fields/igdb:1942"}) || id.String() != "ap:fields/igdb:1942" {
		t.Fatalf("id = %+v", id)
	}
	if _, err := record.ParseIdentity("nocolon"); err == nil {
		t.Fatal("expected an error")
	}
}
