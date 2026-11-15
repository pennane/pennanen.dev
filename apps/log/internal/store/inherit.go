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
	nearest, err := s.nearestRelated(ctx, ids, relations)
	if err != nil {
		return nil, err
	}
	relatedIDs := []int64{}
	for _, toID := range nearest {
		relatedIDs = append(relatedIDs, toID)
	}
	verdictsOf, err := s.verdictsOf(ctx, slices.Compact(slices.Sorted(slices.Values(relatedIDs))))
	if err != nil {
		return nil, err
	}
	for at, toID := range nearest {
		related, present := verdictsOf[toID]
		if !present {
			continue
		}
		if decided[at.fromID] == nil {
			decided[at.fromID] = map[string]verdict.Verdicts{}
		}
		decided[at.fromID][at.relation] = related
	}
	return decided, nil
}

func (s *Store) nearestRelated(ctx context.Context, fromIDs []int64, relations []string) (map[inheritance]int64, error) {
	nearest := map[inheritance]int64{}
	err := forChunks(fromIDs, func(chunk []int64) error {
		rows, err := s.db.QueryContext(ctx, `
			SELECT from_id, type, min(to_id) FROM relation
			WHERE from_id IN (`+placeholders(len(chunk))+`) AND type IN (`+placeholders(len(relations))+`)
			GROUP BY from_id, type`, append(int64sAsArgs(chunk), stringsAsArgs(relations)...)...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var at inheritance
			var toID int64
			if err := rows.Scan(&at.fromID, &at.relation, &toID); err != nil {
				return err
			}
			nearest[at] = toID
		}
		return rows.Err()
	})
	return nearest, err
}

const externalIDUnusedByDecide = ""

func (s *Store) verdictsOf(ctx context.Context, itemIDs []int64) (map[int64]verdict.Verdicts, error) {
	kinds := map[int64]kind.Kind{}
	readings := map[int64][]verdict.Reading{}
	err := forChunks(itemIDs, func(chunk []int64) error {
		rows, err := s.db.QueryContext(ctx, `
			SELECT r.item_id, i.kind, r.source, r.fetched_at, r.reading
			FROM claim_reading r JOIN item i ON i.id = r.item_id
			WHERE r.item_id IN (`+placeholders(len(chunk))+`) ORDER BY r.item_id, r.source`, int64sAsArgs(chunk)...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id int64
			var kindName, source, fetchedAt, stored string
			if err := rows.Scan(&id, &kindName, &source, &fetchedAt, &stored); err != nil {
				return err
			}
			k, known := kind.Get(kindName)
			if !known {
				continue
			}
			reading, err := verdict.Decode(k, source, externalIDUnusedByDecide, fetchedAt, json.RawMessage(stored))
			if err != nil {
				return fmt.Errorf("decode reading of item %d %s: %w", id, source, err)
			}
			kinds[id] = k
			readings[id] = append(readings[id], reading)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	decided := make(map[int64]verdict.Verdicts, len(readings))
	for id, rs := range readings {
		decided[id] = verdict.Decide(kinds[id], rs)
	}
	return decided, nil
}
