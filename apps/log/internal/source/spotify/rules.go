package spotify

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/rule"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

const (
	RuleArtist = "spotify-artist"
	RuleAlbum  = "spotify-album"
)

const secondLayout = "2006-01-02T15:04:05Z"

var Proposer = rule.Proposer{Source: record.SourceSpotify, Propose: loadAndPropose}

type observation struct {
	itemID     int64
	signal     string
	key        string
	observedAt string
}

type mainArtist struct {
	albumID  int64
	artistID int64
}

type heardTimes struct {
	firstSaved string
	firstHeard string
}

type listedTimes struct {
	firstListed string
	shortTerm   map[string]struct{}
}

func propose(observations []observation, mainArtists []mainArtist) []rule.Proposal {
	albums := map[int64]heardTimes{}
	artists := map[int64]listedTimes{}
	var snapshots []string
	shortTermPrefix := fmt.Sprintf("%s:%s:", record.SignalTopArtist, ShortTerm)
	shortTermListingPrefix := fmt.Sprintf("%s:%s:", record.SignalTopArtistsListed, ShortTerm)
	for _, o := range observations {
		at := atSecond(o.observedAt)
		switch o.signal {
		case record.SignalTopArtistsListed:
			if strings.HasPrefix(o.key, shortTermListingPrefix) {
				snapshots = append(snapshots, at)
			}
		case record.SignalSaved:
			album := albums[o.itemID]
			album.firstSaved = earlier(album.firstSaved, at)
			albums[o.itemID] = album
		case record.SignalTopTrack, record.SignalPlay:
			album := albums[o.itemID]
			album.firstHeard = earlier(album.firstHeard, at)
			albums[o.itemID] = album
		case record.SignalTopArtist:
			artist := artists[o.itemID]
			artist.firstListed = earlier(artist.firstListed, at)
			if strings.HasPrefix(o.key, shortTermPrefix) {
				if artist.shortTerm == nil {
					artist.shortTerm = map[string]struct{}{}
				}
				artist.shortTerm[at] = struct{}{}
				snapshots = append(snapshots, at)
			}
			artists[o.itemID] = artist
		}
	}
	snapshots = slices.Compact(slices.Sorted(slices.Values(snapshots)))
	firstSavedBehind := map[int64]string{}
	for _, link := range mainArtists {
		if saved := albums[link.albumID].firstSaved; saved != "" {
			firstSavedBehind[link.artistID] = earlier(firstSavedBehind[link.artistID], saved)
		}
	}
	byItem := map[int64][]rule.Proposal{}
	for itemID, album := range albums {
		byItem[itemID] = append(byItem[itemID], albumProposals(itemID, album)...)
	}
	for itemID, artist := range artists {
		byItem[itemID] = append(byItem[itemID], artistProposals(itemID, artist, snapshots, firstSavedBehind[itemID])...)
	}
	for itemID, saved := range firstSavedBehind {
		if _, listed := artists[itemID]; !listed {
			byItem[itemID] = append(byItem[itemID], artistProposals(itemID, listedTimes{}, snapshots, saved)...)
		}
	}
	var proposals []rule.Proposal
	for _, itemID := range slices.Sorted(maps.Keys(byItem)) {
		proposals = append(proposals, byItem[itemID]...)
	}
	return proposals
}

func albumProposals(itemID int64, album heardTimes) []rule.Proposal {
	var proposals []rule.Proposal
	if album.firstSaved != "" {
		proposals = append(proposals, rule.Proposal{ItemID: itemID, Rule: RuleAlbum, Subject: rule.SubjectEntry, Value: rule.EntryLogged, Strength: rule.Applying})
	}
	switch {
	case album.firstHeard != "":
		proposals = append(proposals, applyingStatus(itemID, RuleAlbum, kind.Listened, album.firstHeard))
	case album.firstSaved != "":
		proposals = append(proposals, applyingStatus(itemID, RuleAlbum, kind.Want, album.firstSaved))
	}
	return proposals
}

func artistProposals(itemID int64, artist listedTimes, snapshots []string, firstSavedBehind string) []rule.Proposal {
	if artist.firstListed == "" && firstSavedBehind == "" {
		return nil
	}
	logged := rule.Proposal{ItemID: itemID, Rule: RuleArtist, Subject: rule.SubjectEntry, Value: rule.EntryLogged, Strength: rule.Applying}
	return []rule.Proposal{logged, artistStatus(itemID, artist, snapshots, firstSavedBehind)}
}

func artistStatus(itemID int64, artist listedTimes, snapshots []string, firstSavedBehind string) rule.Proposal {
	inSnapshot := func(i int) bool {
		_, in := artist.shortTerm[snapshots[i]]
		return in
	}
	newest := len(snapshots) - 1
	if newest >= 0 && inSnapshot(newest) {
		runStart := newest
		for runStart > 0 && inSnapshot(runStart-1) {
			runStart--
		}
		return applyingStatus(itemID, RuleArtist, kind.Listening, snapshots[runStart])
	}
	for i := newest; i >= 0; i-- {
		if inSnapshot(i) {
			return applyingStatus(itemID, RuleArtist, kind.Listened, snapshots[i+1])
		}
	}
	return applyingStatus(itemID, RuleArtist, kind.Listened, earlier(artist.firstListed, firstSavedBehind))
}

func applyingStatus(itemID int64, ruleName, status, at string) rule.Proposal {
	return rule.Proposal{ItemID: itemID, Rule: ruleName, Subject: rule.SubjectStatus, Value: status, Strength: rule.Applying, ChangedAt: at}
}

func earlier(a, b string) string {
	if a == "" || (b != "" && b < a) {
		return b
	}
	return a
}

func atSecond(at string) string {
	t, err := time.Parse(time.RFC3339Nano, at)
	if err != nil {
		return rule.Normalized(at)
	}
	return t.UTC().Format(secondLayout)
}

func loadAndPropose(ctx context.Context, q record.Queryer) ([]rule.Proposal, error) {
	observations, err := spotifyObservations(ctx, q)
	if err != nil {
		return nil, err
	}
	listings, err := topArtistListings(ctx, q)
	if err != nil {
		return nil, err
	}
	observations = append(observations, listings...)
	links, err := mainArtistsOfSaved(ctx, q)
	if err != nil {
		return nil, err
	}
	return propose(observations, links), nil
}

func spotifyObservations(ctx context.Context, q record.Queryer) ([]observation, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT item_id, signal, key, observed_at FROM observation
		WHERE source = ? AND signal IN (?, ?, ?, ?)
		ORDER BY item_id, observed_at, id`,
		ObservationSource, record.SignalTopArtist, record.SignalTopTrack, record.SignalPlay, record.SignalSaved)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	observations := []observation{}
	for rows.Next() {
		var o observation
		if err := rows.Scan(&o.itemID, &o.signal, &o.key, &o.observedAt); err != nil {
			return nil, err
		}
		observations = append(observations, o)
	}
	return observations, rows.Err()
}

func topArtistListings(ctx context.Context, q record.Queryer) ([]observation, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT signal, key, observed_at FROM record_observation
		WHERE source = ? AND signal = ?
		ORDER BY observed_at, id`,
		ObservationSource, record.SignalTopArtistsListed)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	listings := []observation{}
	for rows.Next() {
		var o observation
		if err := rows.Scan(&o.signal, &o.key, &o.observedAt); err != nil {
			return nil, err
		}
		listings = append(listings, o)
	}
	return listings, rows.Err()
}

func mainArtistsOfSaved(ctx context.Context, q record.Queryer) ([]mainArtist, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT DISTINCT from_id, to_id FROM relation
		WHERE type = ? AND from_id IN (SELECT item_id FROM observation WHERE source = ? AND signal = ?)
		ORDER BY from_id, to_id`,
		kind.LinkBy, ObservationSource, record.SignalSaved)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	links := []mainArtist{}
	for rows.Next() {
		var link mainArtist
		if err := rows.Scan(&link.albumID, &link.artistID); err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, rows.Err()
}
