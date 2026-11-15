package kind

import (
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
)

const (
	Game = "game"

	TagAnotherShot = "another-shot"
	TagIdled       = "idled"
)

var steamTake = Take{
	Source:        record.SourceSteamTake,
	PrivateFields: []string{"hours_at_review"},
	Fields: []Field{
		{Name: "reviewed_on", Type: PartialDate},
	},
}

func init() {
	register(Kind{
		Name:            Game,
		DefaultTitle:    "games",
		Statuses:        []string{Untriaged, "never-played", "maybe-played", "playing", "played", "finished", Dropped},
		StatusesToScore: []string{"playing", "played", "finished", Dropped},
		StatusesShown:   []string{"maybe-played", "playing", "played", "finished", Dropped},
		DefaultAxes:     []string{Prestige, Quality},
		RankedAxis:      Prestige,
		IdentitySource:  record.SourceIGDB,
		UnrankedOrder:   "playtime_minutes",
		Fields: []Field{
			{Name: FieldTitle, Type: Text, SourcesBestFirstAfterAP: []string{record.SourceIGDB, record.SourceSteamStore, record.SourceSteam}},
			{Name: "cover", Type: Picture, Shape: Wide, SourcesBestFirstAfterAP: []string{record.SourceSteamStore, record.SourceIGDB}},
			{Name: "poster", Type: Picture, Shape: Tall, SourcesBestFirstAfterAP: []string{record.SourceIGDB}},
			{Name: "released", Type: PartialDate, SourcesBestFirstAfterAP: []string{record.SourceIGDB, record.SourceSteamStore}},
			{Name: "developers", Type: StringList, SourcesBestFirstAfterAP: []string{record.SourceIGDB, record.SourceSteamStore}},
			{Name: "genres", Type: StringList, SourcesBestFirstAfterAP: []string{record.SourceIGDB, record.SourceSteamStore}},
			{Name: "steam_categories", Type: StringList, SourcesBestFirstAfterAP: []string{record.SourceSteamStore}},
			{Name: "platform", Type: Text},
			{Name: "platforms", Type: StringList, SourcesBestFirstAfterAP: []string{record.SourceIGDB}},
			{Name: "link", Type: URL},
			{Name: "steam_appid", Type: ExternalID, SourcesBestFirstAfterAP: []string{record.SourceSteam}},
			{Name: "playtime_minutes", Type: Int, SourcesBestFirstAfterAP: []string{record.SourceSteam}, Reach: OnlyAP, About: AboutAP},
			{Name: "last_played", Type: PartialDate, SourcesBestFirstAfterAP: []string{record.SourceSteam}},
			{Name: "achievements_total", Type: Int, SourcesBestFirstAfterAP: []string{record.SourceSteamAchievements}},
			{Name: "achievements_unlocked", Type: Int, SourcesBestFirstAfterAP: []string{record.SourceSteamAchievements}, About: AboutAP},
		},
		Takes: []Take{steamTake},
	})
}
