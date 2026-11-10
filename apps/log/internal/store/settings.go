package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
)

type Division struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type Axis struct {
	Key       string     `json:"key"`
	Name      string     `json:"name"`
	Divisions []Division `json:"divisions"`
}

type Axes []Axis

type KindSettings struct {
	StatusesShown []string `json:"statusesShown"`
	Axes          []string `json:"axes"`
}

type TagTitles map[string]string

var tiers = []Division{
	{Key: "s", Name: "S"},
	{Key: "a", Name: "A"},
	{Key: "b", Name: "B"},
	{Key: "c", Name: "C"},
	{Key: "d", Name: "D"},
	{Key: "f", Name: "F"},
}

func defaultAxes() Axes {
	return Axes{
		{Key: kind.Prestige, Name: kind.Prestige, Divisions: slices.Clone(tiers)},
		{Key: kind.Quality, Name: kind.Quality, Divisions: slices.Clone(tiers)},
	}
}

var defaultTagTitles = TagTitles{
	"another-shot": "worth another shot",
	"idled":        "idled",
}

const (
	tagTitlesKey = "tag-titles"
	axesKey      = "axes"
)

func kindSettingsKey(k kind.Kind) string { return "kind:" + k.Name }

func (a Axes) Find(key string) (Axis, bool) {
	i := slices.IndexFunc(a, func(axis Axis) bool { return axis.Key == key })
	if i < 0 {
		return Axis{}, false
	}
	return a[i], true
}

func (a Axes) Has(key string) bool {
	_, found := a.Find(key)
	return found
}

func (a Axes) DivisionIndex(axis, division string) (int, bool) {
	found, _ := a.Find(axis)
	i := slices.IndexFunc(found.Divisions, func(d Division) bool { return d.Key == division })
	return i, i >= 0
}

func (s *Store) Axes(ctx context.Context) (Axes, error) {
	axes := defaultAxes()
	_, err := s.readSetting(ctx, axesKey, &axes)
	return axes, err
}

func (s *Store) SetAxes(ctx context.Context, axes Axes) (Axes, error) {
	axes, err := validAxes(axes)
	if err != nil {
		return axes, err
	}
	if err := s.refuseDroppingScored(ctx, axes); err != nil {
		return axes, err
	}
	return axes, s.writeSetting(ctx, axesKey, axes)
}

func validAxes(axes Axes) (Axes, error) {
	out := make(Axes, 0, len(axes))
	for _, axis := range axes {
		axis.Name = strings.TrimSpace(axis.Name)
		axis.Key = keyFrom(axis.Key)
		if axis.Key == "" {
			axis.Key = keyFrom(axis.Name)
		}
		if axis.Name == "" || axis.Key == "" {
			return nil, fmt.Errorf("%w: each axis needs a name", ErrInvalid)
		}
		if out.Has(axis.Key) {
			return nil, fmt.Errorf("%w: two axes share the key %q", ErrInvalid, axis.Key)
		}
		divisions, err := validDivisions(axis.Key, axis.Divisions)
		if err != nil {
			return nil, err
		}
		axis.Divisions = divisions
		out = append(out, axis)
	}
	if !out.Has(kind.Prestige) {
		return nil, fmt.Errorf("%w: %s orders things and can't be removed", ErrInvalid, kind.Prestige)
	}
	return out, nil
}

func validDivisions(axis string, divisions []Division) ([]Division, error) {
	if len(divisions) == 0 {
		return nil, fmt.Errorf("%w: %s needs at least one division", ErrInvalid, axis)
	}
	out := make([]Division, 0, len(divisions))
	for _, d := range divisions {
		d.Name = strings.TrimSpace(d.Name)
		d.Key = keyFrom(d.Key)
		if d.Key == "" {
			d.Key = keyFrom(d.Name)
		}
		if d.Name == "" || d.Key == "" {
			return nil, fmt.Errorf("%w: each %s division needs a name", ErrInvalid, axis)
		}
		if slices.ContainsFunc(out, func(o Division) bool { return o.Key == d.Key }) {
			return nil, fmt.Errorf("%w: two %s divisions share the key %q", ErrInvalid, axis, d.Key)
		}
		out = append(out, d)
	}
	return out, nil
}

func keyFrom(text string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.ReplaceAll(text, "/", " "))), "-")
}

func (s *Store) refuseDroppingScored(ctx context.Context, axes Axes) error {
	rows, err := s.db.QueryContext(ctx, `SELECT axis, division, count(*) FROM score GROUP BY axis, division`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var axis, division string
		var entries int
		if err := rows.Scan(&axis, &division, &entries); err != nil {
			return err
		}
		if !axes.Has(axis) {
			return fmt.Errorf("%w: %d entries are scored on %s; clear them before removing it", ErrConflict, entries, axis)
		}
		if _, kept := axes.DivisionIndex(axis, division); !kept {
			return fmt.Errorf("%w: %d entries sit in %s %q; move them before removing it", ErrConflict, entries, axis, division)
		}
	}
	return rows.Err()
}

func DefaultKindSettings(k kind.Kind) KindSettings {
	return KindSettings{
		StatusesShown: slices.Clone(k.StatusesShown),
		Axes:          slices.Clone(k.DefaultAxes),
	}
}

func (ks KindSettings) Shows(status string) bool { return slices.Contains(ks.StatusesShown, status) }
func (ks KindSettings) Uses(axis string) bool    { return slices.Contains(ks.Axes, axis) }

func (s *Store) KindSettings(ctx context.Context, k kind.Kind) (KindSettings, error) {
	axes, err := s.Axes(ctx)
	if err != nil {
		return KindSettings{}, err
	}
	return s.kindSettings(ctx, k, axes)
}

func (s *Store) kindSettings(ctx context.Context, k kind.Kind, axes Axes) (KindSettings, error) {
	settings := DefaultKindSettings(k)
	if _, err := s.readSetting(ctx, kindSettingsKey(k), &settings); err != nil {
		return settings, err
	}
	settings.Axes = slices.DeleteFunc(settings.Axes, func(axis string) bool { return !axes.Has(axis) })
	return settings, nil
}

func (s *Store) AllKindSettings(ctx context.Context) (map[string]KindSettings, error) {
	axes, err := s.Axes(ctx)
	if err != nil {
		return nil, err
	}
	all := map[string]KindSettings{}
	for _, k := range kind.All() {
		settings, err := s.kindSettings(ctx, k, axes)
		if err != nil {
			return nil, err
		}
		all[k.Name] = settings
	}
	return all, nil
}

func (s *Store) SetKindSettings(ctx context.Context, k kind.Kind, settings KindSettings) (KindSettings, error) {
	axes, err := s.Axes(ctx)
	if err != nil {
		return settings, err
	}
	if err := validKindSettings(k, axes, settings); err != nil {
		return settings, err
	}
	if err := s.refuseDroppingKindAxes(ctx, k, settings); err != nil {
		return settings, err
	}
	return settings, s.writeSetting(ctx, kindSettingsKey(k), settings)
}

func validKindSettings(k kind.Kind, axes Axes, settings KindSettings) error {
	for _, status := range settings.StatusesShown {
		if !k.HasStatus(status) || status == kind.Untriaged || status == kind.Dismissed {
			return fmt.Errorf("%w: %q is not a status %s can show", ErrInvalid, status, k.Name)
		}
	}
	for i, axis := range settings.Axes {
		if !axes.Has(axis) {
			return fmt.Errorf("%w: there is no axis %q", ErrInvalid, axis)
		}
		if slices.Contains(settings.Axes[:i], axis) {
			return fmt.Errorf("%w: %s uses %q twice", ErrInvalid, k.Name, axis)
		}
	}
	if k.RankedAxis != "" && !settings.Uses(k.RankedAxis) {
		return fmt.Errorf("%w: %s ranks by %s, so it keeps that axis", ErrInvalid, k.Name, k.RankedAxis)
	}
	return nil
}

func (s *Store) refuseDroppingKindAxes(ctx context.Context, k kind.Kind, settings KindSettings) error {
	rows, err := s.db.QueryContext(ctx, `
		SELECT s.axis, count(*) FROM score s JOIN item i ON i.id = s.item_id
		WHERE i.kind = ? GROUP BY s.axis`, k.Name)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var axis string
		var entries int
		if err := rows.Scan(&axis, &entries); err != nil {
			return err
		}
		if !settings.Uses(axis) {
			return fmt.Errorf("%w: %d %s entries are scored on %s; clear them before removing it", ErrConflict, entries, k.Name, axis)
		}
	}
	return rows.Err()
}

func (s *Store) TagTitles(ctx context.Context) (TagTitles, error) {
	titles := TagTitles{}
	found, err := s.readSetting(ctx, tagTitlesKey, &titles)
	if !found {
		return maps.Clone(defaultTagTitles), err
	}
	return titles, err
}

func (s *Store) SetTagTitles(ctx context.Context, titles TagTitles) (TagTitles, error) {
	clean := TagTitles{}
	for tag, title := range titles {
		tag, title = NormalizeTag(tag), strings.TrimSpace(title)
		if tag == "" {
			return nil, fmt.Errorf("%w: a tag title needs a tag", ErrInvalid)
		}
		if title != "" {
			clean[tag] = title
		}
	}
	return clean, s.writeSetting(ctx, tagTitlesKey, clean)
}

func (s *Store) readSetting(ctx context.Context, key string, into any) (bool, error) {
	var value string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM setting WHERE key = ?`, key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal([]byte(value), into)
}

func (s *Store) writeSetting(ctx context.Context, key string, value any) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO setting (key, value) VALUES (?, ?) ON CONFLICT (key) DO UPDATE SET value = excluded.value`, key, string(encoded))
	return err
}
