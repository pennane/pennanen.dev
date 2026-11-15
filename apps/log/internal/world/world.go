package world

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"slices"
)

const DeriveVersion = 1

var ErrNotFound = errors.New("not found")

type World struct {
	db      *sql.DB
	records *record.Store
}

func New(db *sql.DB, records *record.Store) *World {
	return &World{db: db, records: records}
}

type Move struct {
	ItemID   int64
	From, To string
}

type Overlap struct {
	Identity            record.Identity
	ItemID, OtherItemID int64
}

type Report struct {
	Created  []int64
	Moves    []Move
	Overlaps []Overlap
}

func (r Report) and(other Report) Report {
	return Report{
		Created:  append(r.Created, other.Created...),
		Moves:    append(r.Moves, other.Moves...),
		Overlaps: append(r.Overlaps, other.Overlaps...),
	}
}

type settling struct {
	ctx      context.Context
	q        record.Queryer
	records  *record.Store
	readings map[record.Identity]record.Reading
	fetched  map[record.Identity]struct{}
	touched  map[int64]struct{}
	related  map[int64]struct{}
	owns     map[int64]decision
	pulled   []record.Identity
	suppress []AssertedLink
	loaded   suppressionsLoad
	report   Report
}

type suppressionsLoad string

const (
	suppressionsUnread suppressionsLoad = "unread"
	suppressionsRead   suppressionsLoad = "read"
)

func (w *World) Settle(ctx context.Context, q record.Queryer, dirty []record.Identity) (Report, error) {
	if len(dirty) == 0 {
		return Report{}, nil
	}
	s := &settling{
		ctx:      ctx,
		q:        q,
		records:  w.records,
		readings: map[record.Identity]record.Reading{},
		fetched:  map[record.Identity]struct{}{},
		touched:  map[int64]struct{}{},
		related:  map[int64]struct{}{},
		owns:     map[int64]decision{},
		loaded:   suppressionsUnread,
	}
	dirty = uniqueSorted(dirty)
	if err := s.group(dirty); err != nil {
		return Report{}, err
	}
	if err := s.storeLinks(dirty); err != nil {
		return Report{}, err
	}
	if err := s.decideKinds(); err != nil {
		return Report{}, err
	}
	if err := s.decideVerdicts(); err != nil {
		return Report{}, err
	}
	return s.report, nil
}

func (s *settling) fetch(ids []record.Identity) error {
	missing := []record.Identity{}
	for _, id := range ids {
		if _, known := s.fetched[id]; !known {
			missing = append(missing, id)
			s.fetched[id] = struct{}{}
		}
	}
	if len(missing) == 0 {
		return nil
	}
	readings, err := s.records.Readings(s.ctx, s.q, missing)
	if err != nil {
		return err
	}
	for _, r := range readings {
		s.readings[r.Identity] = r
	}
	return nil
}

func (s *settling) derivationOf(id record.Identity) (Derivation, error) {
	if err := s.fetch([]record.Identity{id}); err != nil {
		return Derivation{}, err
	}
	reading, read := s.readings[id]
	if !read {
		return Derivation{}, nil
	}
	return Derive(id, reading.Values), nil
}

func (s *settling) readingsOf(ids []record.Identity) ([]record.Reading, error) {
	if err := s.fetch(ids); err != nil {
		return nil, err
	}
	readings := []record.Reading{}
	for _, id := range ids {
		if r, read := s.readings[id]; read {
			readings = append(readings, r)
		}
	}
	return readings, nil
}

func (w *World) ItemOf(ctx context.Context, q record.Queryer, id record.Identity) (int64, error) {
	return itemOf(ctx, q, id)
}

func itemOf(ctx context.Context, q record.Queryer, id record.Identity) (int64, error) {
	var itemID int64
	err := q.QueryRowContext(ctx,
		`SELECT item_id FROM item_identity WHERE source = ? AND external_id = ?`,
		id.Source, id.ExternalID).Scan(&itemID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	return itemID, err
}

func (w *World) KeyOf(ctx context.Context, q record.Queryer, itemID int64) (record.Identity, error) {
	var key record.Identity
	err := q.QueryRowContext(ctx,
		`SELECT source, external_id FROM item_identity WHERE item_id = ? ORDER BY id LIMIT 1`,
		itemID).Scan(&key.Source, &key.ExternalID)
	if errors.Is(err, sql.ErrNoRows) {
		return record.Identity{}, ErrNotFound
	}
	return key, err
}

func compareIdentities(a, b record.Identity) int {
	return cmp.Or(cmp.Compare(a.Source, b.Source), cmp.Compare(a.ExternalID, b.ExternalID))
}

func uniqueSorted(ids []record.Identity) []record.Identity {
	sorted := slices.SortedFunc(slices.Values(ids), compareIdentities)
	return slices.Compact(sorted)
}

func sortedItems(set map[int64]struct{}) []int64 {
	ids := make([]int64, 0, len(set))
	for id := range set {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	return ids
}
