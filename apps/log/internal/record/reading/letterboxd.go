package reading

import (
	"encoding/json"
	"fmt"
)

type LetterboxdReaction string

const (
	NoReaction LetterboxdReaction = ""
	Liked      LetterboxdReaction = "liked"
)

type LetterboxdTake struct {
	Rating           *float64           `json:"rating,omitempty"`
	Reaction         LetterboxdReaction `json:"reaction,omitempty"`
	Review           string             `json:"review,omitempty"`
	LastWatched      string             `json:"lastWatched,omitempty"`
	OnWatchlistSince string             `json:"onWatchlistSince,omitempty"`
}

func readLetterboxdTake(raw json.RawMessage) (map[string]any, error) {
	var take LetterboxdTake
	if err := json.Unmarshal(raw, &take); err != nil {
		return nil, err
	}
	values := map[string]any{
		"reaction":           string(take.Reaction),
		"review":             take.Review,
		"last_watched":       take.LastWatched,
		"on_watchlist_since": take.OnWatchlistSince,
	}
	if take.Rating != nil {
		values["rating"] = *take.Rating
	}
	return withoutEmpty(values), nil
}

type LetterboxdFilm struct {
	ID       string   `json:"id"`
	Slug     string   `json:"slug,omitempty"`
	Title    string   `json:"title"`
	Year     int      `json:"year,omitempty"`
	TMDBID   int64    `json:"tmdbId,omitempty"`
	TMDBType TMDBType `json:"tmdbType,omitempty"`
}

func (f LetterboxdFilm) TMDB() (TMDBLink, bool) {
	if f.TMDBID == 0 {
		return TMDBLink{}, false
	}
	if f.TMDBType == "" {
		return TMDBLink{Type: MovieOnTMDB, ID: f.TMDBID}, true
	}
	return TMDBLink{Type: f.TMDBType, ID: f.TMDBID}, true
}

func (f LetterboxdFilm) Linked(link TMDBLink) LetterboxdFilm {
	f.TMDBID, f.TMDBType = link.ID, link.Type
	return f
}

func (f LetterboxdFilm) Link() string {
	if f.Slug != "" {
		return "https://letterboxd.com/film/" + f.Slug + "/"
	}
	return "https://boxd.it/" + f.ID
}

func readLetterboxdFilm(raw json.RawMessage) (map[string]any, error) {
	var film LetterboxdFilm
	if err := json.Unmarshal(raw, &film); err != nil {
		return nil, err
	}
	values := map[string]any{"title": film.Title, "link": film.Link()}
	if link, linked := film.TMDB(); linked {
		values["tmdb_link"] = link.Type.Source() + ":" + link.ExternalID()
	}
	if film.Year > 0 {
		values["released"] = fmt.Sprintf("%04d", film.Year)
	}
	return withoutEmpty(values), nil
}
