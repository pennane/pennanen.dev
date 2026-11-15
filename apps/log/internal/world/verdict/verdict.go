package verdict

import (
	"encoding/json"
	"fmt"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

type Verdicts struct {
	Values   map[string]any    `json:"values"`
	SourceOf map[string]string `json:"sourceOf"`
}

func ReadAll(k kind.Kind, claims []reading.Claim) []reading.Reading {
	readings := make([]reading.Reading, 0, len(claims))
	for _, c := range claims {
		if r, readability := reading.ReadClaim(c); readability == reading.Readable {
			r.Values = normalizeValues(k, r.Source, r.Values)
			readings = append(readings, r)
		}
	}
	return readings
}

func Decode(k kind.Kind, source, externalID, fetchedAt string, raw json.RawMessage) (reading.Reading, error) {
	var values map[string]any
	if err := json.Unmarshal(raw, &values); err != nil {
		return reading.Reading{}, err
	}
	if values == nil {
		return reading.Reading{}, fmt.Errorf("reading of %s is null", source)
	}
	return reading.Reading{Source: source, ExternalID: externalID, FetchedAt: fetchedAt, Values: normalizeValues(k, source, values)}, nil
}

func DecodeVerdicts(k kind.Kind, raw json.RawMessage) (Verdicts, error) {
	decoded := Verdicts{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return Verdicts{}, err
	}
	if decoded.Values == nil {
		decoded.Values = map[string]any{}
	}
	if decoded.SourceOf == nil {
		decoded.SourceOf = map[string]string{}
	}
	for _, field := range k.Fields {
		if value, decided := decoded.Values[field.Name]; decided {
			decoded.Values[field.Name] = normalize(decoded.SourceOf[field.Name], field.Type, value)
		}
	}
	return decoded, nil
}

func Decide(k kind.Kind, readings []reading.Reading) Verdicts {
	valuesBySource := map[string]map[string]any{}
	for _, r := range readings {
		valuesBySource[r.Source] = r.Values
	}

	decided := Verdicts{Values: map[string]any{}, SourceOf: map[string]string{}}
	for _, field := range k.Fields {
		for _, source := range append([]string{record.SourceAP}, field.SourcesBestFirstAfterAP...) {
			if value, present := valuesBySource[source][field.Name]; present && value != nil {
				decided.Values[field.Name] = value
				decided.SourceOf[field.Name] = source
				break
			}
		}
	}
	return decided
}

func Inherit(k kind.Kind, own Verdicts, relatedThrough map[string]Verdicts) Verdicts {
	for _, field := range k.Fields {
		if field.InheritedThrough == "" {
			continue
		}
		if _, decided := own.Values[field.Name]; decided {
			continue
		}
		related, present := relatedThrough[field.InheritedThrough]
		if !present {
			continue
		}
		if value, given := related.Values[field.Name]; given {
			own.Values[field.Name] = value
			own.SourceOf[field.Name] = kind.InheritedSource(field.InheritedThrough)
		}
	}
	return own
}

type Take struct {
	Source    string         `json:"source"`
	Values    map[string]any `json:"values"`
	FetchedAt string         `json:"fetchedAt"`
}

func Takes(k kind.Kind, readings []reading.Reading) []Take {
	takes := []Take{}
	for _, r := range readings {
		if _, isTake := k.Take(r.Source); !isTake {
			continue
		}
		takes = append(takes, Take{Source: r.Source, Values: r.Values, FetchedAt: r.FetchedAt})
	}
	return takes
}

func Number(k kind.Kind, readings []reading.Reading, source, field string) (float64, bool) {
	for _, r := range readings {
		if r.Source != source || source == record.SourceAP {
			continue
		}
		return asNumber(r.Values[field])
	}
	return 0, false
}

func asNumber(value any) (float64, bool) {
	switch n := value.(type) {
	case float64:
		return n, true
	case int64:
		return float64(n), true
	case int:
		return float64(n), true
	}
	return 0, false
}

func fieldTypesOf(k kind.Kind, source string) map[string]kind.FieldType {
	fields := k.Fields
	if take, isTake := k.Take(source); isTake {
		fields = take.Fields
	}
	types := make(map[string]kind.FieldType, len(fields))
	for _, f := range fields {
		types[f.Name] = f.Type
	}
	return types
}

func normalizeValues(k kind.Kind, source string, values map[string]any) map[string]any {
	types := fieldTypesOf(k, source)
	for name, value := range values {
		if t, declared := types[name]; declared {
			values[name] = normalize(source, t, value)
		}
	}
	return values
}

func normalize(source string, t kind.FieldType, value any) any {
	switch t {
	case kind.Int, kind.ExternalID:
		if n, isNumber := value.(float64); isNumber {
			return int64(n)
		}
	case kind.StringList:
		if value == nil && source != record.SourceAP {
			return []string(nil)
		}
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
