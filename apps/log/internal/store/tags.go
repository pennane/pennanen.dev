package store

import (
	"context"
	"fmt"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
)

type TagPresence int

const (
	TagAbsent TagPresence = iota
	TagPresent
)

type TagOutcome string

const (
	TagAdded     TagOutcome = "added"
	TagRemoved   TagOutcome = "removed"
	TagUnchanged TagOutcome = "unchanged"
)

func (s *Store) SetTagFromSource(ctx context.Context, itemID int64, source, tag string, presence TagPresence) (TagOutcome, error) {
	if source == record.SourceAP {
		return "", fmt.Errorf("%w: ap tags by changing the entry", ErrInvalid)
	}
	tag = NormalizeTag(tag)
	if tag == "" {
		return "", fmt.Errorf("%w: empty tag", ErrInvalid)
	}
	query := `DELETE FROM tag WHERE item_id = ? AND tag = ? AND added_by = ?`
	if presence == TagPresent {
		query = `INSERT INTO tag (item_id, tag, added_by) VALUES (?, ?, ?) ON CONFLICT DO NOTHING`
	}
	result, err := s.db.ExecContext(ctx, query, itemID, tag, source)
	if err != nil {
		return "", err
	}
	changed, err := result.RowsAffected()
	switch {
	case err != nil:
		return "", err
	case changed == 0:
		return TagUnchanged, nil
	case presence == TagPresent:
		return TagAdded, nil
	default:
		return TagRemoved, nil
	}
}
