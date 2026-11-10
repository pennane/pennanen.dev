package store

import (
	"errors"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
)

var (
	ErrNotFound = errors.New("not found")
	ErrInvalid  = errors.New("invalid")
	ErrConflict = errors.New("conflict")
)

type Visibility string

const (
	Auto    Visibility = "auto"
	Private Visibility = "private"
	Public  Visibility = "public"
)

var Visibilities = []Visibility{Auto, Public, Private}

const NewEntryVisibility = Auto

type TimeSpent string

const (
	TimeSpentUnsaid TimeSpent = ""
	NotEnough       TimeSpent = "not-enough"
	ABit            TimeSpent = "a-bit"
	JustRight       TimeSpent = "just-right"
	ALot            TimeSpent = "a-lot"
	TooMuch         TimeSpent = "too-much"
)

var TimeSpentChoices = []TimeSpent{NotEnough, ABit, JustRight, ALot, TooMuch}

type Item struct {
	ID        int64  `json:"id"`
	Kind      string `json:"kind"`
	Slug      string `json:"slug"`
	CreatedAt string `json:"createdAt"`
}

type Entry struct {
	Status     string            `json:"status"`
	Visibility Visibility        `json:"visibility"`
	Blurb      string            `json:"blurb"`
	Prose      string            `json:"prose"`
	TimeSpent  TimeSpent         `json:"timeSpent,omitempty"`
	Scores     map[string]string `json:"scores"`
	Rank       int               `json:"rank,omitempty"`
	Tags       []string          `json:"tags"`
	UpdatedAt  string            `json:"updatedAt"`
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
	Relations []Relation       `json:"relations,omitempty"`
}

type EntryChange struct {
	Status     *string            `json:"status"`
	Visibility *Visibility        `json:"visibility"`
	Blurb      *string            `json:"blurb"`
	Prose      *string            `json:"prose"`
	TimeSpent  *TimeSpent         `json:"timeSpent"`
	Scores     map[string]*string `json:"scores"`
	AddTags    []string           `json:"addTags"`
	RemoveTags []string           `json:"removeTags"`
}

type ItemFilter struct {
	Kind       string
	Status     string
	Visibility Visibility
	Hint       kind.Hint
	Tag        string
	Query      string
	Needs      Need
}

type Need string

const (
	NeedsNothing Need = ""
	NeedsScore   Need = "score"
	NeedsMatch   Need = "match"
)
