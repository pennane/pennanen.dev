package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
)

const firstPosition = 1

func (s *Store) ReplaceParts(ctx context.Context, wholeID int64, partIDs []int64) error {
	for _, partID := range partIDs {
		if partID == wholeID {
			return fmt.Errorf("%w: item %d can't be a part of itself", ErrInvalid, wholeID)
		}
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM part WHERE whole_id = ?`, wholeID); err != nil {
			return err
		}
		for i, partID := range partIDs {
			if _, err := tx.ExecContext(ctx, `INSERT INTO part (whole_id, position, part_id) VALUES (?, ?, ?)`,
				wholeID, firstPosition+i, partID); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) PartsOf(ctx context.Context, wholeID int64) ([]int64, error) {
	return s.itemIDs(ctx, `SELECT part_id FROM part WHERE whole_id = ? ORDER BY position`, wholeID)
}

func (s *Store) LoggedItemsNotDismissed(ctx context.Context, kindName string) ([]int64, error) {
	return s.itemIDs(ctx, `
		SELECT e.item_id FROM entry e JOIN item i ON i.id = e.item_id
		WHERE i.kind = ? AND e.status <> ? ORDER BY e.item_id`, kindName, kind.Dismissed)
}
