package store

import (
	"context"
	"database/sql"
	"errors"
)

func (s *Store) Secret(ctx context.Context, name string) (string, error) {
	var value string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM secret WHERE name = ?`, name).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return value, err
}

func (s *Store) PutSecret(ctx context.Context, name, value string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO secret (name, value) VALUES (?, ?)
		ON CONFLICT (name) DO UPDATE SET
			value = excluded.value,
			updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')`, name, value)
	return err
}
