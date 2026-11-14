package kind

type FieldUse string

const (
	FilterUse FieldUse = "filter"
	ColumnUse FieldUse = "column"
	PageUse   FieldUse = "page"
)

var FieldUses = []FieldUse{FilterUse, ColumnUse, PageUse}

type PublicField struct {
	Key         string     `json:"key"`
	Name        string     `json:"name"`
	Type        FieldType  `json:"type"`
	Take        string     `json:"take,omitempty"`
	DefaultUses []FieldUse `json:"defaultUses"`
}

const takeFieldSeparator = "."

var placedFields = map[string]struct{}{
	FieldTitle: {},
	"link":     {},
	"cover":    {},
	"poster":   {},
	"image":    {},
}

func (k Kind) PublicFields() []PublicField {
	fields := []PublicField{}
	for _, f := range k.Fields {
		if _, placed := placedFields[f.Name]; placed || f.Reach == OnlyAP {
			continue
		}
		fields = append(fields, publicField(f.Name, f))
	}
	for _, t := range k.Takes {
		for _, f := range t.Fields {
			if f.Reach == OnlyAP || t.IsPrivate(f.Name) {
				continue
			}
			field := publicField(t.Source+takeFieldSeparator+f.Name, f)
			field.Take = t.Source
			fields = append(fields, field)
		}
	}
	return fields
}

func (k Kind) PublicField(key string) (PublicField, bool) {
	for _, f := range k.PublicFields() {
		if f.Key == key {
			return f, true
		}
	}
	return PublicField{}, false
}

func publicField(key string, f Field) PublicField {
	return PublicField{Key: key, Name: f.Name, Type: f.Type, DefaultUses: f.Type.defaultUses()}
}

func (t FieldType) defaultUses() []FieldUse {
	switch t {
	case PartialDate, Number:
		return []FieldUse{FilterUse, ColumnUse, PageUse}
	case Int, StringList:
		return []FieldUse{FilterUse, PageUse}
	case Text:
		return []FieldUse{PageUse}
	}
	return []FieldUse{}
}

func (t FieldType) Sortable() bool {
	switch t {
	case PartialDate, Number, Int, Text:
		return true
	}
	return false
}

func (k Kind) PrivateFieldNames() []string {
	names := []string{}
	for _, f := range k.Fields {
		if f.Reach == OnlyAP {
			names = append(names, f.Name)
		}
	}
	return names
}
