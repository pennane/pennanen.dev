package kind

import (
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"slices"
)

const (
	Game = "game"

	SteamTradingCardsCategory = "Steam Trading Cards"

	TagAnotherShot = "another-shot"
)

const (
	HintLikelyIdled       Hint = "likely-idled"
	HintLikelyPlayed      Hint = "likely-played"
	HintLikelyNeverPlayed Hint = "likely-never-played"
	HintLikelyNotAGame    Hint = "likely-not-a-game"
)

var steamSoftwareGenres = []string{
	"Utilities",
	"Design & Illustration",
	"Animation & Modeling",
	"Video Production",
	"Audio Production",
	"Photo Editing",
	"Software Training",
	"Web Publishing",
	"Game Development",
	"Accounting",
}

const (
	cardDropsUsuallyFinishWithinMinutes = 300
	clearlyPlayedAfterMinutes           = 120
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
		Takes:       []Take{steamTake},
		GuessStatus: guessGameStatus,
	})
}

func guessGameStatus(verdicts map[string]any) Hint {
	playtime, playtimeKnown := verdicts["playtime_minutes"].(int64)
	unlocked, _ := verdicts["achievements_unlocked"].(int64)
	categories, _ := verdicts["steam_categories"].([]string)
	genres, _ := verdicts["genres"].([]string)
	dropsCards := slices.Contains(categories, SteamTradingCardsCategory)
	isSoftware := slices.ContainsFunc(genres, func(g string) bool { return slices.Contains(steamSoftwareGenres, g) })
	switch {
	case isSoftware:
		return HintLikelyNotAGame
	case !playtimeKnown:
		return NoHint
	case playtime == 0:
		return HintLikelyNeverPlayed
	case unlocked > 0:
		return HintLikelyPlayed
	case dropsCards && playtime < cardDropsUsuallyFinishWithinMinutes:
		return HintLikelyIdled
	case playtime >= clearlyPlayedAfterMinutes:
		return HintLikelyPlayed
	}
	return NoHint
}
