package tmdb

import (
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
)

func TestOnlyAFilmWithTheSameTitleAndYearMatches(t *testing.T) {
	ripley := verdict.LetterboxdFilm{Title: "Ripley", Year: 2024}
	if id, matched := onlyMatch(ripley, []Found{{ID: 1, Title: "In the Realm of Ripley", ReleaseDate: "2024-02-01"}}); matched {
		t.Errorf("the year's only film matched a series it isn't: %d", id)
	}
	amelie := verdict.LetterboxdFilm{Title: "Amélie", Year: 2001}
	found := []Found{
		{ID: 2, Title: "Amélie", ReleaseDate: "2001-04-25"},
		{ID: 3, Title: "Amélie", ReleaseDate: "1999-01-01"},
	}
	if id, matched := onlyMatch(amelie, found); !matched || id != 2 {
		t.Errorf("amélie: %d %v", id, matched)
	}
	punctuated := verdict.LetterboxdFilm{Title: "Monty Python's The Meaning of Life", Year: 1983}
	if _, matched := onlyMatch(punctuated, []Found{{ID: 4, Title: "Monty Python’s The Meaning of Life", ReleaseDate: "1983-03-31"}}); !matched {
		t.Errorf("punctuation kept the same film apart")
	}
}
