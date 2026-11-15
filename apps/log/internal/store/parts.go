package store

import (
	"context"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

func (s *Store) PartsOf(ctx context.Context, wholeID int64) ([]int64, error) {
	return s.itemIDs(ctx, `SELECT part_id FROM part WHERE whole_id = ? ORDER BY position`, wholeID)
}

func (s *Store) LoggedItemsNotDismissed(ctx context.Context, kindName string) ([]int64, error) {
	return s.itemIDs(ctx, `
		SELECT e.item_id FROM entry e JOIN item i ON i.id = e.item_id
		WHERE i.kind = ? AND e.status <> ? ORDER BY e.item_id`, kindName, kind.Dismissed)
}
