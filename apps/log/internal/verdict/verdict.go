package verdict

import (
	"encoding/json"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
)

type Claim struct {
	Source     string          `json:"source"`
	ExternalID string          `json:"externalId,omitempty"`
	Data       json.RawMessage `json:"data"`
	FetchedAt  string          `json:"fetchedAt"`
}

type Verdicts struct {
	Values   map[string]any    `json:"values"`
	SourceOf map[string]string `json:"sourceOf"`
}

type readClaim func(raw json.RawMessage) (map[string]any, error)

var claimReaders = map[string]readClaim{
	kind.SourceSteam:             readSteamOwnedGame,
	kind.SourceSteamStore:        readSteamStoreDetails,
	kind.SourceSteamAchievements: readSteamAchievements,
}

func Decide(k kind.Kind, claims []Claim) Verdicts {
	valuesBySource := map[string]map[string]any{}
	for _, c := range claims {
		values, err := readValues(k, c)
		if err == nil {
			valuesBySource[c.Source] = values
		}
	}

	decided := Verdicts{Values: map[string]any{}, SourceOf: map[string]string{}}
	for _, field := range k.Fields {
		for _, source := range append([]string{kind.SourceAP}, field.SourcesBestFirstAfterAP...) {
			if value, present := valuesBySource[source][field.Name]; present && value != nil {
				decided.Values[field.Name] = value
				decided.SourceOf[field.Name] = source
				break
			}
		}
	}
	if k.UndeclaredFields == kind.AcceptUndeclaredFields {
		for name, value := range valuesBySource[kind.SourceAP] {
			if _, declared := k.Field(name); !declared && value != nil {
				decided.Values[name] = value
				decided.SourceOf[name] = kind.SourceAP
			}
		}
	}
	return decided
}

func readValues(k kind.Kind, c Claim) (map[string]any, error) {
	if c.Source == kind.SourceAP {
		return readOverride(k, c.Data)
	}
	read, known := claimReaders[c.Source]
	if !known {
		return nil, nil
	}
	return read(c.Data)
}

func readOverride(k kind.Kind, raw json.RawMessage) (map[string]any, error) {
	var override map[string]any
	if err := json.Unmarshal(raw, &override); err != nil {
		return nil, err
	}
	for name, value := range override {
		if field, declared := k.Field(name); declared {
			override[name] = normalize(field.Type, value)
		}
	}
	return override, nil
}

func normalize(t kind.FieldType, value any) any {
	switch t {
	case kind.Int:
		if n, isNumber := value.(float64); isNumber {
			return int64(n)
		}
	case kind.StringList:
		if xs, isList := value.([]any); isList {
			strings := make([]string, 0, len(xs))
			for _, x := range xs {
				if s, isString := x.(string); isString {
					strings = append(strings, s)
				}
			}
			return strings
		}
	}
	return value
}

func Title(v Verdicts) string {
	title, _ := v.Values[kind.FieldTitle].(string)
	return title
}
