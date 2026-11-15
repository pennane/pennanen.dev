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

func TestAFilmPageNamesTheFilmByItsShortLink(t *testing.T) {
	id, err := idIn(`<a href="https://boxd.it/abc1">a review</a><input id="url-field-film-51171" type="text" value="https://boxd.it/297o" readonly spellcheck="false">`)
	if err != nil || id != "297o" {
		t.Errorf("short link: %q %v", id, err)
	}
	if _, err := idIn(`<p>nothing</p>`); err != ErrNoShortLink {
		t.Errorf("no short link: %v", err)
	}
}
