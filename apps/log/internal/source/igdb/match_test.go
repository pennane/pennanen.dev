package igdb_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/platform/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
	"github.com/pennane/pennanen.dev/apps/log/internal/rule"
	"github.com/pennane/pennanen.dev/apps/log/internal/source/igdb"
	"github.com/pennane/pennanen.dev/apps/log/internal/source/steam"
	"github.com/pennane/pennanen.dev/apps/log/internal/source/syncs"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/world"
)

func TestTheMatcherWritesAnIGDBClaimAboutTheSteamGame(t *testing.T) {
	ctx := context.Background()
	d, err := db.Open(ctx, filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	records := record.New(d, reading.ReadRecord, reading.ReadersVersion)
	s := store.New(d, records, world.New(d, records), []rule.Proposer{steam.Proposer})
	steamGame := record.Identity{Source: record.SourceSteam, ExternalID: "10"}
	if err := s.PutClaim(ctx, steamGame, json.RawMessage(`{"appid":10,"name":"Hades"}`)); err != nil {
		t.Fatal(err)
	}

	if err := s.ProposeAll(ctx); err != nil {
		t.Fatal(err)
	}
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			fmt.Fprint(w, `{"access_token":"t","expires_in":86400}`)
		case "/external_games":
			fmt.Fprint(w, `[{"uid":"10","game":777}]`)
		case "/games":
			fmt.Fprint(w, `[{"id":777,"name":"Hades"}]`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer fake.Close()
	client := igdb.NewClient("id", "secret")
	client.TokenURL, client.APIURL = fake.URL+"/token", fake.URL
	syncer, err := syncs.New(ctx, igdb.SyncName, record.SourceSteam, igdb.NewMatcher(s, client).Work, s, nil)
	if err != nil {
		t.Fatal(err)
	}
	if report := syncer.Run(ctx); report.Counts[igdb.CountMatched] != 1 || len(report.Problems) != 0 {
		t.Fatalf("report %+v", report)
	}
	claim, err := s.ClaimOf(ctx, steamGame, record.SourceIGDB)
	if err != nil || claim.ExternalID != "777" {
		t.Errorf("igdb claim about the steam game: %+v %v", claim, err)
	}
}
