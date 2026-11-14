package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
)

type StatusOutcome string

const (
	StatusLogged    StatusOutcome = "logged"
	StatusFollowed  StatusOutcome = "followed"
	StatusAlreadySo StatusOutcome = "already-so"
	StatusKeptAsSet StatusOutcome = "kept-as-set"
)

func (s *Store) SetStatusFromSource(ctx context.Context, itemID int64, source, status string) (StatusOutcome, error) {
	if source == kind.SourceAP {
		return "", fmt.Errorf("%w: ap sets statuses by changing the entry", ErrInvalid)
	}
	var outcome StatusOutcome
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		k, err := lookupKind(ctx, tx, itemID)
		if err != nil {
			return err
		}
		if !k.HasStatus(status) {
			return fmt.Errorf("%w: %s has no status %q", ErrInvalid, k.Name, status)
		}
		var current, setBy string
		err = tx.QueryRowContext(ctx, `SELECT status, status_set_by FROM entry WHERE item_id = ?`, itemID).Scan(&current, &setBy)
		switch {
		case isNoRows(err):
			outcome = StatusLogged
			_, err = tx.ExecContext(ctx, `INSERT INTO entry (item_id, visibility, status, status_set_by) VALUES (?, ?, ?, ?)`, itemID, NewEntryVisibility, status, source)
			return err
		case err != nil:
			return err
		case !sourceMayChange(source, current, setBy):
			outcome = StatusKeptAsSet
			return nil
		case current == status && setBy == source:
			outcome = StatusAlreadySo
			return nil
		}
		outcome = StatusFollowed
		_, err = tx.ExecContext(ctx, `
			UPDATE entry SET status = ?, status_set_by = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
			WHERE item_id = ?`, status, source, itemID)
		return err
	})
	return outcome, err
}

func sourceMayChange(source, current, setBy string) bool {
	return setBy == source || (setBy != kind.SourceAP && current == kind.Untriaged)
}
