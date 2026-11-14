package verdict

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
)

func TestReadSpotifyArtist(t *testing.T) {
	raw := json.RawMessage(`{
		"id": "4tZwfgrHOc3mvqYlEYSvVi",
		"name": "Daft Punk",
		"genres": ["electro", "french house"],
		"popularity": 80,
		"followers": {"href": null, "total": 9000000},
		"external_urls": {"spotify": "https://open.spotify.com/artist/4tZwfgrHOc3mvqYlEYSvVi"},
		"images": [
			{"url": "https://i.scdn.co/small", "height": 160, "width": 160},
			{"url": "https://i.scdn.co/large", "height": 640, "width": 640}
		]
	}`)
	got, err := Read(kind.SourceSpotifyArtist, raw)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"title":      "Daft Punk",
		"genres":     []string{"electro", "french house"},
		"image":      "https://i.scdn.co/large",
		"popularity": int64(80),
		"followers":  int64(9000000),
		"spotify_id": "4tZwfgrHOc3mvqYlEYSvVi",
		"link":       "https://open.spotify.com/artist/4tZwfgrHOc3mvqYlEYSvVi",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("artist = %v, want %v", got, want)
	}
}

func TestReadSpotifyAlbum(t *testing.T) {
	raw := json.RawMessage(`{
		"id": "2noRn2Aes5aoNVsU6iWThc",
		"name": "Homework",
		"album_type": "album",
		"total_tracks": 16,
		"release_date": "1997",
		"release_date_precision": "year",
		"label": "Virgin",
		"genres": [],
		"popularity": 60,
		"artists": [{"id": "4tZwfgrHOc3mvqYlEYSvVi", "name": "Daft Punk"}],
		"external_urls": {"spotify": "https://open.spotify.com/album/2noRn2Aes5aoNVsU6iWThc"},
		"images": [
			{"url": "https://i.scdn.co/cover-big", "height": 640, "width": 640},
			{"url": "https://i.scdn.co/cover-small", "height": 64, "width": 64}
		]
	}`)
	got, err := Read(kind.SourceSpotifyAlbum, raw)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"title":      "Homework",
		"artists":    []string{"Daft Punk"},
		"released":   "1997",
		"album_type": "album",
		"tracks":     int64(16),
		"label":      "Virgin",
		"popularity": int64(60),
		"cover":      "https://i.scdn.co/cover-big",
		"spotify_id": "2noRn2Aes5aoNVsU6iWThc",
		"link":       "https://open.spotify.com/album/2noRn2Aes5aoNVsU6iWThc",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("album = %v, want %v", got, want)
	}
}
