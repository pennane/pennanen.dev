package store

import (
	"context"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
)

func (s *Store) PartsOf(ctx context.Context, wholeID int64) ([]int64, error) {
	return s.itemIDs(ctx, `SELECT part_id FROM part WHERE whole_id = ? ORDER BY position`, wholeID)
}

func (s *Store) LoggedItemsNotDismissed(ctx context.Context, kindName string) ([]int64, error) {
	return s.itemIDs(ctx, `
		SELECT e.item_id FROM shown_entry e JOIN item i ON i.id = e.item_id
		WHERE i.kind = ? ORDER BY e.item_id`, kindName)
}

func (s *Store) LoggedIdentitiesNotDismissed(ctx context.Context, kindName, source string) ([]record.Identity, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.source, c.external_id FROM claim c
		JOIN shown_entry e ON e.item_id = c.item_id JOIN item i ON i.id = e.item_id
		WHERE i.kind = ? AND c.source = ? ORDER BY c.external_id`, kindName, source)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []record.Identity{}
	for rows.Next() {
		var id record.Identity
		if err := rows.Scan(&id.Source, &id.ExternalID); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
