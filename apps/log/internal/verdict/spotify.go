package verdict

import "encoding/json"

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
	Popularity int64    `json:"popularity"`
	Followers  struct {
		Total int64 `json:"total"`
	} `json:"followers"`
	Images       []SpotifyImage      `json:"images"`
	ExternalURLs SpotifyExternalURLs `json:"external_urls"`
}

type SpotifyAlbum struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	AlbumType   string   `json:"album_type"`
	TotalTracks int64    `json:"total_tracks"`
	ReleaseDate string   `json:"release_date"`
	Label       string   `json:"label"`
	Genres      []string `json:"genres"`
	Popularity  int64    `json:"popularity"`
	Artists     []struct {
		Name string `json:"name"`
	} `json:"artists"`
	Images       []SpotifyImage      `json:"images"`
	ExternalURLs SpotifyExternalURLs `json:"external_urls"`
}

func readSpotifyArtist(raw json.RawMessage) (map[string]any, error) {
	var a SpotifyArtist
	if err := json.Unmarshal(raw, &a); err != nil {
		return nil, err
	}
	return withoutEmpty(map[string]any{
		"title":      a.Name,
		"genres":     a.Genres,
		"image":      largestImage(a.Images),
		"popularity": a.Popularity,
		"followers":  a.Followers.Total,
		"spotify_id": a.ID,
		"link":       a.ExternalURLs.Spotify,
	}), nil
}

func readSpotifyAlbum(raw json.RawMessage) (map[string]any, error) {
	var a SpotifyAlbum
	if err := json.Unmarshal(raw, &a); err != nil {
		return nil, err
	}
	credited := make([]string, 0, len(a.Artists))
	for _, artist := range a.Artists {
		credited = append(credited, artist.Name)
	}
	return withoutEmpty(map[string]any{
		"title":      a.Name,
		"artists":    credited,
		"released":   a.ReleaseDate,
		"album_type": a.AlbumType,
		"tracks":     a.TotalTracks,
		"label":      a.Label,
		"genres":     a.Genres,
		"popularity": a.Popularity,
		"cover":      largestImage(a.Images),
		"spotify_id": a.ID,
		"link":       a.ExternalURLs.Spotify,
	}), nil
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
