package world

import (
	"context"
	"errors"
	"slices"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
)

type Release string

const (
	Released  Release = "released"
	StillHeld Release = "still-held"
	NotHeld   Release = "not-held"
)

func (w *World) Release(ctx context.Context, q record.Queryer, id record.Identity) (Release, error) {
	itemID, err := itemOf(ctx, q, id)
	if errors.Is(err, ErrNotFound) {
		return NotHeld, nil
	}
	if err != nil {
		return "", err
	}
	key, err := w.KeyOf(ctx, q, itemID)
	if err != nil {
		return "", err
	}
	if key == id {
		return StillHeld, nil
	}
	var claims int
	if err := q.QueryRowContext(ctx,
		`SELECT count(*) FROM record_claim WHERE source = ? AND external_id = ?`, id.Source, id.ExternalID).Scan(&claims); err != nil {
		return "", err
	}
	if claims > 0 {
		return StillHeld, nil
	}
	held, err := w.heldByARecordOn(ctx, q, itemID, id)
	if err != nil || held == StillHeld {
		return held, err
	}
	if _, err := q.ExecContext(ctx,
		`DELETE FROM item_identity WHERE source = ? AND external_id = ?`, id.Source, id.ExternalID); err != nil {
		return "", err
	}
	return Released, nil
}

func (w *World) heldByARecordOn(ctx context.Context, q record.Queryer, itemID int64, id record.Identity) (Release, error) {
	ids, err := identitiesOn(ctx, q, itemID)
	if err != nil {
		return "", err
	}
	readings, err := w.records.Readings(ctx, q, ids)
	if err != nil {
		return "", err
	}
	for _, r := range readings {
		if slices.Contains(verdict.Derive(r.Identity, r.Values).About, id) {
			return StillHeld, nil
		}
	}
	return NotHeld, nil
}
