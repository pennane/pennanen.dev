package store

import (
	"context"
	"fmt"
	"slices"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
)

type TakeView struct {
	verdict.Take
	Hidden []string `json:"hidden"`
}

func (t TakeView) Public() verdict.Take {
	public := t.Take
	public.Values = map[string]any{}
	for field, value := range t.Values {
		if !slices.Contains(t.Hidden, field) {
			public.Values[field] = value
		}
	}
	return public
}

type TakeFieldVisibility string

const (
	TakeFieldShown  TakeFieldVisibility = "shown"
	TakeFieldHidden TakeFieldVisibility = "hidden"
)

func (s *Store) SetTakeFieldVisibility(ctx context.Context, itemID int64, source, field string, visibility TakeFieldVisibility) error {
	k, err := lookupKind(ctx, s.db, itemID)
	if err != nil {
		return err
	}
	take, isTake := k.Take(source)
	if !isTake || !take.IsPrivate(field) {
		return fmt.Errorf("%w: %s %s is not a private field of %s", ErrInvalid, source, field, k.Name)
	}
	switch visibility {
	case TakeFieldShown:
		_, err = s.db.ExecContext(ctx, `INSERT INTO take_field_shown (item_id, source, field) VALUES (?, ?, ?) ON CONFLICT DO NOTHING`, itemID, source, field)
	case TakeFieldHidden:
		_, err = s.db.ExecContext(ctx, `DELETE FROM take_field_shown WHERE item_id = ? AND source = ? AND field = ?`, itemID, source, field)
	default:
		err = fmt.Errorf("%w: a take field is %s or %s", ErrInvalid, TakeFieldShown, TakeFieldHidden)
	}
	return err
}

type shownField struct{ source, field string }

func (s *Store) attachTakes(ctx context.Context, claimsByItem map[int64][]verdict.Claim, byID map[int64]*ItemView) error {
	shown, err := s.shownTakeFields(ctx, byID)
	if err != nil {
		return err
	}
	for id, v := range byID {
		k, known := kind.Get(v.Kind)
		if !known {
			continue
		}
		v.Takes = []TakeView{}
		for _, take := range verdict.Takes(k, verdict.ReadAll(k, claimsByItem[id])) {
			declared, _ := k.Take(take.Source)
			hidden := []string{}
			for field := range take.Values {
				if declared.IsPrivate(field) && !shown[id][shownField{take.Source, field}] {
					hidden = append(hidden, field)
				}
			}
			slices.Sort(hidden)
			v.Takes = append(v.Takes, TakeView{Take: take, Hidden: hidden})
		}
	}
	return nil
}

func (s *Store) shownTakeFields(ctx context.Context, byID map[int64]*ItemView) (map[int64]map[shownField]bool, error) {
	ids := make([]int64, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	shown := map[int64]map[shownField]bool{}
	err := forChunks(ids, func(chunk []int64) error {
		rows, err := s.db.QueryContext(ctx, `SELECT item_id, source, field FROM take_field_shown WHERE item_id IN (`+placeholders(len(chunk))+`)`, int64sAsArgs(chunk)...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id int64
			var f shownField
			if err := rows.Scan(&id, &f.source, &f.field); err != nil {
				return err
			}
			if shown[id] == nil {
				shown[id] = map[shownField]bool{}
			}
			shown[id][f] = true
		}
		return rows.Err()
	})
	return shown, err
}
