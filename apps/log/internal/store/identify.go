package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
)

type Identity struct {
	Kind       string
	Source     string
	ExternalID string
	Data       any
}

func (s *Store) Identify(ctx context.Context, itemID int64, identity Identity, replacing []string) error {
	to, known := kind.Get(identity.Kind)
	if !known {
		return fmt.Errorf("%w: unknown kind %q", ErrInvalid, identity.Kind)
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		if err := putClaim(ctx, tx, itemID, identity.Source, identity.ExternalID, identity.Data); err != nil {
			return err
		}
		for _, source := range replacing {
			if source == identity.Source {
				continue
			}
			if _, err := tx.ExecContext(ctx, `DELETE FROM claim WHERE item_id = ? AND source = ?`, itemID, source); err != nil {
				return err
			}
		}
		return moveToKind(ctx, tx, itemID, to)
	})
}

func moveToKind(ctx context.Context, tx *sql.Tx, itemID int64, to kind.Kind) error {
	from, err := lookupKind(ctx, tx, itemID)
	if err != nil || from.Name == to.Name {
		return err
	}
	var status, slug string
	if err := tx.QueryRowContext(ctx, `
		SELECT coalesce(e.status, ?), i.slug FROM item i LEFT JOIN entry e ON e.item_id = i.id WHERE i.id = ?`,
		kind.Untriaged, itemID).Scan(&status, &slug); err != nil {
		return err
	}
	if !to.HasStatus(status) {
		return fmt.Errorf("%w: %s has no status %q to move %s into", ErrInvalid, to.Name, status, slug)
	}
	if err := leaveRanking(ctx, tx, from, itemID); err != nil {
		return err
	}
	if slug, err = freeSlug(ctx, tx, to.Name, slug); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE item SET kind = ?, slug = ? WHERE id = ?`, to.Name, slug, itemID); err != nil {
		return err
	}
	return joinRankingAtBottom(ctx, tx, to, itemID)
}

func leaveRanking(ctx context.Context, tx *sql.Tx, k kind.Kind, itemID int64) error {
	if k.RankedAxis == "" {
		return nil
	}
	current, placed, err := currentSpot(ctx, tx, itemID, k.RankedAxis)
	if err != nil || !placed {
		return err
	}
	if err := closeGap(ctx, tx, k, current); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE score SET position = NULL WHERE item_id = ? AND axis = ?`, itemID, k.RankedAxis)
	return err
}

func joinRankingAtBottom(ctx context.Context, tx *sql.Tx, k kind.Kind, itemID int64) error {
	if k.RankedAxis == "" {
		return nil
	}
	var division string
	err := tx.QueryRowContext(ctx, `SELECT division FROM score WHERE item_id = ? AND axis = ?`, itemID, k.RankedAxis).Scan(&division)
	if isNoRows(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return moveTo(ctx, tx, k, itemID, division, AtBottom)
}
