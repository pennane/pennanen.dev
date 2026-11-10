package store

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"strings"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
)

func (s *Store) ChangeEntry(ctx context.Context, itemID int64, change EntryChange) error {
	return s.ChangeEntries(ctx, []int64{itemID}, change)
}

func (s *Store) ChangeEntries(ctx context.Context, itemIDs []int64, change EntryChange) error {
	change, err := normalizedChange(change)
	if err != nil {
		return err
	}
	axes, err := s.Axes(ctx)
	if err != nil {
		return err
	}
	settings, err := s.AllKindSettings(ctx)
	if err != nil {
		return err
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		for _, id := range itemIDs {
			if err := changeEntry(ctx, tx, axes, settings, id, change); err != nil {
				return fmt.Errorf("item %d: %w", id, err)
			}
		}
		return nil
	})
}

func normalizedChange(change EntryChange) (EntryChange, error) {
	if change.Visibility != nil && !slices.Contains(Visibilities, *change.Visibility) {
		return change, fmt.Errorf("%w: visibility must be one of %v", ErrInvalid, Visibilities)
	}
	if change.TimeSpent != nil && *change.TimeSpent != TimeSpentUnsaid && !slices.Contains(TimeSpentChoices, *change.TimeSpent) {
		return change, fmt.Errorf("%w: time spent must be one of %v, or empty to unsay it", ErrInvalid, TimeSpentChoices)
	}
	for axis, division := range change.Scores {
		if division != nil {
			key := keyFrom(*division)
			change.Scores[axis] = &key
		}
	}
	var err error
	if change.AddTags, err = normalizedTags(change.AddTags); err != nil {
		return change, err
	}
	if change.RemoveTags, err = normalizedTags(change.RemoveTags); err != nil {
		return change, err
	}
	return change, nil
}

func NormalizeTag(tag string) string {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(tag)), "/")
	kept := parts[:0]
	for _, p := range parts {
		if p = strings.Join(strings.Fields(p), "-"); p != "" {
			kept = append(kept, p)
		}
	}
	return strings.Join(kept, "/")
}

func normalizedTags(tags []string) ([]string, error) {
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		normalized := NormalizeTag(t)
		if normalized == "" {
			return nil, fmt.Errorf("%w: empty tag", ErrInvalid)
		}
		out = append(out, normalized)
	}
	return out, nil
}

func changeEntry(ctx context.Context, tx *sql.Tx, axes Axes, settings map[string]KindSettings, itemID int64, change EntryChange) error {
	k, err := lookupKind(ctx, tx, itemID)
	if err != nil {
		return err
	}
	if change.Status != nil && !k.HasStatus(*change.Status) {
		return fmt.Errorf("%w: %s has no status %q", ErrInvalid, k.Name, *change.Status)
	}
	for axis, division := range change.Scores {
		if !settings[k.Name].Uses(axis) {
			return fmt.Errorf("%w: %s has no axis %q", ErrInvalid, k.Name, axis)
		}
		if _, known := axes.DivisionIndex(axis, deref(division)); division != nil && !known {
			return fmt.Errorf("%w: %s has no %s division %q", ErrInvalid, k.Name, axis, *division)
		}
	}
	if err := ensureEntry(ctx, tx, itemID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE entry SET
			status = coalesce(?, status),
			visibility = coalesce(?, visibility),
			blurb = coalesce(?, blurb),
			time_spent = CASE WHEN ? IS NULL THEN time_spent ELSE nullif(?, '') END,
			updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		WHERE item_id = ?`,
		change.Status, change.Visibility, change.Blurb, change.TimeSpent, change.TimeSpent, itemID); err != nil {
		return err
	}
	if change.Status != nil && *change.Status == kind.Dismissed {
		if err := clearScores(ctx, tx, k, itemID); err != nil {
			return err
		}
	}
	for axis, division := range change.Scores {
		if err := setScore(ctx, tx, k, itemID, axis, division); err != nil {
			return err
		}
	}
	for _, t := range change.AddTags {
		if _, err := tx.ExecContext(ctx, `INSERT INTO tag (item_id, tag) VALUES (?, ?) ON CONFLICT DO NOTHING`, itemID, t); err != nil {
			return err
		}
	}
	for _, t := range change.RemoveTags {
		if _, err := tx.ExecContext(ctx, `DELETE FROM tag WHERE item_id = ? AND tag = ?`, itemID, t); err != nil {
			return err
		}
	}
	return nil
}

func deref(text *string) string {
	if text == nil {
		return ""
	}
	return *text
}

func lookupKind(ctx context.Context, q queryer, itemID int64) (kind.Kind, error) {
	var name string
	if err := q.QueryRowContext(ctx, `SELECT kind FROM item WHERE id = ?`, itemID).Scan(&name); err != nil {
		if isNoRows(err) {
			return kind.Kind{}, ErrNotFound
		}
		return kind.Kind{}, err
	}
	k, known := kind.Get(name)
	if !known {
		return kind.Kind{}, fmt.Errorf("item %d has unknown kind %q", itemID, name)
	}
	return k, nil
}
