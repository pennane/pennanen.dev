package reading

import (
	"encoding/json"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"strconv"
)

const (
	tmdbPosterBase = "https://image.tmdb.org/t/p/w500"
	tmdbCoverBase  = "https://image.tmdb.org/t/p/w1280"
	tmdbDirector   = "Director"
)

type TMDBMovie struct {
	ID            int64   `json:"id"`
	Title         string  `json:"title"`
	OriginalTitle string  `json:"original_title"`
	ReleaseDate   string  `json:"release_date"`
	PosterPath    string  `json:"poster_path"`
	BackdropPath  string  `json:"backdrop_path"`
	Runtime       int64   `json:"runtime"`
	VoteAverage   float64 `json:"vote_average"`
	VoteCount     int64   `json:"vote_count"`
	Genres        []struct {
		Name string `json:"name"`
	} `json:"genres"`
	ProductionCountries []struct {
		Name string `json:"name"`
	} `json:"production_countries"`
	Credits struct {
		Crew []struct {
			Name string `json:"name"`
			Job  string `json:"job"`
		} `json:"crew"`
	} `json:"credits"`
}

func readTMDBMovie(raw json.RawMessage) (map[string]any, error) {
	var m TMDBMovie
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	values := map[string]any{
		"title":          m.Title,
		"original_title": m.OriginalTitle,
		"released":       m.ReleaseDate,
		"poster":         prefixed(tmdbPosterBase, m.PosterPath),
		"cover":          prefixed(tmdbCoverBase, m.BackdropPath),
		"directors":      directors(m),
		"genres":         names(m.Genres),
		"countries":      names(m.ProductionCountries),
	}
	if m.ID > 0 {
		values["tmdb_id"] = m.ID
	}
	if m.Runtime > 0 {
		values["runtime_minutes"] = m.Runtime
	}
	if m.VoteCount > 0 {
		values["vote_average"] = m.VoteAverage
		values["vote_count"] = m.VoteCount
	}
	return withoutEmpty(values), nil
}

func prefixed(base, path string) string {
	if path == "" {
		return ""
	}
	return base + path
}

func directors(m TMDBMovie) []string {
	out := []string{}
	for _, c := range m.Credits.Crew {
		if c.Job == tmdbDirector {
			out = append(out, c.Name)
		}
	}
	return out
}

func names(xs []struct {
	Name string `json:"name"`
}) []string {
	out := make([]string, 0, len(xs))
	for _, x := range xs {
		out = append(out, x.Name)
	}
	return out
}

type TMDBType string

const (
	MovieOnTMDB  TMDBType = "movie"
	SeriesOnTMDB TMDBType = "tv"
)

var TMDBTypes = []TMDBType{MovieOnTMDB, SeriesOnTMDB}

func (t TMDBType) Source() string {
	if t == SeriesOnTMDB {
		return record.SourceTMDBSeries
	}
	return record.SourceTMDB
}

type TMDBLink struct {
	Type TMDBType `json:"tmdbType"`
	ID   int64    `json:"tmdbId"`
}

func (l TMDBLink) ExternalID() string { return strconv.FormatInt(l.ID, 10) }

type TMDBSeries struct {
	ID               int64   `json:"id"`
	Name             string  `json:"name"`
	OriginalName     string  `json:"original_name"`
	FirstAirDate     string  `json:"first_air_date"`
	PosterPath       string  `json:"poster_path"`
	BackdropPath     string  `json:"backdrop_path"`
	NumberOfSeasons  int64   `json:"number_of_seasons"`
	NumberOfEpisodes int64   `json:"number_of_episodes"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int64   `json:"vote_count"`
	CreatedBy        []struct {
		Name string `json:"name"`
	} `json:"created_by"`
	Genres []struct {
		Name string `json:"name"`
	} `json:"genres"`
	ProductionCountries []struct {
		Name string `json:"name"`
	} `json:"production_countries"`
}

func readTMDBSeries(raw json.RawMessage) (map[string]any, error) {
	var s TMDBSeries
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, err
	}
	values := map[string]any{
		"title":          s.Name,
		"original_title": s.OriginalName,
		"released":       s.FirstAirDate,
		"poster":         prefixed(tmdbPosterBase, s.PosterPath),
		"cover":          prefixed(tmdbCoverBase, s.BackdropPath),
		"creators":       names(s.CreatedBy),
		"genres":         names(s.Genres),
		"countries":      names(s.ProductionCountries),
	}
	if s.ID > 0 {
		values["tmdb_id"] = s.ID
	}
	if s.NumberOfSeasons > 0 {
		values["seasons"] = s.NumberOfSeasons
	}
	if s.NumberOfEpisodes > 0 {
		values["episodes"] = s.NumberOfEpisodes
	}
	if s.VoteCount > 0 {
		values["vote_average"] = s.VoteAverage
		values["vote_count"] = s.VoteCount
	}
	return withoutEmpty(values), nil
}

func TMDBPosterURL(path string) string { return prefixed(tmdbPosterBase, path) }
