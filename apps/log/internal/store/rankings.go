package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
)

type RankingLayout string

const (
	Numbered RankingLayout = "numbered"
	Tiered   RankingLayout = "tiered"
)

type Placement struct {
	ItemID int64  `json:"itemId"`
	Note   string `json:"note"`
}

type Tier[P any] struct {
	Name       string `json:"name"`
	Placements []P    `json:"placements"`
}

type Order[P any] struct {
	Layout     RankingLayout `json:"layout"`
	Placements []P           `json:"placements"`
	Tiers      []Tier[P]     `json:"tiers"`
}

type RankingOrder = Order[Placement]

func (o Order[P]) MarshalJSON() ([]byte, error) {
	if o.Layout == Tiered {
		tiers := orEmpty(o.Tiers)
		for i := range tiers {
			tiers[i].Placements = orEmpty(tiers[i].Placements)
		}
		return json.Marshal(struct {
			Layout RankingLayout `json:"layout"`
			Tiers  []Tier[P]     `json:"tiers"`
		}{o.Layout, tiers})
	}
	return json.Marshal(struct {
		Layout     RankingLayout `json:"layout"`
		Placements []P           `json:"placements"`
	}{Numbered, orEmpty(o.Placements)})
}

func (o Order[P]) All() []P {
	all := append([]P{}, o.Placements...)
	for _, t := range o.Tiers {
		all = append(all, t.Placements...)
	}
	return all
}

func MapOrder[P, Q any](o Order[P], convert func(P) (Q, bool)) Order[Q] {
	keep := func(ps []P) []Q {
		out := []Q{}
		for _, p := range ps {
			if q, kept := convert(p); kept {
				out = append(out, q)
			}
		}
		return out
	}
	mapped := Order[Q]{Layout: o.Layout, Placements: keep(o.Placements)}
	for _, t := range o.Tiers {
		mapped.Tiers = append(mapped.Tiers, Tier[Q]{Name: t.Name, Placements: keep(t.Placements)})
	}
	return mapped
}

type Ranking struct {
	ID          int64        `json:"id"`
	Slug        string       `json:"slug"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	UpdatedAt   string       `json:"updatedAt"`
	Order       RankingOrder `json:"order"`
}

type RankingSummary struct {
	ID          int64  `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	UpdatedAt   string `json:"updatedAt"`
	Placed      int    `json:"placed"`
}

type RankingChange struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

const nowUTC = `strftime('%Y-%m-%dT%H:%M:%SZ', 'now')`

func (s *Store) CreateRanking(ctx context.Context, name, description string) (int64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, fmt.Errorf("%w: name is required", ErrInvalid)
	}
	base := Slugify(name)
	for attempt := 1; ; attempt++ {
		slug := base
		if attempt > 1 {
			slug = fmt.Sprintf("%s-%d", base, attempt)
		}
		result, err := s.db.ExecContext(ctx,
			`INSERT INTO ranking (slug, name, description) VALUES (?, ?, ?) ON CONFLICT (slug) DO NOTHING`,
			slug, name, strings.TrimSpace(description))
		if err != nil {
			return 0, err
		}
		if inserted, _ := result.RowsAffected(); inserted == 0 {
			continue
		}
		return result.LastInsertId()
	}
}

func (s *Store) ChangeRanking(ctx context.Context, id int64, change RankingChange) error {
	sets, args := []string{"updated_at = " + nowUTC}, []any{}
	if change.Name != nil {
		name := strings.TrimSpace(*change.Name)
		if name == "" {
			return fmt.Errorf("%w: name is required", ErrInvalid)
		}
		sets, args = append(sets, "name = ?"), append(args, name)
	}
	if change.Description != nil {
		sets, args = append(sets, "description = ?"), append(args, strings.TrimSpace(*change.Description))
	}
	result, err := s.db.ExecContext(ctx, `UPDATE ranking SET `+strings.Join(sets, ", ")+` WHERE id = ?`, append(args, id)...)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func (s *Store) DeleteRanking(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM ranking WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func (s *Store) ListRankings(ctx context.Context) ([]RankingSummary, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT r.id, r.slug, r.name, r.description, r.updated_at,
		       (SELECT count(*) FROM placement p WHERE p.ranking_id = r.id)
		FROM ranking r
		ORDER BY r.updated_at DESC, r.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	summaries := []RankingSummary{}
	for rows.Next() {
		var r RankingSummary
		if err := rows.Scan(&r.ID, &r.Slug, &r.Name, &r.Description, &r.UpdatedAt, &r.Placed); err != nil {
			return nil, err
		}
		summaries = append(summaries, r)
	}
	return summaries, rows.Err()
}

func (s *Store) GetRanking(ctx context.Context, id int64) (Ranking, error) {
	return s.getRanking(ctx, "id = ?", id)
}

func (s *Store) GetRankingBySlug(ctx context.Context, slug string) (Ranking, error) {
	return s.getRanking(ctx, "slug = ?", slug)
}

func (s *Store) getRanking(ctx context.Context, where string, arg any) (Ranking, error) {
	var r Ranking
	err := s.db.QueryRowContext(ctx, `SELECT id, slug, name, description, updated_at, layout FROM ranking WHERE `+where, arg).
		Scan(&r.ID, &r.Slug, &r.Name, &r.Description, &r.UpdatedAt, &r.Order.Layout)
	if errors.Is(err, sql.ErrNoRows) {
		return r, ErrNotFound
	}
	if err != nil {
		return r, err
	}
	tierAt, err := s.readTiers(ctx, &r)
	if err != nil {
		return r, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT item_id, note, tier_position FROM placement WHERE ranking_id = ? ORDER BY position`, r.ID)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		var p Placement
		var tierPosition sql.NullInt64
		if err := rows.Scan(&p.ItemID, &p.Note, &tierPosition); err != nil {
			return r, err
		}
		if i, tiered := tierAt[tierPosition.Int64]; tierPosition.Valid && tiered {
			r.Order.Tiers[i].Placements = append(r.Order.Tiers[i].Placements, p)
		} else {
			r.Order.Placements = append(r.Order.Placements, p)
		}
	}
	return r, rows.Err()
}

func (s *Store) readTiers(ctx context.Context, r *Ranking) (map[int64]int, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT position, name FROM ranking_tier WHERE ranking_id = ? ORDER BY position`, r.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tierAt := map[int64]int{}
	for rows.Next() {
		var position int64
		var t Tier[Placement]
		if err := rows.Scan(&position, &t.Name); err != nil {
			return nil, err
		}
		tierAt[position] = len(r.Order.Tiers)
		r.Order.Tiers = append(r.Order.Tiers, t)
	}
	return tierAt, rows.Err()
}

func (s *Store) PutRankingOrder(ctx context.Context, id int64, order RankingOrder) error {
	if err := validateOrderShape(order); err != nil {
		return err
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `UPDATE ranking SET layout = ?, updated_at = `+nowUTC+` WHERE id = ?`, order.Layout, id)
		if err != nil {
			return err
		}
		if err := requireAffected(result); err != nil {
			return err
		}
		if err := requirePlaceable(ctx, tx, order.All()); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM placement WHERE ranking_id = ?`, id); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM ranking_tier WHERE ranking_id = ?`, id); err != nil {
			return err
		}
		return insertOrder(ctx, tx, id, order)
	})
}

func validateOrderShape(order RankingOrder) error {
	switch order.Layout {
	case Numbered:
		if len(order.Tiers) > 0 {
			return fmt.Errorf("%w: a numbered ranking has no tiers", ErrInvalid)
		}
	case Tiered:
		if len(order.Placements) > 0 {
			return fmt.Errorf("%w: a tiered ranking places items inside tiers", ErrInvalid)
		}
		names := map[string]struct{}{}
		for _, t := range order.Tiers {
			name := strings.TrimSpace(t.Name)
			if name == "" {
				return fmt.Errorf("%w: tiers need a name", ErrInvalid)
			}
			if _, seen := names[name]; seen {
				return fmt.Errorf("%w: tier %q appears twice", ErrInvalid, name)
			}
			names[name] = struct{}{}
		}
	default:
		return fmt.Errorf("%w: layout must be %q or %q", ErrInvalid, Numbered, Tiered)
	}
	placed := map[int64]struct{}{}
	for _, p := range order.All() {
		if _, seen := placed[p.ItemID]; seen {
			return fmt.Errorf("%w: item %d is placed twice", ErrInvalid, p.ItemID)
		}
		placed[p.ItemID] = struct{}{}
	}
	return nil
}

func requirePlaceable(ctx context.Context, q queryer, placements []Placement) error {
	if len(placements) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(placements))
	for _, p := range placements {
		ids = append(ids, p.ItemID)
	}
	rows, err := q.QueryContext(ctx, `
		SELECT i.id, coalesce(e.status, '') FROM item i LEFT JOIN entry e ON e.item_id = i.id
		WHERE i.id IN (`+placeholders(len(ids))+`)`, int64sAsArgs(ids)...)
	if err != nil {
		return err
	}
	defer rows.Close()
	found := 0
	for rows.Next() {
		var id int64
		var status string
		if err := rows.Scan(&id, &status); err != nil {
			return err
		}
		if status == kind.Dismissed {
			return fmt.Errorf("%w: item %d is dismissed", ErrInvalid, id)
		}
		found++
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if found != len(ids) {
		return fmt.Errorf("%w: unknown item in ranking", ErrInvalid)
	}
	return nil
}

func insertOrder(ctx context.Context, q queryer, rankingID int64, order RankingOrder) error {
	position := 0
	place := func(p Placement, tierPosition sql.NullInt64) error {
		position++
		_, err := q.ExecContext(ctx,
			`INSERT INTO placement (ranking_id, item_id, position, tier_position, note) VALUES (?, ?, ?, ?, ?)`,
			rankingID, p.ItemID, position, tierPosition, strings.TrimSpace(p.Note))
		return err
	}
	for _, p := range order.Placements {
		if err := place(p, sql.NullInt64{}); err != nil {
			return err
		}
	}
	for i, t := range order.Tiers {
		tierPosition := sql.NullInt64{Int64: int64(i + 1), Valid: true}
		if _, err := q.ExecContext(ctx, `INSERT INTO ranking_tier (ranking_id, position, name) VALUES (?, ?, ?)`,
			rankingID, tierPosition.Int64, strings.TrimSpace(t.Name)); err != nil {
			return err
		}
		for _, p := range t.Placements {
			if err := place(p, tierPosition); err != nil {
				return err
			}
		}
	}
	return nil
}

func orEmpty[T any](xs []T) []T {
	if xs == nil {
		return []T{}
	}
	return xs
}
