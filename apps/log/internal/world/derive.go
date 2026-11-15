package world

import (
	"encoding/json"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
	"strings"
)

var DeriveKeys = []string{"about", "kind", "links", "suppressed", "artist_ids", "track_ids", "tmdb_link"}

type AssertedLink struct {
	From     record.Identity
	Type     string
	To       record.Identity
	Position int
}

type Derivation struct {
	About      []record.Identity
	Kind       string
	Links      []AssertedLink
	Suppressed []AssertedLink
}

const (
	apItemPrefix   = "item/"
	apFieldsPrefix = "fields/"
	apMatchPrefix  = "match/"
	apLinkPrefix   = "link/"

	apEditAdded      = "added"
	apEditSuppressed = "suppressed"
)

type linkJSON struct {
	From     string `json:"from"`
	Type     string `json:"type"`
	To       string `json:"to"`
	Position int    `json:"position,omitempty"`
}

func Derive(id record.Identity, reading map[string]any) Derivation {
	switch id.Source {
	case record.SourceSteam, record.SourceSteamStore, record.SourceSteamAchievements, record.SourceSteamTake:
		return Derivation{About: identitiesOf(record.SourceSteam, id.ExternalID), Kind: kind.Game}
	case record.SourceIGDB:
		return Derivation{Kind: kind.Game}
	case record.SourceTMDB:
		return Derivation{Kind: kind.Movie}
	case record.SourceTMDBSeries:
		return Derivation{Kind: kind.Series}
	case record.SourceLetterboxd:
		return deriveLetterboxd(reading)
	case record.SourceLetterboxdTake, record.SourceSpotifyListening:
		return Derivation{About: parseAll(id.ExternalID)}
	case record.SourceSpotifyArtist:
		return Derivation{Kind: kind.Artist}
	case record.SourceSpotifyAlbum:
		return Derivation{Kind: kind.Album, Links: albumLinks(id, reading)}
	case record.SourceSpotifyTrack:
		spotifyID, _ := reading["spotify_id"].(string)
		return Derivation{Kind: kind.Track, About: identitiesOf(record.SourceSpotifyTrackID, spotifyID)}
	case record.SourceAP:
		return deriveAp(id, reading)
	case record.SourceMatched, record.SourceLegacy:
		return Derivation{
			About:      parseAll(stringsOf(reading["about"])...),
			Kind:       stringOf(reading["kind"]),
			Links:      linksOf(reading["links"]),
			Suppressed: linksOf(reading["suppressed"]),
		}
	}
	return Derivation{}
}

func deriveLetterboxd(reading map[string]any) Derivation {
	linked := parseAll(stringOf(reading["tmdb_link"]))
	if len(linked) == 0 {
		return Derivation{Kind: kind.Movie}
	}
	derived := Derivation{About: linked, Kind: kind.Movie}
	if linked[0].Source == record.SourceTMDBSeries {
		derived.Kind = kind.Series
	}
	return derived
}

func albumLinks(album record.Identity, reading map[string]any) []AssertedLink {
	links := []AssertedLink{}
	for i, artistID := range stringsOf(reading["artist_ids"]) {
		linkType := kind.LinkWith
		if i == 0 && artistID != kind.VariousArtists {
			linkType = kind.LinkBy
		}
		links = append(links, AssertedLink{From: album, Type: linkType, To: record.Identity{Source: record.SourceSpotifyArtist, ExternalID: artistID}})
	}
	for i, trackID := range stringsOf(reading["track_ids"]) {
		links = append(links, AssertedLink{
			From:     record.Identity{Source: record.SourceSpotifyTrackID, ExternalID: trackID},
			Type:     kind.LinkPartOf,
			To:       album,
			Position: firstPosition + i,
		})
	}
	return links
}

func deriveAp(id record.Identity, reading map[string]any) Derivation {
	switch {
	case strings.HasPrefix(id.ExternalID, apItemPrefix):
		return Derivation{Kind: stringOf(reading["kind"])}
	case strings.HasPrefix(id.ExternalID, apFieldsPrefix):
		return Derivation{About: parseAll(strings.TrimPrefix(id.ExternalID, apFieldsPrefix))}
	case strings.HasPrefix(id.ExternalID, apMatchPrefix):
		return deriveApMatch(strings.TrimPrefix(id.ExternalID, apMatchPrefix), reading)
	case strings.HasPrefix(id.ExternalID, apLinkPrefix):
		return deriveApLink(reading)
	}
	return Derivation{}
}

func deriveApMatch(rest string, reading map[string]any) Derivation {
	matchSource, about, found := strings.Cut(rest, "/")
	if !found {
		return Derivation{}
	}
	matched := parseAll(about)
	if to := stringOf(reading["to"]); to != "" {
		matched = append(matched, record.Identity{Source: matchSource, ExternalID: to})
	}
	return Derivation{About: matched}
}

func deriveApLink(reading map[string]any) Derivation {
	links := linksOf([]any{reading})
	switch stringOf(reading["edit"]) {
	case apEditAdded:
		return Derivation{Links: links}
	case apEditSuppressed:
		return Derivation{Suppressed: links}
	}
	return Derivation{}
}

func linksOf(value any) []AssertedLink {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	var given []linkJSON
	if err := json.Unmarshal(encoded, &given); err != nil {
		return nil
	}
	links := []AssertedLink{}
	for _, l := range given {
		from, fromErr := record.ParseIdentity(l.From)
		to, toErr := record.ParseIdentity(l.To)
		if fromErr != nil || toErr != nil {
			continue
		}
		links = append(links, AssertedLink{From: from, Type: l.Type, To: to, Position: l.Position})
	}
	return links
}

func parseAll(identities ...string) []record.Identity {
	parsed := []record.Identity{}
	for _, s := range identities {
		if id, err := record.ParseIdentity(s); err == nil {
			parsed = append(parsed, id)
		}
	}
	return parsed
}

func identitiesOf(source, externalID string) []record.Identity {
	if externalID == "" {
		return []record.Identity{}
	}
	return []record.Identity{{Source: source, ExternalID: externalID}}
}

func stringOf(value any) string {
	s, _ := value.(string)
	return s
}

func stringsOf(value any) []string {
	items, _ := value.([]any)
	strs := make([]string, 0, len(items))
	for _, item := range items {
		if s, isString := item.(string); isString {
			strs = append(strs, s)
		}
	}
	return strs
}
