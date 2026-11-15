package store

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
)

func (s *Store) AddRelation(ctx context.Context, r Relation) error {
	if r.Type == "" || r.FromID == r.ToID {
		return fmt.Errorf("%w: relation needs a type and two different items", ErrInvalid)
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO relation (from_id, type, to_id) VALUES (?, ?, ?) ON CONFLICT DO NOTHING`, r.FromID, r.Type, r.ToID)
	return err
}

func (s *Store) RemoveRelation(ctx context.Context, r Relation) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM relation WHERE from_id = ? AND type = ? AND to_id = ?`, r.FromID, r.Type, r.ToID)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func (s *Store) RelatedTo(ctx context.Context, fromIDs []int64, relation string) ([]int64, error) {
	related := []int64{}
	err := forChunks(fromIDs, func(chunk []int64) error {
		ids, err := s.itemIDs(ctx, `
			SELECT DISTINCT to_id FROM relation WHERE type = ? AND from_id IN (`+placeholders(len(chunk))+`)`,
			append([]any{relation}, int64sAsArgs(chunk)...)...)
		related = append(related, ids...)
		return err
	})
	return slices.Compact(slices.Sorted(slices.Values(related))), err
}

func (s *Store) ReplaceRelationsFrom(ctx context.Context, fromID int64, types []string, relations []Relation) error {
	for _, r := range relations {
		if r.FromID != fromID || !slices.Contains(types, r.Type) || r.FromID == r.ToID {
			return fmt.Errorf("%w: relation %v is not one of %v from %d", ErrInvalid, r, types, fromID)
		}
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM relation WHERE from_id = ? AND type IN (`+placeholders(len(types))+`)`,
			append([]any{fromID}, stringsAsArgs(types)...)...); err != nil {
			return err
		}
		for _, r := range relations {
			if _, err := tx.ExecContext(ctx, `INSERT INTO relation (from_id, type, to_id) VALUES (?, ?, ?) ON CONFLICT DO NOTHING`, r.FromID, r.Type, r.ToID); err != nil {
				return err
			}
		}
		return nil
	})
}
