package steam

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/rule"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/verdict"
)

const (
	RuleOwned             = "steam-owned"
	RuleLikelyPlayed      = "likely-played"
	RuleLikelyNeverPlayed = "likely-never-played"
	RuleLikelyIdled       = "likely-idled"
	RuleLikelyNotAGame    = "likely-not-a-game"
)

const TagIdled = "idled"

const (
	statusPlayed      = "played"
	statusNeverPlayed = "never-played"
)

const noHint = ""

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

var Proposer = rule.Proposer{Source: record.SourceSteam, Propose: loadAndPropose}

type gameSigns struct {
	playtimeMinutes      *int64
	achievementsUnlocked int64
	steamCategories      []string
	genres               []string
}

type loggedGame struct {
	itemID int64
	status string
	signs  gameSigns
}

func propose(owned []int64, logged []loggedGame) []rule.Proposal {
	proposals := []rule.Proposal{}
	for _, itemID := range owned {
		proposals = append(proposals, rule.Proposal{ItemID: itemID, Rule: RuleOwned, Subject: rule.SubjectEntry, Value: rule.EntryLogged, Strength: rule.Applying})
	}
	for _, game := range logged {
		if game.status != kind.Untriaged {
			continue
		}
		proposals = append(proposals, hintProposals(game.itemID, hintFor(game.signs))...)
	}
	if len(proposals) == 0 {
		return nil
	}
	return proposals
}

func hintFor(signs gameSigns) string {
	dropsCards := slices.Contains(signs.steamCategories, kind.SteamTradingCardsCategory)
	isSoftware := slices.ContainsFunc(signs.genres, func(g string) bool { return slices.Contains(steamSoftwareGenres, g) })
	switch {
	case isSoftware:
		return RuleLikelyNotAGame
	case signs.playtimeMinutes == nil:
		return noHint
	case *signs.playtimeMinutes == 0:
		return RuleLikelyNeverPlayed
	case signs.achievementsUnlocked > 0:
		return RuleLikelyPlayed
	case dropsCards && *signs.playtimeMinutes < cardDropsUsuallyFinishWithinMinutes:
		return RuleLikelyIdled
	case *signs.playtimeMinutes >= clearlyPlayedAfterMinutes:
		return RuleLikelyPlayed
	}
	return noHint
}

func hintProposals(itemID int64, hint string) []rule.Proposal {
	suggest := func(subject rule.Subject, value string) rule.Proposal {
		return rule.Proposal{ItemID: itemID, Rule: hint, Subject: subject, Value: value, Strength: rule.Suggesting}
	}
	switch hint {
	case RuleLikelyPlayed:
		return []rule.Proposal{suggest(rule.SubjectStatus, statusPlayed)}
	case RuleLikelyNeverPlayed:
		return []rule.Proposal{suggest(rule.SubjectStatus, statusNeverPlayed)}
	case RuleLikelyIdled:
		return []rule.Proposal{suggest(rule.SubjectTag, TagIdled), suggest(rule.SubjectStatus, statusNeverPlayed)}
	case RuleLikelyNotAGame:
		return []rule.Proposal{suggest(rule.SubjectDismissal, kind.Dismissed)}
	}
	return nil
}

func loadAndPropose(ctx context.Context, q record.Queryer) ([]rule.Proposal, error) {
	owned, err := ownedGames(ctx, q)
	if err != nil {
		return nil, err
	}
	logged, err := loggedGames(ctx, q)
	if err != nil {
		return nil, err
	}
	return propose(owned, logged), nil
}

func ownedGames(ctx context.Context, q record.Queryer) ([]int64, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT DISTINCT item_id FROM claim_reading WHERE source = ? ORDER BY item_id`, record.SourceSteam)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	owned := []int64{}
	for rows.Next() {
		var itemID int64
		if err := rows.Scan(&itemID); err != nil {
			return nil, err
		}
		owned = append(owned, itemID)
	}
	return owned, rows.Err()
}

func loggedGames(ctx context.Context, q record.Queryer) ([]loggedGame, error) {
	game, known := kind.Get(kind.Game)
	if !known {
		return nil, fmt.Errorf("kind %s is not registered", kind.Game)
	}
	rows, err := q.QueryContext(ctx, `
		SELECT s.item_id, s.status, coalesce(v.verdicts, '{}')
		FROM shown_entry e
		JOIN item i ON i.id = e.item_id
		JOIN entry_status s ON s.item_id = e.item_id
		LEFT JOIN item_verdict v ON v.item_id = e.item_id
		WHERE i.kind = ?
		ORDER BY s.item_id`, kind.Game)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	logged := []loggedGame{}
	for rows.Next() {
		var g loggedGame
		var stored string
		if err := rows.Scan(&g.itemID, &g.status, &stored); err != nil {
			return nil, err
		}
		verdicts, err := verdict.DecodeVerdicts(game, json.RawMessage(stored))
		if err != nil {
			return nil, fmt.Errorf("verdicts of item %d: %w", g.itemID, err)
		}
		g.signs = signsFrom(verdicts.Values)
		logged = append(logged, g)
	}
	return logged, rows.Err()
}

func signsFrom(values map[string]any) gameSigns {
	var signs gameSigns
	if playtime, known := values["playtime_minutes"].(int64); known {
		signs.playtimeMinutes = &playtime
	}
	signs.achievementsUnlocked, _ = values["achievements_unlocked"].(int64)
	signs.steamCategories, _ = values["steam_categories"].([]string)
	signs.genres, _ = values["genres"].([]string)
	return signs
}
