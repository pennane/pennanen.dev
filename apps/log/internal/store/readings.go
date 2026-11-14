package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
)

const rereadBatchSize = 200

func putReading(ctx context.Context, q queryer, itemID int64, source string, data json.RawMessage) error {
	reading, readability := verdict.ReadClaim(verdict.Claim{Source: source, Data: data})
	if readability == verdict.Unreadable {
		_, err := q.ExecContext(ctx, `DELETE FROM claim_reading WHERE item_id = ? AND source = ?`, itemID, source)
		return err
	}
	encoded, err := verdict.Encode(reading)
	if err != nil {
		return err
	}
	_, err = q.ExecContext(ctx, `
		INSERT INTO claim_reading (item_id, source, version, reading) VALUES (?, ?, ?, ?)
		ON CONFLICT (item_id, source) DO UPDATE SET version = excluded.version, reading = excluded.reading`,
		itemID, source, verdict.ReadersVersion, string(encoded))
	return err
}

type claimPosition struct {
	itemID int64
	source string
}

type unreadClaim struct {
	claimPosition
	data json.RawMessage
}

func (s *Store) RereadClaims(ctx context.Context) (int, error) {
	reread, after := 0, claimPosition{}
	for {
		var batch []unreadClaim
		err := s.inTx(ctx, func(tx *sql.Tx) error {
			var err error
			if batch, err = unreadClaimsAfter(ctx, tx, after); err != nil {
				return err
			}
			for _, c := range batch {
				if err := putReading(ctx, tx, c.itemID, c.source, c.data); err != nil {
					return fmt.Errorf("read claim %d %s: %w", c.itemID, c.source, err)
				}
			}
			return nil
		})
		if err != nil {
			return reread, err
		}
		if len(batch) == 0 {
			return reread, nil
		}
		reread += len(batch)
		after = batch[len(batch)-1].claimPosition
		slog.Info("claims re-read", "total", reread)
	}
}

func unreadClaimsAfter(ctx context.Context, q queryer, after claimPosition) ([]unreadClaim, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT c.item_id, c.source, c.data FROM claim c
		LEFT JOIN claim_reading r ON r.item_id = c.item_id AND r.source = c.source
		WHERE (r.item_id IS NULL OR r.version <> ?) AND (c.item_id, c.source) > (?, ?)
		ORDER BY c.item_id, c.source LIMIT ?`,
		verdict.ReadersVersion, after.itemID, after.source, rereadBatchSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var batch []unreadClaim
	for rows.Next() {
		var c unreadClaim
		var data string
		if err := rows.Scan(&c.itemID, &c.source, &data); err != nil {
			return nil, err
		}
		c.data = json.RawMessage(data)
		batch = append(batch, c)
	}
	return batch, rows.Err()
}

func (s *Store) readingsByItem(ctx context.Context, byID map[int64]*ItemView) (map[int64][]verdict.Reading, error) {
	ids := make([]int64, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	byItem := map[int64][]verdict.Reading{}
	err := forChunks(ids, func(chunk []int64) error {
		rows, err := s.db.QueryContext(ctx, `
			SELECT r.item_id, r.source, coalesce(c.external_id, ''), c.fetched_at, r.reading
			FROM claim_reading r JOIN claim c ON c.item_id = r.item_id AND c.source = r.source
			WHERE r.item_id IN (`+placeholders(len(chunk))+`) ORDER BY r.item_id, r.source`, int64sAsArgs(chunk)...)
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
