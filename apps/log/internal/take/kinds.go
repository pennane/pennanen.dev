package take

import (
	"context"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

func WriteKindStatuses(ctx context.Context, q record.Queryer, kinds []kind.Kind) error {
	if _, err := q.ExecContext(ctx, `DELETE FROM kind_status`); err != nil {
		return err
	}
	for _, k := range kinds {
		for _, status := range k.Statuses {
			if _, err := q.ExecContext(ctx, `INSERT OR IGNORE INTO kind_status (kind, status) VALUES (?, ?)`, k.Name, status); err != nil {
				return err
			}
		}
	}
	return nil
}
