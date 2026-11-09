package kind

import "slices"

const (
	SourceSteam             = "steam"
	SourceSteamStore        = "steam-store"
	SourceSteamAchievements = "steam-achievements"
	SourceIGDB              = "igdb"

	SteamTradingCardsCategory = "Steam Trading Cards"
)

const (
	HintLikelyIdled       Hint = "likely-idled"
	HintLikelyPlayed      Hint = "likely-played"
	HintLikelyNeverPlayed Hint = "likely-never-played"
)

const (
	cardDropsUsuallyFinishWithinMinutes = 300
	clearlyPlayedAfterMinutes           = 120
)

func init() {
	register(Kind{
		Name:             "game",
		Statuses:         []string{Untriaged, "never-played", "idled", "playing", "played", "finished", "dropped"},
		Axes:             []string{Enjoyment, Quality},
		UndeclaredFields: RejectUndeclaredFields,
		Fields: []Field{
			{Name: FieldTitle, Type: Text, SourcesBestFirstAfterAP: []string{SourceIGDB, SourceSteamStore, SourceSteam}},
			{Name: "cover", Type: URL, SourcesBestFirstAfterAP: []string{SourceIGDB, SourceSteamStore}},
			{Name: "released", Type: PartialDate, SourcesBestFirstAfterAP: []string{SourceIGDB, SourceSteamStore}},
			{Name: "description", Type: Text, SourcesBestFirstAfterAP: []string{SourceIGDB, SourceSteamStore}},
			{Name: "developers", Type: StringList, SourcesBestFirstAfterAP: []string{SourceIGDB, SourceSteamStore}},
			{Name: "genres", Type: StringList, SourcesBestFirstAfterAP: []string{SourceIGDB, SourceSteamStore}},
			{Name: "steam_categories", Type: StringList, SourcesBestFirstAfterAP: []string{SourceSteamStore}},
			{Name: "platform", Type: Text},
			{Name: "link", Type: URL},
			{Name: "steam_appid", Type: Int, SourcesBestFirstAfterAP: []string{SourceSteam}},
			{Name: "playtime_minutes", Type: Int, SourcesBestFirstAfterAP: []string{SourceSteam}},
			{Name: "last_played", Type: PartialDate, SourcesBestFirstAfterAP: []string{SourceSteam}},
			{Name: "achievements_total", Type: Int, SourcesBestFirstAfterAP: []string{SourceSteamAchievements}},
			{Name: "achievements_unlocked", Type: Int, SourcesBestFirstAfterAP: []string{SourceSteamAchievements}},
		},
		GuessStatus: guessGameStatus,
	})

	register(Kind{
		Name:             "thing",
		Statuses:         []string{Untriaged, "seen"},
		Axes:             []string{Enjoyment, Quality},
		UndeclaredFields: AcceptUndeclaredFields,
		Fields: []Field{
			{Name: FieldTitle, Type: Text},
			{Name: "description", Type: Text},
			{Name: "link", Type: URL},
			{Name: "image", Type: URL},
		},
	})
}

func guessGameStatus(verdicts map[string]any) Hint {
	playtime, playtimeKnown := verdicts["playtime_minutes"].(int64)
	unlocked, _ := verdicts["achievements_unlocked"].(int64)
	categories, _ := verdicts["steam_categories"].([]string)
	dropsCards := slices.Contains(categories, SteamTradingCardsCategory)
	switch {
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
