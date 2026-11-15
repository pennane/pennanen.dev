package reading

import (
	"encoding/json"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
)

func TestAnUnreadableClaimIsNotRead(t *testing.T) {
	_, broken := ReadClaim(Claim{Source: record.SourceTMDB, Data: json.RawMessage(`[1]`)})
	_, unknown := ReadClaim(Claim{Source: "nobody", Data: json.RawMessage(`{}`)})
	if broken != Unreadable || unknown != Unreadable {
		t.Errorf("broken = %v, unknown = %v, want both unreadable", broken, unknown)
	}
}
