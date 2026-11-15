package letterboxd

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/rule"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

const RuleLetterboxd = "letterboxd"

var Proposer = rule.Proposer{Source: record.SourceLetterboxd, Propose: loadAndPropose}

type shelvedTake struct {
	itemID              int64
	LastWatched         string `json:"last_watched"`
	OnWatchlistSince    string `json:"on_watchlist_since"`
	OnWatchedShelfSince string `json:"on_watched_shelf_since"`
}

func propose(takes []shelvedTake) []rule.Proposal {
	var proposals []rule.Proposal
	for _, take := range takesByItem(takes) {
		proposals = append(proposals, proposalsFor(take)...)
	}
	return proposals
}

func takesByItem(takes []shelvedTake) []shelvedTake {
	byItem := map[int64]shelvedTake{}
	for _, take := range takes {
		merged := byItem[take.itemID]
		merged.itemID = take.itemID
		merged.LastWatched = later(merged.LastWatched, take.LastWatched)
		merged.OnWatchlistSince = later(merged.OnWatchlistSince, take.OnWatchlistSince)
		merged.OnWatchedShelfSince = later(merged.OnWatchedShelfSince, take.OnWatchedShelfSince)
		byItem[take.itemID] = merged
	}
	merged := make([]shelvedTake, 0, len(byItem))
	for _, take := range byItem {
		merged = append(merged, take)
	}
	slices.SortFunc(merged, func(a, b shelvedTake) int { return cmp.Compare(a.itemID, b.itemID) })
	return merged
}

func later(a, b string) string {
	if rule.Normalized(b) > rule.Normalized(a) {
		return b
	}
	return a
}

func proposalsFor(take shelvedTake) []rule.Proposal {
	apply := func(subject rule.Subject, value, at string) rule.Proposal {
		return rule.Proposal{ItemID: take.itemID, Rule: RuleLetterboxd, Subject: subject, Value: value, Strength: rule.Applying, ChangedAt: at}
	}
	logged := apply(rule.SubjectEntry, rule.EntryLogged, "")
	watchedAt := later(take.LastWatched, take.OnWatchedShelfSince)
	switch {
	case watchedAt != "" && take.OnWatchlistSince != "":
		return []rule.Proposal{logged, apply(rule.SubjectStatus, kind.Watched, watchedAt), apply(rule.SubjectTag, kind.TagAnotherShot, take.OnWatchlistSince)}
	case watchedAt != "":
		return []rule.Proposal{logged, apply(rule.SubjectStatus, kind.Watched, watchedAt)}
	case take.OnWatchlistSince != "":
		return []rule.Proposal{logged, apply(rule.SubjectStatus, kind.Want, take.OnWatchlistSince)}
	}
	return nil
}

func loadAndPropose(ctx context.Context, q record.Queryer) ([]rule.Proposal, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT item_id, reading FROM claim_reading WHERE source = ? ORDER BY item_id`, record.SourceLetterboxdTake)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	takes := []shelvedTake{}
	for rows.Next() {
		var take shelvedTake
		var stored string
		if err := rows.Scan(&take.itemID, &stored); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(stored), &take); err != nil {
			return nil, fmt.Errorf("letterboxd take of item %d: %w", take.itemID, err)
		}
		takes = append(takes, take)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return propose(takes), nil
}
