package store

import (
	"context"
	"encoding/json"
	"fmt"
)

type Observation struct {
	ItemID     int64           `json:"itemId"`
	Source     string          `json:"source"`
	Signal     string          `json:"signal"`
	Key        string          `json:"key"`
	ObservedAt string          `json:"observedAt"`
	Data       json.RawMessage `json:"data"`
}

type ObserveOutcome string

const (
	Observed     ObserveOutcome = "observed"
	AlreadyKnown ObserveOutcome = "already-known"
)

func (s *Store) Observe(ctx context.Context, o Observation) (ObserveOutcome, error) {
	if o.Source == "" || o.Signal == "" || o.Key == "" || o.ObservedAt == "" || !json.Valid(o.Data) {
		return "", fmt.Errorf("%w: an observation needs a source, a signal, a key, a time and json data", ErrInvalid)
	}
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO observation (item_id, source, signal, key, observed_at, data) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (source, key) DO NOTHING`,
		o.ItemID, o.Source, o.Signal, o.Key, o.ObservedAt, string(o.Data))
	if err != nil {
		return "", err
	}
	if added, err := result.RowsAffected(); err != nil || added == 0 {
		return AlreadyKnown, err
	}
	return Observed, nil
}

func (s *Store) Observations(ctx context.Context, itemID int64, source, signal string) ([]Observation, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT item_id, source, signal, key, observed_at, data FROM observation
		WHERE item_id = ? AND source = ? AND signal = ?
		ORDER BY observed_at, id`, itemID, source, signal)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	observations := []Observation{}
	for rows.Next() {
		var o Observation
		var data string
		if err := rows.Scan(&o.ItemID, &o.Source, &o.Signal, &o.Key, &o.ObservedAt, &data); err != nil {
			return nil, err
		}
		o.Data = json.RawMessage(data)
		observations = append(observations, o)
	}
	return observations, rows.Err()
}
