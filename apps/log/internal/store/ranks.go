package store

import (
	"cmp"
	"context"
	"database/sql"
	"fmt"
	"math"
	"slices"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
)

const AtBottom = math.MaxInt

type spot struct {
	Division string
	Position int
}

func (s *Store) Place(ctx context.Context, itemID int64, division string, index int) error {
	k, err := lookupKind(ctx, s.db, itemID)
	if err != nil {
		return err
	}
	if k.RankedAxis == "" {
		return fmt.Errorf("%w: %s has no ranked axis", ErrInvalid, k.Name)
	}
	settings, err := s.KindSettings(ctx, k)
	if err != nil {
		return err
	}
	if _, known := settings.DivisionIndex(k.RankedAxis, division); !known {
		return fmt.Errorf("%w: %s has no %s division %q", ErrInvalid, k.Name, k.RankedAxis, division)
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		if err := ensureEntry(ctx, tx, itemID); err != nil {
			return err
		}
		return moveTo(ctx, tx, k, itemID, division, index)
	})
}

func setScore(ctx context.Context, tx *sql.Tx, k kind.Kind, itemID int64, axis string, division *string) error {
	if axis != k.RankedAxis {
		return setUnrankedScore(ctx, tx, itemID, axis, division)
	}
	current, placed, err := currentSpot(ctx, tx, itemID, axis)
	if err != nil {
		return err
	}
	if division == nil {
		return removeRankedScore(ctx, tx, k, itemID, current, placed)
	}
	if placed && current.Division == *division {
		return nil
	}
	return moveTo(ctx, tx, k, itemID, *division, AtBottom)
}

func setUnrankedScore(ctx context.Context, tx *sql.Tx, itemID int64, axis string, division *string) error {
	if division == nil {
		_, err := tx.ExecContext(ctx, `DELETE FROM score WHERE item_id = ? AND axis = ?`, itemID, axis)
		return err
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO score (item_id, axis, division, position) VALUES (?, ?, ?, NULL)
		ON CONFLICT (item_id, axis) DO UPDATE SET division = excluded.division, position = NULL`, itemID, axis, *division)
	return err
}

func removeRankedScore(ctx context.Context, tx *sql.Tx, k kind.Kind, itemID int64, current spot, placed bool) error {
	if placed {
		if err := closeGap(ctx, tx, k, current); err != nil {
			return err
		}
	}
	_, err := tx.ExecContext(ctx, `DELETE FROM score WHERE item_id = ? AND axis = ?`, itemID, k.RankedAxis)
	return err
}

func clearScores(ctx context.Context, tx *sql.Tx, k kind.Kind, itemID int64) error {
	for _, axis := range k.Axes {
		if err := setScore(ctx, tx, k, itemID, axis, nil); err != nil {
			return err
		}
	}
	return nil
}

func currentSpot(ctx context.Context, tx *sql.Tx, itemID int64, axis string) (spot, bool, error) {
	var division string
	var position sql.NullInt64
	err := tx.QueryRowContext(ctx, `SELECT division, position FROM score WHERE item_id = ? AND axis = ?`, itemID, axis).Scan(&division, &position)
	if isNoRows(err) {
		return spot{}, false, nil
	}
	if err != nil {
		return spot{}, false, err
	}
	return spot{Division: division, Position: int(position.Int64)}, position.Valid, nil
}

func moveTo(ctx context.Context, tx *sql.Tx, k kind.Kind, itemID int64, division string, index int) error {
	current, placed, err := currentSpot(ctx, tx, itemID, k.RankedAxis)
	if err != nil {
		return err
	}
	if placed {
		if err := closeGap(ctx, tx, k, current); err != nil {
			return err
		}
	}
	var others int
	if err := tx.QueryRowContext(ctx, `
		SELECT count(*) FROM score
		WHERE axis = ? AND division = ? AND item_id <> ? AND item_id IN (SELECT id FROM item WHERE kind = ?)`,
		k.RankedAxis, division, itemID, k.Name).Scan(&others); err != nil {
		return err
	}
	index = min(max(index, 0), others)
	if _, err := tx.ExecContext(ctx, `
		UPDATE score SET position = position + 1
		WHERE axis = ? AND division = ? AND position >= ? AND item_id <> ? AND item_id IN (SELECT id FROM item WHERE kind = ?)`,
		k.RankedAxis, division, index, itemID, k.Name); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO score (item_id, axis, division, position) VALUES (?, ?, ?, ?)
		ON CONFLICT (item_id, axis) DO UPDATE SET division = excluded.division, position = excluded.position`,
		itemID, k.RankedAxis, division, index)
	return err
}

func closeGap(ctx context.Context, tx *sql.Tx, k kind.Kind, left spot) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE score SET position = position - 1
		WHERE axis = ? AND division = ? AND position > ? AND item_id IN (SELECT id FROM item WHERE kind = ?)`,
		k.RankedAxis, left.Division, left.Position, k.Name)
	return err
}

type rankedRow struct {
	itemID   int64
	division int
	position int
}

func (s *Store) ranksOf(ctx context.Context, k kind.Kind) (map[int64]int, error) {
	ranks := map[int64]int{}
	if k.RankedAxis == "" {
		return ranks, nil
	}
	settings, err := s.KindSettings(ctx, k)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT s.item_id, s.division, s.position FROM score s JOIN item i ON i.id = s.item_id
		WHERE i.kind = ? AND s.axis = ? AND s.position IS NOT NULL`, k.Name, k.RankedAxis)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var order []rankedRow
	for rows.Next() {
		var row rankedRow
		var division string
		if err := rows.Scan(&row.itemID, &division, &row.position); err != nil {
			return nil, err
		}
		index, known := settings.DivisionIndex(k.RankedAxis, division)
		row.division = index
		if !known {
			row.division = math.MaxInt
		}
		order = append(order, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	slices.SortFunc(order, func(a, b rankedRow) int {
		return cmp.Or(cmp.Compare(a.division, b.division), cmp.Compare(a.position, b.position), cmp.Compare(a.itemID, b.itemID))
	})
	for i, row := range order {
		ranks[row.itemID] = i + 1
	}
	return ranks, nil
}

func (s *Store) attachRanks(ctx context.Context, views []ItemView) error {
	ranksByKind := map[string]map[int64]int{}
	for i := range views {
		v := &views[i]
		if v.Entry == nil {
			continue
		}
		ranks, computed := ranksByKind[v.Kind]
		if !computed {
			k, known := kind.Get(v.Kind)
			if !known {
				continue
			}
			var err error
			if ranks, err = s.ranksOf(ctx, k); err != nil {
				return err
			}
			ranksByKind[v.Kind] = ranks
		}
		v.Entry.Rank = ranks[v.ID]
	}
	return nil
}

type Standing struct {
	ItemID     int64
	Status     string
	Visibility Visibility
	Rank       int
}

func (s *Store) Standings(ctx context.Context, k kind.Kind) ([]Standing, error) {
	ranks, err := s.ranksOf(ctx, k)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT e.item_id, e.status, e.visibility FROM entry e JOIN item i ON i.id = e.item_id
		WHERE i.kind = ?`, k.Name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var standings []Standing
	for rows.Next() {
		var st Standing
		if err := rows.Scan(&st.ItemID, &st.Status, &st.Visibility); err != nil {
			return nil, err
		}
		st.Rank = ranks[st.ItemID]
		standings = append(standings, st)
	}
	return standings, rows.Err()
}
