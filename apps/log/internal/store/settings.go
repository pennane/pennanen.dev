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

type KindSettings struct {
	StatusesShown []string              `json:"statusesShown"`
	AxisLabels    map[string]string     `json:"axisLabels"`
	Divisions     map[string][]Division `json:"divisions"`
}

type TagTitles map[string]string

func divisionsNamed(names ...string) []Division {
	divisions := make([]Division, len(names))
	for i, name := range names {
		divisions[i] = Division{Key: name, Name: name}
	}
	return divisions
}

var defaultDivisions = map[string][]Division{
	kind.Enjoyment: divisionsNamed("loved", "liked", "fine", "meh", "disliked"),
	kind.Quality:   divisionsNamed("masterpiece", "great", "solid", "rough", "broken"),
}

var fallbackDivisions = divisionsNamed("good", "fine", "bad")

func defaultDivisionsOf(axis string) []Division {
	if divisions, known := defaultDivisions[axis]; known {
		return slices.Clone(divisions)
	}
	return slices.Clone(fallbackDivisions)
}

var defaultTagTitles = TagTitles{
	"another-shot": "worth another shot",
	"idled":        "idled",
}

const tagTitlesKey = "tag-titles"

func kindSettingsKey(k kind.Kind) string { return "kind:" + k.Name }

func DefaultKindSettings(k kind.Kind) KindSettings {
	labels := map[string]string{}
	divisions := map[string][]Division{}
	for _, axis := range k.Axes {
		labels[axis] = axis
		divisions[axis] = defaultDivisionsOf(axis)
	}
	return KindSettings{
		StatusesShown: slices.Clone(k.StatusesShown),
		AxisLabels:    labels,
		Divisions:     divisions,
	}
}

func (ks KindSettings) Shows(status string) bool { return slices.Contains(ks.StatusesShown, status) }

func (ks KindSettings) DivisionIndex(axis, key string) (int, bool) {
	i := slices.IndexFunc(ks.Divisions[axis], func(d Division) bool { return d.Key == key })
	return i, i >= 0
}

func (s *Store) KindSettings(ctx context.Context, k kind.Kind) (KindSettings, error) {
	settings := DefaultKindSettings(k)
	if _, err := s.readSetting(ctx, kindSettingsKey(k), &settings); err != nil {
		return settings, err
	}
	for _, axis := range k.Axes {
		if settings.AxisLabels[axis] == "" {
			settings.AxisLabels[axis] = axis
		}
		if len(settings.Divisions[axis]) == 0 {
			settings.Divisions[axis] = defaultDivisionsOf(axis)
		}
	}
	return settings, nil
}

func (s *Store) AllKindSettings(ctx context.Context) (map[string]KindSettings, error) {
	all := map[string]KindSettings{}
	for _, k := range kind.All() {
		settings, err := s.KindSettings(ctx, k)
		if err != nil {
			return nil, err
		}
		all[k.Name] = settings
	}
	return all, nil
}

func (s *Store) SetKindSettings(ctx context.Context, k kind.Kind, settings KindSettings) (KindSettings, error) {
	settings, err := validKindSettings(k, settings)
	if err != nil {
		return settings, err
	}
	if err := s.refuseDroppingUsedDivisions(ctx, k, settings); err != nil {
		return settings, err
	}
	return settings, s.writeSetting(ctx, kindSettingsKey(k), settings)
}

func validKindSettings(k kind.Kind, settings KindSettings) (KindSettings, error) {
	for _, status := range settings.StatusesShown {
		if !k.HasStatus(status) || status == kind.Untriaged || status == kind.Dismissed {
			return settings, fmt.Errorf("%w: %q is not a status %s can show", ErrInvalid, status, k.Name)
		}
	}
	labels := map[string]string{}
	for axis, label := range settings.AxisLabels {
		if !k.HasAxis(axis) {
			return settings, fmt.Errorf("%w: %s has no axis %q", ErrInvalid, k.Name, axis)
		}
		if label = strings.TrimSpace(label); label != "" {
			labels[axis] = label
		}
	}
	settings.AxisLabels = labels
	divisions := map[string][]Division{}
	for _, axis := range k.Axes {
		valid, err := validDivisions(axis, settings.Divisions[axis])
		if err != nil {
			return settings, err
		}
		divisions[axis] = valid
	}
	for axis := range settings.Divisions {
		if !k.HasAxis(axis) {
			return settings, fmt.Errorf("%w: %s has no axis %q", ErrInvalid, k.Name, axis)
		}
	}
	settings.Divisions = divisions
	return settings, nil
}

func validDivisions(axis string, divisions []Division) ([]Division, error) {
	if len(divisions) == 0 {
		return nil, fmt.Errorf("%w: %s needs at least one division", ErrInvalid, axis)
	}
	out := make([]Division, 0, len(divisions))
	for _, d := range divisions {
		d.Name = strings.TrimSpace(d.Name)
		d.Key = DivisionKey(d.Key)
		if d.Key == "" {
			d.Key = DivisionKey(d.Name)
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

func DivisionKey(text string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.ReplaceAll(text, "/", " "))), "-")
}

func (s *Store) refuseDroppingUsedDivisions(ctx context.Context, k kind.Kind, settings KindSettings) error {
	rows, err := s.db.QueryContext(ctx, `
		SELECT s.axis, s.division, count(*) FROM score s JOIN item i ON i.id = s.item_id
		WHERE i.kind = ? GROUP BY s.axis, s.division`, k.Name)
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
		if _, kept := settings.DivisionIndex(axis, division); !kept {
			return fmt.Errorf("%w: %d %s entries sit in %s %q; move them before removing it", ErrConflict, entries, k.Name, axis, division)
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
