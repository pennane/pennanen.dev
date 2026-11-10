package api

import (
	"net/http"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
)

type AdminRanking struct {
	store.Ranking
	Items []store.ItemView `json:"items"`
}

type PublicRankingSummary struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	UpdatedAt   string `json:"updatedAt"`
}

type PublicRanking struct {
	PublicRankingSummary
	Order store.Order[PublicPlacement] `json:"order"`
}

type PublicPlacement struct {
	Kind  string      `json:"kind"`
	Title string      `json:"title"`
	Cover string      `json:"cover,omitempty"`
	Note  string      `json:"note"`
	Page  *PublicPage `json:"page,omitempty"`
}

type PublicPage struct {
	Kind string `json:"kind"`
	Slug string `json:"slug"`
}

func (s *Server) listRankings(w http.ResponseWriter, r *http.Request) {
	rankings, err := s.Store.ListRankings(r.Context())
	respond(w, http.StatusOK, rankings, err)
}

func (s *Server) createRanking(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	id, err := s.Store.CreateRanking(r.Context(), body.Name, body.Description)
	if err != nil {
		respond(w, 0, nil, err)
		return
	}
	s.changed()
	ranking, err := s.adminRanking(r, id)
	respond(w, http.StatusCreated, ranking, err)
}

func (s *Server) getRanking(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	ranking, err := s.adminRanking(r, id)
	respond(w, http.StatusOK, ranking, err)
}

func (s *Server) changeRanking(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	var change store.RankingChange
	if !ok || !readJSON(w, r, &change) {
		return
	}
	s.mutatedRanking(w, r, id, s.Store.ChangeRanking(r.Context(), id, change))
}

func (s *Server) putRankingOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	var order store.RankingOrder
	if !ok || !readJSON(w, r, &order) {
		return
	}
	s.mutatedRanking(w, r, id, s.Store.PutRankingOrder(r.Context(), id, order))
}

func (s *Server) deleteRanking(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	s.mutated(w, s.Store.DeleteRanking(r.Context(), id))
}

func (s *Server) mutatedRanking(w http.ResponseWriter, r *http.Request, id int64, err error) {
	if err != nil {
		respond(w, 0, nil, err)
		return
	}
	s.changed()
	ranking, err := s.adminRanking(r, id)
	respond(w, http.StatusOK, ranking, err)
}

func (s *Server) adminRanking(r *http.Request, id int64) (AdminRanking, error) {
	ranking, err := s.Store.GetRanking(r.Context(), id)
	if err != nil {
		return AdminRanking{}, err
	}
	items, err := s.Store.GetItems(r.Context(), placedItemIDs(ranking))
	return AdminRanking{Ranking: ranking, Items: orEmpty(items)}, err
}

func (s *Server) publicRankings(w http.ResponseWriter, r *http.Request) {
	rankings, err := s.Store.ListRankings(r.Context())
	if err != nil {
		respond(w, 0, nil, err)
		return
	}
	summaries := make([]PublicRankingSummary, 0, len(rankings))
	for _, ranking := range rankings {
		summaries = append(summaries, PublicRankingSummary{ranking.Slug, ranking.Name, ranking.Description, ranking.UpdatedAt})
	}
	writeJSON(w, http.StatusOK, summaries)
}

func (s *Server) publicRanking(w http.ResponseWriter, r *http.Request) {
	ranking, err := s.Store.GetRankingBySlug(r.Context(), r.PathValue("slug"))
	if err != nil {
		respond(w, 0, nil, err)
		return
	}
	isPublic, err := s.publicRule(r)
	if err != nil {
		respond(w, 0, nil, err)
		return
	}
	views, err := s.Store.GetItems(r.Context(), placedItemIDs(ranking))
	if err != nil {
		respond(w, 0, nil, err)
		return
	}
	byID := map[int64]store.ItemView{}
	for _, v := range views {
		byID[v.ID] = v
	}
	order := store.MapOrder(ranking.Order, func(p store.Placement) (PublicPlacement, bool) {
		v, found := byID[p.ItemID]
		if !found || isDismissed(v) {
			return PublicPlacement{}, false
		}
		placement := PublicPlacement{Kind: v.Kind, Title: verdict.Title(v.Verdicts), Cover: coverOf(v), Note: p.Note}
		if isPublic(v) {
			placement.Page = &PublicPage{Kind: v.Kind, Slug: v.Slug}
		}
		return placement, true
	})
	writeJSON(w, http.StatusOK, PublicRanking{
		PublicRankingSummary: PublicRankingSummary{ranking.Slug, ranking.Name, ranking.Description, ranking.UpdatedAt},
		Order:                order,
	})
}

func placedItemIDs(ranking store.Ranking) []int64 {
	placements := ranking.Order.All()
	ids := make([]int64, 0, len(placements))
	for _, p := range placements {
		ids = append(ids, p.ItemID)
	}
	return ids
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

func isDismissed(v store.ItemView) bool {
	return v.Entry != nil && v.Entry.Status == kind.Dismissed
}
