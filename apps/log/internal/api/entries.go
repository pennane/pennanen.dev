package api

import (
	"net/http"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
)

type listedEntry struct {
	ID         int64             `json:"id"`
	Title      string            `json:"title"`
	Cover      string            `json:"cover,omitempty"`
	Status     string            `json:"status"`
	Visibility store.Visibility  `json:"visibility"`
	Scores     map[string]string `json:"scores"`
	Rank       int               `json:"rank,omitempty"`
	Tags       []string          `json:"tags"`
}

func (s *Server) listEntries(w http.ResponseWriter, r *http.Request) {
	views, err := s.Store.ListItems(r.Context(), store.ItemFilter{Kind: r.URL.Query().Get("kind")})
	entries := []listedEntry{}
	for _, v := range views {
		if isListed(v.Entry) {
			entries = append(entries, toListedEntry(v))
		}
	}
	respond(w, http.StatusOK, entries, err)
}

func isListed(e *store.Entry) bool {
	return e != nil && e.Status != kind.Untriaged && e.Status != kind.Dismissed
}

func toListedEntry(v store.ItemView) listedEntry {
	title := verdict.Title(v.Verdicts)
	if title == "" {
		title = v.Slug
	}
	return listedEntry{
		ID:         v.ID,
		Title:      title,
		Cover:      coverOf(v),
		Status:     v.Entry.Status,
		Visibility: v.Entry.Visibility,
		Scores:     v.Entry.Scores,
		Rank:       v.Entry.Rank,
		Tags:       v.Entry.Tags,
	}
}

var coverFields = []string{"cover", "image"}

func coverOf(v store.ItemView) string {
	for _, field := range coverFields {
		if url, isString := v.Verdicts.Values[field].(string); isString && url != "" {
			return url
		}
	}
	return ""
}

func (s *Server) placeEntry(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var body struct {
		Division string `json:"division"`
		Index    int    `json:"index"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	err := s.Store.Place(r.Context(), id, body.Division, body.Index)
	if err == nil {
		s.changed()
	}
	respond(w, http.StatusNoContent, nil, err)
}
