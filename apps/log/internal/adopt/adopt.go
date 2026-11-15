package adopt

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
	"github.com/pennane/pennanen.dev/apps/log/internal/world"
)

type Outcome string

const (
	Adopted        Outcome = "adopted"
	NothingToAdopt Outcome = "nothing-to-adopt"
)

type Adoption struct {
	Outcome      Outcome
	Items        int
	Claims       int
	Observations int
	Links        int
	Overlaps     int
}

const (
	batchSize        = 200
	timeLayout       = "2006-01-02T15:04:05Z"
	apItemPrefix     = "item/"
	apFieldsPrefix   = "fields/"
	apMatchPrefix    = "match/"
	legacyItemPrefix = "item/"
	noMatchAnswer    = "none"
)

var legacyTables = []string{
	"legacy_claim",
	"legacy_observation",
	"legacy_relation",
	"legacy_part",
	"moment_archive",
	"moment_ref_archive",
	"number_score_archive",
}

type legacyPresence string

const (
	legacyPresent legacyPresence = "present"
	legacyGone    legacyPresence = "gone"
)

type joining string

const (
	joinedByDerives joining = "joined-by-derives"
	apartInDerives  joining = "apart-in-derives"
)

type keyRecord string

const (
	keyClaimed   keyRecord = "claimed"
	keyUnclaimed keyRecord = "unclaimed"
)

type legacyItem struct {
	id        int64
	kind      string
	createdAt string
}

type legacyClaim struct {
	source     string
	externalID string
	data       json.RawMessage
	fetchedAt  string
}

type placedIdentity struct {
	itemID int64
	id     record.Identity
}

type apItem struct {
	Kind   string          `json:"kind"`
	Fields json.RawMessage `json:"fields"`
}

type apFields struct {
	Fields json.RawMessage `json:"fields"`
}

type apMatchAnswer struct {
	To *string `json:"to"`
}

type legacyLink struct {
	From     string `json:"from"`
	Type     string `json:"type"`
	To       string `json:"to"`
	Position int    `json:"position,omitempty"`
}

type legacyRecord struct {
	About []string     `json:"about"`
	Links []legacyLink `json:"links"`
}

type adopting struct {
	ctx      context.Context
	tx       *sql.Tx
	records  *record.Store
	world    *world.World
	at       string
	keys     map[int64]record.Identity
	placed   []placedIdentity
	adoption Adoption
}

func Run(ctx context.Context, db *sql.DB, records *record.Store, w *world.World) (Adoption, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return Adoption{}, err
	}
	defer tx.Rollback()
	presence, err := presenceOfLegacy(ctx, tx)
	if err != nil || presence == legacyGone {
		return Adoption{Outcome: NothingToAdopt}, err
	}
	a := &adopting{
		ctx:      ctx,
		tx:       tx,
		records:  records,
		world:    w,
		at:       time.Now().UTC().Format(timeLayout),
		keys:     map[int64]record.Identity{},
		adoption: Adoption{Outcome: Adopted},
	}
	for _, step := range []func() error{a.adoptItems, a.adoptObservations, a.rebuild, a.keepLegacyLinks, a.check, a.dropLegacy} {
		if err := step(); err != nil {
			return Adoption{}, fmt.Errorf("adopt: %w", err)
		}
	}
	return a.adoption, tx.Commit()
}

func presenceOfLegacy(ctx context.Context, q record.Queryer) (legacyPresence, error) {
	var tables int
	if err := q.QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = 'legacy_claim'`).Scan(&tables); err != nil {
		return "", err
	}
	if tables == 0 {
		return legacyGone, nil
	}
	return legacyPresent, nil
}

func (a *adopting) adoptItems() error {
	var after int64
	for {
		batch, err := a.legacyItemsAfter(after)
		if err != nil || len(batch) == 0 {
			return err
		}
		for _, it := range batch {
			if err := a.adoptItem(it); err != nil {
				return fmt.Errorf("item %d: %w", it.id, err)
			}
		}
		after = batch[len(batch)-1].id
	}
}

func (a *adopting) legacyItemsAfter(after int64) ([]legacyItem, error) {
	rows, err := a.tx.QueryContext(a.ctx, `SELECT id, kind, created_at FROM item WHERE id > ? ORDER BY id LIMIT ?`, after, batchSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []legacyItem{}
	for rows.Next() {
		var it legacyItem
		if err := rows.Scan(&it.id, &it.kind, &it.createdAt); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (a *adopting) legacyClaimsOf(itemID int64) ([]legacyClaim, error) {
	rows, err := a.tx.QueryContext(a.ctx,
		`SELECT source, coalesce(external_id, ''), data, fetched_at FROM legacy_claim WHERE item_id = ? ORDER BY source`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	claims := []legacyClaim{}
	for rows.Next() {
		var c legacyClaim
		var data string
		if err := rows.Scan(&c.source, &c.externalID, &data, &c.fetchedAt); err != nil {
			return nil, err
		}
		c.data = json.RawMessage(data)
		claims = append(claims, c)
	}
	return claims, rows.Err()
}

func (a *adopting) adoptItem(it legacyItem) error {
	claims, err := a.legacyClaimsOf(it.id)
	if err != nil {
		return err
	}
	key := keyOf(it, claims)
	ids := []record.Identity{key}
	keyRecord := keyUnclaimed
	for _, c := range claims {
		id, data, err := adoptedForm(it, key, c)
		if err != nil {
			return err
		}
		if _, err := a.records.PutClaimAt(a.ctx, a.tx, id, data, c.fetchedAt); err != nil {
			return err
		}
		if id == key {
			keyRecord = keyClaimed
		} else {
			ids = append(ids, id)
		}
		a.adoption.Claims++
	}
	if keyRecord == keyUnclaimed {
		if err := a.putHandMadeKey(it, key); err != nil {
			return err
		}
	}
	if err := a.world.Seed(a.ctx, a.tx, it.id, ids); err != nil {
		return err
	}
	join, err := a.joiningOf(ids)
	if err != nil {
		return err
	}
	if join == apartInDerives {
		legacy := legacyIdentity(it.id)
		if err := a.putLegacyRecord(legacy, legacyRecord{About: identityStrings(ids), Links: []legacyLink{}}); err != nil {
			return err
		}
		if err := a.world.Seed(a.ctx, a.tx, it.id, []record.Identity{legacy}); err != nil {
			return err
		}
		ids = append(ids, legacy)
	}
	for _, id := range ids {
		a.placed = append(a.placed, placedIdentity{itemID: it.id, id: id})
	}
	a.keys[it.id] = key
	a.adoption.Items++
	return nil
}

func keyOf(it legacyItem, claims []legacyClaim) record.Identity {
	if k, known := kind.Get(it.kind); known {
		for _, c := range claims {
			if c.source == k.IdentitySource && c.externalID != "" {
				return record.Identity{Source: c.source, ExternalID: c.externalID}
			}
		}
	}
	for _, c := range claims {
		if c.externalID != "" {
			return record.Identity{Source: c.source, ExternalID: c.externalID}
		}
	}
	return handMadeIdentity(it.id)
}

func handMadeIdentity(itemID int64) record.Identity {
	return record.Identity{Source: kind.SourceAP, ExternalID: apItemPrefix + strconv.FormatInt(itemID, 10)}
}

func legacyIdentity(itemID int64) record.Identity {
	return record.Identity{Source: kind.SourceLegacy, ExternalID: legacyItemPrefix + strconv.FormatInt(itemID, 10)}
}

func adoptedForm(it legacyItem, key record.Identity, c legacyClaim) (record.Identity, json.RawMessage, error) {
	switch {
	case c.source == kind.SourceAP && c.externalID != "":
		return record.Identity{}, nil, fmt.Errorf("an ap claim with external id %q has no adopted form", c.externalID)
	case c.source == kind.SourceAP && key == handMadeIdentity(it.id):
		data, err := json.Marshal(apItem{Kind: it.kind, Fields: c.data})
		return key, data, err
	case c.source == kind.SourceAP:
		data, err := json.Marshal(apFields{Fields: c.data})
		return record.Identity{Source: kind.SourceAP, ExternalID: apFieldsPrefix + key.String()}, data, err
	case c.externalID != "":
		return record.Identity{Source: c.source, ExternalID: c.externalID}, c.data, nil
	case answersNoMatch(c.data):
		data, err := json.Marshal(apMatchAnswer{})
		return record.Identity{Source: kind.SourceAP, ExternalID: apMatchPrefix + c.source + "/" + key.String()}, data, err
	}
	return record.Identity{Source: c.source, ExternalID: key.String()}, c.data, nil
}

func answersNoMatch(data json.RawMessage) bool {
	var values map[string]any
	if err := json.Unmarshal(data, &values); err != nil {
		return false
	}
	return len(values) == 1 && values["match"] == noMatchAnswer
}

func (a *adopting) putHandMadeKey(it legacyItem, key record.Identity) error {
	if key != handMadeIdentity(it.id) {
		return fmt.Errorf("key %s has no record", key)
	}
	data, err := json.Marshal(apItem{Kind: it.kind, Fields: json.RawMessage(`{}`)})
	if err != nil {
		return err
	}
	_, err = a.records.PutClaimAt(a.ctx, a.tx, key, data, it.createdAt)
	return err
}

func (a *adopting) putLegacyRecord(id record.Identity, r legacyRecord) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	_, err = a.records.PutClaimAt(a.ctx, a.tx, id, data, a.at)
	return err
}

func identityStrings(ids []record.Identity) []string {
	strs := make([]string, 0, len(ids))
	for _, id := range ids {
		strs = append(strs, id.String())
	}
	return strs
}

func (a *adopting) joiningOf(ids []record.Identity) (joining, error) {
	readings, err := a.records.Readings(a.ctx, a.tx, ids)
	if err != nil {
		return "", err
	}
	parent := map[record.Identity]record.Identity{}
	var root func(record.Identity) record.Identity
	root = func(id record.Identity) record.Identity {
		above, seen := parent[id]
		if !seen || above == id {
			parent[id] = id
			return id
		}
		top := root(above)
		parent[id] = top
		return top
	}
	for _, r := range readings {
		for _, about := range verdict.Derive(r.Identity, r.Values).About {
			parent[root(about)] = root(r.Identity)
		}
	}
	first := root(ids[0])
	for _, id := range ids[1:] {
		if root(id) != first {
			return apartInDerives, nil
		}
	}
	return joinedByDerives, nil
}

type legacyObservation struct {
	id       int64
	itemID   int64
	observed record.Observation
}

func (a *adopting) adoptObservations() error {
	var after int64
	for {
		batch, err := a.legacyObservationsAfter(after)
		if err != nil || len(batch) == 0 {
			return err
		}
		for _, o := range batch {
			key, known := a.keys[o.itemID]
			if !known {
				return fmt.Errorf("observation %d is on item %d, which has no key", o.id, o.itemID)
			}
			o.observed.About = key
			outcome, err := a.records.Observe(a.ctx, a.tx, o.observed)
			if err != nil {
				return err
			}
			if outcome != record.Observed {
				return fmt.Errorf("observation %d: %s %s was already known", o.id, o.observed.Source, o.observed.Key)
			}
			a.adoption.Observations++
		}
		after = batch[len(batch)-1].id
	}
}

func (a *adopting) legacyObservationsAfter(after int64) ([]legacyObservation, error) {
	rows, err := a.tx.QueryContext(a.ctx, `
		SELECT id, item_id, source, signal, key, observed_at, data FROM legacy_observation
		WHERE id > ? ORDER BY id LIMIT ?`, after, batchSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	observations := []legacyObservation{}
	for rows.Next() {
		var o legacyObservation
		var data string
		if err := rows.Scan(&o.id, &o.itemID, &o.observed.Source, &o.observed.Signal, &o.observed.Key, &o.observed.ObservedAt, &data); err != nil {
			return nil, err
		}
		o.observed.Data = json.RawMessage(data)
		observations = append(observations, o)
	}
	return observations, rows.Err()
}

func (a *adopting) rebuild() error {
	report, err := a.world.RebuildWithin(a.ctx, a.tx)
	if err != nil {
		return err
	}
	return a.accept(report)
}

func (a *adopting) accept(report world.Report) error {
	if len(report.Moves) > 0 {
		return fmt.Errorf("the world would move %d items, the first %d from %s to %s",
			len(report.Moves), report.Moves[0].ItemID, report.Moves[0].From, report.Moves[0].To)
	}
	if len(report.Created) > 0 {
		return fmt.Errorf("the world would create %d items no legacy item holds, the first %d", len(report.Created), report.Created[0])
	}
	a.adoption.Overlaps += len(report.Overlaps)
	return nil
}

func (a *adopting) dropLegacy() error {
	for _, table := range legacyTables {
		if _, err := a.tx.ExecContext(a.ctx, `DROP TABLE `+table); err != nil {
			return err
		}
	}
	return nil
}
