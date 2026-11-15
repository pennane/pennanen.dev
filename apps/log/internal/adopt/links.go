package adopt

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
)

const (
	relationsMissingFromTheView = `
		SELECT from_id, type, to_id FROM legacy_relation
		EXCEPT SELECT from_id, type, to_id FROM relation`
	partsWithoutTheirLink = `
		SELECT lp.whole_id, lp.position, lp.part_id FROM legacy_part lp
		WHERE NOT EXISTS (
			SELECT 1 FROM link l
			JOIN item_identity f ON f.source = l.from_source AND f.external_id = l.from_external_id
			JOIN item_identity t ON t.source = l.to_source AND t.external_id = l.to_external_id
			WHERE l.type = ? AND l.position = lp.position AND f.item_id = lp.part_id AND t.item_id = lp.whole_id)`
	partsMissingFromTheView = `
		SELECT whole_id, part_id FROM legacy_part
		EXCEPT SELECT whole_id, part_id FROM part`
)

type legacyRelation struct {
	from     int64
	linkType string
	to       int64
}

type legacyPart struct {
	whole    int64
	position int
	part     int64
}

func (a *adopting) keepLegacyLinks() error {
	relations, err := a.missingRelations()
	if err != nil {
		return err
	}
	parts, err := a.partsWithoutLinks()
	if err != nil {
		return err
	}
	linksOf := map[int64][]legacyLink{}
	for _, r := range relations {
		from, to, err := a.keysOf(r.from, r.to)
		if err != nil {
			return fmt.Errorf("relation %d %s %d: %w", r.from, r.linkType, r.to, err)
		}
		linksOf[r.from] = append(linksOf[r.from], legacyLink{From: from.String(), Type: r.linkType, To: to.String()})
	}
	for _, p := range parts {
		part, whole, err := a.keysOf(p.part, p.whole)
		if err != nil {
			return fmt.Errorf("part %d of %d: %w", p.part, p.whole, err)
		}
		linksOf[p.part] = append(linksOf[p.part], legacyLink{From: part.String(), Type: kind.LinkPartOf, To: whole.String(), Position: p.position})
	}
	dirty := []record.Identity{}
	for _, itemID := range slices.Sorted(maps.Keys(linksOf)) {
		legacy, err := a.addLegacyLinks(itemID, linksOf[itemID])
		if err != nil {
			return fmt.Errorf("item %d: %w", itemID, err)
		}
		dirty = append(dirty, legacy)
		a.adoption.Links += len(linksOf[itemID])
	}
	for batch := range slices.Chunk(dirty, batchSize) {
		report, err := a.world.Settle(a.ctx, a.tx, batch)
		if err != nil {
			return err
		}
		if err := a.accept(report); err != nil {
			return err
		}
	}
	return nil
}

func (a *adopting) keysOf(from, to int64) (record.Identity, record.Identity, error) {
	fromKey, fromKnown := a.keys[from]
	toKey, toKnown := a.keys[to]
	if !fromKnown || !toKnown {
		return record.Identity{}, record.Identity{}, fmt.Errorf("an end has no key")
	}
	return fromKey, toKey, nil
}

func (a *adopting) addLegacyLinks(itemID int64, links []legacyLink) (record.Identity, error) {
	legacy := legacyIdentity(itemID)
	stored, err := a.storedLegacyRecord(legacy, a.keys[itemID])
	if err != nil {
		return record.Identity{}, err
	}
	stored.Links = append(stored.Links, links...)
	if err := a.putLegacyRecord(legacy, stored); err != nil {
		return record.Identity{}, err
	}
	return legacy, a.world.Seed(a.ctx, a.tx, itemID, []record.Identity{legacy})
}

func (a *adopting) storedLegacyRecord(legacy, key record.Identity) (legacyRecord, error) {
	var raw string
	err := a.tx.QueryRowContext(a.ctx, `SELECT data FROM record_claim WHERE source = ? AND external_id = ?`,
		legacy.Source, legacy.ExternalID).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return legacyRecord{About: []string{key.String()}, Links: []legacyLink{}}, nil
	}
	if err != nil {
		return legacyRecord{}, err
	}
	var stored legacyRecord
	return stored, json.Unmarshal([]byte(raw), &stored)
}

func (a *adopting) missingRelations() ([]legacyRelation, error) {
	rows, err := a.tx.QueryContext(a.ctx, relationsMissingFromTheView+` ORDER BY 1, 2, 3`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	relations := []legacyRelation{}
	for rows.Next() {
		var r legacyRelation
		if err := rows.Scan(&r.from, &r.linkType, &r.to); err != nil {
			return nil, err
		}
		relations = append(relations, r)
	}
	return relations, rows.Err()
}

func (a *adopting) partsWithoutLinks() ([]legacyPart, error) {
	rows, err := a.tx.QueryContext(a.ctx, partsWithoutTheirLink+` ORDER BY lp.whole_id, lp.position`, kind.LinkPartOf)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	parts := []legacyPart{}
	for rows.Next() {
		var p legacyPart
		if err := rows.Scan(&p.whole, &p.position, &p.part); err != nil {
			return nil, err
		}
		parts = append(parts, p)
	}
	return parts, rows.Err()
}
