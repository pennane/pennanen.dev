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

func TestLetterboxdTakeReadsTheWatchedShelf(t *testing.T) {
	values, err := Read(record.SourceLetterboxdTake, json.RawMessage(`{"onWatchedShelfSince":"2026-01-02"}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(values) != 1 || values["on_watched_shelf_since"] != "2026-01-02" {
		t.Errorf("values = %v, want only on_watched_shelf_since", values)
	}
	empty, err := Read(record.SourceLetterboxdTake, json.RawMessage(`{"rating":4}`))
	if err != nil {
		t.Fatal(err)
	}
	if _, present := empty["on_watched_shelf_since"]; present {
		t.Errorf("values = %v, want no watched shelf", empty)
	}
}
