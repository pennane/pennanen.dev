package store

import (
	"context"

	"github.com/pennane/pennanen.dev/apps/log/internal/take"
)

type StatusOutcome string

const (
	StatusLogged    StatusOutcome = "logged"
	StatusFollowed  StatusOutcome = "followed"
	StatusAlreadySo StatusOutcome = "already-so"
	StatusKeptAsSet StatusOutcome = "kept-as-set"
)

func (s *Store) SetStatusFromSource(ctx context.Context, itemID int64, source, status string) (StatusOutcome, error) {
	var dismissed int
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM dismissal WHERE item_id = ?)`, itemID).Scan(&dismissed); err != nil {
		return "", err
	}
	if dismissed == 1 {
		return StatusKeptAsSet, nil
	}
	logging, err := take.Log(ctx, s.db, itemID, source, string(NewEntryVisibility))
	if err != nil || logging == take.AlreadyLogged {
		return StatusAlreadySo, err
	}
	return StatusLogged, nil
}

func (s *Store) ItemsWithStatusSetBy(ctx context.Context, source, kindName string) ([]int64, error) {
	return nil, nil
}
