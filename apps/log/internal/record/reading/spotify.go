package reading

import (
	"cmp"
	"encoding/json"
	"slices"
)

type SpotifyImage struct {
	URL    string `json:"url"`
	Height int64  `json:"height"`
	Width  int64  `json:"width"`
}

type SpotifyExternalURLs struct {
	Spotify string `json:"spotify"`
}

type SpotifyArtist struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Genres     []string `json:"genres"`
	Popularity *int64   `json:"popularity"`
	Followers  struct {
		Total *int64 `json:"total"`
	} `json:"followers"`
	Images       []SpotifyImage      `json:"images"`
	ExternalURLs SpotifyExternalURLs `json:"external_urls"`
}

type SpotifyAlbum struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	AlbumType   string   `json:"album_type"`
	TotalTracks *int64   `json:"total_tracks"`
	ReleaseDate string   `json:"release_date"`
	Label       string   `json:"label"`
	Genres      []string `json:"genres"`
	Popularity  *int64   `json:"popularity"`
	Artists     []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"artists"`
	Tracks struct {
		Items []TracklistEntry `json:"items"`
	} `json:"tracks"`
	Images       []SpotifyImage      `json:"images"`
	ExternalURLs SpotifyExternalURLs `json:"external_urls"`
}

type TracklistEntry struct {
	ID          string `json:"id"`
	DiscNumber  int    `json:"disc_number"`
	TrackNumber int    `json:"track_number"`
}

func InTracklistOrder(entries []TracklistEntry) []string {
	sorted := slices.Clone(entries)
	slices.SortStableFunc(sorted, func(a, b TracklistEntry) int {
		return cmp.Or(cmp.Compare(a.DiscNumber, b.DiscNumber), cmp.Compare(a.TrackNumber, b.TrackNumber))
	})
	ids := make([]string, 0, len(sorted))
	for _, entry := range sorted {
		if entry.ID != "" {
			ids = append(ids, entry.ID)
		}
	}
	return ids
}

func readSpotifyArtist(raw json.RawMessage) (map[string]any, error) {
	var a SpotifyArtist
	if err := json.Unmarshal(raw, &a); err != nil {
		return nil, err
	}
	values := withoutEmpty(map[string]any{
		"title":      a.Name,
		"genres":     a.Genres,
		"image":      largestImage(a.Images),
		"spotify_id": a.ID,
		"link":       a.ExternalURLs.Spotify,
	})
	putWhenGiven(values, "popularity", a.Popularity)
	putWhenGiven(values, "followers", a.Followers.Total)
	return values, nil
}

func readSpotifyAlbum(raw json.RawMessage) (map[string]any, error) {
	var a SpotifyAlbum
	if err := json.Unmarshal(raw, &a); err != nil {
		return nil, err
	}
	credited := make([]string, 0, len(a.Artists))
	creditedIDs := make([]string, 0, len(a.Artists))
	for _, artist := range a.Artists {
		credited = append(credited, artist.Name)
		if artist.ID != "" {
			creditedIDs = append(creditedIDs, artist.ID)
		}
	}
	values := withoutEmpty(map[string]any{
		"title":      a.Name,
		"artists":    credited,
		"artist_ids": creditedIDs,
		"track_ids":  InTracklistOrder(a.Tracks.Items),
		"released":   a.ReleaseDate,
		"album_type": a.AlbumType,
		"label":      a.Label,
		"genres":     a.Genres,
		"cover":      largestImage(a.Images),
		"spotify_id": a.ID,
		"link":       a.ExternalURLs.Spotify,
	})
	putWhenGiven(values, "popularity", a.Popularity)
	putWhenGiven(values, "tracks", a.TotalTracks)
	return values, nil
}

func putWhenGiven(values map[string]any, name string, given *int64) {
	if given != nil {
		values[name] = *given
	}
}

func largestImage(images []SpotifyImage) string {
	var largest SpotifyImage
	for _, image := range images {
		if image.Width*image.Height > largest.Width*largest.Height || largest.URL == "" {
			largest = image
		}
	}
	return largest.URL
}

const listeningField = "listening"

type SpotifyListening struct {
	Listening *int64 `json:"listening"`
}

func readSpotifyListening(raw json.RawMessage) (map[string]any, error) {
	var l SpotifyListening
	if err := json.Unmarshal(raw, &l); err != nil {
		return nil, err
	}
	values := map[string]any{}
	putWhenGiven(values, listeningField, l.Listening)
	return values, nil
}

type SpotifyTrack struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DurationMS  *int64 `json:"duration_ms"`
	Popularity  *int64 `json:"popularity"`
	ExternalIDs struct {
		ISRC string `json:"isrc"`
	} `json:"external_ids"`
	Artists []struct {
		Name string `json:"name"`
	} `json:"artists"`
	ExternalURLs SpotifyExternalURLs `json:"external_urls"`
}

const millisecondsPerSecond = 1000

func readSpotifyTrack(raw json.RawMessage) (map[string]any, error) {
	var t SpotifyTrack
	if err := json.Unmarshal(raw, &t); err != nil {
		return nil, err
	}
	credited := make([]string, 0, len(t.Artists))
	for _, artist := range t.Artists {
		credited = append(credited, artist.Name)
	}
	values := withoutEmpty(map[string]any{
		"title":      t.Name,
		"artists":    credited,
		"isrc":       t.ExternalIDs.ISRC,
		"spotify_id": t.ID,
		"link":       t.ExternalURLs.Spotify,
	})
	if t.DurationMS != nil {
		values["duration"] = (*t.DurationMS + millisecondsPerSecond/2) / millisecondsPerSecond
	}
	putWhenGiven(values, "popularity", t.Popularity)
	return values, nil
}
