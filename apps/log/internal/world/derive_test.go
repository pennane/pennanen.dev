package world_test

import (
	"encoding/json"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
	"github.com/pennane/pennanen.dev/apps/log/internal/world"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
	"slices"
	"testing"
)

func identities(t *testing.T, ss ...string) []record.Identity {
	t.Helper()
	ids := []record.Identity{}
	for _, s := range ss {
		ids = append(ids, identity(t, s))
	}
	return ids
}

func link(t *testing.T, from, linkType, to string, position int) world.AssertedLink {
	t.Helper()
	return world.AssertedLink{From: identity(t, from), Type: linkType, To: identity(t, to), Position: position}
}

func TestEverySourceDerivesWhatItIsAbout(t *testing.T) {
	cases := []struct {
		name    string
		id      string
		reading map[string]any
		about   []string
		kind    string
	}{
		{"steam", "steam:220", nil, []string{"steam:220"}, kind.Game},
		{"steam-store", "steam-store:220", nil, []string{"steam:220"}, kind.Game},
		{"steam-achievements", "steam-achievements:220", nil, []string{"steam:220"}, kind.Game},
		{"steam-take", "steam-take:220", nil, []string{"steam:220"}, kind.Game},
		{"igdb", "igdb:1942", nil, nil, kind.Game},
		{"tmdb", "tmdb:603", nil, nil, kind.Movie},
		{"tmdb-tv", "tmdb-tv:1399", nil, nil, kind.Series},
		{"letterboxd movie", "letterboxd:abc", map[string]any{"tmdb_link": "tmdb:603"}, []string{"tmdb:603"}, kind.Movie},
		{"letterboxd unlinked", "letterboxd:abc", map[string]any{}, nil, kind.Movie},
		{"spotify-artist", "spotify-artist:a", nil, nil, kind.Artist},
		{"spotify-album", "spotify-album:b", map[string]any{}, nil, kind.Album},
		{"spotify-track", "spotify-track:USX", map[string]any{"spotify_id": "t1"}, []string{"spotify-track-id:t1"}, kind.Track},
		{"letterboxd-take", "letterboxd-take:letterboxd:abc", nil, []string{"letterboxd:abc"}, ""},
		{"ap item", "ap:item/3", map[string]any{"kind": "game", "title": "X"}, nil, kind.Game},
		{"ap fields", "ap:fields/igdb:1942", map[string]any{"about": []any{"igdb:1942"}}, []string{"igdb:1942"}, ""},
		{"matched", "matched:steam:220", map[string]any{"about": []any{"steam:220", "igdb:1942"}}, []string{"steam:220", "igdb:1942"}, ""},
		{"legacy", "legacy:item/9", map[string]any{"about": []any{"igdb:1"}}, []string{"igdb:1"}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := world.Derive(identity(t, c.id), c.reading)
			if !slices.Equal(got.About, identities(t, c.about...)) {
				t.Errorf("about = %v", got.About)
			}
			if got.Kind != c.kind {
				t.Errorf("kind = %q", got.Kind)
			}
		})
	}
}

func TestALetterboxdFilmLinkedToATVShowIsASeries(t *testing.T) {
	got := world.Derive(identity(t, "letterboxd:abc"), map[string]any{"tmdb_link": "tmdb-tv:1399"})
	if !slices.Equal(got.About, identities(t, "tmdb-tv:1399")) || got.Kind != kind.Series {
		t.Fatalf("got %+v", got)
	}
}

func TestAnApOverrideIsAboutItsItem(t *testing.T) {
	reading := readingOf(t, "ap", "fields/igdb:1942", `{"fields":{"title":"X"}}`)
	got := world.Derive(identity(t, "ap:fields/igdb:1942"), reading)
	if !slices.Equal(got.About, identities(t, "igdb:1942")) {
		t.Fatalf("about = %v", got.About)
	}
}

func TestNoMatchIsAboutOnlyItsItem(t *testing.T) {
	reading := readingOf(t, "ap", "match/igdb/steam:220", `{"to":null}`)
	got := world.Derive(identity(t, "ap:match/igdb/steam:220"), reading)
	if !slices.Equal(got.About, identities(t, "steam:220")) {
		t.Fatalf("about = %v", got.About)
	}
}

func TestAMatchAnswerIsAboutBothItems(t *testing.T) {
	reading := readingOf(t, "ap", "match/igdb/steam:220", `{"to":"1942"}`)
	got := world.Derive(identity(t, "ap:match/igdb/steam:220"), reading)
	if !slices.Equal(got.About, identities(t, "steam:220", "igdb:1942")) {
		t.Fatalf("about = %v", got.About)
	}
}

func TestALinkEditAddsOrSuppressesALink(t *testing.T) {
	added := world.Derive(identity(t, "ap:link/a:1/with/b:2"), readingOf(t, "ap", "link/a:1/with/b:2", `{"edit":"added","from":"a:1","type":"with","to":"b:2"}`))
	if !slices.Equal(added.Links, []world.AssertedLink{link(t, "a:1", "with", "b:2", 0)}) || len(added.Suppressed) != 0 {
		t.Fatalf("added = %+v", added)
	}
	suppressed := world.Derive(identity(t, "ap:link/a:1/with/b:2"), readingOf(t, "ap", "link/a:1/with/b:2", `{"edit":"suppressed","from":"a:1","type":"with","to":"b:2"}`))
	if !slices.Equal(suppressed.Suppressed, []world.AssertedLink{link(t, "a:1", "with", "b:2", 0)}) || len(suppressed.Links) != 0 {
		t.Fatalf("suppressed = %+v", suppressed)
	}
}

func TestALegacyRecordCarriesItsLinksWithPositions(t *testing.T) {
	reading := map[string]any{
		"about": []any{"spotify-album:b"},
		"links": []any{
			map[string]any{"from": "spotify-track-id:t", "type": "part-of", "to": "spotify-album:b", "position": float64(4)},
			map[string]any{"from": "spotify-album:b", "type": "by", "to": "spotify-artist:a"},
			map[string]any{"from": "spotify-album:b", "type": "nope", "to": "spotify-artist:a"},
			map[string]any{"from": "spotify-track-id:u", "type": "part-of", "to": "spotify-album:b"},
		},
	}
	got := world.Derive(identity(t, "legacy:item/9"), reading)
	want := []world.AssertedLink{
		link(t, "spotify-track-id:t", "part-of", "spotify-album:b", 4),
		link(t, "spotify-album:b", "by", "spotify-artist:a", 0),
		link(t, "spotify-album:b", "nope", "spotify-artist:a", 0),
		link(t, "spotify-track-id:u", "part-of", "spotify-album:b", 0),
	}
	if !slices.Equal(got.Links, want) {
		t.Fatalf("links = %+v", got.Links)
	}
}

func TestALinkEditNamesItsLinkInItsData(t *testing.T) {
	data := `{"edit":"added","from":"ap:item/3","type":"with","to":"b:2"}`
	got := world.Derive(identity(t, "ap:link/ap:item/3/with/b:2"), readingOf(t, "ap", "link/ap:item/3/with/b:2", data))
	if !slices.Equal(got.Links, []world.AssertedLink{link(t, "ap:item/3", "with", "b:2", 0)}) {
		t.Fatalf("links = %+v", got.Links)
	}
}

func TestAnAlbumsMainArtistAndOtherCredits(t *testing.T) {
	reading := map[string]any{"artist_ids": []any{"a", "b", "c"}}
	got := world.Derive(identity(t, "spotify-album:x"), reading)
	want := []world.AssertedLink{
		link(t, "spotify-album:x", "by", "spotify-artist:a", 0),
		link(t, "spotify-album:x", "with", "spotify-artist:b", 0),
		link(t, "spotify-album:x", "with", "spotify-artist:c", 0),
	}
	if !slices.Equal(got.Links, want) {
		t.Fatalf("links = %+v", got.Links)
	}
}

func TestAnAlbumByVariousArtistsHasNoMainArtist(t *testing.T) {
	reading := map[string]any{"artist_ids": []any{kind.VariousArtists, "b"}}
	got := world.Derive(identity(t, "spotify-album:x"), reading)
	want := []world.AssertedLink{
		link(t, "spotify-album:x", "with", "spotify-artist:"+kind.VariousArtists, 0),
		link(t, "spotify-album:x", "with", "spotify-artist:b", 0),
	}
	if !slices.Equal(got.Links, want) {
		t.Fatalf("links = %+v", got.Links)
	}
}

func TestAnAlbumsTracksArePartsInTracklistOrder(t *testing.T) {
	reading := map[string]any{"track_ids": []any{"a", "b"}}
	got := world.Derive(identity(t, "spotify-album:x"), reading)
	want := []world.AssertedLink{
		link(t, "spotify-track-id:a", "part-of", "spotify-album:x", 1),
		link(t, "spotify-track-id:b", "part-of", "spotify-album:x", 2),
	}
	if !slices.Equal(got.Links, want) {
		t.Fatalf("links = %+v", got.Links)
	}
}

func TestTakesAndListeningAreAboutTheIdentityInTheirExternalID(t *testing.T) {
	got := world.Derive(identity(t, "letterboxd-take:letterboxd:abc"), nil)
	if !slices.Equal(got.About, identities(t, "letterboxd:abc")) {
		t.Fatalf("about = %v", got.About)
	}
}

func TestNoKindDeclaresADeriveKeyAsAField(t *testing.T) {
	for _, k := range kind.All() {
		fields := k.Fields
		for _, take := range k.Takes {
			fields = append(slices.Clone(fields), take.Fields...)
		}
		for _, f := range fields {
			if slices.Contains(world.DeriveKeys, f.Name) {
				t.Errorf("%s declares derive key %s", k.Name, f.Name)
			}
		}
	}
}

func TestApReadingsCarryFieldsAtTheTop(t *testing.T) {
	raw, readability := reading.ReadRecord("ap", json.RawMessage(`{"kind":"game","fields":{"title":"X"}}`))
	if readability != record.Readable || string(raw) != `{"kind":"game","title":"X"}` {
		t.Fatalf("raw = %s readability = %s", raw, readability)
	}
}

func readingOf(t *testing.T, source, externalID, data string) map[string]any {
	t.Helper()
	raw, readability := reading.ReadRecord(source, json.RawMessage(data))
	if readability != record.Readable {
		t.Fatalf("%s:%s unreadable", source, externalID)
	}
	var reading map[string]any
	if err := json.Unmarshal(raw, &reading); err != nil {
		t.Fatal(err)
	}
	return reading
}

func TestApFieldsAndMatchAnswersReadAsTheirData(t *testing.T) {
	cases := []struct{ data, want string }{
		{`{"fields":{"title":"X"}}`, `{"title":"X"}`},
		{`{"to":"1942"}`, `{"to":"1942"}`},
		{`{"to":null}`, `{"to":null}`},
		{`{"edit":"added"}`, `{"edit":"added"}`},
	}
	for _, c := range cases {
		raw, readability := reading.ReadRecord("ap", json.RawMessage(c.data))
		if readability != record.Readable || string(raw) != c.want {
			t.Errorf("%s: got %s %s", c.data, raw, readability)
		}
	}
}

func TestAnUnknownApShapeIsUnreadable(t *testing.T) {
	_, readability := reading.ReadRecord("ap", json.RawMessage(`{"nope":1}`))
	if readability != record.Unreadable {
		t.Fatal("readable")
	}
	_, readability = reading.ReadRecord("nowhere", json.RawMessage(`{}`))
	if readability != record.Unreadable {
		t.Fatal("readable")
	}
}

func TestMatchedAndLegacyReadingsAreTheirData(t *testing.T) {
	for _, source := range []string{"matched", "legacy"} {
		raw, readability := reading.ReadRecord(source, json.RawMessage(`{"about":["a:1","b:2"]}`))
		if readability != record.Readable || string(raw) != `{"about":["a:1","b:2"]}` {
			t.Errorf("%s: %s %s", source, raw, readability)
		}
	}
}
