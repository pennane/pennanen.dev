package store

import (
	"context"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
)

func KeyOf(ctx context.Context, s *Store, itemID int64) (record.Identity, error) {
	return s.keyOf(ctx, s.db, itemID)
}
