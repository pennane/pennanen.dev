package spotify

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/pennane/pennanen.dev/apps/log/internal/platform/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/rule"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

const (
	pendulum  int64 = 1
	daftPunk  int64 = 2
	discovery int64 = 3
	strangers int64 = 4
)

func logged(itemID int64, ruleName string) rule.Proposal {
	return rule.Proposal{ItemID: itemID, Rule: ruleName, Subject: rule.SubjectEntry, Value: rule.EntryLogged, Strength: rule.Applying}
}

func status(itemID int64, ruleName, value, at string) rule.Proposal {
	return rule.Proposal{ItemID: itemID, Rule: ruleName, Subject: rule.SubjectStatus, Value: value, Strength: rule.Applying, ChangedAt: at}
}

func assertProposals(t *testing.T, got, want []rule.Proposal) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("proposals = %+v\nwant %+v", got, want)
	}
}

func snapshotAt(n int) string {
	return time.Date(2026, 1, 1, 3, 0, 0, 0, time.UTC).AddDate(0, 0, n).Format(time.RFC3339)
}

func inTopArtists(itemID int64, timeRange TimeRange, n int) observation {
	at := snapshotAt(n)
	return observation{itemID: itemID, signal: record.SignalTopArtist, key: fmt.Sprintf("%s:%s:%s:%d", record.SignalTopArtist, timeRange, at[:10], itemID), observedAt: at}
}

func snapshots(itemID int64, from, through int) []observation {
	var observed []observation
	for n := from; n <= through; n++ {
		observed = append(observed, inTopArtists(itemID, ShortTerm, n))
	}
	return observed
}

func albumObserved(itemID int64, signal, at string) observation {
	return observation{itemID: itemID, signal: signal, key: signal + ":" + at, observedAt: at}
}

func TestSavedAlbumIsWantedUntilHeard(t *testing.T) {
	got := propose([]observation{albumObserved(discovery, record.SignalSaved, "2025-05-01T10:00:00Z")}, nil)
	assertProposals(t, got, []rule.Proposal{logged(discovery, RuleAlbum), status(discovery, RuleAlbum, kind.Want, "2025-05-01T10:00:00Z")})
}

func TestHeardAlbumIsListenedFromFirstHearing(t *testing.T) {
	got := propose([]observation{
		albumObserved(discovery, record.SignalSaved, "2025-05-01T10:00:00Z"),
		albumObserved(discovery, record.SignalPlay, "2025-07-01T10:00:00.589Z"),
		albumObserved(discovery, record.SignalTopTrack, "2025-06-01T03:00:00Z"),
		albumObserved(discovery, record.SignalPlay, "2025-08-01T10:00:00Z"),
	}, nil)
	assertProposals(t, got, []rule.Proposal{logged(discovery, RuleAlbum), status(discovery, RuleAlbum, kind.Listened, "2025-06-01T03:00:00Z")})

	unsaved := propose([]observation{albumObserved(discovery, record.SignalPlay, "2025-07-01T10:00:00.589Z")}, nil)
	assertProposals(t, unsaved, []rule.Proposal{status(discovery, RuleAlbum, kind.Listened, "2025-07-01T10:00:00Z")})
}

func TestArtistInEveryRecentSnapshotIsListeningSinceTheRunBegan(t *testing.T) {
	observed := append(snapshots(daftPunk, 1, 30), snapshots(pendulum, 21, 30)...)
	observed = append(observed, inTopArtists(pendulum, ShortTerm, 5))
	got := propose(observed, nil)
	listeningSince21st := status(pendulum, RuleArtist, kind.Listening, snapshotAt(21))
	assertProposals(t, got, []rule.Proposal{
		logged(pendulum, RuleArtist), listeningSince21st,
		logged(daftPunk, RuleArtist), status(daftPunk, RuleArtist, kind.Listening, snapshotAt(1)),
	})

	again := propose(append(observed, inTopArtists(daftPunk, ShortTerm, 31), inTopArtists(pendulum, ShortTerm, 31)), nil)
	assertProposals(t, again, got)
}

func TestArtistThatDroppedOutIsListenedFromTheNextSnapshot(t *testing.T) {
	observed := append(snapshots(daftPunk, 1, 8), snapshots(pendulum, 1, 5)...)
	observed = append(observed, inTopArtists(pendulum, LongTerm, 8))
	got := propose(observed, nil)
	assertProposals(t, got, []rule.Proposal{
		logged(pendulum, RuleArtist), status(pendulum, RuleArtist, kind.Listened, snapshotAt(6)),
		logged(daftPunk, RuleArtist), status(daftPunk, RuleArtist, kind.Listening, snapshotAt(1)),
	})
}

func TestArtistOnlyBehindASavedAlbumIsListenedFromTheSave(t *testing.T) {
	got := propose([]observation{
		albumObserved(discovery, record.SignalSaved, "2025-05-01T10:00:00Z"),
		albumObserved(discovery, record.SignalSaved, "2026-02-01T03:00:00Z"),
	}, []mainArtist{{albumID: discovery, artistID: daftPunk}})
	assertProposals(t, got, []rule.Proposal{
		logged(daftPunk, RuleArtist), status(daftPunk, RuleArtist, kind.Listened, "2025-05-01T10:00:00Z"),
		logged(discovery, RuleAlbum), status(discovery, RuleAlbum, kind.Want, "2025-05-01T10:00:00Z"),
	})
}

func TestArtistOnlyInLongerListsIsListenedFromItsFirstListing(t *testing.T) {
	observed := append(snapshots(daftPunk, 1, 3), inTopArtists(pendulum, MediumTerm, 2), inTopArtists(pendulum, LongTerm, 3))
	got := propose(observed, nil)
	assertProposals(t, got, []rule.Proposal{
		logged(pendulum, RuleArtist), status(pendulum, RuleArtist, kind.Listened, snapshotAt(2)),
		logged(daftPunk, RuleArtist), status(daftPunk, RuleArtist, kind.Listening, snapshotAt(1)),
	})
}

func TestStrangerArtistSaysNothing(t *testing.T) {
	got := propose([]observation{albumObserved(discovery, record.SignalPlay, "2025-07-01T10:00:00Z")}, []mainArtist{{albumID: discovery, artistID: strangers}})
	assertProposals(t, got, []rule.Proposal{status(discovery, RuleAlbum, kind.Listened, "2025-07-01T10:00:00Z")})
}

func TestSpotifyProposerReadsStoredRecords(t *testing.T) {
	ctx := context.Background()
	d, err := db.Open(ctx, filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	seed := `
		INSERT INTO item (id, kind, slug) VALUES (1, 'artist', 'pendulum'), (2, 'artist', 'daft-punk'), (3, 'album', 'discovery'), (4, 'artist', 'strangers');
		INSERT INTO item_identity (item_id, source, external_id) VALUES
			(1, 'spotify-artist', 'pend'), (2, 'spotify-artist', 'daft'), (3, 'spotify-album', 'disc'), (4, 'spotify-artist', 'str');
		INSERT INTO record_claim (source, external_id, data, fetched_at, changed_at) VALUES
			('spotify-album', 'disc', '{}', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z');
		INSERT INTO link (asserted_source, asserted_external_id, from_source, from_external_id, type, to_source, to_external_id) VALUES
			('spotify-album', 'disc', 'spotify-album', 'disc', 'by', 'spotify-artist', 'daft');
		INSERT INTO record_observation (about_source, about_external_id, source, signal, key, observed_at, data) VALUES
			('spotify-artist', 'pend', 'spotify', 'top-artist', 'top-artist:short_term:2026-01-01:pend', '2026-01-01T03:00:00Z', '{}'),
			('spotify-artist', 'pend', 'spotify', 'top-artist', 'top-artist:short_term:2026-01-02:pend', '2026-01-02T03:00:00Z', '{}'),
			('spotify-album', 'disc', 'spotify', 'saved', 'saved:disc:2025-05-01T10:00:00Z', '2025-05-01T10:00:00Z', '{}'),
			('spotify-album', 'disc', 'spotify', 'unsaved', 'unsaved:disc:2025-09-01', '2025-09-01T03:00:00Z', '{}'),
			('spotify-album', 'disc', 'spotify', 'play', 'play:2025-07-01T10:00:00.589Z:t1', '2025-07-01T10:00:00.589Z', '{}'),
			('spotify-artist', 'str', 'letterboxd', 'top-artist', 'top-artist:short_term:2026-01-01:str', '2026-01-01T03:00:00Z', '{}');`
	if _, err := d.Exec(seed); err != nil {
		t.Fatal(err)
	}
	got, err := Proposer.Propose(ctx, d)
	if err != nil {
		t.Fatal(err)
	}
	assertProposals(t, got, []rule.Proposal{
		logged(pendulum, RuleArtist), status(pendulum, RuleArtist, kind.Listening, "2026-01-01T03:00:00Z"),
		logged(daftPunk, RuleArtist), status(daftPunk, RuleArtist, kind.Listened, "2025-05-01T10:00:00Z"),
		logged(discovery, RuleAlbum), status(discovery, RuleAlbum, kind.Listened, "2025-07-01T10:00:00Z"),
	})
	if Proposer.Source != "spotify" {
		t.Errorf("source = %s", Proposer.Source)
	}
	if _, err := rule.Run(ctx, d, Proposer, "public"); err != nil {
		t.Errorf("running the proposals = %v", err)
	}
}

func listed(timeRange TimeRange, n int) observation {
	at := snapshotAt(n)
	return observation{signal: record.SignalTopArtistsListed, key: fmt.Sprintf("%s:%s:%s", record.SignalTopArtistsListed, timeRange, at[:10]), observedAt: at}
}

func TestAnEmptyShortTermListingIsASnapshotTheArtistDroppedOutOf(t *testing.T) {
	observed := append(snapshots(pendulum, 1, 3), listed(ShortTerm, 1), listed(ShortTerm, 2), listed(ShortTerm, 3), listed(ShortTerm, 4), listed(LongTerm, 5))
	got := propose(observed, nil)
	assertProposals(t, got, []rule.Proposal{
		logged(pendulum, RuleArtist), status(pendulum, RuleArtist, kind.Listened, snapshotAt(4)),
	})
}

const (
	artistA    int64 = 1
	artistB    int64 = 2
	artistC    int64 = 3
	albumD     int64 = 4
	artistE    int64 = 5
	albumF     int64 = 6
	listenedAt       = "2026-01-01T03:00:00Z"
)

func openListening(t *testing.T) (context.Context, *sql.DB) {
	t.Helper()
	ctx := context.Background()
	d, err := db.Open(ctx, filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	seed := `
		INSERT INTO item (id, kind, slug) VALUES
			(1, 'artist', 'a'), (2, 'artist', 'b'), (3, 'artist', 'c'), (4, 'album', 'd'), (5, 'artist', 'e'), (6, 'album', 'f');
		INSERT INTO item_identity (item_id, source, external_id) VALUES
			(1, 'spotify-artist', 'a'), (2, 'spotify-artist', 'b'), (3, 'spotify-artist', 'c'),
			(4, 'spotify-album', 'd'), (5, 'spotify-artist', 'e'), (6, 'spotify-album', 'f');
		INSERT INTO record_claim (source, external_id, data, fetched_at, changed_at) VALUES
			('spotify-album', 'd', '{}', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z');
		INSERT INTO link (asserted_source, asserted_external_id, from_source, from_external_id, type, to_source, to_external_id) VALUES
			('spotify-album', 'd', 'spotify-album', 'd', 'by', 'spotify-artist', 'a');
		INSERT INTO record_observation (about_source, about_external_id, source, signal, key, observed_at, data) VALUES
			('spotify-album', 'd', 'spotify', 'top-track', 'top-track:medium_term:2026-01-01:d', '2026-01-01T03:00:00Z', '{"range":"medium_term","day":"2026-01-01","rank":1}'),
			('spotify-album', 'd', 'spotify', 'play', 'play:1', '2026-01-01T03:00:00Z', '{}'),
			('spotify-album', 'd', 'spotify', 'play', 'play:2', '2026-01-01T03:00:00Z', '{}'),
			('spotify-album', 'd', 'spotify', 'play', 'play:3', '2026-01-01T03:00:00Z', '{}');
		INSERT INTO entry (item_id) VALUES (1), (2), (3), (4), (6);`
	if _, err := d.ExecContext(ctx, seed); err != nil {
		t.Fatal(err)
	}
	return ctx, d
}

func listedInShortTerm(t *testing.T, d *sql.DB, artists ...string) {
	t.Helper()
	for i, artist := range artists {
		if _, err := d.Exec(`
			INSERT INTO record_observation (about_source, about_external_id, source, signal, key, observed_at, data)
			VALUES ('spotify-artist', ?, 'spotify', 'top-artist', ?, ?, json_object('range', 'short_term', 'day', '2026-01-01', 'rank', ?))`,
			artist, "top-artist:short_term:2026-01-01:"+artist, listenedAt, i+1); err != nil {
			t.Fatal(err)
		}
	}
}

func listening(itemID, value int64) rule.Measure {
	return rule.Measure{ItemID: itemID, Field: kind.FieldListening, Value: value}
}

func assertMeasures(t *testing.T, got, want []rule.Measure) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("measures = %+v\nwant %+v", got, want)
	}
}

func TestListeningMeasureMatchesOldPercentiles(t *testing.T) {
	ctx, d := openListening(t)
	listedInShortTerm(t, d, "a", "b")
	got, err := Proposer.Measure(ctx, d)
	if err != nil {
		t.Fatal(err)
	}
	assertMeasures(t, got, []rule.Measure{
		listening(artistA, 100), listening(artistB, 50), listening(artistC, 0),
		listening(albumD, 100), listening(albumF, 0),
	})
}

func TestListeningSkipsDismissed(t *testing.T) {
	ctx, d := openListening(t)
	listedInShortTerm(t, d, "a", "e", "b")
	if _, err := d.Exec(`INSERT INTO entry (item_id) VALUES (5); INSERT INTO dismissal (item_id, dismissed_at) VALUES (5, '2026-01-02T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	got, err := Proposer.Measure(ctx, d)
	if err != nil {
		t.Fatal(err)
	}
	assertMeasures(t, got, []rule.Measure{
		listening(artistA, 100), listening(artistB, 50), listening(artistC, 0),
		listening(albumD, 100), listening(albumF, 0),
	})
}

func TestListeningIsOnlyForAlbumsAndArtists(t *testing.T) {
	ctx, d := openListening(t)
	if _, err := d.Exec(`INSERT INTO item (id, kind, slug) VALUES (7, 'track', 'g'); INSERT INTO entry (item_id) VALUES (7)`); err != nil {
		t.Fatal(err)
	}
	got, err := Proposer.Measure(ctx, d)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range got {
		if m.ItemID == 7 {
			t.Errorf("a track was measured: %+v", m)
		}
	}
}
