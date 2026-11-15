package verdict

import (
	"encoding/json"
	"fmt"

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
	kind.SourceIGDB:              readIGDBGame,
	kind.SourceSteamTake:         readSteamTake,
	kind.SourceLetterboxdTake:    readLetterboxdTake,
	kind.SourceLetterboxd:        readLetterboxdFilm,
	kind.SourceTMDB:              readTMDBMovie,
	kind.SourceTMDBSeries:        readTMDBSeries,
	kind.SourceSpotifyArtist:     readSpotifyArtist,
	kind.SourceSpotifyAlbum:      readSpotifyAlbum,
	kind.SourceSpotifyListening:  readSpotifyListening,
}

func Read(source string, raw json.RawMessage) (map[string]any, error) {
	read, known := claimReaders[source]
	if !known {
		return nil, fmt.Errorf("no reader for %s", source)
	}
	return read(raw)
}

const ReadersVersion = 3

type Reading struct {
	Source     string
	ExternalID string
	FetchedAt  string
	Values     map[string]any
}

type Readability int

const (
	Unreadable Readability = iota
	Readable
)

func ReadClaim(c Claim) (Reading, Readability) {
	values, err := readValues(c)
	if err != nil || values == nil {
		return Reading{}, Unreadable
	}
	return Reading{Source: c.Source, ExternalID: c.ExternalID, FetchedAt: c.FetchedAt, Values: values}, Readable
}

func ReadAll(k kind.Kind, claims []Claim) []Reading {
	readings := make([]Reading, 0, len(claims))
	for _, c := range claims {
		if r, readability := ReadClaim(c); readability == Readable {
			r.Values = normalizeValues(k, r.Source, r.Values)
			readings = append(readings, r)
		}
	}
	return readings
}

func Encode(r Reading) (json.RawMessage, error) {
	return json.Marshal(r.Values)
}

func Decode(k kind.Kind, source, externalID, fetchedAt string, raw json.RawMessage) (Reading, error) {
	var values map[string]any
	if err := json.Unmarshal(raw, &values); err != nil {
		return Reading{}, err
	}
	if values == nil {
		return Reading{}, fmt.Errorf("reading of %s is null", source)
	}
	return Reading{Source: source, ExternalID: externalID, FetchedAt: fetchedAt, Values: normalizeValues(k, source, values)}, nil
}

func Decide(k kind.Kind, readings []Reading) Verdicts {
	valuesBySource := map[string]map[string]any{}
	for _, r := range readings {
		valuesBySource[r.Source] = r.Values
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

func Takes(k kind.Kind, readings []Reading) []Take {
	takes := []Take{}
	for _, r := range readings {
		if _, isTake := k.Take(r.Source); !isTake {
			continue
		}
		takes = append(takes, Take{Source: r.Source, Values: r.Values, FetchedAt: r.FetchedAt})
	}
	return takes
}

func Number(k kind.Kind, readings []Reading, source, field string) (float64, bool) {
	for _, r := range readings {
		if r.Source != source || source == kind.SourceAP {
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

func readValues(c Claim) (map[string]any, error) {
	if c.Source == kind.SourceAP {
		var override map[string]any
		err := json.Unmarshal(c.Data, &override)
		return override, err
	}
	read, known := claimReaders[c.Source]
	if !known {
		return nil, nil
	}
	return read(c.Data)
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
		if value == nil && source != kind.SourceAP {
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
