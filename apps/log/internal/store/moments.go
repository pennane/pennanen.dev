package store

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
)

var isoDate = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func (s *Store) AddMoment(ctx context.Context, itemID int64, m NewMoment) (int64, error) {
	if m.Visibility == "" {
		m.Visibility = InheritFromEntry
	}
	if m.ScoresEffect == "" {
		m.ScoresEffect = MomentScoresUpdateEntry
	}
	if err := validateMoment(m); err != nil {
		return 0, err
	}
	var momentID int64
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		k, err := lookupKind(ctx, tx, itemID)
		if err != nil {
			return err
		}
		for axis := range m.Scores {
			if !k.HasAxis(axis) {
				return fmt.Errorf("%w: %s has no axis %q", ErrInvalid, k.Name, axis)
			}
		}
		if err := ensureEntry(ctx, tx, itemID); err != nil {
			return err
		}
		result, err := tx.ExecContext(ctx, `INSERT INTO moment (item_id, at, note, visibility) VALUES (?, ?, ?, ?)`,
			itemID, m.At, m.Note, m.Visibility)
		if err != nil {
			return err
		}
		if momentID, err = result.LastInsertId(); err != nil {
			return err
		}
		for axis, value := range m.Scores {
			if _, err := tx.ExecContext(ctx, `INSERT INTO moment_score (moment_id, axis, value) VALUES (?, ?, ?)`, momentID, axis, value); err != nil {
				return err
			}
			if m.ScoresEffect == MomentScoresUpdateEntry {
				if err := setScore(ctx, tx, itemID, axis, &value); err != nil {
					return err
				}
			}
		}
		for _, ref := range m.Refs {
			if _, err := tx.ExecContext(ctx, `INSERT INTO moment_ref (moment_id, item_id) VALUES (?, ?) ON CONFLICT DO NOTHING`, momentID, ref); err != nil {
				return fmt.Errorf("%w: ref %d: %v", ErrInvalid, ref, err)
			}
		}
		return nil
	})
	return momentID, err
}

func validateMoment(m NewMoment) error {
	if !isoDate.MatchString(m.At) {
		return fmt.Errorf("%w: moment date must be YYYY-MM-DD", ErrInvalid)
	}
	if m.Visibility != InheritFromEntry && m.Visibility != MomentPrivate {
		return fmt.Errorf("%w: moment visibility must be inherit or private", ErrInvalid)
	}
	if m.ScoresEffect != MomentScoresUpdateEntry && m.ScoresEffect != MomentScoresStayOnMoment {
		return fmt.Errorf("%w: scoresEffect must be update-entry or stay-on-moment", ErrInvalid)
	}
	for axis, value := range m.Scores {
		if value < 0 || value > 100 {
			return fmt.Errorf("%w: score %q must be within 0–100", ErrInvalid, axis)
		}
	}
	return nil
}

func (s *Store) DeleteMoment(ctx context.Context, momentID int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM moment WHERE id = ?`, momentID)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func (s *Store) momentsOf(ctx context.Context, itemID int64) ([]Moment, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, item_id, at, note, visibility FROM moment WHERE item_id = ? ORDER BY at DESC, id DESC`, itemID)
	if err != nil {
		return nil, err
	}
	var moments []Moment
	for rows.Next() {
		m := Moment{Scores: map[string]int{}, Refs: []int64{}}
		if err := rows.Scan(&m.ID, &m.ItemID, &m.At, &m.Note, &m.Visibility); err != nil {
			rows.Close()
			return nil, err
		}
		moments = append(moments, m)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range moments {
		if err := s.attachMomentDetails(ctx, &moments[i]); err != nil {
			return nil, err
		}
	}
	return moments, nil
}

func (s *Store) attachMomentDetails(ctx context.Context, m *Moment) error {
	rows, err := s.db.QueryContext(ctx, `SELECT axis, value FROM moment_score WHERE moment_id = ?`, m.ID)
	if err != nil {
		return err
	}
	for rows.Next() {
		var axis string
		var value int
		if err := rows.Scan(&axis, &value); err != nil {
			rows.Close()
			return err
		}
		m.Scores[axis] = value
	}
	rows.Close()
	rows, err = s.db.QueryContext(ctx, `SELECT item_id FROM moment_ref WHERE moment_id = ? ORDER BY item_id`, m.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var ref int64
		if err := rows.Scan(&ref); err != nil {
			return err
		}
		m.Refs = append(m.Refs, ref)
	}
	return rows.Err()
}

func (s *Store) AddRelation(ctx context.Context, r Relation) error {
	if r.Type == "" || r.FromID == r.ToID {
		return fmt.Errorf("%w: relation needs a type and two different items", ErrInvalid)
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO relation (from_id, type, to_id) VALUES (?, ?, ?) ON CONFLICT DO NOTHING`, r.FromID, r.Type, r.ToID)
	return err
}

func (s *Store) RemoveRelation(ctx context.Context, r Relation) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM relation WHERE from_id = ? AND type = ? AND to_id = ?`, r.FromID, r.Type, r.ToID)
	if err != nil {
		return err
	}
	return requireAffected(result)
}
