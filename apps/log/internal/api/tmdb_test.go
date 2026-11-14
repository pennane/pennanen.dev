package api_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/api"
	"github.com/pennane/pennanen.dev/apps/log/internal/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/tmdb"
)

func TestApFixesAWrongTMDBMatch(t *testing.T) {
	ctx := context.Background()
	d, err := db.Open(ctx, filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/3/movie/1" {
			io.WriteString(w, `{"id":1,"title":"Purple Noon","release_date":"1960-03-10","poster_path":"/p.jpg"}`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(upstream.Close)
	client := tmdb.NewClient("t")
	client.BaseURL = upstream.URL
	st := store.New(d)
	srv := httptest.NewServer((&api.Server{Store: st, AdminToken: token, TMDB: client}).Handler())
	t.Cleanup(srv.Close)

	id, _ := st.CreateItemFromSource(ctx, "movie", "Ripley", kind.SourceLetterboxd, "aaa1", json.RawMessage(`{"id":"aaa1","title":"Ripley","year":2024}`))
	wrong := json.RawMessage(`{"id":9,"title":"In the Realm of Ripley","poster_path":"/wrong.jpg"}`)
	if err := st.PutClaim(ctx, id, kind.SourceTMDB, "9", wrong); err != nil {
		t.Fatal(err)
	}

	if status, body := call(t, srv, "PUT", "/admin/items/"+itoa(id)+"/tmdb/none", token, ""); status != http.StatusOK {
		t.Fatalf("no match: %d %s", status, body)
	}
	v, _ := st.GetItem(ctx, id)
	if v.Verdicts.Values["title"] != "Ripley" || v.Verdicts.Values["poster"] != nil {
		t.Errorf("after no match: %v", v.Verdicts.Values)
	}
	lacking, _ := st.IdentitiesLackingSources(ctx, kind.SourceLetterboxd, kind.SourceTMDB, kind.SourceTMDBSeries)
	if len(lacking) != 0 {
		t.Errorf("the nightly match would try again: %v", lacking)
	}

	if status, body := call(t, srv, "PUT", "/admin/items/"+itoa(id)+"/tmdb", token, `{"tmdbId":1}`); status != http.StatusOK {
		t.Fatalf("pick: %d %s", status, body)
	}
	v, _ = st.GetItem(ctx, id)
	if v.Verdicts.Values["title"] != "Purple Noon" {
		t.Errorf("after picking: %v", v.Verdicts.Values)
	}
	if status, _ := call(t, srv, "PUT", "/admin/items/"+itoa(id)+"/tmdb", token, `{"tmdbId":2}`); status != http.StatusUnprocessableEntity {
		t.Errorf("a movie tmdb doesn't have: %d", status)
	}
}

func TestApAddsASeriesFromTMDBAndMovesAFilmThatIsOne(t *testing.T) {
	ctx := context.Background()
	d, err := db.Open(ctx, filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/3/search/tv":
			io.WriteString(w, `{"results":[{"id":1920,"name":"Twin Peaks","first_air_date":"1990-04-08","poster_path":"/tp.jpg"}]}`)
		case "/3/tv/1920":
			io.WriteString(w, `{"id":1920,"name":"Twin Peaks","first_air_date":"1990-04-08","number_of_seasons":2,"number_of_episodes":30,"created_by":[{"name":"David Lynch"},{"name":"Mark Frost"}]}`)
		case "/3/tv/94028":
			io.WriteString(w, `{"id":94028,"name":"Ripley","first_air_date":"2024-04-04","number_of_seasons":1,"poster_path":"/r.jpg"}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(upstream.Close)
	client := tmdb.NewClient("t")
	client.BaseURL = upstream.URL
	st := store.New(d)
	srv := httptest.NewServer((&api.Server{Store: st, AdminToken: token, TMDB: client}).Handler())
	t.Cleanup(srv.Close)

	status, body := call(t, srv, "GET", "/admin/tmdb/search?type=tv&q=twin", token, "")
	var found []api.TMDBResult
	json.Unmarshal([]byte(body), &found)
	if status != http.StatusOK || len(found) != 1 || found[0].Name != "Twin Peaks" || found[0].Year != "1990" || found[0].Poster == "" {
		t.Fatalf("search: %d %s", status, body)
	}
	status, body = call(t, srv, "POST", "/admin/items/from-tmdb", token, `{"tmdbId":1920,"tmdbType":"tv"}`)
	var added store.ItemView
	json.Unmarshal([]byte(body), &added)
	if status != http.StatusCreated || added.Kind != kind.Series || added.Slug != "twin-peaks" || added.Verdicts.Values["episodes"] != float64(30) {
		t.Fatalf("add: %d %s", status, body)
	}
	if status, _ := call(t, srv, "POST", "/admin/items/from-tmdb", token, `{"tmdbId":1920,"tmdbType":"tv"}`); status != http.StatusOK {
		t.Errorf("adding it again: %d", status)
	}
	if status, _ := call(t, srv, "GET", "/admin/tmdb/search?type=book&q=x", token, ""); status != http.StatusBadRequest {
		t.Errorf("a type tmdb doesn't have: %d", status)
	}

	id, _ := st.CreateItemFromSource(ctx, kind.Movie, "Ripley", kind.SourceLetterboxd, "rip1", json.RawMessage(`{"id":"rip1","title":"Ripley","year":2024}`))
	st.PutClaim(ctx, id, kind.SourceTMDB, "", tmdb.NoMatch)
	if status, body := call(t, srv, "PUT", "/admin/items/"+itoa(id)+"/tmdb", token, `{"tmdbId":94028,"tmdbType":"tv"}`); status != http.StatusOK {
		t.Fatalf("pick a series: %d %s", status, body)
	}
	v, _ := st.GetItem(ctx, id)
	if v.Kind != kind.Series || v.Verdicts.Values["poster"] != "https://image.tmdb.org/t/p/w500/r.jpg" || v.Verdicts.Values["seasons"] != int64(1) {
		t.Errorf("moved film: %s %v", v.Kind, v.Verdicts.Values)
	}
	if _, err := st.Claim(ctx, id, kind.SourceTMDB); err == nil {
		t.Error("the film's no-match answer stayed")
	}
}
