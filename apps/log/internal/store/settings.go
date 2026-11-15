package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
	"maps"
	"slices"
	"strings"
)

type Division struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type Axis struct {
	Key       string     `json:"key"`
	Name      string     `json:"name"`
	Divisions []Division `json:"divisions"`
	Reading   *Reading   `json:"reading,omitempty"`
}

type AxisSort string

const (
	OwnAxis      AxisSort = "own"
	OutsideAxis  AxisSort = "outside"
	CrowdAxis    AxisSort = "crowd"
	ActivityAxis AxisSort = "activity"
)

var readFromBySort = map[AxisSort]kind.ReadFrom{
	OutsideAxis:  kind.FromTake,
	CrowdAxis:    kind.FromDescription,
	ActivityAxis: kind.FromActivity,
}

type Reading struct {
	Sort   AxisSort  `json:"sort"`
	Source string    `json:"source"`
	Field  string    `json:"field"`
	Floors []float64 `json:"floors"`
}

func (a Axis) Sort() AxisSort {
	if a.Reading == nil {
		return OwnAxis
	}
	return a.Reading.Sort
}

func (a Axis) DivisionReadAs(value float64) (string, bool) {
	if a.Reading == nil {
		return "", false
	}
	i := slices.IndexFunc(a.Reading.Floors, func(floor float64) bool { return value >= floor })
	if i < 0 || i >= len(a.Divisions) {
		return "", false
	}
	return a.Divisions[i].Key, true
}

type Axes []Axis

type KindSettings struct {
	Title              string                     `json:"title"`
	Description        string                     `json:"description"`
	StatusesShown      []string                   `json:"statusesShown"`
	Axes               []string                   `json:"axes"`
	DefaultView        View                       `json:"defaultView"`
	AxesWithoutTiers   []string                   `json:"axesWithoutTiers"`
	AxesWithoutColumns []string                   `json:"axesWithoutColumns"`
	FieldUses          map[string][]kind.FieldUse `json:"fieldUses"`
	FieldNames         map[string]string          `json:"fieldNames"`
	Overview           Overview                   `json:"overview"`
}

type View string

const (
	AutoView  View = "auto"
	TiersView View = "tiers"
	ListView  View = "list"
)

var Views = []View{AutoView, TiersView, ListView}

type Overview struct {
	Sort  string `json:"sort"`
	Count int    `json:"count"`
}

const (
	AutoSort        = ""
	RankSort        = "rank"
	TitleSort       = "title"
	ScoreSortPrefix = "score-"

	DefaultOverviewCount = 5
	MostInOverview       = 20
)

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
	kind.TagAnotherShot: "worth another shot",
	kind.TagIdled:       "idled",
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
		if axis.Reading != nil {
			if err := validReading(axis); err != nil {
				return nil, err
			}
		}
		out = append(out, axis)
	}
	prestige, found := out.Find(kind.Prestige)
	if !found {
		return nil, fmt.Errorf("%w: %s orders things and can't be removed", ErrInvalid, kind.Prestige)
	}
	if prestige.Sort() != OwnAxis {
		return nil, fmt.Errorf("%w: %s is ap's own and can't be read from a source", ErrInvalid, kind.Prestige)
	}
	return out, nil
}

func validReading(axis Axis) error {
	r := axis.Reading
	r.Source = strings.TrimSpace(r.Source)
	r.Field = strings.TrimSpace(r.Field)
	if _, readSort := readFromBySort[r.Sort]; !readSort {
		return fmt.Errorf("%w: %s is read from a source, so it is %s, %s or %s", ErrInvalid, axis.Key, OutsideAxis, CrowdAxis, ActivityAxis)
	}
	if r.Source == "" || r.Source == record.SourceAP || r.Field == "" {
		return fmt.Errorf("%w: %s needs a source other than ap and a field to read", ErrInvalid, axis.Key)
	}
	if !kind.IsReadable(r.Source, r.Field, readFromBySort[r.Sort]) {
		return fmt.Errorf("%w: %s reads %s %s, which no kind has as a number for a %s axis", ErrInvalid, axis.Key, r.Source, r.Field, r.Sort)
	}
	if len(r.Floors) != len(axis.Divisions) {
		return fmt.Errorf("%w: %s needs a floor for each division", ErrInvalid, axis.Key)
	}
	for i := 1; i < len(r.Floors); i++ {
		if r.Floors[i] >= r.Floors[i-1] {
			return fmt.Errorf("%w: %s floors must fall from the best division down", ErrInvalid, axis.Key)
		}
	}
	return nil
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
		found, kept := axes.Find(axis)
		if !kept {
			return fmt.Errorf("%w: %d entries are scored on %s; clear them before removing it", ErrConflict, entries, axis)
		}
		if found.Sort() != OwnAxis {
			return fmt.Errorf("%w: %d entries are scored on %s by hand; clear them before reading it from a source", ErrConflict, entries, axis)
		}
		if _, kept := axes.DivisionIndex(axis, division); !kept {
			return fmt.Errorf("%w: %d entries sit in %s %q; move them before removing it", ErrConflict, entries, axis, division)
		}
	}
	return rows.Err()
}

func DefaultKindSettings(k kind.Kind) KindSettings {
	uses := map[string][]kind.FieldUse{}
	for _, f := range k.PublicFields() {
		uses[f.Key] = f.DefaultUses
	}
	return KindSettings{
		Title:              k.DefaultTitle,
		StatusesShown:      slices.Clone(k.StatusesShown),
		Axes:               slices.Clone(k.DefaultAxes),
		DefaultView:        AutoView,
		AxesWithoutTiers:   []string{},
		AxesWithoutColumns: []string{},
		FieldUses:          uses,
		FieldNames:         map[string]string{},
		Overview:           Overview{Sort: AutoSort, Count: DefaultOverviewCount},
	}
}

func (ks KindSettings) filledFrom(defaults KindSettings) KindSettings {
	if ks.DefaultView == "" {
		ks.DefaultView = defaults.DefaultView
	}
	if ks.AxesWithoutTiers == nil {
		ks.AxesWithoutTiers = defaults.AxesWithoutTiers
	}
	if ks.AxesWithoutColumns == nil {
		ks.AxesWithoutColumns = defaults.AxesWithoutColumns
	}
	if ks.Overview.Count == 0 {
		ks.Overview.Count = defaults.Overview.Count
	}
	uses := defaults.FieldUses
	maps.Copy(uses, ks.FieldUses)
	ks.FieldUses = uses
	names := map[string]string{}
	for key, name := range ks.FieldNames {
		if name = strings.TrimSpace(name); name != "" {
			names[key] = name
		}
	}
	ks.FieldNames = names
	return ks
}

func (ks KindSettings) withoutStale(k kind.Kind, axes Axes) KindSettings {
	ks.Axes = slices.DeleteFunc(ks.Axes, func(axis string) bool { return !axes.Has(axis) })
	ks.AxesWithoutTiers = slices.DeleteFunc(ks.AxesWithoutTiers, func(axis string) bool { return !ks.Uses(axis) })
	ks.AxesWithoutColumns = slices.DeleteFunc(ks.AxesWithoutColumns, func(axis string) bool { return !ks.Uses(axis) })
	maps.DeleteFunc(ks.FieldUses, func(key string, _ []kind.FieldUse) bool { _, known := k.PublicField(key); return !known })
	maps.DeleteFunc(ks.FieldNames, func(key string, _ string) bool { _, known := k.PublicField(key); return !known })
	if validOverviewSort(k, ks, ks.Overview.Sort) != nil {
		ks.Overview.Sort = AutoSort
	}
	return ks
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
	return settings.withoutStale(k, axes), nil
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
	settings = settings.filledFrom(DefaultKindSettings(k))
	if err := validKindSettings(k, axes, settings); err != nil {
		return settings, err
	}
	if err := s.refuseDroppingKindAxes(ctx, k, settings); err != nil {
		return settings, err
	}
	return settings, s.writeSetting(ctx, kindSettingsKey(k), settings)
}

func validKindSettings(k kind.Kind, axes Axes, settings KindSettings) error {
	if strings.TrimSpace(settings.Title) == "" {
		return fmt.Errorf("%w: %s needs a title", ErrInvalid, k.Name)
	}
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
	return validPublicSettings(k, settings)
}

func validPublicSettings(k kind.Kind, settings KindSettings) error {
	if !slices.Contains(Views, settings.DefaultView) {
		return fmt.Errorf("%w: the default view must be one of %v", ErrInvalid, Views)
	}
	for _, axis := range settings.AxesWithoutTiers {
		if !settings.Uses(axis) {
			return fmt.Errorf("%w: %s has no axis %q to leave without tiers", ErrInvalid, k.Name, axis)
		}
	}
	for _, axis := range settings.AxesWithoutColumns {
		if !settings.Uses(axis) {
			return fmt.Errorf("%w: %s has no axis %q to leave out of the list", ErrInvalid, k.Name, axis)
		}
	}
	for key, uses := range settings.FieldUses {
		if _, known := k.PublicField(key); !known {
			return fmt.Errorf("%w: %s has no public field %q", ErrInvalid, k.Name, key)
		}
		for _, use := range uses {
			if !slices.Contains(kind.FieldUses, use) {
				return fmt.Errorf("%w: a field is used as one of %v, not %q", ErrInvalid, kind.FieldUses, use)
			}
		}
	}
	for key := range settings.FieldNames {
		if _, known := k.PublicField(key); !known {
			return fmt.Errorf("%w: %s has no public field %q to name", ErrInvalid, k.Name, key)
		}
	}
	if settings.Overview.Count < 1 || settings.Overview.Count > MostInOverview {
		return fmt.Errorf("%w: the overview shows 1 to %d entries", ErrInvalid, MostInOverview)
	}
	return validOverviewSort(k, settings, settings.Overview.Sort)
}

func validOverviewSort(k kind.Kind, settings KindSettings, sort string) error {
	switch sort {
	case AutoSort, RankSort, TitleSort:
		return nil
	}
	if axis, isScore := strings.CutPrefix(sort, ScoreSortPrefix); isScore && settings.Uses(axis) {
		return nil
	}
	if f, known := k.PublicField(sort); known && f.Type.Sortable() {
		return nil
	}
	return fmt.Errorf("%w: %s can't be sorted by %q", ErrInvalid, k.Name, sort)
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
