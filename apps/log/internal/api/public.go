package api

import (
	"net/http"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
)

type PublicItem struct {
	Kind      string         `json:"kind"`
	Slug      string         `json:"slug"`
	Fields    map[string]any `json:"fields"`
	Status    string         `json:"status"`
	Scores    map[string]int `json:"scores"`
	Tags      []string       `json:"tags"`
	Blurb     string         `json:"blurb"`
	UpdatedAt string         `json:"updatedAt"`
	Moments   []PublicMoment `json:"moments,omitempty"`
}

type PublicMoment struct {
	At     string         `json:"at"`
	Note   string         `json:"note"`
	Scores map[string]int `json:"scores"`
	Refs   []PublicRef    `json:"refs"`
}

type PublicRef struct {
	Kind  string `json:"kind"`
	Slug  string `json:"slug"`
	Title string `json:"title"`
}

func (s *Server) publicItems(w http.ResponseWriter, r *http.Request) {
	views, err := s.Store.ListItems(r.Context(), store.ItemFilter{Kind: r.URL.Query().Get("kind"), Visibility: store.Public})
	if err != nil {
		respond(w, 0, nil, err)
		return
	}
	items := make([]PublicItem, 0, len(views))
	for _, v := range views {
		if isPublic(v) {
			items = append(items, publicItem(v))
		}
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) publicItem(w http.ResponseWriter, r *http.Request) {
	view, err := s.Store.GetItemBySlug(r.Context(), r.PathValue("kind"), r.PathValue("slug"))
	if err == nil && !isPublic(view) {
		err = store.ErrNotFound
	}
	if err != nil {
		respond(w, 0, nil, err)
		return
	}
	item := publicItem(view)
	refs, err := s.publicRefs(r, view.Moments)
	if err != nil {
		respond(w, 0, nil, err)
		return
	}
	for _, m := range view.Moments {
		if m.Visibility == store.MomentPrivate {
			continue
		}
		pm := PublicMoment{At: m.At, Note: m.Note, Scores: m.Scores, Refs: []PublicRef{}}
		for _, id := range m.Refs {
			if ref, shown := refs[id]; shown {
				pm.Refs = append(pm.Refs, ref)
			}
		}
		item.Moments = append(item.Moments, pm)
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) publicRefs(r *http.Request, moments []store.Moment) (map[int64]PublicRef, error) {
	var ids []int64
	for _, m := range moments {
		ids = append(ids, m.Refs...)
	}
	views, err := s.Store.GetItems(r.Context(), ids)
	if err != nil {
		return nil, err
	}
	refs := map[int64]PublicRef{}
	for _, v := range views {
		if isPublic(v) {
			refs[v.ID] = PublicRef{Kind: v.Kind, Slug: v.Slug, Title: verdict.Title(v.Verdicts)}
		}
	}
	return refs, nil
}

func isPublic(v store.ItemView) bool {
	return v.Entry != nil && v.Entry.Visibility == store.Public && v.Entry.Status != kind.Dismissed
}

func publicItem(v store.ItemView) PublicItem {
	return PublicItem{
		Kind:      v.Kind,
		Slug:      v.Slug,
		Fields:    v.Verdicts.Values,
		Status:    v.Entry.Status,
		Scores:    v.Entry.Scores,
		Tags:      v.Entry.Tags,
		Blurb:     v.Entry.Blurb,
		UpdatedAt: v.Entry.UpdatedAt,
	}
}
