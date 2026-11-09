package verdict

import (
	"encoding/json"
	"strings"
	"time"
)

type SteamOwnedGame struct {
	AppID           int64  `json:"appid"`
	Name            string `json:"name"`
	PlaytimeForever int64  `json:"playtime_forever"`
	RTimeLastPlayed int64  `json:"rtime_last_played"`
}

type SteamStoreDetails struct {
	Name             string   `json:"name"`
	HeaderImage      string   `json:"header_image"`
	ShortDescription string   `json:"short_description"`
	Developers       []string `json:"developers"`
	Genres           []struct {
		Description string `json:"description"`
	} `json:"genres"`
	Categories []struct {
		Description string `json:"description"`
	} `json:"categories"`
	ReleaseDate struct {
		Date string `json:"date"`
	} `json:"release_date"`
}

type SteamPlayerStats struct {
	Achievements []struct {
		Achieved int `json:"achieved"`
	} `json:"achievements"`
}

func readSteamOwnedGame(raw json.RawMessage) (map[string]any, error) {
	var g SteamOwnedGame
	if err := json.Unmarshal(raw, &g); err != nil {
		return nil, err
	}
	values := map[string]any{
		"title":            g.Name,
		"steam_appid":      g.AppID,
		"playtime_minutes": g.PlaytimeForever,
	}
	if g.RTimeLastPlayed > 0 {
		values["last_played"] = time.Unix(g.RTimeLastPlayed, 0).UTC().Format(time.DateOnly)
	}
	return values, nil
}

func readSteamStoreDetails(raw json.RawMessage) (map[string]any, error) {
	var d SteamStoreDetails
	if err := json.Unmarshal(raw, &d); err != nil {
		return nil, err
	}
	values := map[string]any{
		"title":            d.Name,
		"cover":            d.HeaderImage,
		"description":      d.ShortDescription,
		"developers":       d.Developers,
		"genres":           descriptions(d.Genres),
		"steam_categories": descriptions(d.Categories),
	}
	if released, parsed := steamReleaseDate(d.ReleaseDate.Date); parsed {
		values["released"] = released
	}
	return withoutEmpty(values), nil
}

func readSteamAchievements(raw json.RawMessage) (map[string]any, error) {
	var stats SteamPlayerStats
	if err := json.Unmarshal(raw, &stats); err != nil {
		return nil, err
	}
	var unlocked int64
	for _, a := range stats.Achievements {
		if a.Achieved == 1 {
			unlocked++
		}
	}
	return map[string]any{
		"achievements_total":    int64(len(stats.Achievements)),
		"achievements_unlocked": unlocked,
	}, nil
}

func descriptions(xs []struct {
	Description string `json:"description"`
}) []string {
	out := make([]string, 0, len(xs))
	for _, x := range xs {
		out = append(out, x.Description)
	}
	return out
}

var steamReleaseDateLayouts = []struct {
	layout, format string
}{
	{"2 Jan, 2006", time.DateOnly},
	{"Jan 2, 2006", time.DateOnly},
	{"Jan 2006", "2006-01"},
	{"2006", "2006"},
}

func steamReleaseDate(s string) (string, bool) {
	s = strings.TrimSpace(s)
	for _, l := range steamReleaseDateLayouts {
		if t, err := time.Parse(l.layout, s); err == nil {
			return t.Format(l.format), true
		}
	}
	return "", false
}

func withoutEmpty(values map[string]any) map[string]any {
	for name, value := range values {
		switch v := value.(type) {
		case string:
			if v == "" {
				delete(values, name)
			}
		case []string:
			if len(v) == 0 {
				delete(values, name)
			}
		}
	}
	return values
}
