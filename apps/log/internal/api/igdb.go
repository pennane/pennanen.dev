package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/pennane/pennanen.dev/apps/log/internal/igdb"
	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
)

const igdbKind = "game"

type IGDBResult struct {
	IGDBID    int64    `json:"igdbId"`
	Name      string   `json:"name"`
	Year      string   `json:"year,omitempty"`
	Platforms []string `json:"platforms"`
	Poster    string   `json:"poster,omitempty"`
	ItemID    int64    `json:"itemId,omitempty"`
}

func (s *Server) igdbConfigured(w http.ResponseWriter) bool {
	if s.IGDB == nil {
		writeError(w, http.StatusNotImplemented, "igdb is not configured")
		return false
	}
	return true
}

func (s *Server) searchIGDB(w http.ResponseWriter, r *http.Request) {
	if !s.igdbConfigured(w) {
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		writeJSON(w, http.StatusOK, []IGDBResult{})
		return
	}
	games, err := s.IGDB.Search(r.Context(), query)
	if err != nil {
		respond(w, 0, nil, err)
		return
	}
	results := make([]IGDBResult, 0, len(games))
	for _, raw := range games {
		var g verdict.IGDBGame
		if json.Unmarshal(raw, &g) != nil || !g.Matched() {
			continue
		}
		result := IGDBResult{IGDBID: g.ID, Name: g.Name, Year: g.Year(), Platforms: g.PlatformNames()}
		if g.Cover != nil {
			result.Poster = verdict.IGDBPosterURL(*g.Cover)
		}
		if itemID, err := s.Store.FindByIdentity(r.Context(), kind.SourceIGDB, strconv.FormatInt(g.ID, 10)); err == nil {
			result.ItemID = itemID
		}
		results = append(results, result)
	}
	writeJSON(w, http.StatusOK, results)
}

type igdbPick struct {
	IGDBID int64 `json:"igdbId"`
}

func (s *Server) addFromIGDB(w http.ResponseWriter, r *http.Request) {
	var pick igdbPick
	if !s.igdbConfigured(w) || !readJSON(w, r, &pick) {
		return
	}
	externalID := strconv.FormatInt(pick.IGDBID, 10)
	if existing, err := s.Store.FindByIdentity(r.Context(), kind.SourceIGDB, externalID); err == nil {
		view, err := s.Store.GetItem(r.Context(), existing)
		respond(w, http.StatusOK, view, err)
		return
	}
	game, raw, err := s.fetchIGDBGame(r.Context(), pick.IGDBID)
	if err != nil {
		respond(w, 0, nil, err)
		return
	}
	id, err := s.Store.CreateItemFromSource(r.Context(), igdbKind, game.Name, kind.SourceIGDB, externalID, raw)
	if err != nil {
		respond(w, 0, nil, err)
		return
	}
	s.changed()
	view, err := s.Store.GetItem(r.Context(), id)
	respond(w, http.StatusCreated, view, err)
}

func (s *Server) matchIGDB(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	var pick igdbPick
	if !ok || !s.igdbConfigured(w) || !readJSON(w, r, &pick) {
		return
	}
	_, err := s.Store.GetItem(r.Context(), id)
	var raw json.RawMessage
	if err == nil {
		_, raw, err = s.fetchIGDBGame(r.Context(), pick.IGDBID)
	}
	if err == nil {
		err = s.Store.PutClaim(r.Context(), id, kind.SourceIGDB, strconv.FormatInt(pick.IGDBID, 10), raw)
	}
	s.mutatedItem(w, r, id, err)
}

func (s *Server) matchIGDBNone(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	_, err := s.Store.GetItem(r.Context(), id)
	if err == nil {
		err = s.Store.PutClaim(r.Context(), id, kind.SourceIGDB, "", igdb.NoMatch)
	}
	s.mutatedItem(w, r, id, err)
}

func (s *Server) fetchIGDBGame(ctx context.Context, igdbID int64) (verdict.IGDBGame, json.RawMessage, error) {
	games, err := s.IGDB.Games(ctx, []int64{igdbID})
	if err != nil {
		return verdict.IGDBGame{}, nil, err
	}
	raw, found := games[igdbID]
	if !found {
		return verdict.IGDBGame{}, nil, fmt.Errorf("%w: igdb has no game %d", store.ErrNotFound, igdbID)
	}
	var game verdict.IGDBGame
	return game, raw, json.Unmarshal(raw, &game)
}

func (s *Server) igdbMatchStatus(w http.ResponseWriter, r *http.Request) {
	if s.IGDBMatch == nil {
		s.igdbConfigured(w)
		return
	}
	writeJSON(w, http.StatusOK, s.IGDBMatch.Last())
}

func (s *Server) startIGDBMatch(w http.ResponseWriter, r *http.Request) {
	if s.IGDBMatch == nil {
		s.igdbConfigured(w)
		return
	}
	if err := s.IGDBMatch.Start(context.WithoutCancel(r.Context())); errors.Is(err, igdb.ErrAlreadyMatching) {
		writeJSON(w, http.StatusConflict, s.IGDBMatch.Last())
		return
	}
	writeJSON(w, http.StatusAccepted, s.IGDBMatch.Last())
}
