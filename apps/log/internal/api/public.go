package api

import (
	"cmp"
	"maps"
	"net/http"
	"slices"
	"strings"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
)

type PublicItem struct {
	Kind       string            `json:"kind"`
	Slug       string            `json:"slug"`
	Fields     map[string]any    `json:"fields"`
	Status     string            `json:"status"`
	Scores     map[string]string `json:"scores"`
	Rank       *PublicRank       `json:"rank,omitempty"`
	Order      int               `json:"order,omitempty"`
	Tags       []string          `json:"tags"`
	Blurb      string            `json:"blurb"`
	TimeSpent  store.TimeSpent   `json:"timeSpent,omitempty"`
	UpdatedAt  string            `json:"updatedAt"`
	Prose      string            `json:"prose,omitempty"`
	ProseWords int               `json:"proseWords"`
	Takes      []verdict.Take    `json:"takes"`
	Parts      *PublicParts      `json:"parts,omitempty"`
}

type PublicParts struct {
	Title string       `json:"title"`
	Items []PublicPart `json:"items"`
}

type PublicPart struct {
	Kind   string            `json:"kind"`
	Fields map[string]any    `json:"fields"`
	Scores map[string]string `json:"scores,omitempty"`
}

type PublicRank struct {
	Place int `json:"place"`
	Of    int `json:"of"`
}

func (s *Server) publicRanks(r *http.Request, kindName string) (map[int64]PublicRank, error) {
	ranks := map[int64]PublicRank{}
	k, known := kind.Get(kindName)
	if !known {
		return ranks, nil
	}
	shows, err := s.entryRule(r)
	if err != nil {
		return nil, err
	}
	standings, err := s.Store.Standings(r.Context(), k)
	if err != nil {
		return nil, err
	}
	shown := slices.DeleteFunc(standings, func(st store.Standing) bool {
		return st.Rank == 0 || !shows(k.Name, st.Status, st.Visibility)
	})
	slices.SortFunc(shown, func(a, b store.Standing) int { return cmp.Compare(a.Rank, b.Rank) })
	for i, st := range shown {
		ranks[st.ItemID] = PublicRank{Place: i + 1, Of: len(shown)}
	}
	return ranks, nil
}

func withRank(item PublicItem, rank PublicRank, ranked bool) PublicItem {
	if ranked {
		item.Rank = &rank
	}
	return item
}

type PublicKind struct {
	Name      string `json:"name"`
	Untriaged int    `json:"untriaged"`
}

func (s *Server) publicKind(w http.ResponseWriter, r *http.Request) {
	k, known := kind.Get(r.PathValue("kind"))
	if !known {
		respond(w, 0, nil, store.ErrNotFound)
		return
	}
	untriaged, err := s.Store.CountUntriaged(r.Context(), k.Name)
	respond(w, http.StatusOK, PublicKind{Name: k.Name, Untriaged: untriaged}, err)
}

type Presentation struct {
	Axes      store.Axes                     `json:"axes"`
	Kinds     map[string]store.KindSettings  `json:"kinds"`
	Fields    map[string][]kind.PublicField  `json:"fields"`
	Pictures  map[string][]kind.PictureField `json:"pictures"`
	TagTitles store.TagTitles                `json:"tagTitles"`
	Readables []kind.Readable                `json:"readables"`
	KindOrder store.KindOrder                `json:"kindOrder"`
}

func picturesByKind() map[string][]kind.PictureField {
	pictures := map[string][]kind.PictureField{}
	for _, k := range kind.All() {
		pictures[k.Name] = k.Pictures()
	}
	return pictures
}

func publicFieldsByKind() map[string][]kind.PublicField {
	fields := map[string][]kind.PublicField{}
	for _, k := range kind.All() {
		fields[k.Name] = k.PublicFields()
	}
	return fields
}

func (s *Server) presentation(r *http.Request) (Presentation, error) {
	axes, err := s.Store.Axes(r.Context())
	if err != nil {
		return Presentation{}, err
	}
	kinds, err := s.Store.AllKindSettings(r.Context())
	if err != nil {
		return Presentation{}, err
	}
	titles, err := s.Store.TagTitles(r.Context())
	if err != nil {
		return Presentation{}, err
	}
	order, err := s.Store.KindOrder(r.Context())
	return Presentation{Axes: axes, Kinds: kinds, Fields: publicFieldsByKind(), Pictures: picturesByKind(), TagTitles: titles, Readables: kind.Readables(), KindOrder: order}, err
}

func (s *Server) adminSettings(w http.ResponseWriter, r *http.Request) {
	presentation, err := s.presentation(r)
	respond(w, http.StatusOK, presentation, err)
}

func (s *Server) publicPresentation(w http.ResponseWriter, r *http.Request) {
	presentation, err := s.presentation(r)
	if err == nil {
		err = s.keepKindsWithSomethingPublic(r, &presentation)
	}
	respond(w, http.StatusOK, presentation, err)
}

func (s *Server) keepKindsWithSomethingPublic(r *http.Request, presentation *Presentation) error {
	isPublic, err := s.publicRule(r)
	if err != nil {
		return err
	}
	views, err := s.Store.ListItems(r.Context(), store.ItemFilter{})
	if err != nil {
		return err
	}
	hasPublic := map[string]bool{}
	for _, v := range views {
		if isPublic(v) {
			hasPublic[v.Kind] = true
		}
	}
	presentation.Kinds = maps.Collect(func(yield func(string, store.KindSettings) bool) {
		for name, settings := range presentation.Kinds {
			if hasPublic[name] && !yield(name, settings) {
				return
			}
		}
	})
	return nil
}

type publicRule func(store.ItemView) bool

type entryRule func(kindName, status string, visibility store.Visibility) bool

func (s *Server) entryRule(r *http.Request) (entryRule, error) {
	kinds, err := s.Store.AllKindSettings(r.Context())
	if err != nil {
		return nil, err
	}
	return func(kindName, status string, visibility store.Visibility) bool {
		switch visibility {
		case store.Public:
			return status != kind.Dismissed
		case store.Auto:
			settings, known := kinds[kindName]
			return known && settings.Shows(status)
		default:
			return false
		}
	}, nil
}

func (s *Server) publicRule(r *http.Request) (publicRule, error) {
	shows, err := s.entryRule(r)
	if err != nil {
		return nil, err
	}
	return func(v store.ItemView) bool {
		return v.Entry != nil && shows(v.Kind, v.Entry.Status, v.Entry.Visibility)
	}, nil
}

func (s *Server) publicItems(w http.ResponseWriter, r *http.Request) {
	isPublic, err := s.publicRule(r)
	if err != nil {
		respond(w, 0, nil, err)
		return
	}
	kindName := r.URL.Query().Get("kind")
	views, err := s.Store.ListItems(r.Context(), store.ItemFilter{Kind: kindName})
	if err != nil {
		respond(w, 0, nil, err)
		return
	}
	ranks, err := s.publicRanks(r, kindName)
	if err != nil {
		respond(w, 0, nil, err)
		return
	}
	shown := slices.DeleteFunc(views, func(v store.ItemView) bool { return !isPublic(v) })
	orders := unrankedOrders(shown)
	items := make([]PublicItem, 0, len(shown))
	for _, v := range shown {
		rank, ranked := ranks[v.ID]
		item := withRank(publicItem(v), rank, ranked)
		item.Order = orders[v.ID]
		items = append(items, item)
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) publicItem(w http.ResponseWriter, r *http.Request) {
	isPublic, err := s.publicRule(r)
	if err != nil {
		respond(w, 0, nil, err)
		return
	}
	view, err := s.Store.GetItemBySlug(r.Context(), r.PathValue("kind"), r.PathValue("slug"))
	if err == nil && !isPublic(view) {
		err = store.ErrNotFound
	}
	if err != nil {
		respond(w, 0, nil, err)
		return
	}
	ranks, err := s.publicRanks(r, view.Kind)
	if err != nil {
		respond(w, 0, nil, err)
		return
	}
	rank, ranked := ranks[view.ID]
	item := withRank(publicItem(view), rank, ranked)
	item.Prose = view.Entry.Prose
	if item.Parts, err = s.publicParts(r, view.ID, isPublic); err != nil {
		respond(w, 0, nil, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) publicParts(r *http.Request, wholeID int64, isPublic publicRule) (*PublicParts, error) {
	partIDs, err := s.Store.PartsOf(r.Context(), wholeID)
	if err != nil || len(partIDs) == 0 {
		return nil, err
	}
	views, err := s.Store.GetItems(r.Context(), partIDs)
	if err != nil {
		return nil, err
	}
	byID := map[int64]store.ItemView{}
	for _, v := range views {
		byID[v.ID] = v
	}
	parts := make([]PublicPart, 0, len(partIDs))
	for _, id := range partIDs {
		v := byID[id]
		part := PublicPart{Kind: v.Kind, Fields: withoutPrivateFields(v.Kind, v.Verdicts.Values)}
		if v.Entry != nil && isPublic(v) {
			part.Scores = v.Entry.EveryScore()
		}
		parts = append(parts, part)
	}
	partKind, known := kind.Get(parts[0].Kind)
	if !known {
		return nil, nil
	}
	settings, err := s.Store.KindSettings(r.Context(), partKind)
	if err != nil {
		return nil, err
	}
	return &PublicParts{Title: settings.Title, Items: parts}, nil
}

func publicItem(v store.ItemView) PublicItem {
	return PublicItem{
		Kind:       v.Kind,
		Slug:       v.Slug,
		Fields:     withoutPrivateFields(v.Kind, v.Verdicts.Values),
		Status:     v.Entry.Status,
		Scores:     v.Entry.EveryScore(),
		Tags:       v.Entry.Tags,
		Blurb:      v.Entry.Blurb,
		TimeSpent:  v.Entry.TimeSpent,
		UpdatedAt:  v.Entry.UpdatedAt,
		ProseWords: len(strings.Fields(v.Entry.Prose)),
		Takes:      publicTakes(v.Takes),
	}
}

func withoutPrivateFields(kindName string, values map[string]any) map[string]any {
	k, known := kind.Get(kindName)
	if !known {
		return values
	}
	public := maps.Clone(values)
	for _, name := range k.PrivateFieldNames() {
		delete(public, name)
	}
	return public
}

func publicTakes(takes []store.TakeView) []verdict.Take {
	public := make([]verdict.Take, 0, len(takes))
	for _, t := range takes {
		public = append(public, t.Public())
	}
	return public
}

func unrankedOrders(views []store.ItemView) map[int64]int {
	orders := map[int64]int{}
	values := map[int64]float64{}
	for _, v := range views {
		k, known := kind.Get(v.Kind)
		if !known || k.UnrankedOrder == "" {
			continue
		}
		if value, present := asOrderValue(v.Verdicts.Values[k.UnrankedOrder]); present {
			values[v.ID] = value
		}
	}
	for id, value := range values {
		ahead := 0
		for _, other := range values {
			if other > value {
				ahead++
			}
		}
		orders[id] = ahead + 1
	}
	return orders
}

func asOrderValue(value any) (float64, bool) {
	switch n := value.(type) {
	case int64:
		return float64(n), true
	case float64:
		return n, true
	}
	return 0, false
}
