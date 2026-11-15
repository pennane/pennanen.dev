package world

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
)

const (
	noPosition    = 0
	firstPosition = 1
)

type linkKey struct {
	from     record.Identity
	linkType string
	to       record.Identity
}

func (s *settling) storeLinks(dirty []record.Identity) error {
	asserters, err := s.assertersAround(dirty)
	if err != nil {
		return err
	}
	for _, asserter := range asserters {
		if err := s.storeLinksAssertedBy(asserter); err != nil {
			return err
		}
	}
	return nil
}

func (s *settling) assertersAround(dirty []record.Identity) ([]record.Identity, error) {
	asserters := slices.Clone(dirty)
	for _, id := range dirty {
		if id.Source != kind.SourceAP {
			continue
		}
		derived, err := s.derivationOf(id)
		if err != nil {
			return nil, err
		}
		for _, l := range slices.Concat(derived.Links, derived.Suppressed) {
			for _, end := range []record.Identity{l.From, l.To} {
				ids, err := s.identitiesAt(end)
				if err != nil {
					return nil, err
				}
				asserters = append(asserters, ids...)
			}
		}
	}
	return uniqueSorted(asserters), nil
}

func (s *settling) identitiesAt(end record.Identity) ([]record.Identity, error) {
	itemID, err := itemOf(s.ctx, s.q, end)
	if errors.Is(err, ErrNotFound) {
		return []record.Identity{end}, nil
	}
	if err != nil {
		return nil, err
	}
	return identitiesOn(s.ctx, s.q, itemID)
}

func (s *settling) storeLinksAssertedBy(asserter record.Identity) error {
	derived, err := s.derivationOf(asserter)
	if err != nil {
		return err
	}
	want, err := s.admitted(asserter, derived.Links)
	if err != nil {
		return err
	}
	have, err := s.linksAssertedBy(asserter)
	if err != nil {
		return err
	}
	for key := range have {
		if _, kept := want[key]; kept {
			continue
		}
		if _, err := s.q.ExecContext(s.ctx, `
			DELETE FROM link WHERE asserted_source = ? AND asserted_external_id = ?
			  AND from_source = ? AND from_external_id = ? AND type = ? AND to_source = ? AND to_external_id = ?`,
			asserter.Source, asserter.ExternalID, key.from.Source, key.from.ExternalID, key.linkType, key.to.Source, key.to.ExternalID); err != nil {
			return err
		}
		if err := s.relateThrough(key); err != nil {
			return err
		}
	}
	for key, position := range want {
		stored, had := have[key]
		if had && stored == position {
			continue
		}
		if _, err := s.q.ExecContext(s.ctx, `
			INSERT INTO link (asserted_source, asserted_external_id, from_source, from_external_id, type, to_source, to_external_id, position)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT DO UPDATE SET position = excluded.position`,
			asserter.Source, asserter.ExternalID, key.from.Source, key.from.ExternalID, key.linkType, key.to.Source, key.to.ExternalID,
			positionColumn(position)); err != nil {
			return err
		}
		if err := s.relateThrough(key); err != nil {
			return err
		}
	}
	return nil
}

func positionColumn(position int) any {
	if position == noPosition {
		return nil
	}
	return position
}

func (s *settling) admitted(asserter record.Identity, links []verdict.AssertedLink) (map[linkKey]int, error) {
	admitted := map[linkKey]int{}
	for _, l := range links {
		if err := fitsItsType(asserter, l); err != nil {
			return nil, err
		}
		key := linkKey{from: l.From, linkType: l.Type, to: l.To}
		if _, repeated := admitted[key]; repeated {
			continue
		}
		suppressed, err := s.isSuppressed(l)
		if err != nil {
			return nil, err
		}
		if !suppressed {
			admitted[key] = l.Position
		}
	}
	return admitted, nil
}

func fitsItsType(asserter record.Identity, l verdict.AssertedLink) error {
	linkType, known := kind.LinkTypeOf(l.Type)
	switch {
	case !known:
		return fmt.Errorf("%s asserts a link of unknown type %q from %s to %s", asserter, l.Type, l.From, l.To)
	case linkType.Order == kind.Ordered && l.Position < firstPosition:
		return fmt.Errorf("%s asserts a %s link from %s to %s without a position", asserter, l.Type, l.From, l.To)
	case linkType.Order == kind.Unordered && l.Position != noPosition:
		return fmt.Errorf("%s asserts a %s link from %s to %s at position %d, but %s links have none", asserter, l.Type, l.From, l.To, l.Position, l.Type)
	}
	return nil
}

func (s *settling) isSuppressed(l verdict.AssertedLink) (bool, error) {
	if err := s.loadSuppressions(); err != nil {
		return false, err
	}
	for _, suppression := range s.suppress {
		if suppression.Type != l.Type {
			continue
		}
		sameFrom, err := s.sameEnd(suppression.From, l.From)
		if err != nil {
			return false, err
		}
		sameTo, err := s.sameEnd(suppression.To, l.To)
		if err != nil {
			return false, err
		}
		if sameFrom && sameTo {
			return true, nil
		}
	}
	return false, nil
}

func (s *settling) sameEnd(a, b record.Identity) (bool, error) {
	if a == b {
		return true, nil
	}
	aItem, err := itemOf(s.ctx, s.q, a)
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	bItem, err := itemOf(s.ctx, s.q, b)
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return aItem == bItem, nil
}

func (s *settling) loadSuppressions() error {
	if s.loaded == suppressionsRead {
		return nil
	}
	rows, err := s.q.QueryContext(s.ctx,
		`SELECT external_id, reading FROM record_reading WHERE source = ? ORDER BY external_id`, kind.SourceAP)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		id := record.Identity{Source: kind.SourceAP}
		var raw string
		if err := rows.Scan(&id.ExternalID, &raw); err != nil {
			return err
		}
		var values map[string]any
		if err := json.Unmarshal([]byte(raw), &values); err != nil {
			return fmt.Errorf("reading of %s: %w", id, err)
		}
		s.suppress = append(s.suppress, verdict.Derive(id, values).Suppressed...)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	s.loaded = suppressionsRead
	return nil
}

func (s *settling) linksAssertedBy(asserter record.Identity) (map[linkKey]int, error) {
	rows, err := s.q.QueryContext(s.ctx, `
		SELECT from_source, from_external_id, type, to_source, to_external_id, position
		FROM link WHERE asserted_source = ? AND asserted_external_id = ?`,
		asserter.Source, asserter.ExternalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	links := map[linkKey]int{}
	for rows.Next() {
		var key linkKey
		var position sql.NullInt64
		if err := rows.Scan(&key.from.Source, &key.from.ExternalID, &key.linkType, &key.to.Source, &key.to.ExternalID, &position); err != nil {
			return nil, err
		}
		links[key] = int(position.Int64)
	}
	return links, rows.Err()
}

func (s *settling) relateThrough(key linkKey) error {
	if !slices.Contains(inheritingLinkTypes(), key.linkType) {
		return nil
	}
	itemID, err := itemOf(s.ctx, s.q, key.from)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	s.related[itemID] = struct{}{}
	return nil
}

func inheritingLinkTypes() []string {
	types := []string{}
	for _, k := range kind.All() {
		for _, through := range k.InheritingRelations() {
			if !slices.Contains(types, through) {
				types = append(types, through)
			}
		}
	}
	slices.Sort(types)
	return types
}
