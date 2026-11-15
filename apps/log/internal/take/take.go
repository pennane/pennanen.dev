package take

import (
	"context"
	"slices"
	"strings"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
)

type Logging string

const (
	NewlyLogged   Logging = "newly-logged"
	AlreadyLogged Logging = "already-logged"
)

const Nobody = "nobody"

func Log(ctx context.Context, q record.Queryer, itemID int64, loggedBy, visibility string) (Logging, error) {
	result, err := q.ExecContext(ctx,
		`INSERT INTO entry (item_id, logged_by, visibility) VALUES (?, ?, ?) ON CONFLICT (item_id) DO NOTHING`,
		itemID, loggedBy, visibility)
	if err != nil {
		return "", err
	}
	created, err := result.RowsAffected()
	if err != nil {
		return "", err
	}
	if created == 0 {
		return AlreadyLogged, nil
	}
	return NewlyLogged, nil
}

func ChooseStatus(ctx context.Context, q record.Queryer, itemID int64, status, at string) error {
	_, err := q.ExecContext(ctx,
		`INSERT INTO status_choice (item_id, status, chosen_at) VALUES (?, ?, ?)
		 ON CONFLICT (item_id) DO UPDATE SET status = excluded.status, chosen_at = excluded.chosen_at`,
		itemID, status, at)
	return err
}

func WithdrawStatusChoice(ctx context.Context, q record.Queryer, itemID int64) error {
	_, err := q.ExecContext(ctx, `DELETE FROM status_choice WHERE item_id = ?`, itemID)
	return err
}

type TagPresence string

const (
	TagAdded   TagPresence = "added"
	TagRemoved TagPresence = "removed"
)

func ChooseTag(ctx context.Context, q record.Queryer, itemID int64, tag string, presence TagPresence, at string) error {
	_, err := q.ExecContext(ctx,
		`INSERT INTO tag_choice (item_id, tag, presence, chosen_at) VALUES (?, ?, ?, ?)
		 ON CONFLICT (item_id, tag) DO UPDATE SET presence = excluded.presence, chosen_at = excluded.chosen_at`,
		itemID, tag, presence, at)
	return err
}

func Dismiss(ctx context.Context, q record.Queryer, itemID int64, at string) error {
	_, err := q.ExecContext(ctx,
		`INSERT INTO dismissal (item_id, dismissed_at) VALUES (?, ?) ON CONFLICT (item_id) DO NOTHING`,
		itemID, at)
	return err
}

func Undismiss(ctx context.Context, q record.Queryer, itemID int64) error {
	_, err := q.ExecContext(ctx, `DELETE FROM dismissal WHERE item_id = ?`, itemID)
	return err
}

type StatusChoice struct {
	Status   string
	ChosenAt string
}

func StatusChoices(ctx context.Context, q record.Queryer, ids []int64) (map[int64]StatusChoice, error) {
	choices := map[int64]StatusChoice{}
	err := readInChunks(ctx, q, `SELECT item_id, status, chosen_at FROM status_choice WHERE item_id IN `, ids,
		func(scan func(dest ...any) error) error {
			var itemID int64
			var choice StatusChoice
			if err := scan(&itemID, &choice.Status, &choice.ChosenAt); err != nil {
				return err
			}
			choices[itemID] = choice
			return nil
		})
	return choices, err
}

func Dismissals(ctx context.Context, q record.Queryer, ids []int64) (map[int64]string, error) {
	dismissals := map[int64]string{}
	err := readInChunks(ctx, q, `SELECT item_id, dismissed_at FROM dismissal WHERE item_id IN `, ids,
		func(scan func(dest ...any) error) error {
			var itemID int64
			var dismissedAt string
			if err := scan(&itemID, &dismissedAt); err != nil {
				return err
			}
			dismissals[itemID] = dismissedAt
			return nil
		})
	return dismissals, err
}

const idsPerRead = 500

func readInChunks(ctx context.Context, q record.Queryer, selectWhereIn string, ids []int64, eachRow func(scan func(dest ...any) error) error) error {
	for chunk := range slices.Chunk(ids, idsPerRead) {
		if err := readChunk(ctx, q, selectWhereIn, chunk, eachRow); err != nil {
			return err
		}
	}
	return nil
}

func readChunk(ctx context.Context, q record.Queryer, selectWhereIn string, chunk []int64, eachRow func(scan func(dest ...any) error) error) error {
	args := make([]any, len(chunk))
	for i, id := range chunk {
		args[i] = id
	}
	placeholders := "(" + strings.TrimSuffix(strings.Repeat("?,", len(chunk)), ",") + ")"
	rows, err := q.QueryContext(ctx, selectWhereIn+placeholders, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err := eachRow(rows.Scan); err != nil {
			return err
		}
	}
	return rows.Err()
}
