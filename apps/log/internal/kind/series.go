package kind

const (
	Series           = "series"
	SourceTMDBSeries = "tmdb-tv"

	Dropped = "dropped"
)

func init() {
	register(Kind{
		Name:            Series,
		DefaultTitle:    "series",
		Statuses:        []string{Untriaged, Want, Watched, Dropped},
		StatusesToScore: []string{Watched, Dropped},
		StatusesShown:   []string{Watched, Dropped},
		DefaultAxes:     []string{Prestige, Quality},
		RankedAxis:      Prestige,
		IdentitySource:  SourceTMDBSeries,
		Fields: []Field{
			{Name: FieldTitle, Type: Text, SourcesBestFirstAfterAP: []string{SourceTMDBSeries, SourceLetterboxd}},
			{Name: "original_title", Type: Text, SourcesBestFirstAfterAP: []string{SourceTMDBSeries}},
			{Name: "released", Type: PartialDate, SourcesBestFirstAfterAP: []string{SourceTMDBSeries, SourceLetterboxd}},
			{Name: "poster", Type: URL, SourcesBestFirstAfterAP: []string{SourceTMDBSeries}},
			{Name: "cover", Type: URL, SourcesBestFirstAfterAP: []string{SourceTMDBSeries}},
			{Name: "creators", Type: StringList, SourcesBestFirstAfterAP: []string{SourceTMDBSeries}},
			{Name: "genres", Type: StringList, SourcesBestFirstAfterAP: []string{SourceTMDBSeries}},
			{Name: "countries", Type: StringList, SourcesBestFirstAfterAP: []string{SourceTMDBSeries}},
			{Name: "seasons", Type: Int, SourcesBestFirstAfterAP: []string{SourceTMDBSeries}},
			{Name: "episodes", Type: Int, SourcesBestFirstAfterAP: []string{SourceTMDBSeries}},
			{Name: "tmdb_id", Type: ExternalID, SourcesBestFirstAfterAP: []string{SourceTMDBSeries}},
			{Name: "link", Type: URL, SourcesBestFirstAfterAP: []string{SourceLetterboxd}},
		},
		Takes: []Take{letterboxdTake},
	})
}
