package verdict_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/verdict"
)

type claimFixture struct {
	kind   string
	source string
	data   string
}

var claimFixtures = []claimFixture{
	{"game", record.SourceIGDB, `{"id": 7, "name": "Hades", "first_release_date": 1600000000, "cover": {"image_id": "co1"}, "genres": [{"name": "Roguelike"}], "platforms": [{"name": "PC"}], "involved_companies": [{"developer": true, "company": {"name": "Supergiant"}}]}`},
	{"game", record.SourceSteam, `{"appid": 1145360, "name": "Hades", "playtime_forever": 900, "rtime_last_played": 1700000000}`},
	{"game", record.SourceSteamStore, `{"name": "Hades", "header_image": "https://x.test/h.jpg", "developers": ["Supergiant"], "genres": [{"description": "Action"}], "categories": [{"description": "Single-player"}], "release_date": {"date": "17 Sep, 2020"}}`},
	{"game", record.SourceSteamAchievements, `{"achievements": [{"achieved": 1}, {"achieved": 0}]}`},
	{"game", record.SourceSteamTake, `{"recommendation": "recommended", "review": "good", "reviewedOn": "2024-01-02", "hoursAtReview": 12.5}`},
	{"game", record.SourceAP, `{"title": "Hades II", "playtime_minutes": 300, "genres": ["A", "B"], "steam_appid": 12, "platform": null}`},
	{"movie", record.SourceTMDB, `{"id": 603, "title": "The Matrix", "original_title": "The Matrix", "release_date": "1999-03-30", "poster_path": "/p.jpg", "backdrop_path": "/b.jpg", "runtime": 136, "vote_average": 8.2, "vote_count": 100, "genres": [{"name": "Action"}], "production_countries": [{"name": "USA"}], "credits": {"crew": [{"job": "Director", "name": "Wachowski"}]}}`},
	{"movie", record.SourceLetterboxd, `{"id": "abc", "slug": "the-matrix", "title": "The Matrix", "year": 1999, "tmdbId": 603}`},
	{"movie", record.SourceLetterboxdTake, `{"rating": 4.5, "reaction": "liked", "review": "r", "lastWatched": "2024-05-06", "onWatchlistSince": "2023-01-01"}`},
	{"series", record.SourceTMDBSeries, `{"id": 1, "name": "Show", "first_air_date": "2020-01-01", "number_of_seasons": 3, "number_of_episodes": 30, "vote_average": 7.5, "vote_count": 5, "genres": [{"name": "Drama"}]}`},
	{"artist", record.SourceSpotifyArtist, `{"id": "a1", "name": "Daft Punk", "genres": null, "popularity": 80, "followers": {"total": 9000000}, "external_urls": {"spotify": "https://open.spotify.com/artist/a1"}, "images": []}`},
	{"album", record.SourceSpotifyAlbum, `{"id": "b1", "name": "Discovery", "artists": [{"name": "Daft Punk"}], "release_date": "2001-03-12", "album_type": "album", "total_tracks": 14, "label": "Virgin", "genres": ["house"], "popularity": 70, "external_urls": {"spotify": "https://open.spotify.com/album/b1"}, "images": []}`},
}

func TestAStoredReadingDecidesLikeTheClaim(t *testing.T) {
	byKind := map[string][]reading.Claim{}
	for _, f := range claimFixtures {
		byKind[f.kind] = append(byKind[f.kind], reading.Claim{Source: f.source, ExternalID: "x", Data: json.RawMessage(f.data), FetchedAt: "2026-01-01T00:00:00Z"})
	}
	for name, claims := range byKind {
		k, known := kind.Get(name)
		if !known {
			t.Fatalf("no kind %s", name)
		}
		raw := verdict.ReadAll(k, claims)
		if len(raw) != len(claims) {
			t.Fatalf("%s: %d of %d claims readable", name, len(raw), len(claims))
		}
		stored := make([]reading.Reading, 0, len(raw))
		for _, r := range raw {
			encoded, err := reading.Encode(r)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := verdict.Decode(k, r.Source, r.ExternalID, r.FetchedAt, encoded)
			if err != nil {
				t.Fatal(err)
			}
			stored = append(stored, decoded)
		}
		rawVerdicts, storedVerdicts := verdict.Decide(k, raw), verdict.Decide(k, stored)
		if !reflect.DeepEqual(rawVerdicts, storedVerdicts) {
			t.Errorf("%s verdicts differ:\nraw    %#v\nstored %#v", name, rawVerdicts, storedVerdicts)
		}
		if !reflect.DeepEqual(k.Hint(rawVerdicts.Values), k.Hint(storedVerdicts.Values)) {
			t.Errorf("%s hints differ", name)
		}
		if !reflect.DeepEqual(verdict.Takes(k, raw), verdict.Takes(k, stored)) {
			t.Errorf("%s takes differ", name)
		}
		for _, r := range raw {
			for field := range r.Values {
				rawNumber, rawOK := verdict.Number(k, raw, r.Source, field)
				storedNumber, storedOK := verdict.Number(k, stored, r.Source, field)
				if rawNumber != storedNumber || rawOK != storedOK {
					t.Errorf("%s %s.%s number differs", name, r.Source, field)
				}
			}
		}
	}
}
