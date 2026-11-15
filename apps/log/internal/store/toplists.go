package store

import (
	"context"
)

type ListPoints map[int64]map[string]float64

func (s *Store) TopListPoints(ctx context.Context, source, signal string) (ListPoints, error) {
	rows, err := s.db.QueryContext(ctx, `
		WITH listed AS (
			SELECT item_id,
				json_extract(data, '$.range') AS list_range,
				json_extract(data, '$.day') AS day,
				json_extract(data, '$.rank') AS place
			FROM observation WHERE source = ? AND signal = ?),
		lengths AS (SELECT list_range, day, count(*) AS length FROM listed GROUP BY list_range, day)
		SELECT l.item_id, l.list_range, sum((n.length - l.place + 1) * 1.0 / n.length)
		FROM listed l JOIN lengths n ON n.list_range = l.list_range AND n.day = l.day
		GROUP BY l.item_id, l.list_range`, source, signal)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	points := ListPoints{}
	for rows.Next() {
		var itemID int64
		var listRange string
		var sum float64
		if err := rows.Scan(&itemID, &listRange, &sum); err != nil {
			return nil, err
		}
		if points[itemID] == nil {
			points[itemID] = map[string]float64{}
		}
		points[itemID][listRange] = sum
	}
	return points, rows.Err()
}

func (s *Store) ObservationCounts(ctx context.Context, source, signal string) (map[int64]int, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT item_id, count(*) FROM observation WHERE source = ? AND signal = ? GROUP BY item_id`, source, signal)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	counts := map[int64]int{}
	for rows.Next() {
		var itemID int64
		var count int
		if err := rows.Scan(&itemID, &count); err != nil {
			return nil, err
		}
		counts[itemID] = count
	}
	return counts, rows.Err()
}

func (s *Store) RelationsOfType(ctx context.Context, relation string) ([]Relation, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT from_id, type, to_id FROM relation WHERE type = ? ORDER BY from_id, to_id`, relation)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	relations := []Relation{}
	for rows.Next() {
		var r Relation
		if err := rows.Scan(&r.FromID, &r.Type, &r.ToID); err != nil {
			return nil, err
		}
		relations = append(relations, r)
	}
	return relations, rows.Err()
}

func (s *Store) LoggedItems(ctx context.Context, kindName string) ([]int64, error) {
	return s.itemIDs(ctx, `
		SELECT e.item_id FROM entry e JOIN item i ON i.id = e.item_id
		WHERE i.kind = ? ORDER BY e.item_id`, kindName)
}
