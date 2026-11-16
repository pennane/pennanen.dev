package store

import (
	"context"
	"slices"

	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

func (s *Store) attachMeasures(ctx context.Context, ids []int64, byID map[int64]*ItemView) error {
	return forChunks(ids, func(chunk []int64) error {
		rows, err := s.db.QueryContext(ctx, `
			SELECT item_id, source, field, value FROM measure WHERE item_id IN (`+placeholders(len(chunk))+`)`, int64sAsArgs(chunk)...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id, value int64
			var source, field string
			if err := rows.Scan(&id, &source, &field, &value); err != nil {
				return err
			}
			fillUndecided(byID[id], source, field, value)
		}
		return rows.Err()
	})
}

func fillUndecided(v *ItemView, source, field string, value int64) {
	k, known := kind.Get(v.Kind)
	if !known || !slices.ContainsFunc(k.Fields, func(f kind.Field) bool { return f.Name == field }) {
		return
	}
	if _, decided := v.Verdicts.Values[field]; decided {
		return
	}
	v.Verdicts.Values[field] = value
	v.Verdicts.SourceOf[field] = source
}

type measureKey struct {
	itemID int64
	source string
	field  string
}

func (s *Store) measuresOf(ctx context.Context, ids []int64) (map[measureKey]int64, error) {
	found := map[measureKey]int64{}
	err := forChunks(ids, func(chunk []int64) error {
		rows, err := s.db.QueryContext(ctx, `
			SELECT item_id, source, field, value FROM measure WHERE item_id IN (`+placeholders(len(chunk))+`)`, int64sAsArgs(chunk)...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var key measureKey
			var value int64
			if err := rows.Scan(&key.itemID, &key.source, &key.field, &value); err != nil {
				return err
			}
			found[key] = value
		}
		return rows.Err()
	})
	return found, err
}
