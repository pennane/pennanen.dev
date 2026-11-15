package kind

const (
	Movie = "movie"

	SourceTMDB           = "tmdb"
	SourceLetterboxd     = "letterboxd"
	SourceLetterboxdTake = "letterboxd-take"

	Want    = "want"
	Watched = "watched"
)

var letterboxdTake = Take{
	Source:        SourceLetterboxdTake,
	PrivateFields: []string{"on_watchlist_since"},
	Fields: []Field{
		{Name: "rating", Type: Number},
		{Name: "last_watched", Type: PartialDate},
	},
}

func init() {
	register(Kind{
		Name:            Movie,
		DefaultTitle:    "movies",
		Statuses:        []string{Untriaged, Want, Watched},
		StatusesToScore: []string{Watched},
		StatusesShown:   []string{Watched},
		DefaultAxes:     []string{Prestige, Quality},
		RankedAxis:      Prestige,
		IdentitySource:  SourceTMDB,
		Fields: []Field{
			{Name: FieldTitle, Type: Text, SourcesBestFirstAfterAP: []string{SourceTMDB, SourceLetterboxd}},
			{Name: "original_title", Type: Text, SourcesBestFirstAfterAP: []string{SourceTMDB}},
			{Name: "released", Type: PartialDate, SourcesBestFirstAfterAP: []string{SourceTMDB, SourceLetterboxd}},
			{Name: "poster", Type: Picture, Shape: Tall, SourcesBestFirstAfterAP: []string{SourceTMDB}},
			{Name: "cover", Type: Picture, Shape: Wide, SourcesBestFirstAfterAP: []string{SourceTMDB}},
			{Name: "directors", Type: StringList, SourcesBestFirstAfterAP: []string{SourceTMDB}},
			{Name: "genres", Type: StringList, SourcesBestFirstAfterAP: []string{SourceTMDB}},
			{Name: "countries", Type: StringList, SourcesBestFirstAfterAP: []string{SourceTMDB}},
			{Name: "runtime_minutes", Type: Int, SourcesBestFirstAfterAP: []string{SourceTMDB}},
			{Name: "tmdb_id", Type: ExternalID, SourcesBestFirstAfterAP: []string{SourceTMDB}},
			{Name: "link", Type: URL, SourcesBestFirstAfterAP: []string{SourceLetterboxd}},
		},
		Takes: []Take{letterboxdTake},
	})
}
