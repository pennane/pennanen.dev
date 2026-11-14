package verdict

import (
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
)

func TestEveryTakeSourceCanBeRead(t *testing.T) {
	for _, k := range kind.All() {
		for _, take := range k.Takes {
			if _, readable := claimReaders[take.Source]; !readable {
				t.Errorf("%s takes from %s, which nothing reads", k.Name, take.Source)
			}
		}
	}
}
