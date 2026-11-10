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

	"github.com/pennane/pennanen.dev/apps/log/internal/igdb"
	"github.com/pennane/pennanen.dev/apps/log/internal/images"
	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/steam"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
)

type Server struct {
	Store      *store.Store
	AdminToken string
	Steam      *steam.Syncer
	IGDB       *igdb.Client
	IGDBMatch  *igdb.Matcher
	Backup     func(ctx context.Context) (string, error)
	Images     *images.Store
	Changed    func()
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /public/items", s.publicItems)
	mux.HandleFunc("GET /public/items/{kind}/{slug}", s.publicItem)
	mux.HandleFunc("GET /public/kinds/{kind}", s.publicKind)
	mux.HandleFunc("GET /public/presentation", s.publicPresentation)
	mux.HandleFunc("GET /images/{name}", s.serveImage)
	mux.HandleFunc("OPTIONS /images", s.uploadPreflight)
	mux.HandleFunc("POST /images", s.uploadImage)

	admin := http.NewServeMux()
	admin.HandleFunc("GET /admin/kinds", s.kinds)
	admin.HandleFunc("GET /admin/settings", s.publicPresentation)
	admin.HandleFunc("PUT /admin/settings/axes", s.putAxes)
	admin.HandleFunc("PUT /admin/settings/kinds/{kind}", s.putKindSettings)
	admin.HandleFunc("PUT /admin/settings/tag-titles", s.putTagTitles)
	admin.HandleFunc("GET /admin/items", s.listItems)
	admin.HandleFunc("POST /admin/items", s.createItem)
	admin.HandleFunc("GET /admin/items/{id}", s.getItem)
	admin.HandleFunc("DELETE /admin/items/{id}", s.deleteItem)
	admin.HandleFunc("PATCH /admin/items/{id}/fields", s.changeFields)
	admin.HandleFunc("PATCH /admin/items/{id}/entry", s.changeEntry)
	admin.HandleFunc("PUT /admin/items/{id}/place", s.placeEntry)
	admin.HandleFunc("GET /admin/entries", s.listEntries)
	admin.HandleFunc("POST /admin/entries/change", s.changeEntries)
	admin.HandleFunc("POST /admin/relations", s.addRelation)
	admin.HandleFunc("DELETE /admin/relations", s.removeRelation)
	admin.HandleFunc("GET /admin/igdb/search", s.searchIGDB)
	admin.HandleFunc("POST /admin/items/from-igdb", s.addFromIGDB)
	admin.HandleFunc("PUT /admin/items/{id}/igdb", s.matchIGDB)
	admin.HandleFunc("PUT /admin/items/{id}/igdb/none", s.matchIGDBNone)
	admin.HandleFunc("GET /admin/sync/igdb", s.igdbMatchStatus)
	admin.HandleFunc("POST /admin/sync/igdb", s.startIGDBMatch)
	admin.HandleFunc("GET /admin/sync/steam", s.steamSyncStatus)
	admin.HandleFunc("POST /admin/sync/steam", s.startSteamSync)
	admin.HandleFunc("POST /admin/backup", s.backupNow)
	admin.HandleFunc("POST /admin/images/ticket", s.uploadTicket)
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

type adminKind struct {
	kind.Kind
	Axes store.Axes `json:"axes"`
}

func (s *Server) kinds(w http.ResponseWriter, r *http.Request) {
	axes, err := s.Store.Axes(r.Context())
	if err != nil {
		respond(w, 0, nil, err)
		return
	}
	settings, err := s.Store.AllKindSettings(r.Context())
	if err != nil {
		respond(w, 0, nil, err)
		return
	}
	kinds := []adminKind{}
	for _, k := range kind.All() {
		k.StatusesShown = settings[k.Name].StatusesShown
		used := store.Axes{}
		for _, key := range settings[k.Name].Axes {
			if axis, found := axes.Find(key); found {
				used = append(used, axis)
			}
		}
		kinds = append(kinds, adminKind{Kind: k, Axes: used})
	}
	writeJSON(w, http.StatusOK, kinds)
}

func (s *Server) putAxes(w http.ResponseWriter, r *http.Request) {
	var body store.Axes
	if !readJSON(w, r, &body) {
		return
	}
	axes, err := s.Store.SetAxes(r.Context(), body)
	if err == nil {
		s.changed()
	}
	respond(w, http.StatusOK, axes, err)
}

func (s *Server) putKindSettings(w http.ResponseWriter, r *http.Request) {
	k, known := kind.Get(r.PathValue("kind"))
	if !known {
		respond(w, 0, nil, store.ErrNotFound)
		return
	}
	body, err := s.Store.KindSettings(r.Context(), k)
	if err != nil {
		respond(w, 0, nil, err)
		return
	}
	if !readJSON(w, r, &body) {
		return
	}
	settings, err := s.Store.SetKindSettings(r.Context(), k, body)
	if err == nil {
		s.changed()
	}
	respond(w, http.StatusOK, settings, err)
}

func (s *Server) putTagTitles(w http.ResponseWriter, r *http.Request) {
	var body store.TagTitles
	if !readJSON(w, r, &body) {
		return
	}
	titles, err := s.Store.SetTagTitles(r.Context(), body)
	if err == nil {
		s.changed()
	}
	respond(w, http.StatusOK, titles, err)
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
		Needs:      store.Need(q.Get("needs")),
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
	case errors.Is(err, store.ErrConflict):
		writeError(w, http.StatusConflict, err.Error())
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
