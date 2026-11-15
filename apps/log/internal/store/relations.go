package store

import (
	"context"
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
