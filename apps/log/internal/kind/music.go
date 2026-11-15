package kind

const (
	Artist = "artist"
	Album  = "album"

	SourceSpotify       = "spotify"
	SourceSpotifyArtist = "spotify-artist"
	SourceSpotifyAlbum  = "spotify-album"

	Listening = "listening"
	Listened  = "listened"

	RelationBy = "by"
)

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
		Fields: []Field{
			{Name: FieldTitle, Type: Text, SourcesBestFirstAfterAP: []string{SourceSpotifyArtist}},
			{Name: "genres", Type: StringList, SourcesBestFirstAfterAP: []string{SourceSpotifyArtist}},
			{Name: "image", Type: URL, SourcesBestFirstAfterAP: []string{SourceSpotifyArtist}},
			{Name: "popularity", Type: Int, SourcesBestFirstAfterAP: []string{SourceSpotifyArtist}},
			{Name: "followers", Type: Int, SourcesBestFirstAfterAP: []string{SourceSpotifyArtist}},
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
		Fields: []Field{
			{Name: FieldTitle, Type: Text, SourcesBestFirstAfterAP: []string{SourceSpotifyAlbum}},
			{Name: "artists", Type: StringList, SourcesBestFirstAfterAP: []string{SourceSpotifyAlbum}},
			{Name: "released", Type: PartialDate, SourcesBestFirstAfterAP: []string{SourceSpotifyAlbum}},
			{Name: "album_type", Type: Text, SourcesBestFirstAfterAP: []string{SourceSpotifyAlbum}},
			{Name: "tracks", Type: Int, SourcesBestFirstAfterAP: []string{SourceSpotifyAlbum}},
			{Name: "label", Type: Text, SourcesBestFirstAfterAP: []string{SourceSpotifyAlbum}},
			{Name: "genres", Type: StringList, SourcesBestFirstAfterAP: []string{SourceSpotifyAlbum}},
			{Name: "popularity", Type: Int, SourcesBestFirstAfterAP: []string{SourceSpotifyAlbum}},
			{Name: "cover", Type: URL, SourcesBestFirstAfterAP: []string{SourceSpotifyAlbum}},
			{Name: "spotify_id", Type: Text, SourcesBestFirstAfterAP: []string{SourceSpotifyAlbum}},
			{Name: "link", Type: URL, SourcesBestFirstAfterAP: []string{SourceSpotifyAlbum}},
		},
	})
}
