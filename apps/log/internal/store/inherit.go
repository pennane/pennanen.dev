package store

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
)

type inheritance struct {
	fromID   int64
	relation string
}

type relatedReadings struct {
	kind     kind.Kind
	readings []verdict.Reading
}

func inheritingItems(byID map[int64]*ItemView) ([]int64, []string) {
	ids, relations := []int64{}, []string{}
	for id, v := range byID {
		k, known := kind.Get(v.Kind)
		if !known {
			continue
		}
		through := k.InheritingRelations()
		if len(through) == 0 {
			continue
		}
		ids = append(ids, id)
		for _, relation := range through {
			if !slices.Contains(relations, relation) {
				relations = append(relations, relation)
			}
		}
	}
	slices.Sort(ids)
	return ids, relations
}

func (s *Store) relatedVerdicts(ctx context.Context, byID map[int64]*ItemView) (map[int64]map[string]verdict.Verdicts, error) {
	ids, relations := inheritingItems(byID)
	decided := map[int64]map[string]verdict.Verdicts{}
	if len(ids) == 0 {
		return decided, nil
	}
	found := map[inheritance]*relatedReadings{}
	err := forChunks(ids, func(chunk []int64) error {
		args := append(int64sAsArgs(chunk), stringsAsArgs(relations)...)
		rows, err := s.db.QueryContext(ctx, `
			WITH nearest AS (
				SELECT from_id, type, min(to_id) AS to_id FROM relation
				WHERE from_id IN (`+placeholders(len(chunk))+`) AND type IN (`+placeholders(len(relations))+`)
				GROUP BY from_id, type)
			SELECT n.from_id, n.type, i.kind, r.source, coalesce(c.external_id, ''), r.fetched_at, r.reading
			FROM nearest n
			JOIN item i ON i.id = n.to_id
			JOIN claim_reading r ON r.item_id = n.to_id
			JOIN claim c ON c.item_id = r.item_id AND c.source = r.source
			ORDER BY n.from_id, n.type, r.source`, args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var at inheritance
			var kindName, source, externalID, fetchedAt, stored string
			if err := rows.Scan(&at.fromID, &at.relation, &kindName, &source, &externalID, &fetchedAt, &stored); err != nil {
				return err
			}
			k, known := kind.Get(kindName)
			if !known {
				continue
			}
			reading, err := verdict.Decode(k, source, externalID, fetchedAt, json.RawMessage(stored))
			if err != nil {
				return fmt.Errorf("decode reading of the item %d inherits from: %w", at.fromID, err)
			}
			if found[at] == nil {
				found[at] = &relatedReadings{kind: k}
			}
			found[at].readings = append(found[at].readings, reading)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	for at, related := range found {
		if decided[at.fromID] == nil {
			decided[at.fromID] = map[string]verdict.Verdicts{}
		}
		decided[at.fromID][at.relation] = verdict.Decide(related.kind, related.readings)
	}
	return decided, nil
}
