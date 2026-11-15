package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
	"github.com/pennane/pennanen.dev/apps/log/internal/world"
)

const (
	apItemPrefix   = "item/"
	apFieldsPrefix = "fields/"
	apMatchPrefix  = "match/"
	apLinkPrefix   = "link/"

	handMadeIDBytes = 6
)

type linkEdit string

const (
	linkAdded      linkEdit = "added"
	linkSuppressed linkEdit = "suppressed"
)

type apFields struct {
	Kind   string         `json:"kind,omitempty"`
	Fields map[string]any `json:"fields"`
}

type apMatchAnswer struct {
	To *string `json:"to"`
}

type apLink struct {
	Edit linkEdit `json:"edit"`
	From string   `json:"from"`
	Type string   `json:"type"`
	To   string   `json:"to"`
}

type matchedRecord struct {
	About []string `json:"about"`
}

type placement string

const (
	onTheItem     placement = "on-the-item"
	onAnotherItem placement = "on-another-item"
	nowhere       placement = "nowhere"
)

func Open(db *sql.DB) *Store {
	records := record.New(db, verdict.ReadRecord, verdict.ReadersVersion)
	return New(db, records, world.New(db, records))
}

func identityOf(source, externalID string) record.Identity {
	return record.Identity{Source: source, ExternalID: externalID}
}

func apIdentity(prefix string, parts ...string) record.Identity {
	return identityOf(kind.SourceAP, prefix+strings.Join(parts, "/"))
}

func handMadeIdentity() (record.Identity, error) {
	random := make([]byte, handMadeIDBytes)
	if _, err := rand.Read(random); err != nil {
		return record.Identity{}, err
	}
	return apIdentity(apItemPrefix, hex.EncodeToString(random)), nil
}

func matchedIdentity(id record.Identity) record.Identity {
	return identityOf(kind.SourceMatched, id.String())
}

func matchAnswerIdentity(source string, key record.Identity) record.Identity {
	return apIdentity(apMatchPrefix, source, key.String())
}

func matchAnswerPrefix(source string) string {
	return apMatchPrefix + source + "/"
}

func linkEditIdentity(from record.Identity, linkType string, to record.Identity) record.Identity {
	return apIdentity(apLinkPrefix, from.String(), linkType, to.String())
}

func (s *Store) keyOf(ctx context.Context, q queryer, itemID int64) (record.Identity, error) {
	key, err := s.world.KeyOf(ctx, q, itemID)
	if errors.Is(err, world.ErrNotFound) {
		return record.Identity{}, fmt.Errorf("%w: item %d", ErrNotFound, itemID)
	}
	return key, err
}

func (s *Store) itemOf(ctx context.Context, q queryer, id record.Identity) (int64, error) {
	itemID, err := s.world.ItemOf(ctx, q, id)
	if errors.Is(err, world.ErrNotFound) {
		return 0, ErrNotFound
	}
	return itemID, err
}

func (s *Store) placementOf(ctx context.Context, q queryer, itemID int64, id record.Identity) (placement, error) {
	owner, err := s.itemOf(ctx, q, id)
	switch {
	case errors.Is(err, ErrNotFound):
		return nowhere, nil
	case err != nil:
		return "", err
	case owner == itemID:
		return onTheItem, nil
	}
	return onAnotherItem, nil
}

type landing string

const (
	alreadyOnTheItem   landing = "already-on-the-item"
	derivedOntoTheItem landing = "derived-onto-the-item"
	needsAMatch        landing = "needs-a-match"
	landsOnAnotherItem landing = "on-another-item"
)

func (s *Store) landingOf(ctx context.Context, q queryer, itemID int64, id record.Identity, raw json.RawMessage) (landing, error) {
	place, err := s.placementOf(ctx, q, itemID, id)
	switch {
	case err != nil:
		return "", err
	case place == onTheItem:
		return alreadyOnTheItem, nil
	case place == onAnotherItem:
		return landsOnAnotherItem, conflictOver(id)
	}
	reached := needsAMatch
	for _, about := range aboutOf(id, raw) {
		place, err := s.placementOf(ctx, q, itemID, about)
		switch {
		case err != nil:
			return "", err
		case place == onAnotherItem:
			return landsOnAnotherItem, conflictOver(about)
		case place == onTheItem:
			reached = derivedOntoTheItem
		}
	}
	return reached, nil
}

func conflictOver(id record.Identity) error {
	return fmt.Errorf("%w: another item already has %s", ErrConflict, id)
}

func aboutOf(id record.Identity, raw json.RawMessage) []record.Identity {
	reading, readability := verdict.ReadRecord(id.Source, raw)
	if readability == record.Unreadable {
		return nil
	}
	var values map[string]any
	if err := json.Unmarshal(reading, &values); err != nil {
		return nil
	}
	return verdict.Derive(id, values).About
}

func (s *Store) requireLanded(ctx context.Context, q queryer, itemID int64, id record.Identity) error {
	place, err := s.placementOf(ctx, q, itemID, id)
	if err != nil || place == onTheItem {
		return err
	}
	return fmt.Errorf("%w: %s settled away from item %d", ErrConflict, id, itemID)
}

func (s *Store) put(ctx context.Context, q queryer, id record.Identity, data any) (record.Change, error) {
	raw, err := asJSON(data)
	if err != nil {
		return "", err
	}
	return s.records.PutClaim(ctx, q, id, raw)
}

func (s *Store) settle(ctx context.Context, tx *sql.Tx, dirty ...record.Identity) error {
	report, err := s.world.Settle(ctx, tx, dirty)
	if err != nil {
		return err
	}
	for _, m := range report.Moves {
		if err := applyMove(ctx, tx, m); err != nil {
			return err
		}
	}
	return nil
}

func applyMove(ctx context.Context, tx *sql.Tx, m world.Move) error {
	if from, known := kind.Get(m.From); known {
		if err := leaveRanking(ctx, tx, from, m.ItemID); err != nil {
			return err
		}
	}
	to, known := kind.Get(m.To)
	if !known {
		return fmt.Errorf("item %d moved to unknown kind %q", m.ItemID, m.To)
	}
	if err := requireStatusIn(ctx, tx, to, m.ItemID); err != nil {
		return err
	}
	return joinRankingAtBottom(ctx, tx, to, m.ItemID)
}

func requireStatusIn(ctx context.Context, tx *sql.Tx, to kind.Kind, itemID int64) error {
	var status string
	if err := tx.QueryRowContext(ctx, `
		SELECT coalesce(e.status, ?) FROM item i LEFT JOIN entry e ON e.item_id = i.id WHERE i.id = ?`,
		kind.Untriaged, itemID).Scan(&status); err != nil {
		return err
	}
	if !to.HasStatus(status) {
		return fmt.Errorf("%w: %s has no status %q to move item %d into", ErrInvalid, to.Name, status, itemID)
	}
	return nil
}

func (s *Store) putClaimOn(ctx context.Context, tx *sql.Tx, itemID int64, id record.Identity, raw json.RawMessage) error {
	landing, err := s.landingOf(ctx, tx, itemID, id, raw)
	if err != nil {
		return err
	}
	change, err := s.records.PutClaim(ctx, tx, id, raw)
	if err != nil {
		return err
	}
	dirty := []record.Identity{id}
	switch landing {
	case alreadyOnTheItem:
		if change == record.Unchanged {
			return nil
		}
	case needsAMatch:
		matched, err := s.putMatched(ctx, tx, itemID, id)
		if err != nil {
			return err
		}
		dirty = append(dirty, matched)
	}
	if err := s.settle(ctx, tx, dirty...); err != nil {
		return err
	}
	return s.requireLanded(ctx, tx, itemID, id)
}

func (s *Store) putMatched(ctx context.Context, q queryer, itemID int64, id record.Identity) (record.Identity, error) {
	key, err := s.keyOf(ctx, q, itemID)
	if err != nil {
		return record.Identity{}, err
	}
	matched := matchedIdentity(id)
	_, err = s.put(ctx, q, matched, matchedRecord{About: []string{key.String(), id.String()}})
	return matched, err
}

func claimsOn(ctx context.Context, q queryer, itemID int64) ([]record.Identity, error) {
	rows, err := q.QueryContext(ctx, `SELECT source, external_id FROM claim WHERE item_id = ? ORDER BY source, external_id`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []record.Identity{}
	for rows.Next() {
		var id record.Identity
		if err := rows.Scan(&id.Source, &id.ExternalID); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func claimsFrom(ctx context.Context, q queryer, itemID int64, sources []string, keep record.Identity) ([]record.Identity, error) {
	ids, err := claimsOn(ctx, q, itemID)
	if err != nil {
		return nil, err
	}
	return slices.DeleteFunc(ids, func(id record.Identity) bool {
		return id == keep || !slices.Contains(sources, id.Source)
	}), nil
}

func (s *Store) takeOffTheItem(ctx context.Context, q queryer, ids []record.Identity) error {
	removed := []record.Identity{}
	for _, id := range ids {
		for _, claim := range []record.Identity{id, matchedIdentity(id)} {
			if _, err := s.records.RemoveClaim(ctx, q, claim); err != nil {
				return err
			}
			removed = append(removed, claim)
		}
	}
	for _, id := range removed {
		if _, err := s.world.Release(ctx, q, id); err != nil {
			return err
		}
	}
	return nil
}

type presence string

const (
	present presence = "present"
	absent  presence = "absent"
)

func storedData(ctx context.Context, q queryer, id record.Identity, into any) (presence, error) {
	var raw string
	err := q.QueryRowContext(ctx, `SELECT data FROM record_claim WHERE source = ? AND external_id = ?`, id.Source, id.ExternalID).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return absent, nil
	}
	if err != nil {
		return "", err
	}
	return present, json.Unmarshal([]byte(raw), into)
}
