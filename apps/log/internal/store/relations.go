package store

import (
	"context"
	"fmt"
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
