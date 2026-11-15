package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
	"slices"
)

func (s *Store) AddRelation(ctx context.Context, r Relation) error {
	if err := editable(r); err != nil {
		return err
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		edit, err := s.linkEditOf(ctx, tx, r, linkAdded)
		if err != nil {
			return err
		}
		return s.putLinkEdit(ctx, tx, edit)
	})
}

func (s *Store) RemoveRelation(ctx context.Context, r Relation) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		if err := requireRelation(ctx, tx, r); err != nil {
			return err
		}
		edit, err := s.linkEditOf(ctx, tx, r, linkSuppressed)
		if err != nil {
			return err
		}
		stored := apLink{}
		presence, err := storedData(ctx, tx, edit.id, &stored)
		if err != nil {
			return err
		}
		if presence == present && stored.Edit == linkAdded {
			if _, err := s.records.RemoveClaim(ctx, tx, edit.id); err != nil {
				return err
			}
			if err := s.settle(ctx, tx, edit.id, edit.from); err != nil {
				return err
			}
			err := requireRelation(ctx, tx, r)
			if errors.Is(err, ErrNotFound) {
				return nil
			}
			if err != nil {
				return err
			}
		}
		return s.putLinkEdit(ctx, tx, edit)
	})
}

func editable(r Relation) error {
	linkType, known := kind.LinkTypeOf(r.Type)
	if !known || linkType.Order != kind.Unordered || r.FromID == r.ToID {
		return fmt.Errorf("%w: a relation needs an unordered link type and two different items", ErrInvalid)
	}
	return nil
}

func requireRelation(ctx context.Context, q queryer, r Relation) error {
	var found int
	if err := q.QueryRowContext(ctx, `SELECT count(*) FROM relation WHERE from_id = ? AND type = ? AND to_id = ?`, r.FromID, r.Type, r.ToID).Scan(&found); err != nil {
		return err
	}
	if found == 0 {
		return fmt.Errorf("%w: no %s relation from %d to %d", ErrNotFound, r.Type, r.FromID, r.ToID)
	}
	return nil
}

type pendingLinkEdit struct {
	id   record.Identity
	from record.Identity
	data apLink
}

func (s *Store) linkEditOf(ctx context.Context, q queryer, r Relation, edit linkEdit) (pendingLinkEdit, error) {
	from, err := s.keyOf(ctx, q, r.FromID)
	if err != nil {
		return pendingLinkEdit{}, err
	}
	to, err := s.keyOf(ctx, q, r.ToID)
	if err != nil {
		return pendingLinkEdit{}, err
	}
	return pendingLinkEdit{
		id:   linkEditIdentity(from, r.Type, to),
		from: from,
		data: apLink{Edit: edit, From: from.String(), Type: r.Type, To: to.String()},
	}, nil
}

func (s *Store) putLinkEdit(ctx context.Context, tx *sql.Tx, edit pendingLinkEdit) error {
	change, err := s.put(ctx, tx, edit.id, edit.data)
	if err != nil || change == record.Unchanged {
		return err
	}
	return s.settle(ctx, tx, edit.id)
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
