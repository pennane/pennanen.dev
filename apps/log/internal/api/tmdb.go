package api

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
	"github.com/pennane/pennanen.dev/apps/log/internal/source/tmdb"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

type TMDBResult struct {
	TMDBID   int64            `json:"tmdbId"`
	TMDBType reading.TMDBType `json:"tmdbType"`
	Name     string           `json:"name"`
	Year     string           `json:"year,omitempty"`
	Poster   string           `json:"poster,omitempty"`
	ItemID   int64            `json:"itemId,omitempty"`
}

func (s *Server) tmdbConfigured(w http.ResponseWriter) bool {
	if s.TMDB == nil {
		writeError(w, http.StatusNotImplemented, "tmdb is not configured")
		return false
	}
	return true
}

func readTMDBLink(w http.ResponseWriter, r *http.Request) (reading.TMDBLink, bool) {
	var link reading.TMDBLink
	if !readJSON(w, r, &link) {
		return link, false
	}
	if link.Type == "" {
		link.Type = reading.MovieOnTMDB
	}
	if !slices.Contains(reading.TMDBTypes, link.Type) {
		writeError(w, http.StatusUnprocessableEntity, fmt.Sprintf("tmdb has no type %q", link.Type))
		return link, false
	}
	return link, true
}

func notOnTMDB(err error, link reading.TMDBLink) error {
	if errors.Is(err, tmdb.ErrNoSuchEntry) {
		return fmt.Errorf("%w: tmdb has no %s %d", store.ErrInvalid, link.Type, link.ID)
	}
	return err
}

func (s *Server) searchTMDB(w http.ResponseWriter, r *http.Request) {
	if !s.tmdbConfigured(w) {
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	tmdbType := reading.TMDBType(r.URL.Query().Get("type"))
	if !slices.Contains(reading.TMDBTypes, tmdbType) {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("tmdb has no type %q", tmdbType))
		return
	}
	if query == "" {
		writeJSON(w, http.StatusOK, []TMDBResult{})
		return
	}
	search := func() ([]tmdb.Found, error) { return s.TMDB.SearchMovies(r.Context(), query, 0) }
	if tmdbType == reading.SeriesOnTMDB {
		search = func() ([]tmdb.Found, error) { return s.TMDB.SearchSeries(r.Context(), query) }
	}
	found, err := search()
	if err != nil {
		respond(w, 0, nil, err)
		return
	}
	results := make([]TMDBResult, 0, len(found))
	for _, f := range found {
		result := TMDBResult{TMDBID: f.ID, TMDBType: tmdbType, Name: f.Title, Poster: reading.TMDBPosterURL(f.PosterPath)}
		if year := f.Year(); year > 0 {
			result.Year = fmt.Sprint(year)
		}
		link := reading.TMDBLink{Type: tmdbType, ID: f.ID}
		if itemID, err := s.Store.FindByIdentity(r.Context(), link.Type.Source(), link.ExternalID()); err == nil {
			result.ItemID = itemID
		}
		results = append(results, result)
	}
	writeJSON(w, http.StatusOK, results)
}

func (s *Server) addFromTMDB(w http.ResponseWriter, r *http.Request) {
	if !s.tmdbConfigured(w) {
		return
	}
	link, ok := readTMDBLink(w, r)
	if !ok {
		return
	}
	if existing, err := s.Store.FindByIdentity(r.Context(), link.Type.Source(), link.ExternalID()); err == nil {
		view, err := s.Store.GetItem(r.Context(), existing)
		respond(w, http.StatusOK, view, err)
		return
	}
	raw, err := s.TMDB.Entry(r.Context(), link)
	if err != nil {
		respond(w, 0, nil, notOnTMDB(err, link))
		return
	}
	values, err := reading.Read(link.Type.Source(), raw)
	if err != nil {
		respond(w, 0, nil, err)
		return
	}
	title, _ := values[kind.FieldTitle].(string)
	id, err := s.Store.CreateItemFromSource(r.Context(), kind.OfTMDB(link.Type), title, link.Type.Source(), link.ExternalID(), raw)
	if err != nil {
		respond(w, 0, nil, err)
		return
	}
	s.changed()
	view, err := s.Store.GetItem(r.Context(), id)
	respond(w, http.StatusCreated, view, err)
}

func (s *Server) matchTMDB(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok || !s.tmdbConfigured(w) {
		return
	}
	link, ok := readTMDBLink(w, r)
	if !ok {
		return
	}
	_, err := s.Store.GetItem(r.Context(), id)
	if err == nil {
		err = notOnTMDB(tmdb.Identify(r.Context(), s.Store, s.TMDB, id, link), link)
	}
	s.mutatedItem(w, r, id, err)
}

func (s *Server) matchTMDBNone(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	v, err := s.Store.GetItem(r.Context(), id)
	if err == nil {
		k, _ := kind.Get(v.Kind)
		err = s.Store.AnswerNoMatch(r.Context(), id, k.IdentitySource)
	}
	s.mutatedItem(w, r, id, err)
}
