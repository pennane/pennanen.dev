package world

import (
	"cmp"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/verdict"
	"maps"
	"slices"
	"strings"
)

type ownOrigin string

const (
	decidedNow    ownOrigin = "decided-now"
	storedEarlier ownOrigin = "stored-earlier"
)

type decision struct {
	k      kind.Kind
	own    verdict.Verdicts
	origin ownOrigin
}

const externalIDUnusedByDecide = ""

func unknownKind(itemID int64, kindName string) error {
	return fmt.Errorf("item %d has unknown kind %q", itemID, kindName)
}

func decideOwn(k kind.Kind, readings []record.Reading) (verdict.Verdicts, error) {
	oldestFirst := slices.SortedFunc(slices.Values(readings), func(a, b record.Reading) int {
		return cmp.Or(cmp.Compare(a.ChangedAt, b.ChangedAt), compareIdentities(a.Identity, b.Identity))
	})
	merged := map[string]map[string]any{}
	fetchedAt := map[string]string{}
	for _, r := range oldestFirst {
		if merged[r.Source] == nil {
			merged[r.Source] = map[string]any{}
		}
		maps.Copy(merged[r.Source], r.Values)
		fetchedAt[r.Source] = max(fetchedAt[r.Source], r.FetchedAt)
	}
	decoded := make([]reading.Reading, 0, len(merged))
	for _, source := range slices.Sorted(maps.Keys(merged)) {
		raw, err := json.Marshal(merged[source])
		if err != nil {
			return verdict.Verdicts{}, err
		}
		r, err := verdict.Decode(k, source, externalIDUnusedByDecide, fetchedAt[source], raw)
		if err != nil {
			return verdict.Verdicts{}, err
		}
		decoded = append(decoded, r)
	}
	return verdict.Decide(k, decoded), nil
}

func (s *settling) decideVerdicts() error {
	touched := sortedItems(s.touched)
	for _, itemID := range touched {
		inheriting, err := s.itemsInheritingFrom(itemID)
		if err != nil {
			return err
		}
		for _, id := range inheriting {
			s.related[id] = struct{}{}
		}
	}
	for _, itemID := range touched {
		if err := s.storeVerdicts(itemID); err != nil {
			return err
		}
	}
	for _, itemID := range sortedItems(s.related) {
		if _, isTouched := s.touched[itemID]; isTouched {
			continue
		}
		if err := s.storeVerdicts(itemID); err != nil {
			return err
		}
	}
	return nil
}

func (s *settling) storeVerdicts(itemID int64) error {
	d, err := s.decisionOf(itemID)
	if err != nil {
		return err
	}
	verdicts, err := s.inherited(itemID, d)
	if err != nil {
		return err
	}
	encodedVerdicts, err := json.Marshal(verdicts)
	if err != nil {
		return err
	}
	if d.origin == storedEarlier {
		_, err := s.q.ExecContext(s.ctx,
			`UPDATE item_verdict SET verdicts = ? WHERE item_id = ? AND verdicts IS NOT ?`,
			string(encodedVerdicts), itemID, string(encodedVerdicts))
		return err
	}
	encodedOwn, err := json.Marshal(d.own)
	if err != nil {
		return err
	}
	_, err = s.q.ExecContext(s.ctx, `
		INSERT INTO item_verdict (item_id, own, verdicts) VALUES (?, ?, ?)
		ON CONFLICT (item_id) DO UPDATE SET own = excluded.own, verdicts = excluded.verdicts
		WHERE own IS NOT excluded.own OR verdicts IS NOT excluded.verdicts`,
		itemID, string(encodedOwn), string(encodedVerdicts))
	return err
}

func (s *settling) decisionOf(itemID int64) (decision, error) {
	if d, decided := s.owns[itemID]; decided {
		return d, nil
	}
	var kindName string
	var stored sql.NullString
	if err := s.q.QueryRowContext(s.ctx,
		`SELECT i.kind, v.own FROM item i LEFT JOIN item_verdict v ON v.item_id = i.id WHERE i.id = ?`,
		itemID).Scan(&kindName, &stored); err != nil {
		return decision{}, err
	}
	k, known := kind.Get(kindName)
	if !known {
		return decision{}, unknownKind(itemID, kindName)
	}
	d := decision{k: k, origin: storedEarlier}
	if stored.Valid {
		if err := json.Unmarshal([]byte(stored.String), &d.own); err != nil {
			return decision{}, fmt.Errorf("own verdicts of item %d: %w", itemID, err)
		}
	} else {
		ids, err := identitiesOn(s.ctx, s.q, itemID)
		if err != nil {
			return decision{}, err
		}
		readings, err := s.readingsOf(ids)
		if err != nil {
			return decision{}, err
		}
		if d.own, err = decideOwn(k, readings); err != nil {
			return decision{}, err
		}
		d.origin = decidedNow
	}
	s.owns[itemID] = d
	return d, nil
}

func (s *settling) inherited(itemID int64, d decision) (verdict.Verdicts, error) {
	related := map[string]verdict.Verdicts{}
	for _, through := range d.k.InheritingRelations() {
		farID, err := s.farEnd(itemID, through)
		if errors.Is(err, ErrNotFound) {
			continue
		}
		if err != nil {
			return verdict.Verdicts{}, err
		}
		far, err := s.decisionOf(farID)
		if err != nil {
			return verdict.Verdicts{}, err
		}
		related[through] = far.own
	}
	return verdict.Inherit(d.k, cloned(d.own), related), nil
}

func cloned(v verdict.Verdicts) verdict.Verdicts {
	c := verdict.Verdicts{Values: map[string]any{}, SourceOf: map[string]string{}}
	maps.Copy(c.Values, v.Values)
	maps.Copy(c.SourceOf, v.SourceOf)
	return c
}

func (s *settling) farEnd(itemID int64, through string) (int64, error) {
	var farID int64
	err := s.q.QueryRowContext(s.ctx, `
		SELECT t.item_id FROM item_identity f
		JOIN link l ON l.from_source = f.source AND l.from_external_id = f.external_id
		JOIN item_identity t ON t.source = l.to_source AND t.external_id = l.to_external_id
		WHERE f.item_id = ? AND l.type = ? AND t.item_id <> f.item_id
		ORDER BY t.item_id LIMIT 1`, itemID, through).Scan(&farID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	return farID, err
}

func (s *settling) itemsInheritingFrom(itemID int64) ([]int64, error) {
	types := inheritingLinkTypes()
	if len(types) == 0 {
		return nil, nil
	}
	args := []any{itemID}
	for _, t := range types {
		args = append(args, t)
	}
	rows, err := s.q.QueryContext(s.ctx, `
		SELECT DISTINCT f.item_id FROM item_identity t
		JOIN link l ON l.to_source = t.source AND l.to_external_id = t.external_id
		JOIN item_identity f ON f.source = l.from_source AND f.external_id = l.from_external_id
		WHERE t.item_id = ? AND f.item_id <> t.item_id AND l.type IN (`+placeholders(len(types))+`)`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []int64{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func placeholders(n int) string {
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
}
