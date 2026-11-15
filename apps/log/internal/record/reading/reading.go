package reading

import (
	"encoding/json"
	"fmt"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
)

type Claim struct {
	Source     string          `json:"source"`
	ExternalID string          `json:"externalId,omitempty"`
	Data       json.RawMessage `json:"data"`
	FetchedAt  string          `json:"fetchedAt"`
}

type readClaim func(raw json.RawMessage) (map[string]any, error)

var claimReaders = map[string]readClaim{
	record.SourceSteam:             readSteamOwnedGame,
	record.SourceSteamStore:        readSteamStoreDetails,
	record.SourceSteamAchievements: readSteamAchievements,
	record.SourceIGDB:              readIGDBGame,
	record.SourceSteamTake:         readSteamTake,
	record.SourceLetterboxdTake:    readLetterboxdTake,
	record.SourceLetterboxd:        readLetterboxdFilm,
	record.SourceTMDB:              readTMDBMovie,
	record.SourceTMDBSeries:        readTMDBSeries,
	record.SourceSpotifyArtist:     readSpotifyArtist,
	record.SourceSpotifyAlbum:      readSpotifyAlbum,
	record.SourceSpotifyTrack:      readSpotifyTrack,
	record.SourceSpotifyListening:  readSpotifyListening,
}

func Reads(source string) bool {
	_, known := claimReaders[source]
	return known
}

func Read(source string, raw json.RawMessage) (map[string]any, error) {
	read, known := claimReaders[source]
	if !known {
		return nil, fmt.Errorf("no reader for %s", source)
	}
	return read(raw)
}

const ReadersVersion = 4

type Reading struct {
	Source     string
	ExternalID string
	FetchedAt  string
	Values     map[string]any
}

type Readability int

const (
	Unreadable Readability = iota
	Readable
)

func ReadClaim(c Claim) (Reading, Readability) {
	values, err := readValues(c)
	if err != nil || values == nil {
		return Reading{}, Unreadable
	}
	return Reading{Source: c.Source, ExternalID: c.ExternalID, FetchedAt: c.FetchedAt, Values: values}, Readable
}

func Encode(r Reading) (json.RawMessage, error) {
	return json.Marshal(r.Values)
}

func readValues(c Claim) (map[string]any, error) {
	if c.Source == record.SourceAP {
		var override map[string]any
		err := json.Unmarshal(c.Data, &override)
		return override, err
	}
	read, known := claimReaders[c.Source]
	if !known {
		return nil, nil
	}
	return read(c.Data)
}
