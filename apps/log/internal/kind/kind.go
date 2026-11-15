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
	Number      FieldType = "number"
	ExternalID  FieldType = "external-id"
	PartialDate FieldType = "partial-date"
	URL         FieldType = "url"
	StringList  FieldType = "string-list"
)

type Field struct {
	Name                    string    `json:"name"`
	Type                    FieldType `json:"type"`
	SourcesBestFirstAfterAP []string  `json:"sourcesBestFirstAfterAp"`
	Reach                   Reach     `json:"reach"`
	InheritedThrough        string    `json:"inheritedThrough,omitempty"`
}

const inheritedSourcePrefix = "inherited:"

func InheritedSource(relation string) string { return inheritedSourcePrefix + relation }

type Reach string

const (
	Everyone Reach = ""
	OnlyAP   Reach = "only-ap"
)

type Kind struct {
	Name             string                             `json:"name"`
	DefaultTitle     string                             `json:"-"`
	Statuses         []string                           `json:"statuses"`
	StatusesToScore  []string                           `json:"statusesToScore"`
	StatusesShown    []string                           `json:"statusesShown"`
	DefaultAxes      []string                           `json:"-"`
	RankedAxis       string                             `json:"rankedAxis"`
	Fields           []Field                            `json:"fields"`
	IdentitySource   string                             `json:"identitySource,omitempty"`
	StatusComputedBy string                             `json:"statusComputedBy,omitempty"`
	Takes            []Take                             `json:"takes"`
	GuessStatus      func(verdicts map[string]any) Hint `json:"-"`
}

type Take struct {
	Source        string   `json:"source"`
	PrivateFields []string `json:"privateFields"`
	Fields        []Field  `json:"fields"`
}

func (k Kind) Take(source string) (Take, bool) {
	i := slices.IndexFunc(k.Takes, func(t Take) bool { return t.Source == source })
	if i < 0 {
		return Take{}, false
	}
	return k.Takes[i], true
}

func (t Take) IsPrivate(field string) bool { return slices.Contains(t.PrivateFields, field) }

type Hint string

const NoHint Hint = ""

const (
	SourceAP   = "ap"
	Untriaged  = "untriaged"
	Dismissed  = "dismissed"
	Prestige   = "prestige"
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

func (k Kind) InheritingRelations() []string {
	relations := []string{}
	for _, f := range k.Fields {
		if f.InheritedThrough != "" && !slices.Contains(relations, f.InheritedThrough) {
			relations = append(relations, f.InheritedThrough)
		}
	}
	return relations
}

func (k Kind) HasStatus(status string) bool { return slices.Contains(k.Statuses, status) }
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
	case Int, ExternalID:
		n, isNumber := value.(float64)
		return isNumber && n == float64(int64(n))
	case Number:
		_, isNumber := value.(float64)
		return isNumber
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
	if k.Takes == nil {
		k.Takes = []Take{}
	}
	for i := range k.Takes {
		if k.Takes[i].Fields == nil {
			k.Takes[i].Fields = []Field{}
		}
	}
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
