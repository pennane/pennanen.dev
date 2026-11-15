package kind

import (
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
)

const (
	Series = "series"

	Watching = "watching"
	Dropped  = "dropped"
)

func init() {
	register(Kind{
		Name:            Series,
		DefaultTitle:    "series",
		Statuses:        []string{Untriaged, Want, Watching, Watched, Dropped},
		StatusesToScore: []string{Watching, Watched, Dropped},
		StatusesShown:   []string{Watching, Watched, Dropped},
		DefaultAxes:     []string{Prestige, Quality},
		RankedAxis:      Prestige,
		IdentitySource:  record.SourceTMDBSeries,
		Fields: []Field{
			{Name: FieldTitle, Type: Text, SourcesBestFirstAfterAP: []string{record.SourceTMDBSeries, record.SourceLetterboxd}},
			{Name: "original_title", Type: Text, SourcesBestFirstAfterAP: []string{record.SourceTMDBSeries}},
			{Name: "released", Type: PartialDate, SourcesBestFirstAfterAP: []string{record.SourceTMDBSeries, record.SourceLetterboxd}},
			{Name: "poster", Type: Picture, Shape: Tall, SourcesBestFirstAfterAP: []string{record.SourceTMDBSeries}},
			{Name: "cover", Type: Picture, Shape: Wide, SourcesBestFirstAfterAP: []string{record.SourceTMDBSeries}},
			{Name: "creators", Type: StringList, SourcesBestFirstAfterAP: []string{record.SourceTMDBSeries}},
			{Name: "genres", Type: StringList, SourcesBestFirstAfterAP: []string{record.SourceTMDBSeries}},
			{Name: "countries", Type: StringList, SourcesBestFirstAfterAP: []string{record.SourceTMDBSeries}},
			{Name: "seasons", Type: Int, SourcesBestFirstAfterAP: []string{record.SourceTMDBSeries}},
			{Name: "episodes", Type: Int, SourcesBestFirstAfterAP: []string{record.SourceTMDBSeries}},
			{Name: "tmdb_id", Type: ExternalID, SourcesBestFirstAfterAP: []string{record.SourceTMDBSeries}},
			{Name: "link", Type: URL, SourcesBestFirstAfterAP: []string{record.SourceLetterboxd}},
		},
		Takes: []Take{letterboxdTake},
	})
}

func OfTMDB(t reading.TMDBType) string {
	if t == reading.SeriesOnTMDB {
		return Series
	}
	return Movie
}
