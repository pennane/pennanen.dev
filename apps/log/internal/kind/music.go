package kind

const (
	Artist = "artist"
	Album  = "album"
	Track  = "track"

	SourceSpotify       = "spotify"
	SourceSpotifyArtist = "spotify-artist"
	SourceSpotifyAlbum  = "spotify-album"
	SourceSpotifyTrack  = "spotify-track"

	SourceSpotifyListening = "spotify-listening"
	FieldListening         = "listening"

	Listening = "listening"
	Listened  = "listened"

	VariousArtists = "0LyfQWJT6nXafLPZqxe9Of"
)

var listeningField = Field{
	Name:                    FieldListening,
	Type:                    Int,
	SourcesBestFirstAfterAP: []string{SourceSpotifyListening},
	About:                   AboutAP,
}

func init() {
	register(Kind{
		Name:             Artist,
		DefaultTitle:     "artists",
		Statuses:         []string{Listening, Listened},
		StatusesToScore:  []string{Listening, Listened},
		StatusesShown:    []string{Listening, Listened},
		DefaultAxes:      []string{Prestige, Quality},
		RankedAxis:       Prestige,
		IdentitySource:   SourceSpotifyArtist,
		StatusComputedBy: SourceSpotify,
		UnrankedOrder:    FieldListening,
		Fields: []Field{
			{Name: FieldTitle, Type: Text, SourcesBestFirstAfterAP: []string{SourceSpotifyArtist}},
			{Name: "genres", Type: StringList, SourcesBestFirstAfterAP: []string{SourceSpotifyArtist}},
			{Name: "image", Type: Picture, Shape: Square, SourcesBestFirstAfterAP: []string{SourceSpotifyArtist}},
			{Name: "popularity", Type: Int, SourcesBestFirstAfterAP: []string{SourceSpotifyArtist}},
			{Name: "followers", Type: Int, SourcesBestFirstAfterAP: []string{SourceSpotifyArtist}},
			listeningField,
			{Name: "spotify_id", Type: Text, SourcesBestFirstAfterAP: []string{SourceSpotifyArtist}},
			{Name: "link", Type: URL, SourcesBestFirstAfterAP: []string{SourceSpotifyArtist}},
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
		IdentitySource:  SourceSpotifyAlbum,
		UnrankedOrder:   FieldListening,
		Fields: []Field{
			{Name: FieldTitle, Type: Text, SourcesBestFirstAfterAP: []string{SourceSpotifyAlbum}},
			{Name: "artists", Type: StringList, SourcesBestFirstAfterAP: []string{SourceSpotifyAlbum}},
			{Name: "released", Type: PartialDate, SourcesBestFirstAfterAP: []string{SourceSpotifyAlbum}},
			{Name: "album_type", Type: Text, SourcesBestFirstAfterAP: []string{SourceSpotifyAlbum}},
			{Name: "tracks", Type: Int, SourcesBestFirstAfterAP: []string{SourceSpotifyAlbum}},
			{Name: "label", Type: Text, SourcesBestFirstAfterAP: []string{SourceSpotifyAlbum}},
			{Name: "genres", Type: StringList, SourcesBestFirstAfterAP: []string{SourceSpotifyAlbum}, InheritedThrough: LinkBy},
			{Name: "popularity", Type: Int, SourcesBestFirstAfterAP: []string{SourceSpotifyAlbum}},
			listeningField,
			{Name: "cover", Type: Picture, Shape: Square, SourcesBestFirstAfterAP: []string{SourceSpotifyAlbum}},
			{Name: "spotify_id", Type: Text, SourcesBestFirstAfterAP: []string{SourceSpotifyAlbum}},
			{Name: "link", Type: URL, SourcesBestFirstAfterAP: []string{SourceSpotifyAlbum}},
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
		IdentitySource:  SourceSpotifyTrack,
		Fields: []Field{
			{Name: FieldTitle, Type: Text, SourcesBestFirstAfterAP: []string{SourceSpotifyTrack}},
			{Name: "artists", Type: StringList, SourcesBestFirstAfterAP: []string{SourceSpotifyTrack}},
			{Name: "duration", Type: Int, SourcesBestFirstAfterAP: []string{SourceSpotifyTrack}},
			{Name: "popularity", Type: Int, SourcesBestFirstAfterAP: []string{SourceSpotifyTrack}},
			{Name: "isrc", Type: Text, SourcesBestFirstAfterAP: []string{SourceSpotifyTrack}},
			{Name: "spotify_id", Type: Text, SourcesBestFirstAfterAP: []string{SourceSpotifyTrack}},
			{Name: "link", Type: URL, SourcesBestFirstAfterAP: []string{SourceSpotifyTrack}},
		},
	})
}
