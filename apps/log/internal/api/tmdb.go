package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/tmdb"
)

type tmdbPick struct {
	TMDBID int64 `json:"tmdbId"`
}

func (s *Server) tmdbConfigured(w http.ResponseWriter) bool {
	if s.TMDB == nil {
		writeError(w, http.StatusNotImplemented, "tmdb is not configured")
		return false
	}
	return true
}

func (s *Server) matchTMDB(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	var pick tmdbPick
	if !ok || !s.tmdbConfigured(w) || !readJSON(w, r, &pick) {
		return
	}
	_, err := s.Store.GetItem(r.Context(), id)
	if err == nil {
		var raw []byte
		raw, err = s.TMDB.Movie(r.Context(), pick.TMDBID)
		if errors.Is(err, tmdb.ErrNoSuchMovie) {
			err = fmt.Errorf("%w: tmdb has no movie %d", store.ErrInvalid, pick.TMDBID)
		}
		if err == nil {
			err = s.Store.PutClaim(r.Context(), id, kind.SourceTMDB, strconv.FormatInt(pick.TMDBID, 10), raw)
		}
	}
	s.mutatedItem(w, r, id, err)
}

func (s *Server) matchTMDBNone(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	_, err := s.Store.GetItem(r.Context(), id)
	if err == nil {
		err = s.Store.PutClaim(r.Context(), id, kind.SourceTMDB, "", tmdb.NoMatch)
	}
	s.mutatedItem(w, r, id, err)
}
