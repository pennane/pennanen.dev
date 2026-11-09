package store

import (
	"context"
	"database/sql"
	"fmt"
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
	return s.inTx(ctx, func(tx *sql.Tx) error {
		for _, id := range itemIDs {
			if err := changeEntry(ctx, tx, id, change); err != nil {
				return fmt.Errorf("item %d: %w", id, err)
			}
		}
		return nil
	})
}

func normalizedChange(change EntryChange) (EntryChange, error) {
	if change.Visibility != nil && *change.Visibility != Private && *change.Visibility != Public {
		return change, fmt.Errorf("%w: visibility must be private or public", ErrInvalid)
	}
	for axis, value := range change.Scores {
		if value != nil && (*value < 0 || *value > 100) {
			return change, fmt.Errorf("%w: score %q must be within 0–100", ErrInvalid, axis)
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

func changeEntry(ctx context.Context, tx *sql.Tx, itemID int64, change EntryChange) error {
	k, err := lookupKind(ctx, tx, itemID)
	if err != nil {
		return err
	}
	if change.Status != nil && !k.HasStatus(*change.Status) {
		return fmt.Errorf("%w: %s has no status %q", ErrInvalid, k.Name, *change.Status)
	}
	for axis := range change.Scores {
		if !k.HasAxis(axis) {
			return fmt.Errorf("%w: %s has no axis %q", ErrInvalid, k.Name, axis)
		}
	}
	if err := ensureEntry(ctx, tx, itemID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE entry SET
			status = coalesce(?, status),
			visibility = coalesce(?, visibility),
			notes = coalesce(?, notes),
			blurb = coalesce(?, blurb),
			updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		WHERE item_id = ?`,
		change.Status, change.Visibility, change.Notes, change.Blurb, itemID); err != nil {
		return err
	}
	for axis, value := range change.Scores {
		if err := setScore(ctx, tx, itemID, axis, value); err != nil {
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

func setScore(ctx context.Context, tx *sql.Tx, itemID int64, axis string, value *int) error {
	if value == nil {
		_, err := tx.ExecContext(ctx, `DELETE FROM score WHERE item_id = ? AND axis = ?`, itemID, axis)
		return err
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO score (item_id, axis, value) VALUES (?, ?, ?)
		ON CONFLICT (item_id, axis) DO UPDATE SET value = excluded.value`, itemID, axis, *value)
	return err
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
