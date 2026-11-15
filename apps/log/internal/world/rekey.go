package world

import (
	"context"
	"errors"
	"fmt"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
)

var ErrHeldElsewhere = errors.New("held by another item")

func (w *World) IdentitiesOn(ctx context.Context, q record.Queryer, itemID int64) ([]record.Identity, error) {
	return identitiesOn(ctx, q, itemID)
}

func (w *World) RenameIdentity(ctx context.Context, q record.Queryer, from, to record.Identity) error {
	if from == to {
		return nil
	}
	itemID, err := itemOf(ctx, q, from)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	owner, err := itemOf(ctx, q, to)
	switch {
	case errors.Is(err, ErrNotFound):
	case err != nil:
		return err
	case owner != itemID:
		return fmt.Errorf("%w: %s belongs to item %d", ErrHeldElsewhere, to, owner)
	default:
		if _, err := q.ExecContext(ctx,
			`DELETE FROM item_identity WHERE source = ? AND external_id = ?`, to.Source, to.ExternalID); err != nil {
			return err
		}
	}
	if _, err := q.ExecContext(ctx,
		`UPDATE item_identity SET source = ?, external_id = ? WHERE source = ? AND external_id = ?`,
		to.Source, to.ExternalID, from.Source, from.ExternalID); err != nil {
		return err
	}
	_, err = q.ExecContext(ctx,
		`UPDATE OR REPLACE item_overlap SET source = ?, external_id = ? WHERE source = ? AND external_id = ?`,
		to.Source, to.ExternalID, from.Source, from.ExternalID)
	return err
}

func (w *World) KeyAmong(ctx context.Context, q record.Queryer, itemID int64, candidates []record.Identity) (record.Identity, error) {
	if len(candidates) == 0 {
		return record.Identity{}, ErrNotFound
	}
	var kindName string
	if err := q.QueryRowContext(ctx, `SELECT kind FROM item WHERE id = ?`, itemID).Scan(&kindName); err != nil {
		return record.Identity{}, err
	}
	k, known := kind.Get(kindName)
	if !known {
		return record.Identity{}, unknownKind(itemID, kindName)
	}
	s := &settling{
		ctx:      ctx,
		q:        q,
		records:  w.records,
		readings: map[record.Identity]record.Reading{},
		fetched:  map[record.Identity]struct{}{},
	}
	ordered, err := s.keyFirst(uniqueSorted(candidates), k)
	if err != nil {
		return record.Identity{}, err
	}
	return ordered[0], nil
}
