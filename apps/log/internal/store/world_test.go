package store_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"maps"
	"reflect"
	"regexp"
	"slices"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

type storedClaim struct {
	data      string
	changedAt string
}

func claimsStored(t *testing.T, d *sql.DB) map[string]storedClaim {
	t.Helper()
	rows, err := d.Query(`SELECT source || ':' || external_id, data, changed_at FROM record_claim`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	claims := map[string]storedClaim{}
	for rows.Next() {
		var identity string
		var c storedClaim
		if err := rows.Scan(&identity, &c.data, &c.changedAt); err != nil {
			t.Fatal(err)
		}
		claims[identity] = c
	}
	return claims
}

func claimData(t *testing.T, d *sql.DB, identity string) map[string]any {
	t.Helper()
	c, stored := claimsStored(t, d)[identity]
	if !stored {
		t.Fatalf("no claim %s", identity)
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(c.data), &data); err != nil {
		t.Fatal(err)
	}
	return data
}

func steamGame(t *testing.T, s *store.Store, appID, name string) int64 {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"appid": json.Number(appID), "name": name})
	if err != nil {
		t.Fatal(err)
	}
	id, err := s.CreateItemFromSource(context.Background(), kind.Game, name, record.SourceSteam, appID, json.RawMessage(raw))
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestAResyncWithTheSameClaimDerivesNothing(t *testing.T) {
	ctx := context.Background()
	s, d := openStoreWithDB(t)
	data := json.RawMessage(`{"appid":1145360,"name":"Hades","playtime_forever":3000}`)
	id, err := s.CreateItemFromSource(ctx, kind.Game, "Hades", record.SourceSteam, "1145360", data)
	if err != nil {
		t.Fatal(err)
	}
	const untouched = `{"values":{"title":"untouched"},"sourceOf":{"title":"steam"}}`
	if _, err := d.Exec(`UPDATE item_verdict SET own = ?, verdicts = ? WHERE item_id = ?`, untouched, untouched, id); err != nil {
		t.Fatal(err)
	}
	before := claimsStored(t, d)

	if err := putClaimOnItem(ctx, s, id, record.Identity{Source: record.SourceSteam, ExternalID: "1145360"}, data); err != nil {
		t.Fatal(err)
	}

	var own, verdicts string
	if err := d.QueryRow(`SELECT own, verdicts FROM item_verdict WHERE item_id = ?`, id).Scan(&own, &verdicts); err != nil {
		t.Fatal(err)
	}
	if own != untouched || verdicts != untouched {
		t.Errorf("the resync derived again: own %s, verdicts %s", own, verdicts)
	}
	if after := claimsStored(t, d); !maps.Equal(before, after) {
		t.Errorf("claims %v became %v", before, after)
	}
}

func TestAClaimAnotherItemOwnsIsAConflictAndWritesNothing(t *testing.T) {
	ctx := context.Background()
	s, d := openStoreWithDB(t)
	first := steamGame(t, s, "1", "First")
	second := steamGame(t, s, "2", "Second")
	if err := putClaimOnItem(ctx, s, first, record.Identity{Source: record.SourceIGDB, ExternalID: "1"}, json.RawMessage(`{"id":1,"name":"First"}`)); err != nil {
		t.Fatal(err)
	}
	before := claimsStored(t, d)

	err := putClaimOnItem(ctx, s, second, record.Identity{Source: record.SourceIGDB, ExternalID: "1"}, json.RawMessage(`{"id":1,"name":"Changed"}`))
	if !errors.Is(err, store.ErrConflict) {
		t.Fatalf("err = %v, want a conflict", err)
	}
	if after := claimsStored(t, d); !maps.Equal(before, after) {
		t.Errorf("claims %v became %v", before, after)
	}
	if owner, err := s.FindByIdentity(ctx, record.SourceIGDB, "1"); err != nil || owner != first {
		t.Errorf("igdb:1 belongs to %d, %v; want %d", owner, err, first)
	}
}

func TestASyncMatchIsAMatchedRecord(t *testing.T) {
	ctx := context.Background()
	s, d := openStoreWithDB(t)
	id := steamGame(t, s, "220", "Half-Life 2")

	if err := putClaimOnItem(ctx, s, id, record.Identity{Source: record.SourceIGDB, ExternalID: "7"}, json.RawMessage(`{"id":7,"name":"Half-Life 2: Episode"}`)); err != nil {
		t.Fatal(err)
	}

	if _, claimed := claimsStored(t, d)["igdb:7"]; !claimed {
		t.Error("no claim igdb:7")
	}
	matched := claimData(t, d, "matched:igdb:7")
	if want := map[string]any{"about": []any{"steam:220", "igdb:7"}}; !reflect.DeepEqual(matched, want) {
		t.Errorf("matched = %v, want %v", matched, want)
	}
	if owner, err := s.FindByIdentity(ctx, record.SourceIGDB, "7"); err != nil || owner != id {
		t.Errorf("igdb:7 belongs to %d, %v; want %d", owner, err, id)
	}
	v, err := s.GetItem(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if v.Verdicts.Values["title"] != "Half-Life 2: Episode" || v.Verdicts.SourceOf["title"] != record.SourceIGDB {
		t.Errorf("verdicts = %v", v.Verdicts)
	}
}

func TestApsPickIsAnApMatchAnswer(t *testing.T) {
	ctx := context.Background()
	s, d := openStoreWithDB(t)
	id := steamGame(t, s, "220", "Half-Life 2")

	if err := s.AnswerMatch(ctx, id, record.SourceIGDB, "7", json.RawMessage(`{"id":7,"name":"Half-Life 2"}`)); err != nil {
		t.Fatal(err)
	}

	if answer := claimData(t, d, "ap:match/igdb/steam:220"); !reflect.DeepEqual(answer, map[string]any{"to": "7"}) {
		t.Errorf("answer = %v", answer)
	}
	if _, claimed := claimsStored(t, d)["igdb:7"]; !claimed {
		t.Error("no claim igdb:7")
	}
	if owner, err := s.FindByIdentity(ctx, record.SourceIGDB, "7"); err != nil || owner != id {
		t.Errorf("igdb:7 belongs to %d, %v; want %d", owner, err, id)
	}
	other := steamGame(t, s, "221", "Other")
	if err := s.AnswerMatch(ctx, other, record.SourceIGDB, "7", json.RawMessage(`{"id":7}`)); !errors.Is(err, store.ErrConflict) {
		t.Errorf("picking igdb:7 for a second item: %v", err)
	}
}

func TestNoMatchKeepsAnItemOutOfTheMatchQueue(t *testing.T) {
	ctx := context.Background()
	s, d := openStoreWithDB(t)
	answered := steamGame(t, s, "1", "Answered")
	waiting := steamGame(t, s, "2", "Waiting")

	if err := s.AnswerNoMatch(ctx, answered, record.SourceIGDB); err != nil {
		t.Fatal(err)
	}

	if answer := claimData(t, d, "ap:match/igdb/steam:1"); !reflect.DeepEqual(answer, map[string]any{"to": nil}) {
		t.Errorf("answer = %v", answer)
	}
	queue, err := s.ListItems(ctx, store.ItemFilter{Kind: kind.Game, Needs: store.NeedsMatch})
	if err != nil {
		t.Fatal(err)
	}
	if len(queue) != 1 || queue[0].ID != waiting {
		t.Errorf("match queue = %+v, want only %d", queue, waiting)
	}
	lacking, err := s.IdentitiesLackingSources(ctx, record.SourceSteam, record.SourceIGDB)
	if err != nil {
		t.Fatal(err)
	}
	if want := []record.Identity{{Source: record.SourceSteam, ExternalID: "2"}}; !slices.Equal(lacking, want) {
		t.Errorf("lacking = %v, want %v", lacking, want)
	}
}

func TestAnOverrideIsAnApFieldsRecord(t *testing.T) {
	ctx := context.Background()
	s, d := openStoreWithDB(t)
	id := steamGame(t, s, "1145360", "Hades")

	if err := s.ChangeOverride(ctx, id, map[string]any{"title": "HADES", "released": "2020"}); err != nil {
		t.Fatal(err)
	}
	if err := s.ChangeOverride(ctx, id, map[string]any{"released": nil}); err != nil {
		t.Fatal(err)
	}

	want := map[string]any{"fields": map[string]any{"title": "HADES"}}
	if fields := claimData(t, d, "ap:fields/steam:1145360"); !reflect.DeepEqual(fields, want) {
		t.Errorf("override = %v, want %v", fields, want)
	}
	v, err := s.GetItem(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if v.Verdicts.Values["title"] != "HADES" || v.Verdicts.SourceOf["title"] != record.SourceAP {
		t.Errorf("verdicts = %v", v.Verdicts)
	}
	if _, released := v.Verdicts.Values["released"]; released {
		t.Errorf("the removed override stayed: %v", v.Verdicts.Values)
	}
}

var handMadeIdentity = regexp.MustCompile(`^ap:item/[0-9a-f]{12}$`)

func handMadeRecords(t *testing.T, d *sql.DB) map[string]map[string]any {
	t.Helper()
	records := map[string]map[string]any{}
	for identity := range claimsStored(t, d) {
		if handMadeIdentity.MatchString(identity) {
			records[identity] = claimData(t, d, identity)
		}
	}
	return records
}

func TestAHandMadeItemIsAnApItemRecord(t *testing.T) {
	ctx := context.Background()
	s, d := openStoreWithDB(t)

	id, err := s.CreateItem(ctx, kind.Game, "Tetris", nil)
	if err != nil {
		t.Fatal(err)
	}

	records := handMadeRecords(t, d)
	if len(records) != 1 {
		t.Fatalf("hand-made records = %v", records)
	}
	for _, data := range records {
		if want := map[string]any{"kind": "game", "fields": map[string]any{"title": "Tetris"}}; !reflect.DeepEqual(data, want) {
			t.Errorf("record = %v, want %v", data, want)
		}
	}
	v, err := s.GetItem(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if v.Kind != kind.Game || v.Slug != "tetris" || v.Entry == nil || v.Entry.LoggedBy != record.SourceAP {
		t.Errorf("item = %+v", v)
	}

	if err := s.ChangeOverride(ctx, id, map[string]any{"released": "1984"}); err != nil {
		t.Fatal(err)
	}
	for identity, data := range handMadeRecords(t, d) {
		want := map[string]any{"kind": "game", "fields": map[string]any{"title": "Tetris", "released": "1984"}}
		if !reflect.DeepEqual(data, want) {
			t.Errorf("%s = %v, want %v", identity, data, want)
		}
	}
	for identity := range claimsStored(t, d) {
		if regexp.MustCompile(`^ap:fields/`).MatchString(identity) {
			t.Errorf("a hand-made item's override went to %s", identity)
		}
	}
}

func TestOnlyAHandMadeItemCanBeDeleted(t *testing.T) {
	ctx := context.Background()
	s, d := openStoreWithDB(t)
	fromSteam := steamGame(t, s, "1", "From Steam")
	handMade, err := s.CreateItem(ctx, kind.Game, "Tetris", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ChangeOverride(ctx, fromSteam, map[string]any{"title": "Mine"}); err != nil {
		t.Fatal(err)
	}

	if err := s.DeleteItem(ctx, fromSteam); !errors.Is(err, store.ErrInvalid) {
		t.Errorf("deleting an item from steam: %v", err)
	}
	if _, err := s.GetItem(ctx, fromSteam); err != nil {
		t.Errorf("the item from steam is gone: %v", err)
	}
	if err := s.DeleteItem(ctx, handMade); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetItem(ctx, handMade); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("the hand-made item is still there: %v", err)
	}
	if records := handMadeRecords(t, d); len(records) != 0 {
		t.Errorf("its records stayed: %v", records)
	}
	if err := s.DeleteItem(ctx, handMade); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("deleting it again: %v", err)
	}
}

func TestIdentifyingAFilmAsASeriesMovesItAndItsRank(t *testing.T) {
	ctx := context.Background()
	s, d := openStoreWithDB(t)
	watched := func(externalID, title string) int64 {
		t.Helper()
		id, err := s.CreateItemFromSource(ctx, kind.Movie, title, record.SourceTMDB, externalID, json.RawMessage(`{"id":`+externalID+`,"title":"`+title+`"}`))
		if err != nil {
			t.Fatal(err)
		}
		if err := s.ChangeEntry(ctx, id, store.EntryChange{Status: ptr(kind.Watched), Scores: map[string]*string{kind.Prestige: ptr("a")}}); err != nil {
			t.Fatal(err)
		}
		return id
	}
	first := watched("1", "First Film")
	ripley := watched("603", "Ripley")
	last := watched("2", "Last Film")

	identity := store.Identity{Kind: kind.Series, Source: record.SourceTMDBSeries, ExternalID: "94028", Data: json.RawMessage(`{"id":94028,"name":"Ripley"}`)}
	if err := s.Identify(ctx, ripley, identity, []string{record.SourceTMDB, record.SourceTMDBSeries}); err != nil {
		t.Fatal(err)
	}

	v, err := s.GetItem(ctx, ripley)
	if err != nil {
		t.Fatal(err)
	}
	if v.Kind != kind.Series || v.Slug != "ripley" || v.Entry.Status != kind.Watched || v.Entry.Scores[kind.Prestige] != "a" {
		t.Errorf("moved item = %s/%s %+v", v.Kind, v.Slug, v.Entry)
	}
	if got, want := ranksIn(t, s, kind.Movie), map[int64]int{first: 1, last: 2}; !maps.Equal(got, want) {
		t.Errorf("films = %v, want %v", got, want)
	}
	if got, want := ranksIn(t, s, kind.Series), map[int64]int{ripley: 1}; !maps.Equal(got, want) {
		t.Errorf("series = %v, want %v", got, want)
	}
	if _, kept := claimsStored(t, d)["tmdb:603"]; kept {
		t.Error("the tmdb claim stayed")
	}
	if answer := claimData(t, d, "ap:match/tmdb-tv/tmdb-tv:94028"); !reflect.DeepEqual(answer, map[string]any{"to": "94028"}) {
		t.Errorf("answer = %v", answer)
	}
	if owner, err := s.FindByIdentity(ctx, record.SourceTMDB, "603"); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("tmdb:603 still belongs to %d, %v", owner, err)
	}
	removed, err := s.Observations(ctx, ripley, record.SourceTMDB, "claim-removed")
	if err != nil {
		t.Fatal(err)
	}
	if len(removed) != 1 || string(removed[0].Data) != `{"id":603,"title":"Ripley"}` {
		t.Errorf("observations of the removed claim = %+v", removed)
	}
}

func relationsOf(t *testing.T, s *store.Store, itemID int64) map[store.Relation]struct{} {
	t.Helper()
	v, err := s.GetItem(context.Background(), itemID)
	if err != nil {
		t.Fatal(err)
	}
	relations := map[store.Relation]struct{}{}
	for _, r := range v.Relations {
		relations[r] = struct{}{}
	}
	return relations
}

func TestAnApLinkEditIsARecordAndSuppressionHides(t *testing.T) {
	ctx := context.Background()
	s, d := openStoreWithDB(t)
	ann := knowArtist(t, s, "A", "Ann")
	ben := knowArtist(t, s, "B", "Ben")
	album, err := s.CreateItemFromSource(ctx, kind.Album, "Alpha", record.SourceSpotifyAlbum, "AL",
		json.RawMessage(`{"id":"AL","name":"Alpha","artists":[{"id":"A","name":"Ann"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	by := store.Relation{FromID: album, Type: kind.LinkBy, ToID: ann}
	with := store.Relation{FromID: album, Type: kind.LinkWith, ToID: ben}

	if err := s.AddRelation(ctx, with); err != nil {
		t.Fatal(err)
	}
	added := map[string]any{"edit": "added", "from": "spotify-album:AL", "type": "with", "to": "spotify-artist:B"}
	if edit := claimData(t, d, "ap:link/spotify-album:AL/with/spotify-artist:B"); !reflect.DeepEqual(edit, added) {
		t.Errorf("edit = %v, want %v", edit, added)
	}
	if got, want := relationsOf(t, s, album), map[store.Relation]struct{}{by: {}, with: {}}; !maps.Equal(got, want) {
		t.Errorf("relations = %v, want %v", got, want)
	}

	if err := s.RemoveRelation(ctx, by); err != nil {
		t.Fatal(err)
	}
	suppressed := map[string]any{"edit": "suppressed", "from": "spotify-album:AL", "type": "by", "to": "spotify-artist:A"}
	if edit := claimData(t, d, "ap:link/spotify-album:AL/by/spotify-artist:A"); !reflect.DeepEqual(edit, suppressed) {
		t.Errorf("edit = %v, want %v", edit, suppressed)
	}
	if err := s.RemoveRelation(ctx, with); err != nil {
		t.Fatal(err)
	}
	if _, kept := claimsStored(t, d)["ap:link/spotify-album:AL/with/spotify-artist:B"]; kept {
		t.Error("the added link's record stayed")
	}
	if got := relationsOf(t, s, album); len(got) != 0 {
		t.Errorf("relations = %v, want none", got)
	}
	if err := s.RemoveRelation(ctx, by); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("removing it again: %v", err)
	}
}

func TestCorrectingASyncMatchFreesTheWrongIdentity(t *testing.T) {
	corrections := map[string]func(s *store.Store, item int64) error{
		"a pick": func(s *store.Store, item int64) error {
			return s.AnswerMatch(context.Background(), item, record.SourceIGDB, "10", json.RawMessage(`{"id":10,"name":"Right"}`))
		},
		"no match": func(s *store.Store, item int64) error {
			return s.AnswerNoMatch(context.Background(), item, record.SourceIGDB)
		},
	}
	for name, correct := range corrections {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			s, d := openStoreWithDB(t)
			first := steamGame(t, s, "1", "First")
			if err := putClaimOnItem(ctx, s, first, record.Identity{Source: record.SourceIGDB, ExternalID: "9"}, json.RawMessage(`{"id":9,"name":"Wrong"}`)); err != nil {
				t.Fatal(err)
			}

			if err := correct(s, first); err != nil {
				t.Fatal(err)
			}

			claims := claimsStored(t, d)
			for _, gone := range []string{"igdb:9", "matched:igdb:9"} {
				if _, kept := claims[gone]; kept {
					t.Errorf("%s stayed", gone)
				}
			}
			if owner, err := s.FindByIdentity(ctx, record.SourceIGDB, "9"); !errors.Is(err, store.ErrNotFound) {
				t.Errorf("igdb:9 still belongs to %d, %v", owner, err)
			}
			second := steamGame(t, s, "2", "Second")
			if err := putClaimOnItem(ctx, s, second, record.Identity{Source: record.SourceIGDB, ExternalID: "9"}, json.RawMessage(`{"id":9,"name":"Wrong"}`)); err != nil {
				t.Fatalf("the second item can't take igdb:9: %v", err)
			}
			if owner, err := s.FindByIdentity(ctx, record.SourceIGDB, "9"); err != nil || owner != second {
				t.Errorf("igdb:9 belongs to %d, %v; want %d", owner, err, second)
			}
		})
	}
}

func TestIdentifyingAFilmFreesItsWrongTMDBMatch(t *testing.T) {
	ctx := context.Background()
	s, _ := openStoreWithDB(t)
	film := func(id, title string) int64 {
		t.Helper()
		item, err := s.CreateItemFromSource(ctx, kind.Movie, title, record.SourceLetterboxd, id, json.RawMessage(`{"id":"`+id+`","title":"`+title+`"}`))
		if err != nil {
			t.Fatal(err)
		}
		return item
	}
	ripley := film("rip1", "Ripley")
	if err := putClaimOnItem(ctx, s, ripley, record.Identity{Source: record.SourceTMDB, ExternalID: "9"}, json.RawMessage(`{"id":9,"title":"Wrong"}`)); err != nil {
		t.Fatal(err)
	}

	identity := store.Identity{Kind: kind.Series, Source: record.SourceTMDBSeries, ExternalID: "94028", Data: json.RawMessage(`{"id":94028,"name":"Ripley"}`)}
	if err := s.Identify(ctx, ripley, identity, []string{record.SourceTMDB, record.SourceTMDBSeries}); err != nil {
		t.Fatal(err)
	}

	wrong := film("wrong1", "Wrong")
	if err := putClaimOnItem(ctx, s, wrong, record.Identity{Source: record.SourceTMDB, ExternalID: "9"}, json.RawMessage(`{"id":9,"title":"Wrong"}`)); err != nil {
		t.Fatalf("another film can't take tmdb:9: %v", err)
	}
	if owner, err := s.FindByIdentity(ctx, record.SourceTMDB, "9"); err != nil || owner != wrong {
		t.Errorf("tmdb:9 belongs to %d, %v; want %d", owner, err, wrong)
	}
}

func TestAClaimWhoseDeriveLandsOnAnotherItemIsAConflict(t *testing.T) {
	ctx := context.Background()
	s, d := openStoreWithDB(t)
	first := steamGame(t, s, "1", "First")
	steamGame(t, s, "2", "Second")
	before := claimsStored(t, d)

	err := putClaimOnItem(ctx, s, first, record.Identity{Source: record.SourceSteamStore, ExternalID: "2"}, json.RawMessage(`{"name":"Second"}`))
	if !errors.Is(err, store.ErrConflict) {
		t.Fatalf("err = %v, want a conflict", err)
	}
	if after := claimsStored(t, d); !maps.Equal(before, after) {
		t.Errorf("claims %v became %v", before, after)
	}
}

func TestDeletingAHandMadeItemRemovesTheLinkEditsNamingIt(t *testing.T) {
	ctx := context.Background()
	s, d := openStoreWithDB(t)
	ann, err := s.CreateItem(ctx, kind.Artist, "Ann", nil)
	if err != nil {
		t.Fatal(err)
	}
	album, err := s.CreateItem(ctx, kind.Album, "Alpha", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.AddRelation(ctx, store.Relation{FromID: album, Type: kind.LinkBy, ToID: ann}); err != nil {
		t.Fatal(err)
	}

	if err := s.DeleteItem(ctx, ann); err != nil {
		t.Fatal(err)
	}

	for identity := range claimsStored(t, d) {
		if regexp.MustCompile(`^ap:link/`).MatchString(identity) {
			t.Errorf("%s stayed", identity)
		}
	}
	if got := relationsOf(t, s, album); len(got) != 0 {
		t.Errorf("relations = %v, want none", got)
	}
}
