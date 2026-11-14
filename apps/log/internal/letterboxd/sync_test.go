package letterboxd_test

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/letterboxd"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/syncs"
	"github.com/pennane/pennanen.dev/apps/log/internal/tmdb"
)

func gridItem(lid, slug, name, viewing string) string {
	return fmt.Sprintf(`<li class="griditem">
	<div class="react-component" data-component-class="LazyPoster" data-item-name="%s" data-item-slug="%s" data-item-link="/film/%s/" data-postered-identifier='{&quot;lid&quot;:&quot;%s&quot;,&quot;uid&quot;:&quot;film:1&quot;,&quot;type&quot;:&quot;film&quot;}'> <div class="poster film-poster"></div> </div> %s
</li>`, name, slug, slug, lid, viewing)
}

const rss = `<?xml version='1.0' encoding='utf-8'?>
<rss version="2.0" xmlns:letterboxd="https://letterboxd.com" xmlns:tmdb="https://themoviedb.org"><channel>
<item> <title>Decent Film, 2024 - ★★★½</title> <link>https://letterboxd.com/me/film/decent-film/</link> <guid isPermaLink="false">letterboxd-review-9</guid> <pubDate>Wed, 11 Nov 2026 11:09:02 +1300</pubDate> <letterboxd:watchedDate>2026-11-10</letterboxd:watchedDate> <letterboxd:rewatch>No</letterboxd:rewatch> <letterboxd:filmTitle>Decent Film</letterboxd:filmTitle> <letterboxd:filmYear>2024</letterboxd:filmYear> <letterboxd:memberRating>3.5</letterboxd:memberRating> <letterboxd:memberLike>No</letterboxd:memberLike> <tmdb:movieId>44</tmdb:movieId> <description><![CDATA[ <p><img src="https://a.ltrbxd.com/p.jpg"/></p> <p>Better than I &amp; you thought</p> <p>second thought</p> ]]></description> </item>
<item> <title>Some list</title> <link>https://letterboxd.com/me/list/x/</link> <guid isPermaLink="false">letterboxd-list-1</guid> <description>a list</description> </item>
</channel></rss>`

func exportZip(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	z := zip.NewWriter(&buf)
	for name, content := range files {
		f, err := z.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		f.Write([]byte(content))
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestExportThenNightlySyncThenTMDBMatch(t *testing.T) {
	ctx := context.Background()
	d, err := db.Open(ctx, filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	s := store.New(d)

	export, err := letterboxd.ParseExport(exportZip(t, map[string]string{
		"watched.csv":     "Date,Name,Year,Letterboxd URI\n2026-03-23,Stalker,1979,https://boxd.it/aaa1\n2026-03-23,Heat,1995,https://boxd.it/bbb2\n",
		"ratings.csv":     "Date,Name,Year,Letterboxd URI,Rating\n2026-03-23,Stalker,1979,https://boxd.it/aaa1,4.5\n",
		"likes/films.csv": "Date,Name,Year,Letterboxd URI\n2026-03-23,Stalker,1979,https://boxd.it/aaa1\n",
		"watchlist.csv":   "Date,Name,Year,Letterboxd URI\n2026-03-23,Memento,2000,https://boxd.it/ccc3\n",
		"diary.csv":       "Date,Name,Year,Letterboxd URI,Rating,Rewatch,Tags,Watched Date\n2026-03-24,Stalker,1979,https://boxd.it/entry1,4.5,Yes,,2026-03-23\n",
		"reviews.csv":     "Date,Name,Year,Letterboxd URI,Rating,Rewatch,Review,Tags,Watched Date\n2026-03-24,Stalker,1979,https://boxd.it/entry1,4.5,Yes,\"slow, then not\",,2026-03-23\n",
	}))
	if err != nil {
		t.Fatal(err)
	}

	searches := 0
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/me/films/":
			fmt.Fprint(w, gridItem("ddd4", "decent-film", "Decent Film (2024)", `<p class="poster-viewingdata"><span class="rating -micro rated-7">★★★½</span></p>`)+
				gridItem("ccc3", "memento", "Memento (2000)", `<p class="poster-viewingdata"><span class="like liked-micro has-icon icon-liked"></span></p>`))
		case r.URL.Path == "/me/watchlist/":
			fmt.Fprint(w, `<div class="poster-grid"><ul class="grid"></ul></div>`)
		case r.URL.Path == "/me/rss/":
			fmt.Fprint(w, rss)
		case r.URL.Path == "/3/search/movie":
			searches++
			if r.Header.Get("Authorization") != "Bearer token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			switch r.URL.Query().Get("query") {
			case "Stalker":
				fmt.Fprint(w, `{"results":[{"id":1398,"title":"Stalker","release_date":"1979-05-25"},{"id":7,"title":"Stalker","release_date":"2010-01-01"}]}`)
			case "Heat":
				fmt.Fprint(w, `{"results":[{"id":949,"title":"Heat","release_date":"1995-12-15"},{"id":950,"title":"Heat","release_date":"1995-01-01"}]}`)
			default:
				fmt.Fprint(w, `{"results":[]}`)
			}
		case strings.HasPrefix(r.URL.Path, "/3/movie/"):
			id := strings.TrimPrefix(r.URL.Path, "/3/movie/")
			fmt.Fprintf(w, `{"id":%s,"title":"TMDB %s","release_date":"1979-05-25","poster_path":"/p.jpg","runtime":161,"vote_average":8.1,"vote_count":2000,
				"genres":[{"name":"Drama"}],"credits":{"crew":[{"name":"Andrei Tarkovsky","job":"Director"},{"name":"Someone","job":"Editor"}]}}`, id, id)
		default:
			http.NotFound(w, r)
		}
	}))
	defer fake.Close()

	client := letterboxd.NewClient("me")
	client.BaseURL = fake.URL
	syncer := letterboxd.NewSyncer(s, client)

	imported, err := syncs.New(ctx, "letterboxd-export", func(ctx context.Context, p *syncs.Progress) error {
		return syncer.Import(ctx, p, export)
	}, s, nil)
	if err != nil {
		t.Fatal(err)
	}
	report := imported.Run(ctx)
	if report.State != syncs.Succeeded || report.Counts[letterboxd.CountFilmsAdded] != 3 || report.Counts[letterboxd.CountWatchesObserved] != 1 || len(report.Problems) != 0 {
		t.Fatalf("import: %+v", report)
	}
	stalker := item(t, s, "aaa1")
	if stalker.Entry.Status != kind.Watched || stalker.Entry.StatusSetBy != kind.SourceLetterboxd {
		t.Errorf("stalker entry: %+v", stalker.Entry)
	}
	if len(stalker.Takes) != 1 {
		t.Fatalf("stalker takes: %+v", stalker.Takes)
	}
	take := stalker.Takes[0].Values
	if take["rating"] != 4.5 || take["reaction"] != "liked" || take["review"] != "slow, then not" || take["last_watched"] != "2026-03-23" {
		t.Errorf("stalker take: %v", take)
	}
	if heat := item(t, s, "bbb2"); len(heat.Takes) != 0 {
		t.Errorf("an unrated, unliked film got a take: %+v", heat.Takes)
	}
	if memento := item(t, s, "ccc3"); memento.Entry.Status != kind.Want {
		t.Errorf("watchlist: %+v", memento.Entry)
	}

	nightly, err := syncs.New(ctx, letterboxd.SyncName, syncer.Work, s, nil)
	if err != nil {
		t.Fatal(err)
	}
	report = nightly.Run(ctx)
	if report.State != syncs.Succeeded || report.Counts[letterboxd.CountFilmsAdded] != 1 || len(report.Problems) != 0 {
		t.Fatalf("nightly: %+v", report)
	}
	if memento := item(t, s, "ccc3"); memento.Entry.Status != kind.Watched || memento.Takes[0].Values["reaction"] != "liked" {
		t.Errorf("watched off the watchlist: %+v %+v", memento.Entry, memento.Takes)
	}
	decent := item(t, s, "ddd4")
	if decent.Takes[0].Values["rating"] != 3.5 || decent.Takes[0].Values["review"] != "Better than I & you thought\n\nsecond thought" {
		t.Errorf("decent take: %v", decent.Takes[0].Values)
	}
	if decent.Verdicts.Values["link"] != "https://letterboxd.com/film/decent-film/" {
		t.Errorf("decent link: %v", decent.Verdicts.Values)
	}
	if again := nightly.Run(ctx); again.Counts[letterboxd.CountWatchesObserved] != 0 || again.Counts[letterboxd.CountFilmsAdded] != 0 {
		t.Errorf("a second night saw old things as new: %+v", again)
	}

	tmdbClient := tmdb.NewClient("token")
	tmdbClient.BaseURL = fake.URL
	matcher := tmdb.NewMatcher(s, tmdbClient)
	matcher.RequestInterval = 1
	matching, err := syncs.New(ctx, tmdb.SyncName, matcher.Work, s, nil)
	if err != nil {
		t.Fatal(err)
	}
	report = matching.Run(ctx)
	if report.Counts[tmdb.CountMatched] != 2 || report.Counts[tmdb.CountUnmatched] != 2 || len(report.Problems) != 0 {
		t.Fatalf("tmdb match: %+v", report)
	}
	if searches != 3 {
		t.Errorf("searched %d times; the film with a tmdb id from the diary needs no search", searches)
	}
	stalker = item(t, s, "aaa1")
	values := stalker.Verdicts.Values
	if values["title"] != "TMDB 1398" || values["poster"] != "https://image.tmdb.org/t/p/w500/p.jpg" || values["runtime_minutes"] != int64(161) {
		t.Errorf("stalker verdicts: %v", values)
	}
	if directors, _ := values["directors"].([]string); len(directors) != 1 || directors[0] != "Andrei Tarkovsky" {
		t.Errorf("directors: %v", values["directors"])
	}
	if decent := item(t, s, "ddd4"); decent.Verdicts.Values["tmdb_id"] != int64(44) {
		t.Errorf("decent tmdb: %v", decent.Verdicts.Values)
	}
}

func item(t *testing.T, s *store.Store, lid string) store.ItemView {
	t.Helper()
	ctx := context.Background()
	id, err := s.FindByIdentity(ctx, kind.SourceLetterboxd, lid)
	if err != nil {
		t.Fatalf("%s: %v", lid, err)
	}
	v, err := s.GetItem(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
