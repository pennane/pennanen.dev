package store

import "context"

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
	return TagUnchanged, nil
}
