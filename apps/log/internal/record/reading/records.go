package reading

import (
	"encoding/json"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
)

type apHandMade struct {
	Kind   string         `json:"kind"`
	Fields map[string]any `json:"fields"`
}

func ReadRecord(source string, raw json.RawMessage) (json.RawMessage, record.Readability) {
	values, readability := readRecordValues(source, raw)
	if readability == record.Unreadable {
		return nil, record.Unreadable
	}
	reading, err := json.Marshal(values)
	if err != nil {
		return nil, record.Unreadable
	}
	return reading, record.Readable
}

func readRecordValues(source string, raw json.RawMessage) (map[string]any, record.Readability) {
	switch source {
	case record.SourceAP:
		return readApRecord(raw)
	case record.SourceMatched, record.SourceLegacy:
		return readAsGiven(raw)
	}
	values, err := Read(source, raw)
	if err != nil || values == nil {
		return nil, record.Unreadable
	}
	return values, record.Readable
}

func readAsGiven(raw json.RawMessage) (map[string]any, record.Readability) {
	var values map[string]any
	if err := json.Unmarshal(raw, &values); err != nil || values == nil {
		return nil, record.Unreadable
	}
	return values, record.Readable
}

func readApRecord(raw json.RawMessage) (map[string]any, record.Readability) {
	given, readability := readAsGiven(raw)
	if readability == record.Unreadable {
		return nil, record.Unreadable
	}
	_, hasFields := given["fields"]
	_, hasAnswer := given["to"]
	_, hasEdit := given["edit"]
	switch {
	case hasFields:
		var item apHandMade
		if err := json.Unmarshal(raw, &item); err != nil || item.Fields == nil {
			return nil, record.Unreadable
		}
		if item.Kind != "" {
			item.Fields["kind"] = item.Kind
		}
		return item.Fields, record.Readable
	case hasAnswer, hasEdit:
		return given, record.Readable
	}
	return nil, record.Unreadable
}
