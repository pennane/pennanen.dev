package verdict

import (
	"encoding/json"
	"time"
)

type IGDBImage struct {
	ImageID string `json:"image_id"`
}

type IGDBGame struct {
	ID                int64       `json:"id"`
	Name              string      `json:"name"`
	FirstReleaseDate  int64       `json:"first_release_date"`
	Cover             *IGDBImage  `json:"cover"`
	Artworks          []IGDBImage `json:"artworks"`
	Screenshots       []IGDBImage `json:"screenshots"`
	Genres            []igdbNamed `json:"genres"`
	Platforms         []igdbNamed `json:"platforms"`
	InvolvedCompanies []struct {
		Developer bool      `json:"developer"`
		Company   igdbNamed `json:"company"`
	} `json:"involved_companies"`
}

type igdbNamed struct {
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
}

const (
	igdbImageBase  = "https://images.igdb.com/igdb/image/upload/"
	igdbPosterSize = "t_cover_big"
	igdbWideSize   = "t_screenshot_big"
)

func IGDBImageURL(size string, image IGDBImage) string {
	return igdbImageBase + size + "/" + image.ImageID + ".jpg"
}

func IGDBPosterURL(image IGDBImage) string { return IGDBImageURL(igdbPosterSize, image) }

func (g IGDBGame) Matched() bool { return g.ID != 0 }

func (g IGDBGame) Year() string {
	if g.FirstReleaseDate == 0 {
		return ""
	}
	return time.Unix(g.FirstReleaseDate, 0).UTC().Format("2006")
}

func (g IGDBGame) PlatformNames() []string {
	names := make([]string, 0, len(g.Platforms))
	for _, p := range g.Platforms {
		if p.Abbreviation != "" {
			names = append(names, p.Abbreviation)
		} else if p.Name != "" {
			names = append(names, p.Name)
		}
	}
	return names
}

func readIGDBGame(raw json.RawMessage) (map[string]any, error) {
	var g IGDBGame
	if err := json.Unmarshal(raw, &g); err != nil {
		return nil, err
	}
	if !g.Matched() {
		return map[string]any{}, nil
	}
	values := map[string]any{"title": g.Name}
	if g.FirstReleaseDate != 0 {
		values["released"] = time.Unix(g.FirstReleaseDate, 0).UTC().Format(time.DateOnly)
	}
	if g.Cover != nil && g.Cover.ImageID != "" {
		values["poster"] = IGDBPosterURL(*g.Cover)
	}
	for _, wide := range [][]IGDBImage{g.Artworks, g.Screenshots} {
		if len(wide) > 0 && wide[0].ImageID != "" {
			values["cover"] = IGDBImageURL(igdbWideSize, wide[0])
			break
		}
	}
	if genres := namesOf(g.Genres); len(genres) > 0 {
		values["genres"] = genres
	}
	if platforms := g.PlatformNames(); len(platforms) > 0 {
		values["platforms"] = platforms
	}
	var developers []string
	for _, c := range g.InvolvedCompanies {
		if c.Developer && c.Company.Name != "" {
			developers = append(developers, c.Company.Name)
		}
	}
	if len(developers) > 0 {
		values["developers"] = developers
	}
	return values, nil
}

func namesOf(named []igdbNamed) []string {
	names := make([]string, 0, len(named))
	for _, n := range named {
		if n.Name != "" {
			names = append(names, n.Name)
		}
	}
	return names
}
