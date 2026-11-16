package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
	"github.com/pennane/pennanen.dev/apps/log/internal/rule"
	"github.com/pennane/pennanen.dev/apps/log/internal/take"
	"github.com/pennane/pennanen.dev/apps/log/internal/world"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
	"strings"
)

type Store struct {
	db        *sql.DB
	records   *record.Store
	world     *world.World
	proposers []rule.Proposer
}

func New(db *sql.DB, records *record.Store, world *world.World, proposers []rule.Proposer) *Store {
	return &Store{db: db, records: records, world: world, proposers: proposers}
}

type queryer = record.Queryer

func (s *Store) inTx(ctx context.Context, work func(tx *sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := work(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) CreateItem(ctx context.Context, kindName, title string, override map[string]any) (int64, error) {
	k, known := kind.Get(kindName)
	if !known {
		return 0, fmt.Errorf("%w: unknown kind %q", ErrInvalid, kindName)
	}
	if strings.TrimSpace(title) == "" {
		return 0, fmt.Errorf("%w: title is required", ErrInvalid)
	}
	if override == nil {
		override = map[string]any{}
	}
	override[kind.FieldTitle] = title
	for name, value := range override {
		if err := k.ValidateOverride(name, value); err != nil {
			return 0, fmt.Errorf("%w: %v", ErrInvalid, err)
		}
	}
	var id int64
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		handMade, err := handMadeIdentity()
		if err != nil {
			return err
		}
		if _, err := s.put(ctx, tx, handMade, apFields{Kind: kindName, Fields: override}); err != nil {
			return err
		}
		if err := s.settle(ctx, tx, handMade); err != nil {
			return err
		}
		if id, err = s.itemOf(ctx, tx, handMade); err != nil {
			return err
		}
		return logByAp(ctx, tx, id)
	})
	return id, err
}

func (s *Store) CreateItemFromSource(ctx context.Context, kindName, title, source, externalID string, data any) (int64, error) {
	var id int64
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		var err error
		if id, err = s.itemFromSource(ctx, tx, source, externalID, data); err != nil {
			return err
		}
		return logByAp(ctx, tx, id)
	})
	return id, err
}

func (s *Store) itemFromSource(ctx context.Context, tx *sql.Tx, source, externalID string, data any) (int64, error) {
	if externalID == "" {
		return 0, fmt.Errorf("%w: an item from %s needs an external id", ErrInvalid, source)
	}
	id := identityOf(source, externalID)
	if _, err := s.itemOf(ctx, tx, id); err == nil {
		return 0, fmt.Errorf("%w: another item already has %s", ErrConflict, id)
	} else if !errors.Is(err, ErrNotFound) {
		return 0, err
	}
	if _, err := s.put(ctx, tx, id, data); err != nil {
		return 0, err
	}
	if err := s.settle(ctx, tx, id); err != nil {
		return 0, err
	}
	itemID, err := s.itemOf(ctx, tx, id)
	if errors.Is(err, ErrNotFound) {
		return 0, fmt.Errorf("%w: %s fits no kind", ErrInvalid, id)
	}
	return itemID, err
}

func logByAp(ctx context.Context, q queryer, itemID int64) error {
	_, err := take.Log(ctx, q, itemID, record.SourceAP, string(NewEntryVisibility))
	return err
}

func Slugify(title string) string {
	return world.Slugify(title)
}

func (s *Store) FindByIdentity(ctx context.Context, source, externalID string) (int64, error) {
	return s.itemOf(ctx, s.db, identityOf(source, externalID))
}

func (s *Store) PutClaim(ctx context.Context, id record.Identity, data any) error {
	if id.ExternalID == "" {
		return fmt.Errorf("%w: a claim from %s needs an external id", ErrInvalid, id.Source)
	}
	raw, err := asJSON(data)
	if err != nil {
		return err
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		itemID, err := s.itemOf(ctx, tx, id)
		if err == nil {
			return s.putClaimOn(ctx, tx, itemID, id, raw)
		}
		if !errors.Is(err, ErrNotFound) {
			return err
		}
		if _, err := s.put(ctx, tx, id, raw); err != nil {
			return err
		}
		if err := s.settle(ctx, tx, id); err != nil {
			return err
		}
		_, err = s.itemOf(ctx, tx, id)
		if errors.Is(err, ErrNotFound) {
			return fmt.Errorf("%w: %s fits no kind", ErrInvalid, id)
		}
		return err
	})
}

func (s *Store) PutClaimAbout(ctx context.Context, id, about record.Identity, data any) error {
	if id.ExternalID == "" {
		return fmt.Errorf("%w: a claim from %s needs an external id", ErrInvalid, id.Source)
	}
	raw, err := asJSON(data)
	if err != nil {
		return err
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		itemID, err := s.itemOf(ctx, tx, about)
		if err != nil {
			return err
		}
		return s.putClaimOn(ctx, tx, itemID, id, raw)
	})
}

func (s *Store) RemoveClaim(ctx context.Context, id record.Identity) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		return s.takeOffTheItem(ctx, tx, []record.Identity{id})
	})
}

func (s *Store) ClaimOf(ctx context.Context, id record.Identity, source string) (reading.Claim, error) {
	itemID, err := s.itemOf(ctx, s.db, id)
	if err != nil {
		return reading.Claim{Source: source}, err
	}
	return s.claimOn(ctx, itemID, source)
}

func (s *Store) IdentitiesLackingSources(ctx context.Context, have string, lacking ...string) ([]record.Identity, error) {
	claimedOrAnswered, answeredArgs := claimedOrAnsweredFrom("l", lacking)
	args := append([]any{have}, answeredArgs...)
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.source, c.external_id FROM claim c
		JOIN shown_entry e ON e.item_id = c.item_id
		WHERE c.source = ?
		  AND NOT EXISTS (SELECT 1 FROM claim l WHERE l.item_id = c.item_id AND (`+claimedOrAnswered+`))
		ORDER BY c.external_id`,
		args...)
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

func (s *Store) AnswerMatch(ctx context.Context, itemID int64, source, externalID string, data any) error {
	if externalID == "" {
		return fmt.Errorf("%w: a match needs an external id", ErrInvalid)
	}
	return s.inTxThenPropose(ctx, func(tx *sql.Tx) error {
		return s.answer(ctx, tx, itemID, identityOf(source, externalID), data, []string{source})
	})
}

func (s *Store) AnswerNoMatch(ctx context.Context, itemID int64, source string) error {
	return s.inTxThenPropose(ctx, func(tx *sql.Tx) error {
		key, err := s.keyOf(ctx, tx, itemID)
		if err != nil {
			return err
		}
		answer := matchAnswerIdentity(source, key)
		if _, err := s.put(ctx, tx, answer, apMatchAnswer{}); err != nil {
			return err
		}
		replaced, err := claimsFrom(ctx, tx, itemID, []string{source}, answer)
		if err != nil {
			return err
		}
		if err := s.takeOffTheItem(ctx, tx, replaced); err != nil {
			return err
		}
		if err := s.settle(ctx, tx, answer, key); err != nil {
			return err
		}
		return s.rekeyOffAReplacedKey(ctx, tx, itemID, key, replaced, s.nextKey)
	})
}

func (s *Store) answer(ctx context.Context, tx *sql.Tx, itemID int64, id record.Identity, data any, replacing []string) error {
	raw, err := asJSON(data)
	if err != nil {
		return err
	}
	key, err := s.keyOf(ctx, tx, itemID)
	if err != nil {
		return err
	}
	if _, err := s.landingOf(ctx, tx, itemID, id, raw); err != nil {
		return err
	}
	if _, err := s.records.PutClaim(ctx, tx, id, raw); err != nil {
		return err
	}
	answer := matchAnswerIdentity(id.Source, key)
	if _, err := s.put(ctx, tx, answer, apMatchAnswer{To: &id.ExternalID}); err != nil {
		return err
	}
	replaced, err := claimsFrom(ctx, tx, itemID, replacing, id)
	if err != nil {
		return err
	}
	for _, source := range replacing {
		if source != id.Source {
			replaced = append(replaced, matchAnswerIdentity(source, key))
		}
	}
	if err := s.takeOffTheItem(ctx, tx, replaced); err != nil {
		return err
	}
	if err := s.settle(ctx, tx, id, answer, key); err != nil {
		return err
	}
	if err := s.requireLanded(ctx, tx, itemID, id); err != nil {
		return err
	}
	return s.rekeyOffAReplacedKey(ctx, tx, itemID, key, replaced, theAnswer(id))
}

func claimedOrAnsweredFrom(alias string, sources []string) (string, []any) {
	conditions := make([]string, 0, len(sources))
	args := []any{}
	for _, source := range sources {
		prefix := matchAnswerPrefix(source)
		conditions = append(conditions, alias+`.source = ? OR (`+alias+`.source = ? AND substr(`+alias+`.external_id, 1, length(?)) = ?)`)
		args = append(args, source, record.SourceAP, prefix, prefix)
	}
	return strings.Join(conditions, " OR "), args
}

func asJSON(data any) ([]byte, error) {
	switch d := data.(type) {
	case json.RawMessage:
		return d, nil
	case []byte:
		return d, nil
	}
	return json.Marshal(data)
}

func (s *Store) claimOn(ctx context.Context, itemID int64, source string) (reading.Claim, error) {
	c := reading.Claim{Source: source}
	var data string
	err := s.db.QueryRowContext(ctx, `
		SELECT external_id, data, fetched_at FROM claim WHERE item_id = ? AND source = ?
		ORDER BY fetched_at DESC, external_id LIMIT 1`, itemID, source).
		Scan(&c.ExternalID, &data, &c.FetchedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return c, ErrNotFound
	}
	c.Data = json.RawMessage(data)
	return c, err
}

func (s *Store) ChangeOverride(ctx context.Context, itemID int64, changes map[string]any) error {
	k, err := lookupKind(ctx, s.db, itemID)
	if err != nil {
		return err
	}
	for name, value := range changes {
		if err := k.ValidateOverride(name, value); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalid, err)
		}
	}
	return s.inTxThenPropose(ctx, func(tx *sql.Tx) error {
		id, err := s.overrideOf(ctx, tx, itemID)
		if err != nil {
			return err
		}
		override := apFields{}
		if _, err := storedData(ctx, tx, id, &override); err != nil {
			return err
		}
		if override.Fields == nil {
			override.Fields = map[string]any{}
		}
		for name, value := range changes {
			if kind.IsRemoval(value) {
				delete(override.Fields, name)
			} else {
				override.Fields[name] = value
			}
		}
		change, err := s.put(ctx, tx, id, override)
		if err != nil || change == record.Unchanged {
			return err
		}
		return s.settle(ctx, tx, id)
	})
}

func (s *Store) overrideOf(ctx context.Context, q queryer, itemID int64) (record.Identity, error) {
	handMade := identityOf(record.SourceAP, "")
	err := q.QueryRowContext(ctx, `
		SELECT external_id FROM claim
		WHERE item_id = ? AND source = ? AND substr(external_id, 1, length(?)) = ?
		ORDER BY external_id LIMIT 1`, itemID, record.SourceAP, apItemPrefix, apItemPrefix).Scan(&handMade.ExternalID)
	if err == nil {
		return handMade, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return record.Identity{}, err
	}
	key, err := s.keyOf(ctx, q, itemID)
	if err != nil {
		return record.Identity{}, err
	}
	return apIdentity(apFieldsPrefix, key.String()), nil
}

func (s *Store) DeleteItem(ctx context.Context, itemID int64) error {
	return s.inTxThenPropose(ctx, func(tx *sql.Tx) error {
		if _, err := lookupKind(ctx, tx, itemID); err != nil {
			return err
		}
		claims, err := claimsOn(ctx, tx, itemID)
		if err != nil {
			return err
		}
		for _, id := range claims {
			if id.Source != record.SourceAP {
				return fmt.Errorf("%w: item %d has a claim from %s; only an item made by hand can be deleted", ErrInvalid, itemID, id.Source)
			}
		}
		inheritors, err := s.itemIDsWith(ctx, tx, `SELECT DISTINCT from_id FROM relation WHERE to_id = ? ORDER BY from_id`, itemID)
		if err != nil {
			return err
		}
		key, err := s.keyOf(ctx, tx, itemID)
		if err != nil {
			return err
		}
		edits, err := linkEditsNaming(ctx, tx, key)
		if err != nil {
			return err
		}
		for _, id := range append(claims, edits...) {
			if _, err := s.records.RemoveClaim(ctx, tx, id); err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM item WHERE id = ?`, itemID); err != nil {
			return err
		}
		keys := make([]record.Identity, 0, len(inheritors))
		for _, inheritor := range inheritors {
			key, err := s.keyOf(ctx, tx, inheritor)
			if err != nil {
				return err
			}
			keys = append(keys, key)
		}
		return s.settle(ctx, tx, append(keys, edits...)...)
	})
}

func linkEditsNaming(ctx context.Context, q queryer, key record.Identity) ([]record.Identity, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT external_id FROM record_claim
		WHERE source = ? AND substr(external_id, 1, length(?)) = ?
		  AND (json_extract(data, '$.from') = ? OR json_extract(data, '$.to') = ?)
		ORDER BY external_id`,
		record.SourceAP, apLinkPrefix, apLinkPrefix, key.String(), key.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	edits := []record.Identity{}
	for rows.Next() {
		edit := identityOf(record.SourceAP, "")
		if err := rows.Scan(&edit.ExternalID); err != nil {
			return nil, err
		}
		edits = append(edits, edit)
	}
	return edits, rows.Err()
}
