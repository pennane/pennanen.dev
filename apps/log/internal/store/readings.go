package store

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
)

func (s *Store) RereadClaims(ctx context.Context) (int, error) {
	reread, err := s.records.Reread(ctx)
	if err != nil {
		return reread, err
	}
	report, err := s.world.RebuildIfStale(ctx)
	if err != nil {
		return reread, err
	}
	if len(report.Created) > 0 || len(report.Moves) > 0 || len(report.Overlaps) > 0 {
		slog.Info("world rebuilt", "created", len(report.Created), "moves", len(report.Moves), "overlaps", len(report.Overlaps))
	}
	return reread, nil
}

func (s *Store) readingsByItem(ctx context.Context, byID map[int64]*ItemView) (map[int64][]verdict.Reading, error) {
	ids := make([]int64, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	byItem := map[int64][]verdict.Reading{}
	err := forChunks(ids, func(chunk []int64) error {
		rows, err := s.db.QueryContext(ctx, `
			SELECT i.item_id, r.source, r.external_id, c.fetched_at, r.reading
			FROM item_identity i
			JOIN record_reading r ON r.source = i.source AND r.external_id = i.external_id
			JOIN record_claim c ON c.source = r.source AND c.external_id = r.external_id
			WHERE i.item_id IN (`+placeholders(len(chunk))+`) ORDER BY i.item_id, r.source, r.external_id`, int64sAsArgs(chunk)...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id int64
			var source, externalID, fetchedAt, stored string
			if err := rows.Scan(&id, &source, &externalID, &fetchedAt, &stored); err != nil {
				return err
			}
			k, known := kind.Get(byID[id].Kind)
			if !known {
				continue
			}
			reading, err := verdict.Decode(k, source, externalID, fetchedAt, json.RawMessage(stored))
			if err != nil {
				return fmt.Errorf("decode reading of item %d %s: %w", id, source, err)
			}
			byItem[id] = append(byItem[id], reading)
		}
		return rows.Err()
	})
	return byItem, err
}
