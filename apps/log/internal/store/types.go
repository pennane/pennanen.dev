package store

import (
	"errors"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
)

var (
	ErrNotFound = errors.New("not found")
	ErrInvalid  = errors.New("invalid")
)

type Visibility string

const (
	Private Visibility = "private"
	Public  Visibility = "public"
)

type MomentVisibility string

const (
	InheritFromEntry MomentVisibility = "inherit"
	MomentPrivate    MomentVisibility = "private"
)

type MomentScoresEffect string

const (
	MomentScoresUpdateEntry  MomentScoresEffect = "update-entry"
	MomentScoresStayOnMoment MomentScoresEffect = "stay-on-moment"
)

type Item struct {
	ID        int64  `json:"id"`
	Kind      string `json:"kind"`
	Slug      string `json:"slug"`
	CreatedAt string `json:"createdAt"`
}

type Entry struct {
	Status     string         `json:"status"`
	Visibility Visibility     `json:"visibility"`
	Notes      string         `json:"notes"`
	Blurb      string         `json:"blurb"`
	Scores     map[string]int `json:"scores"`
	Tags       []string       `json:"tags"`
	UpdatedAt  string         `json:"updatedAt"`
}

type Moment struct {
	ID         int64            `json:"id"`
	ItemID     int64            `json:"itemId"`
	At         string           `json:"at"`
	Note       string           `json:"note"`
	Visibility MomentVisibility `json:"visibility"`
	Scores     map[string]int   `json:"scores"`
	Refs       []int64          `json:"refs"`
}

type Relation struct {
	FromID int64  `json:"fromId"`
	Type   string `json:"type"`
	ToID   int64  `json:"toId"`
}

type ItemView struct {
	Item
	Verdicts  verdict.Verdicts `json:"verdicts"`
	Entry     *Entry           `json:"entry"`
	Hint      kind.Hint        `json:"hint,omitempty"`
	Claims    []verdict.Claim  `json:"claims,omitempty"`
	Moments   []Moment         `json:"moments,omitempty"`
	Relations []Relation       `json:"relations,omitempty"`
}

type EntryChange struct {
	Status     *string         `json:"status"`
	Visibility *Visibility     `json:"visibility"`
	Notes      *string         `json:"notes"`
	Blurb      *string         `json:"blurb"`
	Scores     map[string]*int `json:"scores"`
	AddTags    []string        `json:"addTags"`
	RemoveTags []string        `json:"removeTags"`
}

type NewMoment struct {
	At           string             `json:"at"`
	Note         string             `json:"note"`
	Visibility   MomentVisibility   `json:"visibility"`
	Scores       map[string]int     `json:"scores"`
	Refs         []int64            `json:"refs"`
	ScoresEffect MomentScoresEffect `json:"scoresEffect"`
}

type ItemFilter struct {
	Kind       string
	Status     string
	Visibility Visibility
	Hint       kind.Hint
	Tag        string
	Query      string
}
