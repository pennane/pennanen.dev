package world

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
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

func (w *World) Rebuild(ctx context.Context) (Report, error) {
	report := Report{}
	for _, claims := range []string{everyClaim, everyUnresolvedClaim} {
		settled, err := w.settleInBatches(ctx, claims)
		if err != nil {
			return report, err
		}
		report = report.and(settled)
	}
	_, err := w.db.ExecContext(ctx,
		`INSERT INTO setting (key, value) VALUES (?, ?) ON CONFLICT (key) DO UPDATE SET value = excluded.value`,
		worldVersionSetting, strconv.Itoa(DeriveVersion))
	return report, err
}

func (w *World) settleInBatches(ctx context.Context, claims string) (Report, error) {
	report := Report{}
	var after record.Identity
	for {
		batch, settled, err := w.settleBatch(ctx, claims, after)
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

func (w *World) settleBatch(ctx context.Context, claims string, after record.Identity) ([]record.Identity, Report, error) {
	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, Report{}, err
	}
	defer tx.Rollback()
	batch, err := identitiesFrom(ctx, tx, claims, after)
	if err != nil {
		return nil, Report{}, err
	}
	settled, err := w.Settle(ctx, tx, batch)
	if err != nil {
		return nil, Report{}, err
	}
	return batch, settled, tx.Commit()
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
