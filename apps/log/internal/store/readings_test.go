package store_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/platform/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

func openStoreWithDB(t *testing.T) (*store.Store, *sql.DB) {
	t.Helper()
	d, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	return store.Open(d), d
}

const matrix = `{"id":603,"title":"The Matrix","release_date":"1999-03-30"}`

func loggedMovie(t *testing.T, s *store.Store, source, data string) int64 {
	t.Helper()
	ctx := context.Background()
	id, err := s.CreateItemFromSource(ctx, kind.Movie, "The Matrix", source, "603", json.RawMessage(data))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ChangeEntry(ctx, id, store.EntryChange{Status: ptr(kind.Watched)}); err != nil {
		t.Fatal(err)
	}
	return id
}

func readingVersions(t *testing.T, d *sql.DB, itemID int64) map[string]int {
	t.Helper()
	rows, err := d.Query(`
		SELECT r.source, r.version FROM record_reading r
		JOIN item_identity i ON i.source = r.source AND i.external_id = r.external_id
		WHERE i.item_id = ?`, itemID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	versions := map[string]int{}
	for rows.Next() {
		var source string
		var version int
		if err := rows.Scan(&source, &version); err != nil {
			t.Fatal(err)
		}
		versions[source] = version
	}
	return versions
}

func TestAClaimIsReadWhenWritten(t *testing.T) {
	s, d := openStoreWithDB(t)
	id := loggedMovie(t, s, record.SourceTMDB, matrix)
	if got := readingVersions(t, d, id); got[record.SourceTMDB] != reading.ReadersVersion {
		t.Errorf("readings = %v, want tmdb at version %d", got, reading.ReadersVersion)
	}
}

func TestListsDecideFromReadingsAlone(t *testing.T) {
	ctx := context.Background()
	s, d := openStoreWithDB(t)
	id := loggedMovie(t, s, record.SourceTMDB, matrix)
	if _, err := d.Exec(`UPDATE record_claim SET data = '{}' WHERE source = ? AND external_id = ?`, record.SourceTMDB, "603"); err != nil {
		t.Fatal(err)
	}
	views, err := s.ListItems(ctx, store.ItemFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(views) != 1 || views[0].Verdicts.Values["title"] != "The Matrix" || views[0].Verdicts.Values["released"] != "1999-03-30" {
		t.Fatalf("list = %+v", views)
	}
	one, err := s.GetItem(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if len(one.Claims) != 1 || string(one.Claims[0].Data) != "{}" {
		t.Errorf("claims = %+v", one.Claims)
	}
}

func TestRereadFillsMissingAndOldReadings(t *testing.T) {
	ctx := context.Background()
	s, d := openStoreWithDB(t)
	id := loggedMovie(t, s, record.SourceTMDB, matrix)
	if err := s.PutClaim(ctx, id, record.SourceLetterboxd, "abc", json.RawMessage(`{"id":"abc","title":"The Matrix","year":1999}`)); err != nil {
		t.Fatal(err)
	}
	if err := s.ChangeOverride(ctx, id, map[string]any{"title": "Matrix"}); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Exec(`DELETE FROM record_reading WHERE source = ? AND external_id = ?`, record.SourceTMDB, "603"); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Exec(`UPDATE record_reading SET version = 0 WHERE source = ? AND external_id = ?`, record.SourceLetterboxd, "abc"); err != nil {
		t.Fatal(err)
	}
	reread, err := s.RereadClaims(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if reread != 2 {
		t.Errorf("re-read %d claims, want 2", reread)
	}
	got := readingVersions(t, d, id)
	for _, source := range []string{record.SourceTMDB, record.SourceLetterboxd, record.SourceAP} {
		if got[source] != reading.ReadersVersion {
			t.Errorf("%s reading version = %d, want %d", source, got[source], reading.ReadersVersion)
		}
	}
	again, err := s.RereadClaims(ctx)
	if err != nil || again != 0 {
		t.Errorf("second re-read = %d, %v; want 0", again, err)
	}
}

func TestRereadWalksPastBatches(t *testing.T) {
	ctx := context.Background()
	s, d := openStoreWithDB(t)
	for range 450 {
		if _, err := s.CreateItem(ctx, kind.Movie, "Film", map[string]any{"title": "Film"}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := d.Exec(`DELETE FROM record_reading`); err != nil {
		t.Fatal(err)
	}
	reread, err := s.RereadClaims(ctx)
	if err != nil || reread != 450 {
		t.Fatalf("re-read %d, %v; want 450", reread, err)
	}
	var stored int
	if err := d.QueryRow(`SELECT count(*) FROM record_reading`).Scan(&stored); err != nil || stored != 450 {
		t.Errorf("stored %d readings, %v", stored, err)
	}
}

func TestAnUnreadableClaimHasNoReading(t *testing.T) {
	ctx := context.Background()
	s, d := openStoreWithDB(t)
	id := loggedMovie(t, s, record.SourceTMDB, matrix)
	if err := s.PutClaim(ctx, id, record.SourceTMDB, "603", json.RawMessage(`[1]`)); err != nil {
		t.Fatal(err)
	}
	if got := readingVersions(t, d, id); len(got) != 0 {
		t.Errorf("readings = %v, want none", got)
	}
	views, err := s.ListItems(ctx, store.ItemFilter{})
	if err != nil || len(views) != 1 {
		t.Fatalf("list = %v, %v", views, err)
	}
	if _, titled := views[0].Verdicts.Values["title"]; titled {
		t.Errorf("values = %v, want none from the unreadable claim", views[0].Verdicts.Values)
	}
}

func TestRemovingAClaimRemovesItsReading(t *testing.T) {
	ctx := context.Background()
	s, d := openStoreWithDB(t)
	id := loggedMovie(t, s, record.SourceTMDB, matrix)
	identity := store.Identity{Kind: kind.Series, Source: record.SourceTMDBSeries, ExternalID: "94028", Data: json.RawMessage(`{"id":94028,"name":"Ripley"}`)}
	if err := s.Identify(ctx, id, identity, []string{record.SourceTMDB, record.SourceTMDBSeries}); err != nil {
		t.Fatal(err)
	}
	got := readingVersions(t, d, id)
	if _, stale := got[record.SourceTMDB]; stale || got[record.SourceTMDBSeries] != reading.ReadersVersion {
		t.Errorf("readings = %v, want only tmdb-series", got)
	}
}

func TestListedTakesKeepTheirClaimsFetchTime(t *testing.T) {
	ctx := context.Background()
	s, d := openStoreWithDB(t)
	id := loggedMovie(t, s, record.SourceTMDB, matrix)
	if err := s.PutClaim(ctx, id, record.SourceLetterboxdTake, "matrix", json.RawMessage(`{"rating":4.5}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Exec(`UPDATE record_claim SET fetched_at = '2020-01-02T03:04:05Z' WHERE source = ?`, record.SourceLetterboxdTake); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Exec(`DELETE FROM record_reading`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RereadClaims(ctx); err != nil {
		t.Fatal(err)
	}
	views, err := s.ListItems(ctx, store.ItemFilter{})
	if err != nil || len(views) != 1 || len(views[0].Takes) != 1 {
		t.Fatalf("list = %+v, %v", views, err)
	}
	if got := views[0].Takes[0].FetchedAt; got != "2020-01-02T03:04:05Z" {
		t.Errorf("take fetchedAt = %q", got)
	}
}
