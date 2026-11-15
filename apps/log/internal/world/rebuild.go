package world

import (
	"context"
	"database/sql"
	"errors"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"strconv"
)

const (
	rebuildBatchSize    = 200
	worldVersionSetting = "world-version"

	everyClaim = `
		SELECT source, external_id FROM record_claim
		WHERE (source, external_id) > (?, ?)
		ORDER BY source, external_id LIMIT ?`
	everyUnresolvedClaim = `
		SELECT c.source, c.external_id FROM record_claim c
		WHERE NOT EXISTS (SELECT 1 FROM item_identity i WHERE i.source = c.source AND i.external_id = c.external_id)
		  AND (c.source, c.external_id) > (?, ?)
		ORDER BY c.source, c.external_id LIMIT ?`
)

func (w *World) RebuildIfStale(ctx context.Context) (Report, error) {
	var built string
	err := w.db.QueryRowContext(ctx, `SELECT value FROM setting WHERE key = ?`, worldVersionSetting).Scan(&built)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Report{}, err
	}
	if built == strconv.Itoa(DeriveVersion) {
		return Report{}, nil
	}
	return w.Rebuild(ctx)
}

type settleBatch func(ctx context.Context, claims string, after record.Identity) ([]record.Identity, Report, error)

func (w *World) Rebuild(ctx context.Context) (Report, error) {
	return w.rebuild(ctx, w.db, w.settleBatchInItsOwnTransaction)
}

func (w *World) RebuildWithin(ctx context.Context, q record.Queryer) (Report, error) {
	return w.rebuild(ctx, q, func(ctx context.Context, claims string, after record.Identity) ([]record.Identity, Report, error) {
		return w.settleBatchWith(ctx, q, claims, after)
	})
}

func (w *World) rebuild(ctx context.Context, q record.Queryer, settle settleBatch) (Report, error) {
	report := Report{}
	for _, claims := range []string{everyClaim, everyUnresolvedClaim} {
		settled, err := settleInBatches(ctx, claims, settle)
		if err != nil {
			return report, err
		}
		report = report.and(settled)
	}
	_, err := q.ExecContext(ctx,
		`INSERT INTO setting (key, value) VALUES (?, ?) ON CONFLICT (key) DO UPDATE SET value = excluded.value`,
		worldVersionSetting, strconv.Itoa(DeriveVersion))
	return report, err
}

func settleInBatches(ctx context.Context, claims string, settle settleBatch) (Report, error) {
	report := Report{}
	var after record.Identity
	for {
		batch, settled, err := settle(ctx, claims, after)
		if err != nil {
			return report, err
		}
		if len(batch) == 0 {
			return report, nil
		}
		report = report.and(settled)
		after = batch[len(batch)-1]
	}
}

func (w *World) settleBatchInItsOwnTransaction(ctx context.Context, claims string, after record.Identity) ([]record.Identity, Report, error) {
	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, Report{}, err
	}
	defer tx.Rollback()
	batch, settled, err := w.settleBatchWith(ctx, tx, claims, after)
	if err != nil {
		return nil, Report{}, err
	}
	return batch, settled, tx.Commit()
}

func (w *World) settleBatchWith(ctx context.Context, q record.Queryer, claims string, after record.Identity) ([]record.Identity, Report, error) {
	batch, err := identitiesFrom(ctx, q, claims, after)
	if err != nil {
		return nil, Report{}, err
	}
	settled, err := w.Settle(ctx, q, batch)
	if err != nil {
		return nil, Report{}, err
	}
	return batch, settled, nil
}

func identitiesFrom(ctx context.Context, q record.Queryer, claims string, after record.Identity) ([]record.Identity, error) {
	rows, err := q.QueryContext(ctx, claims, after.Source, after.ExternalID, rebuildBatchSize)
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
