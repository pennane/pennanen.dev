package letterboxd

import (
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
)

func TestAFilmPageTellsWhetherItIsAMovieOrASeriesOnTMDB(t *testing.T) {
	series, err := tmdbLinkIn(`<a href="https://www.themoviedb.org/tv/94028/" class="micro-button track-event" data-track-action="TMDB">TMDB</a>`)
	if err != nil || series != (verdict.TMDBLink{Type: verdict.SeriesOnTMDB, ID: 94028}) {
		t.Errorf("series: %+v %v", series, err)
	}
	movie, err := tmdbLinkIn(`<body data-tmdb-type="movie" data-tmdb-id="1398"><a href="https://www.themoviedb.org/movie/1398/">TMDB</a>`)
	if err != nil || movie != (verdict.TMDBLink{Type: verdict.MovieOnTMDB, ID: 1398}) {
		t.Errorf("movie: %+v %v", movie, err)
	}
	if _, err := tmdbLinkIn(`<p>nothing</p>`); err != ErrNoTMDBLink {
		t.Errorf("no link: %v", err)
	}
}
