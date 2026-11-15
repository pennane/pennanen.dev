package adopt

import (
	"errors"
	"fmt"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/world"
)

const (
	claimsWithoutTheirRecord = `
		SELECT 1 FROM legacy_claim lc
		WHERE coalesce(lc.external_id, '') <> '' AND lc.source <> ?
		  AND NOT EXISTS (
			SELECT 1 FROM record_claim rc
			JOIN item_identity i ON i.source = rc.source AND i.external_id = rc.external_id
			WHERE rc.source = lc.source AND rc.external_id = lc.external_id
			  AND rc.data = lc.data AND rc.fetched_at = lc.fetched_at AND i.item_id = lc.item_id)`
	observationsWithoutTheirRecord = `
		SELECT 1 FROM legacy_observation lo
		WHERE NOT EXISTS (
			SELECT 1 FROM record_observation ro
			JOIN item_identity i ON i.source = ro.about_source AND i.external_id = ro.about_external_id
			WHERE ro.source = lo.source AND ro.key = lo.key AND ro.signal = lo.signal
			  AND ro.observed_at = lo.observed_at AND ro.data = lo.data AND i.item_id = lo.item_id)`
	itemsWithoutIdentities = `
		SELECT 1 FROM item WHERE NOT EXISTS (SELECT 1 FROM item_identity i WHERE i.item_id = item.id)`
)

func (a *adopting) check() error {
	for _, check := range []func() error{
		a.everyClaimIsARecord,
		a.everyObservationIsARecord,
		a.everyIdentityStaysOnItsItem,
		a.everyItemKeepsItsKey,
		a.everyRelationShows,
		a.everyPartShows,
	} {
		if err := check(); err != nil {
			return err
		}
	}
	return nil
}

func (a *adopting) count(query string, args ...any) (int, error) {
	var n int
	err := a.tx.QueryRowContext(a.ctx, `SELECT count(*) FROM (`+query+`)`, args...).Scan(&n)
	return n, err
}

func (a *adopting) none(what, query string, args ...any) error {
	n, err := a.count(query, args...)
	if err != nil {
		return err
	}
	if n > 0 {
		return fmt.Errorf("%d %s", n, what)
	}
	return nil
}

func (a *adopting) sameCount(what, legacyTable string, adopted int) error {
	n, err := a.count(`SELECT 1 FROM ` + legacyTable)
	if err != nil {
		return err
	}
	if n != adopted {
		return fmt.Errorf("adopted %d of %d legacy %s", adopted, n, what)
	}
	return nil
}

func (a *adopting) everyClaimIsARecord() error {
	if err := a.sameCount("claims", "legacy_claim", a.adoption.Claims); err != nil {
		return err
	}
	return a.none("legacy claims are not the same record on the same item", claimsWithoutTheirRecord, kind.SourceAP)
}

func (a *adopting) everyObservationIsARecord() error {
	if err := a.sameCount("observations", "legacy_observation", a.adoption.Observations); err != nil {
		return err
	}
	return a.none("legacy observations are not the same observation about the same item", observationsWithoutTheirRecord)
}

func (a *adopting) everyIdentityStaysOnItsItem() error {
	for _, p := range a.placed {
		itemID, err := a.world.ItemOf(a.ctx, a.tx, p.id)
		if errors.Is(err, world.ErrNotFound) {
			return fmt.Errorf("%s left item %d", p.id, p.itemID)
		}
		if err != nil {
			return err
		}
		if itemID != p.itemID {
			return fmt.Errorf("%s moved from item %d to %d", p.id, p.itemID, itemID)
		}
		records, err := a.count(`SELECT 1 FROM record_claim WHERE source = ? AND external_id = ?`, p.id.Source, p.id.ExternalID)
		if err != nil {
			return err
		}
		if records != 1 {
			return fmt.Errorf("%s on item %d has no record", p.id, p.itemID)
		}
	}
	return nil
}

func (a *adopting) everyItemKeepsItsKey() error {
	if err := a.none("items have no identity", itemsWithoutIdentities); err != nil {
		return err
	}
	for itemID, key := range a.keys {
		got, err := a.world.KeyOf(a.ctx, a.tx, itemID)
		if err != nil {
			return err
		}
		if got != key {
			return fmt.Errorf("item %d has key %s, want %s", itemID, got, key)
		}
	}
	return nil
}

func (a *adopting) everyRelationShows() error {
	return a.none("legacy relations are missing from the relation view", relationsMissingFromTheView)
}

func (a *adopting) everyPartShows() error {
	if err := a.none("legacy parts have no part-of link at their position", partsWithoutTheirLink, kind.LinkPartOf); err != nil {
		return err
	}
	return a.none("legacy parts are missing from the part view", partsMissingFromTheView)
}
