package api

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/steam"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
)

type Server struct {
	Store      *store.Store
	AdminToken string
	Steam      *steam.Syncer
	Backup     func(ctx context.Context) (string, error)
	Changed    func()
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /public/items", s.publicItems)
	mux.HandleFunc("GET /public/items/{kind}/{slug}", s.publicItem)

	admin := http.NewServeMux()
	admin.HandleFunc("GET /admin/kinds", s.kinds)
	admin.HandleFunc("GET /admin/items", s.listItems)
	admin.HandleFunc("POST /admin/items", s.createItem)
	admin.HandleFunc("GET /admin/items/{id}", s.getItem)
	admin.HandleFunc("DELETE /admin/items/{id}", s.deleteItem)
	admin.HandleFunc("PATCH /admin/items/{id}/fields", s.changeFields)
	admin.HandleFunc("PATCH /admin/items/{id}/entry", s.changeEntry)
	admin.HandleFunc("POST /admin/entries/change", s.changeEntries)
	admin.HandleFunc("POST /admin/items/{id}/moments", s.addMoment)
	admin.HandleFunc("DELETE /admin/moments/{id}", s.deleteMoment)
	admin.HandleFunc("POST /admin/relations", s.addRelation)
	admin.HandleFunc("DELETE /admin/relations", s.removeRelation)
	admin.HandleFunc("GET /admin/sync/steam", s.steamSyncStatus)
	admin.HandleFunc("POST /admin/sync/steam", s.startSteamSync)
	admin.HandleFunc("POST /admin/backup", s.backupNow)
	mux.Handle("/admin/", s.requireAdmin(admin))

	return mux
}

func (s *Server) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		presented := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if s.AdminToken == "" || subtle.ConstantTimeCompare([]byte(presented), []byte(s.AdminToken)) != 1 {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) kinds(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, kind.All())
}

func (s *Server) listItems(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	views, err := s.Store.ListItems(r.Context(), store.ItemFilter{
		Kind:       q.Get("kind"),
		Status:     q.Get("status"),
		Visibility: store.Visibility(q.Get("visibility")),
		Hint:       kind.Hint(q.Get("hint")),
		Tag:        store.NormalizeTag(q.Get("tag")),
		Query:      q.Get("q"),
	})
	respond(w, http.StatusOK, orEmpty(views), err)
}

func (s *Server) createItem(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Kind   string         `json:"kind"`
		Title  string         `json:"title"`
		Fields map[string]any `json:"fields"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	id, err := s.Store.CreateItem(r.Context(), body.Kind, body.Title, body.Fields)
	if err != nil {
		respond(w, 0, nil, err)
		return
	}
	s.changed()
	view, err := s.Store.GetItem(r.Context(), id)
	respond(w, http.StatusCreated, view, err)
}

func (s *Server) getItem(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	view, err := s.Store.GetItem(r.Context(), id)
	respond(w, http.StatusOK, view, err)
}

func (s *Server) deleteItem(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	s.mutated(w, s.Store.DeleteItem(r.Context(), id))
}

func (s *Server) changeFields(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	var changes map[string]any
	if !ok || !readJSON(w, r, &changes) {
		return
	}
	s.mutatedItem(w, r, id, s.Store.ChangeOverride(r.Context(), id, changes))
}

func (s *Server) changeEntry(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	var change store.EntryChange
	if !ok || !readJSON(w, r, &change) {
		return
	}
	s.mutatedItem(w, r, id, s.Store.ChangeEntry(r.Context(), id, change))
}

func (s *Server) changeEntries(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ItemIDs []int64           `json:"itemIds"`
		Change  store.EntryChange `json:"change"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	s.mutated(w, s.Store.ChangeEntries(r.Context(), body.ItemIDs, body.Change))
}

func (s *Server) addMoment(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	var m store.NewMoment
	if !ok || !readJSON(w, r, &m) {
		return
	}
	_, err := s.Store.AddMoment(r.Context(), id, m)
	s.mutatedItem(w, r, id, err)
}

func (s *Server) deleteMoment(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	s.mutated(w, s.Store.DeleteMoment(r.Context(), id))
}

func (s *Server) addRelation(w http.ResponseWriter, r *http.Request) {
	var rel store.Relation
	if !readJSON(w, r, &rel) {
		return
	}
	s.mutated(w, s.Store.AddRelation(r.Context(), rel))
}

func (s *Server) removeRelation(w http.ResponseWriter, r *http.Request) {
	var rel store.Relation
	if !readJSON(w, r, &rel) {
		return
	}
	s.mutated(w, s.Store.RemoveRelation(r.Context(), rel))
}

func (s *Server) steamSyncStatus(w http.ResponseWriter, r *http.Request) {
	if s.Steam == nil {
		writeError(w, http.StatusNotImplemented, "steam is not configured")
		return
	}
	writeJSON(w, http.StatusOK, s.Steam.Last())
}

func (s *Server) startSteamSync(w http.ResponseWriter, r *http.Request) {
	if s.Steam == nil {
		writeError(w, http.StatusNotImplemented, "steam is not configured")
		return
	}
	if err := s.Steam.Start(context.WithoutCancel(r.Context())); errors.Is(err, steam.ErrAlreadySyncing) {
		writeJSON(w, http.StatusConflict, s.Steam.Last())
		return
	}
	writeJSON(w, http.StatusAccepted, s.Steam.Last())
}

func (s *Server) backupNow(w http.ResponseWriter, r *http.Request) {
	if s.Backup == nil {
		writeError(w, http.StatusNotImplemented, "backups are not configured")
		return
	}
	key, err := s.Backup(r.Context())
	respond(w, http.StatusOK, map[string]string{"key": key}, err)
}

func (s *Server) changed() {
	if s.Changed != nil {
		s.Changed()
	}
}

func (s *Server) mutated(w http.ResponseWriter, err error) {
	if err == nil {
		s.changed()
	}
	respond(w, http.StatusNoContent, nil, err)
}

func (s *Server) mutatedItem(w http.ResponseWriter, r *http.Request, id int64, err error) {
	if err != nil {
		respond(w, 0, nil, err)
		return
	}
	s.changed()
	view, err := s.Store.GetItem(r.Context(), id)
	respond(w, http.StatusOK, view, err)
}

func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id must be a number")
		return 0, false
	}
	return id, true
}

const maxBodyBytes = 1 << 20

func readJSON(w http.ResponseWriter, r *http.Request, into any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(into); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return false
	}
	return true
}

func respond(w http.ResponseWriter, status int, body any, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, store.ErrInvalid):
		writeError(w, http.StatusUnprocessableEntity, err.Error())
	case err != nil:
		slog.Error("request failed", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	case status == http.StatusNoContent:
		w.WriteHeader(status)
	default:
		writeJSON(w, status, body)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func orEmpty[T any](xs []T) []T {
	if xs == nil {
		return []T{}
	}
	return xs
}
