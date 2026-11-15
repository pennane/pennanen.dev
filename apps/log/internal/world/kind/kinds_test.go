package kind

import (
	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
	"testing"
)

func TestEveryKindCanBeDismissed(t *testing.T) {
	for _, k := range All() {
		if !k.HasStatus(Dismissed) {
			t.Errorf("%s cannot be dismissed", k.Name)
		}
	}
}

func TestEveryPictureHasAShapeAndOnlyPicturesDo(t *testing.T) {
	for _, k := range All() {
		for _, f := range k.Fields {
			if (f.Type == Picture) != (f.Shape != NoShape) {
				t.Errorf("%s.%s is %s with shape %q", k.Name, f.Name, f.Type, f.Shape)
			}
		}
	}
}

func TestEveryTakeSourceCanBeRead(t *testing.T) {
	for _, k := range All() {
		for _, take := range k.Takes {
			if !reading.Reads(take.Source) {
				t.Errorf("%s takes from %s, which nothing reads", k.Name, take.Source)
			}
		}
	}
}
