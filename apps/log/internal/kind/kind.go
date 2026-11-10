package kind

import (
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"
)

type FieldType string

const (
	Text        FieldType = "text"
	Int         FieldType = "int"
	PartialDate FieldType = "partial-date"
	URL         FieldType = "url"
	StringList  FieldType = "string-list"
)

type Field struct {
	Name                    string    `json:"name"`
	Type                    FieldType `json:"type"`
	SourcesBestFirstAfterAP []string  `json:"sourcesBestFirstAfterAp"`
}

type UndeclaredFields string

const (
	RejectUndeclaredFields UndeclaredFields = "reject"
	AcceptUndeclaredFields UndeclaredFields = "accept"
)

type Kind struct {
	Name             string                             `json:"name"`
	Statuses         []string                           `json:"statuses"`
	StatusesToScore  []string                           `json:"statusesToScore"`
	StatusesShown    []string                           `json:"statusesShown"`
	Axes             []string                           `json:"axes"`
	RankedAxis       string                             `json:"rankedAxis"`
	Fields           []Field                            `json:"fields"`
	UndeclaredFields UndeclaredFields                   `json:"undeclaredFields"`
	IdentitySource   string                             `json:"identitySource,omitempty"`
	GuessStatus      func(verdicts map[string]any) Hint `json:"-"`
}

type Hint string

const NoHint Hint = ""

const (
	SourceAP   = "ap"
	Untriaged  = "untriaged"
	Dismissed  = "dismissed"
	Enjoyment  = "enjoyment"
	Quality    = "quality"
	FieldTitle = "title"
)

func (k Kind) Field(name string) (Field, bool) {
	i := slices.IndexFunc(k.Fields, func(f Field) bool { return f.Name == name })
	if i < 0 {
		return Field{}, false
	}
	return k.Fields[i], true
}

func (k Kind) HasStatus(status string) bool { return slices.Contains(k.Statuses, status) }
func (k Kind) HasAxis(axis string) bool     { return slices.Contains(k.Axes, axis) }
func (k Kind) Shows(status string) bool     { return slices.Contains(k.StatusesShown, status) }
func (k Kind) Scores(status string) bool    { return slices.Contains(k.StatusesToScore, status) }

func (k Kind) Hint(verdicts map[string]any) Hint {
	if k.GuessStatus == nil {
		return NoHint
	}
	return k.GuessStatus(verdicts)
}

var partialDate = regexp.MustCompile(`^\d{4}(-\d{2}(-\d{2})?)?$`)

func IsRemoval(value any) bool { return value == nil }

func (k Kind) ValidateOverride(name string, value any) error {
	if IsRemoval(value) {
		return nil
	}
	field, declared := k.Field(name)
	if !declared {
		if k.UndeclaredFields == AcceptUndeclaredFields {
			return nil
		}
		return fmt.Errorf("%s has no field %q", k.Name, name)
	}
	if !field.Type.accepts(value) {
		return fmt.Errorf("field %q must be %s", name, field.Type)
	}
	return nil
}

func (t FieldType) accepts(value any) bool {
	switch t {
	case Text:
		_, isString := value.(string)
		return isString
	case Int:
		n, isNumber := value.(float64)
		return isNumber && n == float64(int64(n))
	case PartialDate:
		s, isString := value.(string)
		return isString && partialDate.MatchString(s)
	case URL:
		s, isString := value.(string)
		return isString && isWebURL(s)
	case StringList:
		xs, isList := value.([]any)
		return isList && !slices.ContainsFunc(xs, func(x any) bool { _, isString := x.(string); return !isString })
	}
	return false
}

func isWebURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

var registry = map[string]Kind{}

func register(k Kind) {
	k.Statuses = append(k.Statuses, Dismissed)
	registry[k.Name] = k
}

func Get(name string) (Kind, bool) {
	k, ok := registry[name]
	return k, ok
}

func All() []Kind {
	kinds := make([]Kind, 0, len(registry))
	for _, k := range registry {
		kinds = append(kinds, k)
	}
	slices.SortFunc(kinds, func(a, b Kind) int { return strings.Compare(a.Name, b.Name) })
	return kinds
}
