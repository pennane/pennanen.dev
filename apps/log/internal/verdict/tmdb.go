package verdict

import "encoding/json"

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
