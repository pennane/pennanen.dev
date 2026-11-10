package api

import (
	"net/http"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
)

type listedEntry struct {
	ID         int64            `json:"id"`
	Title      string           `json:"title"`
	Cover      string           `json:"cover,omitempty"`
	Status     string           `json:"status"`
	Visibility store.Visibility `json:"visibility"`
	Scores     map[string]int   `json:"scores"`
	Tags       []string         `json:"tags"`
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
		Tags:       v.Entry.Tags,
	}
}
