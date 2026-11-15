package steam_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/platform/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
	"github.com/pennane/pennanen.dev/apps/log/internal/rule"
	"github.com/pennane/pennanen.dev/apps/log/internal/source/steam"
	"github.com/pennane/pennanen.dev/apps/log/internal/source/syncs"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/take"
	"github.com/pennane/pennanen.dev/apps/log/internal/world"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

func TestSyncImportsHintsAndOnlyRefetchesChangedAchievements(t *testing.T) {
	ctx := context.Background()
	d, err := db.Open(ctx, filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	s := openProposing(t, d)

	playtime := atomic.Int64{}
	playtime.Store(150)
	achievementCalls := atomic.Int64{}
	storeCalls := atomic.Int64{}
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/IPlayerService/GetOwnedGames/v1/":
			fmt.Fprintf(w, `{"response":{"games":[
				{"appid":1,"name":"Card Farm","playtime_forever":%d},
				{"appid":2,"name":"Never Touched","playtime_forever":0}]}}`, playtime.Load())
		case "/api/appdetails":
			storeCalls.Add(1)
			id := r.URL.Query().Get("appids")
			fmt.Fprintf(w, `{"%s":{"success":true,"data":{"name":"Store %s","categories":[{"description":"Steam Trading Cards"}],"release_date":{"date":"4 Dec, 2020"}}}}`, id, id)
		case "/ISteamUserStats/GetPlayerAchievements/v1/":
			achievementCalls.Add(1)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"playerstats":{"error":"Requested app has no stats","success":false}}`)
		case "/profiles/me/recommended/":
			if r.URL.Query().Get("p") == "1" {
				fmt.Fprint(w, `<div class="review_box"><a href="https://steamcommunity.com/app/1"></a><img src="/icon_thumbsDown.png"><div class="content ">farms cards</div></div>`)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer fake.Close()

	client := steam.NewClient("key", "me")
	client.WebAPIBaseURL, client.StoreBaseURL, client.CommunityBaseURL = fake.URL, fake.URL, fake.URL
	steamSyncer := steam.NewSyncer(s, client)
	steamSyncer.StoreRequestInterval = 1
	syncer, err := syncs.New(ctx, steam.SyncName, record.SourceSteam, steamSyncer.Work, s, nil)
	if err != nil {
		t.Fatal(err)
	}

	report := syncer.Run(ctx)
	if report.State != syncs.Succeeded || report.Counts[steam.CountGamesAdded] != 2 || len(report.Problems) != 0 {
		t.Fatalf("first sync: %+v", report)
	}
	idled, _ := s.ListItems(ctx, store.ItemFilter{Suggested: steam.RuleLikelyIdled})
	never, _ := s.ListItems(ctx, store.ItemFilter{Suggested: steam.RuleLikelyNeverPlayed})
	if len(idled) != 1 || len(never) != 1 {
		t.Fatalf("idled %d never %d", len(idled), len(never))
	}
	if idled[0].Verdicts.Values["title"] != "Store 1" || idled[0].Verdicts.Values["released"] != "2020-12-04" {
		t.Errorf("verdicts %v", idled[0].Verdicts.Values)
	}
	if takes := idled[0].Takes; len(takes) != 1 || takes[0].Source != record.SourceSteamTake ||
		takes[0].Values["recommendation"] != "not-recommended" || takes[0].Values["review"] != "farms cards" ||
		report.Counts[steam.CountReviewsRecorded] != 1 {
		t.Errorf("takes %+v counts %v", takes, report.Counts)
	}

	syncer.Run(ctx)
	if achievementCalls.Load() != 1 || storeCalls.Load() != 2 {
		t.Errorf("unchanged resync fetched again: achievements %d store %d", achievementCalls.Load(), storeCalls.Load())
	}
	playtime.Store(400)
	report = syncer.Run(ctx)
	if achievementCalls.Load() != 2 || report.Counts[steam.CountGamesAdded] != 0 {
		t.Errorf("playtime change: achievements %d, report %+v", achievementCalls.Load(), report)
	}
}

func openProposing(t *testing.T, d *sql.DB) *store.Store {
	t.Helper()
	if err := take.WriteKindStatuses(context.Background(), d, kind.All()); err != nil {
		t.Fatal(err)
	}
	records := record.New(d, reading.ReadRecord, reading.ReadersVersion)
	return store.New(d, records, world.New(d, records), []rule.Proposer{steam.Proposer})
}

func TestOwnedSteamGamesAreLoggedUntriaged(t *testing.T) {
	ctx := context.Background()
	d, err := db.Open(ctx, filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	s := openProposing(t, d)
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/IPlayerService/GetOwnedGames/v1/":
			fmt.Fprint(w, `{"response":{"games":[{"appid":7,"name":"Short Session","playtime_forever":35}]}}`)
		case "/api/appdetails":
			fmt.Fprint(w, `{"7":{"success":true,"data":{"name":"Short Session"}}}`)
		case "/ISteamUserStats/GetPlayerAchievements/v1/":
			fmt.Fprint(w, `{"playerstats":{"success":true,"achievements":[]}}`)
		case "/profiles/me/recommended/":
		default:
			http.NotFound(w, r)
		}
	}))
	defer fake.Close()
	client := steam.NewClient("key", "me")
	client.WebAPIBaseURL, client.StoreBaseURL, client.CommunityBaseURL = fake.URL, fake.URL, fake.URL
	steamSyncer := steam.NewSyncer(s, client)
	steamSyncer.StoreRequestInterval = 1
	syncer, err := syncs.New(ctx, steam.SyncName, record.SourceSteam, steamSyncer.Work, s, nil)
	if err != nil {
		t.Fatal(err)
	}
	if report := syncer.Run(ctx); report.State != syncs.Succeeded || len(report.Problems) != 0 {
		t.Fatalf("sync: %+v", report)
	}
	id, err := s.FindByIdentity(ctx, record.SourceSteam, "7")
	if err != nil {
		t.Fatal(err)
	}
	v, err := s.GetItem(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if v.Entry == nil || v.Entry.LoggedBy != steam.RuleOwned || v.Entry.Status != kind.Untriaged || v.Entry.StatusDecidedBy != take.Nobody {
		t.Fatalf("an owned game: %+v", v.Entry)
	}
}
