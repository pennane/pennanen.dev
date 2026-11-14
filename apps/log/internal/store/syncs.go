package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

func (s *Store) SaveSyncReport(ctx context.Context, name string, report any) error {
	encoded, err := json.Marshal(report)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO sync_report (name, report) VALUES (?, ?) ON CONFLICT (name) DO UPDATE SET report = excluded.report`, name, string(encoded))
	return err
}

func (s *Store) SyncReport(ctx context.Context, name string, into any) (bool, error) {
	var report string
	err := s.db.QueryRowContext(ctx, `SELECT report FROM sync_report WHERE name = ?`, name).Scan(&report)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal([]byte(report), into)
}
