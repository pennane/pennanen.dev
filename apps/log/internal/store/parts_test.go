package store_test

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

const twoTrackAlbum = `{"id":"%s","name":"Beta","artists":[{"id":"Y","name":"Yara"}],
	"tracks":{"total":2,"items":[{"id":"t2","disc_number":1,"track_number":2},{"id":"t1","disc_number":1,"track_number":1}]}}`

func TestALinkTwoRecordsAssertIsOnePartAndOneRelation(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	yara := knowArtist(t, s, "Y", "Yara")
	album, err := s.CreateItemFromSource(ctx, kind.Album, "Beta", record.SourceSpotifyAlbum, "B", json.RawMessage(albumWithID("B")))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.PutClaimItemKeyed(ctx, album, record.SourceSpotifyAlbum, "B2", json.RawMessage(albumWithID("B2"))); err != nil {
		t.Fatal(err)
	}
	first, err := s.KnowItemFromSource(ctx, kind.Track, "First", record.SourceSpotifyTrack, "I1", json.RawMessage(`{"id":"t1","name":"First"}`))
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.KnowItemFromSource(ctx, kind.Track, "Second", record.SourceSpotifyTrack, "I2", json.RawMessage(`{"id":"t2","name":"Second"}`))
	if err != nil {
		t.Fatal(err)
	}

	if parts, err := s.PartsOf(ctx, album); err != nil || !slices.Equal(parts, []int64{first, second}) {
		t.Errorf("parts = %v, %v; want %v", parts, err, []int64{first, second})
	}
	if got, want := relationsOf(t, s, album), map[store.Relation]struct{}{{FromID: album, Type: kind.LinkBy, ToID: yara}: {}}; !maps.Equal(got, want) {
		t.Errorf("relations = %v, want %v", got, want)
	}
}

func albumWithID(id string) string {
	return fmt.Sprintf(twoTrackAlbum, id)
}
