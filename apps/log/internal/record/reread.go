package record

import (
	"context"
	"encoding/json"
)

const rereadBatchSize = 200

type pendingReread struct {
	id   Identity
	data string
}

func (s *Store) Reread(ctx context.Context) (int, error) {
	total := 0
	var after Identity
	for {
		batch, err := s.rereadBatch(ctx, after)
		if err != nil {
			return total, err
		}
		if len(batch) == 0 {
			return total, nil
		}
		total += len(batch)
		after = batch[len(batch)-1]
	}
}

func (s *Store) rereadBatch(ctx context.Context, after Identity) ([]Identity, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	pending, err := s.pendingRereads(ctx, tx, after)
	if err != nil {
		return nil, err
	}
	ids := make([]Identity, 0, len(pending))
	for _, p := range pending {
		reading, readability := s.read(p.id.Source, json.RawMessage(p.data))
		if err := s.writeReading(ctx, tx, p.id, reading, readability); err != nil {
			return nil, err
		}
		ids = append(ids, p.id)
	}
	return ids, tx.Commit()
}

func (s *Store) pendingRereads(ctx context.Context, q Queryer, after Identity) ([]pendingReread, error) {
	rows, err := q.QueryContext(ctx,
		`SELECT c.source, c.external_id, c.data
		 FROM record_claim c
		 LEFT JOIN record_reading r ON r.source = c.source AND r.external_id = c.external_id
		 WHERE (r.version IS NULL OR r.version <> ?)
		   AND (c.source, c.external_id) > (?, ?)
		 ORDER BY c.source, c.external_id
		 LIMIT ?`,
		s.readersVersion, after.Source, after.ExternalID, rereadBatchSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var pending []pendingReread
	for rows.Next() {
		var p pendingReread
		if err := rows.Scan(&p.id.Source, &p.id.ExternalID, &p.data); err != nil {
			return nil, err
		}
		pending = append(pending, p)
	}
	return pending, rows.Err()
}
