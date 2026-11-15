package kind

import (
	"cmp"
	"slices"
)

type ReadFrom string

const (
	FromTake        ReadFrom = "take"
	FromDescription ReadFrom = "description"
	FromActivity    ReadFrom = "activity"
)

type Readable struct {
	Source string   `json:"source"`
	Field  string   `json:"field"`
	From   ReadFrom `json:"from"`
}

func (t FieldType) readable() bool {
	return t == Number || t == Int
}

func Readables() []Readable {
	found := []Readable{}
	add := func(r Readable) {
		if !slices.Contains(found, r) {
			found = append(found, r)
		}
	}
	for _, k := range All() {
		for _, f := range k.Fields {
			if !f.Type.readable() {
				continue
			}
			for _, source := range f.SourcesBestFirstAfterAP {
				add(Readable{Source: source, Field: f.Name, From: f.About.readFrom()})
			}
		}
		for _, t := range k.Takes {
			for _, f := range t.Fields {
				if f.Type.readable() {
					add(Readable{Source: t.Source, Field: f.Name, From: FromTake})
				}
			}
		}
	}
	slices.SortFunc(found, func(a, b Readable) int {
		return cmp.Or(cmp.Compare(a.Source, b.Source), cmp.Compare(a.Field, b.Field))
	})
	return found
}

func IsReadable(source, field string, from ReadFrom) bool {
	return slices.Contains(Readables(), Readable{Source: source, Field: field, From: from})
}

func (a About) readFrom() ReadFrom {
	if a == AboutAP {
		return FromActivity
	}
	return FromDescription
}
