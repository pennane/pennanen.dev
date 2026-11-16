package kind

import (
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
)

const (
	Artist = "artist"
	Album  = "album"
	Track  = "track"

	FieldListening = "listening"

	Listening = "listening"
	Listened  = "listened"

	VariousArtists = "0LyfQWJT6nXafLPZqxe9Of"
)

var listeningField = Field{
	Name:                    FieldListening,
	Type:                    Int,
	SourcesBestFirstAfterAP: []string{},
	MeasuredBy:              record.SourceSpotify,
	About:                   AboutAP,
}

func init() {
	register(Kind{
		Name:            Artist,
		DefaultTitle:    "artists",
		Statuses:        []string{Untriaged, Listening, Listened},
		StatusesToScore: []string{Listening, Listened},
		StatusesShown:   []string{Listening, Listened},
		DefaultAxes:     []string{Prestige, Quality},
		RankedAxis:      Prestige,
		IdentitySource:  record.SourceSpotifyArtist,
		UnrankedOrder:   FieldListening,
		Fields: []Field{
			{Name: FieldTitle, Type: Text, SourcesBestFirstAfterAP: []string{record.SourceSpotifyArtist}},
			{Name: "genres", Type: StringList, SourcesBestFirstAfterAP: []string{record.SourceSpotifyArtist}},
			{Name: "image", Type: Picture, Shape: Square, SourcesBestFirstAfterAP: []string{record.SourceSpotifyArtist}},
			{Name: "popularity", Type: Int, SourcesBestFirstAfterAP: []string{record.SourceSpotifyArtist}},
			{Name: "followers", Type: Int, SourcesBestFirstAfterAP: []string{record.SourceSpotifyArtist}},
			listeningField,
			{Name: "spotify_id", Type: Text, SourcesBestFirstAfterAP: []string{record.SourceSpotifyArtist}},
			{Name: "link", Type: URL, SourcesBestFirstAfterAP: []string{record.SourceSpotifyArtist}},
		},
	})
	register(Kind{
		Name:            Album,
		DefaultTitle:    "albums",
		Statuses:        []string{Untriaged, Want, Listened},
		StatusesToScore: []string{Listened},
		StatusesShown:   []string{Listened},
		DefaultAxes:     []string{Prestige, Quality},
		RankedAxis:      Prestige,
		IdentitySource:  record.SourceSpotifyAlbum,
		UnrankedOrder:   FieldListening,
		Fields: []Field{
			{Name: FieldTitle, Type: Text, SourcesBestFirstAfterAP: []string{record.SourceSpotifyAlbum}},
			{Name: "artists", Type: StringList, SourcesBestFirstAfterAP: []string{record.SourceSpotifyAlbum}},
			{Name: "released", Type: PartialDate, SourcesBestFirstAfterAP: []string{record.SourceSpotifyAlbum}},
			{Name: "album_type", Type: Text, SourcesBestFirstAfterAP: []string{record.SourceSpotifyAlbum}},
			{Name: "tracks", Type: Int, SourcesBestFirstAfterAP: []string{record.SourceSpotifyAlbum}},
			{Name: "label", Type: Text, SourcesBestFirstAfterAP: []string{record.SourceSpotifyAlbum}},
			{Name: "genres", Type: StringList, SourcesBestFirstAfterAP: []string{record.SourceSpotifyAlbum}, InheritedThrough: LinkBy},
			{Name: "popularity", Type: Int, SourcesBestFirstAfterAP: []string{record.SourceSpotifyAlbum}},
			listeningField,
			{Name: "cover", Type: Picture, Shape: Square, SourcesBestFirstAfterAP: []string{record.SourceSpotifyAlbum}},
			{Name: "spotify_id", Type: Text, SourcesBestFirstAfterAP: []string{record.SourceSpotifyAlbum}},
			{Name: "link", Type: URL, SourcesBestFirstAfterAP: []string{record.SourceSpotifyAlbum}},
		},
	})
	register(Kind{
		Name:            Track,
		DefaultTitle:    "tracks",
		Statuses:        []string{Untriaged, Listened},
		StatusesToScore: []string{Listened},
		StatusesShown:   []string{Listened},
		DefaultAxes:     []string{Prestige, Quality},
		RankedAxis:      Prestige,
		IdentitySource:  record.SourceSpotifyTrack,
		Fields: []Field{
			{Name: FieldTitle, Type: Text, SourcesBestFirstAfterAP: []string{record.SourceSpotifyTrack}},
			{Name: "artists", Type: StringList, SourcesBestFirstAfterAP: []string{record.SourceSpotifyTrack}},
			{Name: "duration", Type: Int, SourcesBestFirstAfterAP: []string{record.SourceSpotifyTrack}},
			{Name: "popularity", Type: Int, SourcesBestFirstAfterAP: []string{record.SourceSpotifyTrack}},
			{Name: "isrc", Type: Text, SourcesBestFirstAfterAP: []string{record.SourceSpotifyTrack}},
			{Name: "spotify_id", Type: Text, SourcesBestFirstAfterAP: []string{record.SourceSpotifyTrack}},
			{Name: "link", Type: URL, SourcesBestFirstAfterAP: []string{record.SourceSpotifyTrack}},
		},
	})
}
