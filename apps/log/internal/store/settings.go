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

type ScoreDisplay string

const (
	ShowNumber ScoreDisplay = "number"
	ShowWord   ScoreDisplay = "word"
	ShowBoth   ScoreDisplay = "both"
)

var ScoreDisplays = []ScoreDisplay{ShowNumber, ShowWord, ShowBoth}

type ScoreWord struct {
	From int    `json:"from"`
	Word string `json:"word"`
}

type KindSettings struct {
	StatusesShown []string          `json:"statusesShown"`
	AxisLabels    map[string]string `json:"axisLabels"`
	ScoreDisplay  ScoreDisplay      `json:"scoreDisplay"`
	ScoreWords    []ScoreWord       `json:"scoreWords"`
}

type TagTitles map[string]string

var defaultScoreWords = []ScoreWord{
	{From: 0, Word: "disliked"},
	{From: 50, Word: "meh"},
	{From: 70, Word: "liked"},
	{From: 88, Word: "loved"},
}

var defaultTagTitles = TagTitles{
	"another-shot": "worth another shot",
	"idled":        "idled",
}

const tagTitlesKey = "tag-titles"

func kindSettingsKey(k kind.Kind) string { return "kind:" + k.Name }

func DefaultKindSettings(k kind.Kind) KindSettings {
	labels := map[string]string{}
	for _, axis := range k.Axes {
		labels[axis] = axis
	}
	return KindSettings{
		StatusesShown: slices.Clone(k.StatusesShown),
		AxisLabels:    labels,
		ScoreDisplay:  ShowNumber,
		ScoreWords:    slices.Clone(defaultScoreWords),
	}
}

func (ks KindSettings) Shows(status string) bool { return slices.Contains(ks.StatusesShown, status) }

func (s *Store) KindSettings(ctx context.Context, k kind.Kind) (KindSettings, error) {
	settings := DefaultKindSettings(k)
	if _, err := s.readSetting(ctx, kindSettingsKey(k), &settings); err != nil {
		return settings, err
	}
	for _, axis := range k.Axes {
		if settings.AxisLabels[axis] == "" {
			settings.AxisLabels[axis] = axis
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
	if !slices.Contains(ScoreDisplays, settings.ScoreDisplay) {
		return settings, fmt.Errorf("%w: score display must be one of %v", ErrInvalid, ScoreDisplays)
	}
	words, err := validScoreWords(settings.ScoreWords)
	settings.ScoreWords = words
	return settings, err
}

func validScoreWords(words []ScoreWord) ([]ScoreWord, error) {
	out := make([]ScoreWord, 0, len(words))
	for _, w := range words {
		w.Word = strings.TrimSpace(w.Word)
		if w.Word == "" || w.From < 0 || w.From > 100 {
			return nil, fmt.Errorf("%w: each score word needs a word and a start within 0–100", ErrInvalid)
		}
		if slices.ContainsFunc(out, func(o ScoreWord) bool { return o.From == w.From }) {
			return nil, fmt.Errorf("%w: two score words start at %d", ErrInvalid, w.From)
		}
		out = append(out, w)
	}
	slices.SortFunc(out, func(a, b ScoreWord) int { return a.From - b.From })
	return out, nil
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
