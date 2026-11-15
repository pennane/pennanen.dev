package store

import (
	"errors"
	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
	"github.com/pennane/pennanen.dev/apps/log/internal/rule"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/verdict"
	"maps"
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
	Status          string            `json:"status"`
	StatusDecidedBy string            `json:"statusDecidedBy"`
	StatusChoice    *StatusChoice     `json:"statusChoice"`
	StatusProposals []StatusProposal  `json:"statusProposals"`
	Visibility      Visibility        `json:"visibility"`
	Blurb           string            `json:"blurb"`
	Prose           string            `json:"prose"`
	TimeSpent       TimeSpent         `json:"timeSpent,omitempty"`
	Scores          map[string]string `json:"scores"`
	ReadScores      map[string]string `json:"readScores"`
	Rank            int               `json:"rank,omitempty"`
	Tags            []string          `json:"tags"`
	TagsFromRules   map[string]string `json:"tagsFromRules"`
	LoggedBy        string            `json:"loggedBy"`
	UpdatedAt       string            `json:"updatedAt"`
}

type ChoiceStanding string

const (
	ChoiceDecides  ChoiceStanding = "decides"
	ChoiceSetAside ChoiceStanding = "set-aside"
	ChoiceOutvoted ChoiceStanding = "outvoted"
)

type StatusChoice struct {
	Status   string         `json:"status"`
	ChosenAt string         `json:"chosenAt"`
	Standing ChoiceStanding `json:"standing"`
}

type StatusProposal struct {
	Rule      string `json:"rule"`
	Value     string `json:"value"`
	ChangedAt string `json:"changedAt,omitempty"`
}

type Suggestion struct {
	Rule    string       `json:"rule"`
	Subject rule.Subject `json:"subject"`
	Value   string       `json:"value"`
}

type DismissalChange string

const (
	Dismissed DismissalChange = "dismissed"
	Kept      DismissalChange = "kept"
)

func (e Entry) EveryScore() map[string]string {
	every := maps.Clone(e.ReadScores)
	maps.Copy(every, e.Scores)
	return every
}

type Relation struct {
	FromID int64  `json:"fromId"`
	Type   string `json:"type"`
	ToID   int64  `json:"toId"`
}

type ItemView struct {
	Item
	Verdicts    verdict.Verdicts `json:"verdicts"`
	Entry       *Entry           `json:"entry"`
	DismissedAt string           `json:"dismissedAt,omitempty"`
	Suggestions []Suggestion     `json:"suggestions"`
	Takes       []TakeView       `json:"takes"`
	Claims      []reading.Claim  `json:"claims,omitempty"`
	Relations   []Relation       `json:"relations,omitempty"`
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
	Dismissal  *DismissalChange   `json:"dismissal"`
}

type ItemFilter struct {
	Kind       string
	Status     string
	Visibility Visibility
	Suggested  string
	Dismissal  DismissalChange
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
