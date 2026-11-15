package kind

import (
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
)

const (
	Movie = "movie"

	Want    = "want"
	Watched = "watched"
)

var letterboxdTake = Take{
	Source:        record.SourceLetterboxdTake,
	PrivateFields: []string{"on_watchlist_since", "on_watched_shelf_since"},
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
		IdentitySource:  record.SourceTMDB,
		Fields: []Field{
			{Name: FieldTitle, Type: Text, SourcesBestFirstAfterAP: []string{record.SourceTMDB, record.SourceLetterboxd}},
			{Name: "original_title", Type: Text, SourcesBestFirstAfterAP: []string{record.SourceTMDB}},
			{Name: "released", Type: PartialDate, SourcesBestFirstAfterAP: []string{record.SourceTMDB, record.SourceLetterboxd}},
			{Name: "poster", Type: Picture, Shape: Tall, SourcesBestFirstAfterAP: []string{record.SourceTMDB}},
			{Name: "cover", Type: Picture, Shape: Wide, SourcesBestFirstAfterAP: []string{record.SourceTMDB}},
			{Name: "directors", Type: StringList, SourcesBestFirstAfterAP: []string{record.SourceTMDB}},
			{Name: "genres", Type: StringList, SourcesBestFirstAfterAP: []string{record.SourceTMDB}},
			{Name: "countries", Type: StringList, SourcesBestFirstAfterAP: []string{record.SourceTMDB}},
			{Name: "runtime_minutes", Type: Int, SourcesBestFirstAfterAP: []string{record.SourceTMDB}},
			{Name: "tmdb_id", Type: ExternalID, SourcesBestFirstAfterAP: []string{record.SourceTMDB}},
			{Name: "link", Type: URL, SourcesBestFirstAfterAP: []string{record.SourceLetterboxd}},
		},
		Takes: []Take{letterboxdTake},
	})
}
