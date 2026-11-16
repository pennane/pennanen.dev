package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

type Identity struct {
	Kind       string
	Source     string
	ExternalID string
	Data       any
}

func (s *Store) Identify(ctx context.Context, itemID int64, identity Identity, replacing []string) error {
	if _, known := kind.Get(identity.Kind); !known {
		return fmt.Errorf("%w: unknown kind %q", ErrInvalid, identity.Kind)
	}
	if identity.ExternalID == "" {
		return fmt.Errorf("%w: an identity needs an external id", ErrInvalid)
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		return s.answer(ctx, tx, itemID, identityOf(identity.Source, identity.ExternalID), identity.Data, replacing)
	})
}

func (s *Store) IdentifyAbout(ctx context.Context, about record.Identity, identity Identity, replacing []string) error {
	itemID, err := s.itemOf(ctx, s.db, about)
	if err != nil {
		return err
	}
	return s.Identify(ctx, itemID, identity, replacing)
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
