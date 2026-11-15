package store_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
)

const fieldGenres = "genres"

func knowArtist(t *testing.T, s *store.Store, spotifyID, name string, genres ...string) int64 {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"id": spotifyID, "name": name, "genres": genres})
	if err != nil {
		t.Fatal(err)
	}
	id, err := s.KnowItemFromSource(context.Background(), kind.Artist, name, kind.SourceSpotifyArtist, spotifyID, json.RawMessage(raw))
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func logAlbum(t *testing.T, s *store.Store, spotifyID, name string, genres ...string) int64 {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"id": spotifyID, "name": name, "genres": genres})
	if err != nil {
		t.Fatal(err)
	}
	id, err := s.CreateItemFromSource(context.Background(), kind.Album, name, kind.SourceSpotifyAlbum, spotifyID, json.RawMessage(raw))
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func relate(t *testing.T, s *store.Store, from int64, relation string, to int64) {
	t.Helper()
	if err := s.AddRelation(context.Background(), store.Relation{FromID: from, Type: relation, ToID: to}); err != nil {
		t.Fatal(err)
	}
}

func listedAlbum(t *testing.T, s *store.Store, albumID int64) store.ItemView {
	t.Helper()
	views, err := s.ListItems(context.Background(), store.ItemFilter{Kind: kind.Album})
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range views {
		if v.ID == albumID {
			return v
		}
	}
	t.Fatalf("album %d is not listed", albumID)
	return store.ItemView{}
}

func bothReads(t *testing.T, s *store.Store, albumID int64) map[string]store.ItemView {
	t.Helper()
	one, err := s.GetItem(context.Background(), albumID)
	if err != nil {
		t.Fatal(err)
	}
	return map[string]store.ItemView{"GetItem": one, "ListItems": listedAlbum(t, s, albumID)}
}

func TestAnAlbumInheritsGenresFromItsMainArtist(t *testing.T) {
	s := openStore(t)
	artist := knowArtist(t, s, "S", "Slowdive", "shoegaze")
	album := logAlbum(t, s, "V", "Souvlaki")
	relate(t, s, album, kind.RelationBy, artist)

	for read, v := range bothReads(t, s, album) {
		if got := v.Verdicts.Values[fieldGenres]; !reflect.DeepEqual(got, []string{"shoegaze"}) {
			t.Errorf("%s genres = %v", read, got)
		}
		if got := v.Verdicts.SourceOf[fieldGenres]; got != "inherited:by" {
			t.Errorf("%s genres source = %q", read, got)
		}
	}
}

func TestAnAlbumsOwnGenresWin(t *testing.T) {
	s := openStore(t)
	artist := knowArtist(t, s, "S", "Slowdive", "shoegaze")
	album := logAlbum(t, s, "V", "Souvlaki", "dream pop")
	relate(t, s, album, kind.RelationBy, artist)

	for read, v := range bothReads(t, s, album) {
		if got := v.Verdicts.Values[fieldGenres]; !reflect.DeepEqual(got, []string{"dream pop"}) || v.Verdicts.SourceOf[fieldGenres] != kind.SourceSpotifyAlbum {
			t.Errorf("%s genres = %v from %q", read, got, v.Verdicts.SourceOf[fieldGenres])
		}
	}

	if err := s.ChangeOverride(context.Background(), album, map[string]any{fieldGenres: []any{"noise"}}); err != nil {
		t.Fatal(err)
	}
	for read, v := range bothReads(t, s, album) {
		if got := v.Verdicts.Values[fieldGenres]; !reflect.DeepEqual(got, []string{"noise"}) || v.Verdicts.SourceOf[fieldGenres] != kind.SourceAP {
			t.Errorf("%s genres after override = %v from %q", read, got, v.Verdicts.SourceOf[fieldGenres])
		}
	}
}

func TestInheritanceIsOneStep(t *testing.T) {
	s := openStore(t)
	credited := knowArtist(t, s, "S", "Slowdive", "shoegaze")
	album := logAlbum(t, s, "V", "Souvlaki")
	relate(t, s, album, kind.RelationWith, credited)

	for read, v := range bothReads(t, s, album) {
		if got, present := v.Verdicts.Values[fieldGenres]; present {
			t.Errorf("%s genres through with = %v", read, got)
		}
	}
}

func TestAnAlbumInheritsFromItsLowestMainArtistAndNothingFromOneWithoutGenres(t *testing.T) {
	s := openStore(t)
	first := knowArtist(t, s, "S", "Slowdive", "shoegaze")
	second := knowArtist(t, s, "M", "Mojave 3", "slowcore")
	album := logAlbum(t, s, "V", "Souvlaki")
	relate(t, s, album, kind.RelationBy, second)
	relate(t, s, album, kind.RelationBy, first)

	for read, v := range bothReads(t, s, album) {
		if got := v.Verdicts.Values[fieldGenres]; !reflect.DeepEqual(got, []string{"shoegaze"}) {
			t.Errorf("%s genres = %v", read, got)
		}
	}

	bare := knowArtist(t, s, "B", "Bare")
	quiet := logAlbum(t, s, "Q", "Quiet")
	relate(t, s, quiet, kind.RelationBy, bare)
	for read, v := range bothReads(t, s, quiet) {
		if got, present := v.Verdicts.Values[fieldGenres]; present {
			t.Errorf("%s genres from an artist without any = %v", read, got)
		}
		if _, present := v.Verdicts.SourceOf[fieldGenres]; present {
			t.Errorf("%s genres has a source without a value", read)
		}
	}
}
