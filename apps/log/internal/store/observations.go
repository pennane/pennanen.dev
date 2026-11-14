package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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

func (s *Store) ItemsLastObservedAs(ctx context.Context, source string, signals []string, last string) ([]int64, error) {
	if len(signals) == 0 {
		return []int64{}, nil
	}
	marks := strings.TrimSuffix(strings.Repeat("?, ", len(signals)), ", ")
	args := []any{source}
	for _, signal := range signals {
		args = append(args, signal)
	}
	args = append(args, last)
	rows, err := s.db.QueryContext(ctx, `
		SELECT item_id FROM (
			SELECT item_id, signal, ROW_NUMBER() OVER (PARTITION BY item_id ORDER BY observed_at DESC, id DESC) AS recency
			FROM observation WHERE source = ? AND signal IN (`+marks+`)
		) WHERE recency = 1 AND signal = ?
		ORDER BY item_id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	itemIDs := []int64{}
	for rows.Next() {
		var itemID int64
		if err := rows.Scan(&itemID); err != nil {
			return nil, err
		}
		itemIDs = append(itemIDs, itemID)
	}
	return itemIDs, rows.Err()
}

func (s *Store) HasObservationKeyWithPrefix(ctx context.Context, source, prefix string) (bool, error) {
	var found bool
	err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS (SELECT 1 FROM observation WHERE source = ? AND substr(key, 1, length(?)) = ?)`,
		source, prefix, prefix).Scan(&found)
	return found, err
}
