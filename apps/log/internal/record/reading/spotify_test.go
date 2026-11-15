package reading

import (
	"encoding/json"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"reflect"
	"testing"
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
	got, err := Read(record.SourceSpotifyArtist, raw)
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
	got, err := Read(record.SourceSpotifyAlbum, raw)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"title":      "Homework",
		"artists":    []string{"Daft Punk"},
		"artist_ids": []string{"4tZwfgrHOc3mvqYlEYSvVi"},
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

func TestSpotifyPopularityIsMissingWhenAbsent(t *testing.T) {
	absent, err := Read(record.SourceSpotifyArtist, json.RawMessage(`{"id":"a","name":"Simplified"}`))
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"popularity", "followers"} {
		if _, present := absent[key]; present {
			t.Errorf("artist without %s has it in %v", key, absent)
		}
	}

	zero, err := Read(record.SourceSpotifyArtist, json.RawMessage(`{"id":"a","name":"Quiet","popularity":0,"followers":{"total":0}}`))
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"popularity", "followers"} {
		if zero[key] != int64(0) {
			t.Errorf("artist with spotify's 0 has %s = %v, want int64(0)", key, zero[key])
		}
	}

	album, err := Read(record.SourceSpotifyAlbum, json.RawMessage(`{"id":"b","name":"Simplified"}`))
	if err != nil {
		t.Fatal(err)
	}
	if _, present := album["popularity"]; present {
		t.Errorf("album without popularity has it in %v", album)
	}
}

func TestTracksAreMissingWhenAbsent(t *testing.T) {
	album, err := Read(record.SourceSpotifyAlbum, json.RawMessage(`{"id":"b","name":"Simplified"}`))
	if err != nil {
		t.Fatal(err)
	}
	if _, present := album["tracks"]; present {
		t.Errorf("album without total_tracks has tracks in %v", album)
	}
}
