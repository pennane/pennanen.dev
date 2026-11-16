package tmdb_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/pennane/pennanen.dev/apps/log/internal/platform/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
	"github.com/pennane/pennanen.dev/apps/log/internal/rule"
	"github.com/pennane/pennanen.dev/apps/log/internal/source/letterboxd"
	"github.com/pennane/pennanen.dev/apps/log/internal/source/syncs"
	"github.com/pennane/pennanen.dev/apps/log/internal/source/tmdb"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/world"
)

func TestTheMatcherWritesATMDBClaimAboutTheLetterboxdFilm(t *testing.T) {
	ctx := context.Background()
	d, err := db.Open(ctx, filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	records := record.New(d, reading.ReadRecord, reading.ReadersVersion)
	s := store.New(d, records, world.New(d, records), []rule.Proposer{letterboxd.Proposer})
	film := record.Identity{Source: record.SourceLetterboxd, ExternalID: "aaa1"}
	if err := s.PutClaim(ctx, film, json.RawMessage(`{"id":"aaa1","title":"Ripley","year":2024,"tmdb":{"type":"tv","id":94028}}`)); err != nil {
		t.Fatal(err)
	}
	watchlisted := record.Identity{Source: record.SourceLetterboxdTake, ExternalID: film.String()}
	if err := s.PutClaimAbout(ctx, watchlisted, film, json.RawMessage(`{"onWatchlistSince":"2026-01-01"}`)); err != nil {
		t.Fatal(err)
	}
	if err := s.ProposeAll(ctx); err != nil {
		t.Fatal(err)
	}
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/3/movie/1":
			fmt.Fprint(w, `{"id":1,"title":"Ripley","release_date":"2024-04-04"}`)
		case "/3/search/movie":
			fmt.Fprint(w, `{"results":[{"id":1,"title":"Ripley","release_date":"2024-04-04"}]}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer fake.Close()
	client := tmdb.NewClient("t")
	client.BaseURL = fake.URL
	matcher := tmdb.NewMatcher(s, client, nil)
	matcher.RequestInterval = time.Millisecond
	syncer, err := syncs.New(ctx, tmdb.SyncName, record.SourceLetterboxd, matcher.Work, s, nil)
	if err != nil {
		t.Fatal(err)
	}
	if report := syncer.Run(ctx); report.Counts[tmdb.CountMatched] != 1 || len(report.Problems) != 0 {
		t.Fatalf("report %+v", report)
	}
	claim, err := s.ClaimOf(ctx, film, record.SourceTMDB)
	if err != nil || claim.ExternalID != "1" {
		t.Errorf("tmdb claim about the film: %+v %v", claim, err)
	}
}
