package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
)

func (s *Store) ListItems(ctx context.Context, filter ItemFilter) ([]ItemView, error) {
	where, args := []string{"1 = 1"}, []any{}
	if filter.Kind != "" {
		where, args = append(where, "i.kind = ?"), append(args, filter.Kind)
	}
	if filter.Status != "" {
		where, args = append(where, "e.status = ?"), append(args, filter.Status)
	}
	if filter.Visibility != "" {
		where, args = append(where, "e.visibility = ?"), append(args, filter.Visibility)
	}
	if filter.Tag != "" {
		where = append(where, "EXISTS (SELECT 1 FROM tag t WHERE t.item_id = i.id AND (t.tag = ? OR t.tag LIKE ? ESCAPE '\\'))")
		args = append(args, filter.Tag, escapeLike(filter.Tag)+"/%")
	}
	if filter.Needs == NeedsMatch {
		k, known := kind.Get(filter.Kind)
		if !known || k.IdentitySource == "" {
			return nil, fmt.Errorf("%w: needs=match takes a kind that has an identity source", ErrInvalid)
		}
		where = append(where, "e.status <> ? AND NOT EXISTS (SELECT 1 FROM claim m WHERE m.item_id = i.id AND m.source = ?)")
		args = append(args, kind.Dismissed, k.IdentitySource)
	}
	views, err := s.queryViews(ctx, strings.Join(where, " AND "), args...)
	if err != nil {
		return nil, err
	}
	return slices.DeleteFunc(views, func(v ItemView) bool { return !matchesDerived(v, filter) }), nil
}

func matchesDerived(v ItemView, filter ItemFilter) bool {
	if filter.Hint != kind.NoHint && v.Hint != filter.Hint {
		return false
	}
	if filter.Query != "" && !strings.Contains(strings.ToLower(verdict.Title(v.Verdicts)), strings.ToLower(filter.Query)) {
		return false
	}
	if filter.Needs == NeedsScore && !needsScore(v) {
		return false
	}
	return true
}

func needsScore(v ItemView) bool {
	k, known := kind.Get(v.Kind)
	return known && v.Entry != nil && k.Scores(v.Entry.Status) && len(v.Entry.Scores) == 0
}

func (s *Store) CountUntriaged(ctx context.Context, kindName string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT count(*) FROM entry e JOIN item i ON i.id = e.item_id WHERE i.kind = ? AND e.status = ?`,
		kindName, kind.Untriaged).Scan(&count)
	return count, err
}

func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

func (s *Store) GetItem(ctx context.Context, itemID int64) (ItemView, error) {
	return s.getOne(ctx, "i.id = ?", itemID)
}

func (s *Store) GetItemBySlug(ctx context.Context, kindName, slug string) (ItemView, error) {
	return s.getOne(ctx, "i.kind = ? AND i.slug = ?", kindName, slug)
}

func (s *Store) GetItems(ctx context.Context, itemIDs []int64) ([]ItemView, error) {
	if len(itemIDs) == 0 {
		return nil, nil
	}
	return s.queryViews(ctx, "i.id IN ("+placeholders(len(itemIDs))+")", int64sAsArgs(itemIDs)...)
}

func (s *Store) getOne(ctx context.Context, where string, args ...any) (ItemView, error) {
	views, err := s.queryViews(ctx, where, args...)
	if err != nil {
		return ItemView{}, err
	}
	if len(views) == 0 {
		return ItemView{}, ErrNotFound
	}
	v := views[0]
	claims, err := s.claimsOf(ctx, []int64{v.ID})
	if err != nil {
		return v, err
	}
	for _, c := range claims {
		v.Claims = append(v.Claims, c.Claim)
	}
	if v.Moments, err = s.momentsOf(ctx, v.ID); err != nil {
		return v, err
	}
	if v.Relations, err = s.relationsOf(ctx, v.ID); err != nil {
		return v, err
	}
	return v, nil
}

func (s *Store) queryViews(ctx context.Context, where string, args ...any) ([]ItemView, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT i.id, i.kind, i.slug, i.created_at,
		       e.status, e.visibility, e.blurb, e.time_spent, e.updated_at
		FROM item i LEFT JOIN entry e ON e.item_id = i.id
		WHERE `+where+`
		ORDER BY coalesce(e.updated_at, i.created_at) DESC, i.id DESC`, args...)
	if err != nil {
		return nil, err
	}
	var views []ItemView
	byID := map[int64]*ItemView{}
	for rows.Next() {
		var v ItemView
		var status, visibility, blurb, timeSpent, updatedAt sql.NullString
		if err := rows.Scan(&v.ID, &v.Kind, &v.Slug, &v.CreatedAt, &status, &visibility, &blurb, &timeSpent, &updatedAt); err != nil {
			rows.Close()
			return nil, err
		}
		if status.Valid {
			v.Entry = &Entry{
				Status: status.String, Visibility: Visibility(visibility.String),
				Blurb: blurb.String, TimeSpent: TimeSpent(timeSpent.String), UpdatedAt: updatedAt.String,
				Scores: map[string]string{}, Tags: []string{},
			}
		}
		views = append(views, v)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	ids := make([]int64, len(views))
	for i := range views {
		ids[i] = views[i].ID
		byID[views[i].ID] = &views[i]
	}
	if err := s.attachVerdicts(ctx, ids, byID); err != nil {
		return nil, err
	}
	if err := s.attachScoresAndTags(ctx, ids, byID); err != nil {
		return nil, err
	}
	if err := s.attachRanks(ctx, views); err != nil {
		return nil, err
	}
	return views, nil
}

func (s *Store) attachVerdicts(ctx context.Context, ids []int64, byID map[int64]*ItemView) error {
	claims, err := s.claimsOf(ctx, ids)
	if err != nil {
		return err
	}
	claimsByItem := map[int64][]verdict.Claim{}
	for _, c := range claims {
		claimsByItem[c.itemID] = append(claimsByItem[c.itemID], c.Claim)
	}
	for id, v := range byID {
		k, known := kind.Get(v.Kind)
		if !known {
			continue
		}
		v.Verdicts = verdict.Decide(k, claimsByItem[id])
		if v.Entry != nil && v.Entry.Status == kind.Untriaged {
			v.Hint = k.Hint(v.Verdicts.Values)
		}
	}
	return nil
}

type itemClaim struct {
	itemID int64
	verdict.Claim
}

func (s *Store) claimsOf(ctx context.Context, ids []int64) ([]itemClaim, error) {
	var claims []itemClaim
	err := forChunks(ids, func(chunk []int64) error {
		rows, err := s.db.QueryContext(ctx, `
			SELECT item_id, source, coalesce(external_id, ''), data, fetched_at FROM claim
			WHERE item_id IN (`+placeholders(len(chunk))+`) ORDER BY item_id, source`, int64sAsArgs(chunk)...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var c itemClaim
			var data string
			if err := rows.Scan(&c.itemID, &c.Source, &c.ExternalID, &data, &c.FetchedAt); err != nil {
				return err
			}
			c.Data = json.RawMessage(data)
			claims = append(claims, c)
		}
		return rows.Err()
	})
	return claims, err
}

func (s *Store) attachScoresAndTags(ctx context.Context, ids []int64, byID map[int64]*ItemView) error {
	return forChunks(ids, func(chunk []int64) error {
		in, args := placeholders(len(chunk)), int64sAsArgs(chunk)
		rows, err := s.db.QueryContext(ctx, `SELECT item_id, axis, division FROM score WHERE item_id IN (`+in+`)`, args...)
		if err != nil {
			return err
		}
		for rows.Next() {
			var id int64
			var axis, division string
			if err := rows.Scan(&id, &axis, &division); err != nil {
				rows.Close()
				return err
			}
			if v := byID[id]; v.Entry != nil {
				v.Entry.Scores[axis] = division
			}
		}
		rows.Close()
		rows, err = s.db.QueryContext(ctx, `SELECT item_id, tag FROM tag WHERE item_id IN (`+in+`) ORDER BY tag`, args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id int64
			var t string
			if err := rows.Scan(&id, &t); err != nil {
				return err
			}
			if v := byID[id]; v.Entry != nil {
				v.Entry.Tags = append(v.Entry.Tags, t)
			}
		}
		return rows.Err()
	})
}

func (s *Store) relationsOf(ctx context.Context, itemID int64) ([]Relation, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT from_id, type, to_id FROM relation WHERE from_id = ? OR to_id = ? ORDER BY type, from_id, to_id`, itemID, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var relations []Relation
	for rows.Next() {
		var r Relation
		if err := rows.Scan(&r.FromID, &r.Type, &r.ToID); err != nil {
			return nil, err
		}
		relations = append(relations, r)
	}
	return relations, rows.Err()
}

const sqliteVariableLimit = 500

func forChunks(ids []int64, work func(chunk []int64) error) error {
	for chunk := range slices.Chunk(ids, sqliteVariableLimit) {
		if err := work(chunk); err != nil {
			return err
		}
	}
	return nil
}

func placeholders(n int) string {
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
}

func int64sAsArgs(ids []int64) []any {
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	return args
}

func isNoRows(err error) bool { return errors.Is(err, sql.ErrNoRows) }
