package store

import (
	"context"
	"database/sql"
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

type keyChoice func(ctx context.Context, q queryer, itemID int64, candidates []record.Identity) (record.Identity, error)

func theAnswer(answer record.Identity) keyChoice {
	return func(context.Context, queryer, int64, []record.Identity) (record.Identity, error) {
		return answer, nil
	}
}

func (s *Store) nextKey(ctx context.Context, q queryer, itemID int64, candidates []record.Identity) (record.Identity, error) {
	return s.world.KeyAmong(ctx, q, itemID, candidates)
}

type keyHold string

const (
	heldByARecord         keyHold = "held-by-a-record"
	heldByItsKeyedRecords keyHold = "held-by-its-keyed-records"
)

func (s *Store) rekeyOffAReplacedKey(ctx context.Context, tx *sql.Tx, itemID int64, old record.Identity, replaced []record.Identity, choose keyChoice) error {
	if !slices.Contains(replaced, old) {
		return nil
	}
	keyed, hold, err := s.recordsKeyedBy(ctx, tx, itemID, old)
	if err != nil || hold == heldByARecord {
		return err
	}
	ids, err := s.world.IdentitiesOn(ctx, tx, itemID)
	if err != nil {
		return err
	}
	candidates := slices.DeleteFunc(ids, func(id record.Identity) bool {
		return id == old || slices.Contains(keyed, id) || isTheLogsOwn(id)
	})
	key, err := choose(ctx, tx, itemID, candidates)
	if errors.Is(err, world.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := s.renameIdentity(ctx, tx, old, key); err != nil {
		return err
	}
	if err := s.records.RenameAbout(ctx, tx, old, key); err != nil {
		return err
	}
	dirty := []record.Identity{key}
	for _, id := range keyed {
		renamed, err := s.rekeyRecord(ctx, tx, id, old, key)
		if err != nil {
			return err
		}
		dirty = append(dirty, renamed)
	}
	if err := s.settle(ctx, tx, dirty...); err != nil {
		return err
	}
	return s.requireLanded(ctx, tx, itemID, key)
}

func isTheLogsOwn(id record.Identity) bool {
	return id.Source == kind.SourceAP || id.Source == kind.SourceMatched || id.Source == kind.SourceLegacy
}

func (s *Store) recordsKeyedBy(ctx context.Context, q queryer, itemID int64, old record.Identity) ([]record.Identity, keyHold, error) {
	claims, err := claimsOn(ctx, q, itemID)
	if err != nil {
		return nil, "", err
	}
	keyed := []record.Identity{}
	others := []record.Identity{}
	for _, id := range claims {
		if isKeyedBy(id, old) {
			keyed = append(keyed, id)
		} else {
			others = append(others, id)
		}
	}
	readings, err := s.records.Readings(ctx, q, others)
	if err != nil {
		return nil, "", err
	}
	for _, r := range readings {
		if !slices.Contains(verdict.Derive(r.Identity, r.Values).About, old) {
			continue
		}
		if r.Source != kind.SourceMatched && r.Source != kind.SourceLegacy {
			return nil, heldByARecord, nil
		}
		keyed = append(keyed, r.Identity)
	}
	edits, err := linkEditsNaming(ctx, q, old)
	if err != nil {
		return nil, "", err
	}
	legacy, err := legacyRecordsLinking(ctx, q, old)
	if err != nil {
		return nil, "", err
	}
	all := slices.Concat(keyed, edits, legacy)
	slices.SortFunc(all, compareIdentities)
	all = slices.Compact(all)
	asserting, err := recordsLinking(ctx, q, old)
	if err != nil {
		return nil, "", err
	}
	for _, id := range asserting {
		if !slices.Contains(all, id) {
			return nil, heldByARecord, nil
		}
	}
	return all, heldByItsKeyedRecords, nil
}

func recordsLinking(ctx context.Context, q queryer, id record.Identity) ([]record.Identity, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT asserted_source, asserted_external_id FROM link WHERE from_source = ? AND from_external_id = ?
		UNION
		SELECT asserted_source, asserted_external_id FROM link WHERE to_source = ? AND to_external_id = ?`,
		id.Source, id.ExternalID, id.Source, id.ExternalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records := []record.Identity{}
	for rows.Next() {
		var asserting record.Identity
		if err := rows.Scan(&asserting.Source, &asserting.ExternalID); err != nil {
			return nil, err
		}
		records = append(records, asserting)
	}
	return records, rows.Err()
}

func compareIdentities(a, b record.Identity) int {
	return strings.Compare(a.String(), b.String())
}

func legacyRecordsLinking(ctx context.Context, q queryer, id record.Identity) ([]record.Identity, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT DISTINCT asserted_external_id FROM link
		WHERE asserted_source = ?
		  AND ((from_source = ? AND from_external_id = ?) OR (to_source = ? AND to_external_id = ?))
		ORDER BY asserted_external_id`,
		kind.SourceLegacy, id.Source, id.ExternalID, id.Source, id.ExternalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records := []record.Identity{}
	for rows.Next() {
		legacy := identityOf(kind.SourceLegacy, "")
		if err := rows.Scan(&legacy.ExternalID); err != nil {
			return nil, err
		}
		records = append(records, legacy)
	}
	return records, rows.Err()
}

func isKeyedBy(id, old record.Identity) bool {
	switch {
	case id.ExternalID == old.String():
		return true
	case id.Source != kind.SourceAP:
		return false
	case id.ExternalID == apFieldsPrefix+old.String():
		return true
	}
	answered, isAnAnswer := strings.CutPrefix(id.ExternalID, apMatchPrefix)
	_, about, _ := strings.Cut(answered, "/")
	return isAnAnswer && about == old.String()
}

func renamedIdentity(id, old, key record.Identity) record.Identity {
	switch {
	case !isKeyedBy(id, old):
		return id
	case id.ExternalID == old.String():
		return identityOf(id.Source, key.String())
	case id.ExternalID == apFieldsPrefix+old.String():
		return apIdentity(apFieldsPrefix, key.String())
	}
	source, _, _ := strings.Cut(strings.TrimPrefix(id.ExternalID, apMatchPrefix), "/")
	return matchAnswerIdentity(source, key)
}

func (s *Store) rekeyRecord(ctx context.Context, tx *sql.Tx, id, old, key record.Identity) (record.Identity, error) {
	var stored string
	if err := tx.QueryRowContext(ctx, `SELECT data FROM record_claim WHERE source = ? AND external_id = ?`, id.Source, id.ExternalID).Scan(&stored); err != nil {
		return record.Identity{}, err
	}
	data := json.RawMessage(stored)
	renamed := renamedIdentity(id, old, key)
	if namesIdentitiesInItsData(id) {
		var err error
		if data, err = renamedInData(data, old, key); err != nil {
			return record.Identity{}, err
		}
	}
	if isLinkEdit(id) {
		var edit apLink
		if err := json.Unmarshal(data, &edit); err != nil {
			return record.Identity{}, err
		}
		from, err := record.ParseIdentity(edit.From)
		if err != nil {
			return record.Identity{}, err
		}
		to, err := record.ParseIdentity(edit.To)
		if err != nil {
			return record.Identity{}, err
		}
		renamed = linkEditIdentity(from, edit.Type, to)
	}
	if err := s.records.Rename(ctx, tx, id, renamed, data); err != nil {
		if errors.Is(err, record.ErrClaimTaken) {
			return record.Identity{}, fmt.Errorf("%w: %v", ErrConflict, err)
		}
		return record.Identity{}, err
	}
	return renamed, s.renameIdentity(ctx, tx, id, renamed)
}

func isLinkEdit(id record.Identity) bool {
	return id.Source == kind.SourceAP && strings.HasPrefix(id.ExternalID, apLinkPrefix)
}

func namesIdentitiesInItsData(id record.Identity) bool {
	return id.Source == kind.SourceMatched || id.Source == kind.SourceLegacy || isLinkEdit(id)
}

func renamedInData(data json.RawMessage, old, key record.Identity) (json.RawMessage, error) {
	var parsed any
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}
	return json.Marshal(renamedIn(parsed, old.String(), key.String()))
}

func renamedIn(value any, old, key string) any {
	switch v := value.(type) {
	case string:
		if v == old {
			return key
		}
	case []any:
		for i := range v {
			v[i] = renamedIn(v[i], old, key)
		}
	case map[string]any:
		for name := range v {
			v[name] = renamedIn(v[name], old, key)
		}
	}
	return value
}

func (s *Store) renameIdentity(ctx context.Context, q queryer, from, to record.Identity) error {
	err := s.world.RenameIdentity(ctx, q, from, to)
	if errors.Is(err, world.ErrHeldElsewhere) {
		return fmt.Errorf("%w: %v", ErrConflict, err)
	}
	return err
}
