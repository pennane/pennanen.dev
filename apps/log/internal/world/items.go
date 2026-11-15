package world

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/verdict"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const (
	untitled = "untitled"
	noKind   = ""
)

func (s *settling) group(dirty []record.Identity) error {
	for pending := dirty; len(pending) > 0; {
		if err := s.groupOnce(pending); err != nil {
			return err
		}
		pending, s.pulled = uniqueSorted(s.pulled), nil
	}
	return nil
}

func (s *settling) groupOnce(dirty []record.Identity) error {
	parent := map[record.Identity]record.Identity{}
	var root func(record.Identity) record.Identity
	root = func(id record.Identity) record.Identity {
		above, seen := parent[id]
		if !seen {
			parent[id] = id
			return id
		}
		if above == id {
			return id
		}
		top := root(above)
		parent[id] = top
		return top
	}
	for _, id := range dirty {
		derived, err := s.derivationOf(id)
		if err != nil {
			return err
		}
		for _, about := range derived.About {
			parent[root(about)] = root(id)
		}
		root(id)
	}
	members := map[record.Identity][]record.Identity{}
	for id := range parent {
		top := root(id)
		members[top] = append(members[top], id)
	}
	groups := make([][]record.Identity, 0, len(members))
	for _, ids := range members {
		groups = append(groups, uniqueSorted(ids))
	}
	slices.SortFunc(groups, func(a, b []record.Identity) int { return compareIdentities(a[0], b[0]) })
	for _, g := range groups {
		if err := s.settleGroup(g); err != nil {
			return err
		}
	}
	return nil
}

func (s *settling) settleGroup(g []record.Identity) error {
	owners := map[int64][]record.Identity{}
	unresolved := []record.Identity{}
	for _, id := range g {
		itemID, err := itemOf(s.ctx, s.q, id)
		switch {
		case errors.Is(err, ErrNotFound):
			unresolved = append(unresolved, id)
		case err != nil:
			return err
		default:
			owners[itemID] = append(owners[itemID], id)
		}
	}
	if len(owners) == 0 {
		return s.create(g)
	}
	items := sortedItems(setOf(owners))
	target := items[0]
	for _, id := range unresolved {
		if err := s.resolve(target, id); err != nil {
			return err
		}
	}
	for _, itemID := range items {
		s.touched[itemID] = struct{}{}
	}
	for _, other := range items[1:] {
		if err := s.overlap(Overlap{Identity: owners[other][0], ItemID: target, OtherItemID: other}); err != nil {
			return err
		}
	}
	return nil
}

func setOf[V any](m map[int64]V) map[int64]struct{} {
	set := make(map[int64]struct{}, len(m))
	for id := range m {
		set[id] = struct{}{}
	}
	return set
}

func (s *settling) create(g []record.Identity) error {
	readings, err := s.readingsOf(g)
	if err != nil {
		return err
	}
	kindName := kindOf(readings, noKind)
	if kindName == noKind {
		return s.wait(g)
	}
	k, _ := kind.Get(kindName)
	own, err := decideOwn(k, readings)
	if err != nil {
		return err
	}
	slug, err := FreeSlug(s.ctx, s.q, kindName, Slugify(verdict.Title(own)))
	if err != nil {
		return err
	}
	result, err := s.q.ExecContext(s.ctx, `INSERT INTO item (kind, slug) VALUES (?, ?)`, kindName, slug)
	if err != nil {
		return err
	}
	itemID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	ordered, err := s.keyFirst(g, k)
	if err != nil {
		return err
	}
	for _, id := range ordered {
		if err := s.resolve(itemID, id); err != nil {
			return err
		}
	}
	s.report.Created = append(s.report.Created, itemID)
	s.touched[itemID] = struct{}{}
	return nil
}

type keyRank int

const (
	identityOfTheKind keyRank = iota
	namedByARecord
	recordOfAKind
	anyOther
)

func (s *settling) keyFirst(g []record.Identity, k kind.Kind) ([]record.Identity, error) {
	named := map[record.Identity]struct{}{}
	kinded := map[record.Identity]struct{}{}
	for _, id := range g {
		derived, err := s.derivationOf(id)
		if err != nil {
			return nil, err
		}
		for _, about := range derived.About {
			named[about] = struct{}{}
		}
		if derived.Kind != noKind {
			kinded[id] = struct{}{}
		}
	}
	rankOf := func(id record.Identity) keyRank {
		_, isNamed := named[id]
		_, isKinded := kinded[id]
		switch {
		case id.Source == k.IdentitySource:
			return identityOfTheKind
		case isNamed:
			return namedByARecord
		case isKinded:
			return recordOfAKind
		}
		return anyOther
	}
	ordered := slices.Clone(g)
	slices.SortStableFunc(ordered, func(a, b record.Identity) int { return cmp.Compare(rankOf(a), rankOf(b)) })
	return ordered, nil
}

func (s *settling) resolve(itemID int64, id record.Identity) error {
	if _, err := s.q.ExecContext(s.ctx,
		`INSERT INTO item_identity (item_id, source, external_id) VALUES (?, ?, ?)`,
		itemID, id.Source, id.ExternalID); err != nil {
		return err
	}
	return s.pullWaitingOn(id)
}

func (s *settling) wait(g []record.Identity) error {
	for _, waiting := range g {
		derived, err := s.derivationOf(waiting)
		if err != nil {
			return err
		}
		for _, about := range derived.About {
			if about == waiting {
				continue
			}
			if _, err := s.q.ExecContext(s.ctx, `
				INSERT INTO identity_waiting (about_source, about_external_id, waiting_source, waiting_external_id)
				VALUES (?, ?, ?, ?) ON CONFLICT DO NOTHING`,
				about.Source, about.ExternalID, waiting.Source, waiting.ExternalID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *settling) pullWaitingOn(resolved record.Identity) error {
	rows, err := s.q.QueryContext(s.ctx,
		`SELECT waiting_source, waiting_external_id FROM identity_waiting WHERE about_source = ? AND about_external_id = ?`,
		resolved.Source, resolved.ExternalID)
	if err != nil {
		return err
	}
	for rows.Next() {
		var waiting record.Identity
		if err := rows.Scan(&waiting.Source, &waiting.ExternalID); err != nil {
			rows.Close()
			return err
		}
		s.pulled = append(s.pulled, waiting)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	_, err = s.q.ExecContext(s.ctx, `
		DELETE FROM identity_waiting
		WHERE (about_source = ? AND about_external_id = ?) OR (waiting_source = ? AND waiting_external_id = ?)`,
		resolved.Source, resolved.ExternalID, resolved.Source, resolved.ExternalID)
	return err
}

func (s *settling) overlap(o Overlap) error {
	result, err := s.q.ExecContext(s.ctx,
		`INSERT INTO item_overlap (source, external_id, item_id, other_item_id) VALUES (?, ?, ?, ?)
		 ON CONFLICT DO NOTHING`,
		o.Identity.Source, o.Identity.ExternalID, o.ItemID, o.OtherItemID)
	if err != nil {
		return err
	}
	inserted, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if inserted > 0 {
		s.report.Overlaps = append(s.report.Overlaps, o)
	}
	return nil
}

func kindOf(readings []record.Reading, current string) string {
	if name := latestKind(readings, kindOfIdentitySource); name != noKind {
		return name
	}
	if name := latestKind(readings, derivedKind); name != noKind {
		return name
	}
	return current
}

func kindOfIdentitySource(r record.Reading) string {
	k, isIdentitySource := kind.KindOfIdentitySource(r.Source)
	if !isIdentitySource {
		return noKind
	}
	return k.Name
}

func derivedKind(r record.Reading) string {
	return Derive(r.Identity, r.Values).Kind
}

func latestKind(readings []record.Reading, kindFrom func(record.Reading) string) string {
	latest := noKind
	var latestReading record.Reading
	for _, r := range readings {
		name := kindFrom(r)
		if _, known := kind.Get(name); !known {
			continue
		}
		if latest == noKind || changedLater(r, latestReading) {
			latest, latestReading = name, r
		}
	}
	return latest
}

func changedLater(a, b record.Reading) bool {
	return cmp.Or(cmp.Compare(a.ChangedAt, b.ChangedAt), compareIdentities(a.Identity, b.Identity)) > 0
}

func (s *settling) decideKinds() error {
	for _, itemID := range sortedItems(s.touched) {
		ids, err := identitiesOn(s.ctx, s.q, itemID)
		if err != nil {
			return err
		}
		readings, err := s.readingsOf(ids)
		if err != nil {
			return err
		}
		kindName, err := s.moveIfItsKindChanged(itemID, readings)
		if err != nil {
			return err
		}
		k, known := kind.Get(kindName)
		if !known {
			return unknownKind(itemID, kindName)
		}
		own, err := decideOwn(k, readings)
		if err != nil {
			return err
		}
		s.owns[itemID] = decision{k: k, own: own, origin: decidedNow}
	}
	return nil
}

func (s *settling) moveIfItsKindChanged(itemID int64, readings []record.Reading) (string, error) {
	var current, slug string
	if err := s.q.QueryRowContext(s.ctx, `SELECT kind, slug FROM item WHERE id = ?`, itemID).Scan(&current, &slug); err != nil {
		return "", err
	}
	decided := kindOf(readings, current)
	if decided == current {
		return current, nil
	}
	slug, err := FreeSlug(s.ctx, s.q, decided, slug)
	if err != nil {
		return "", err
	}
	if _, err := s.q.ExecContext(s.ctx, `UPDATE item SET kind = ?, slug = ? WHERE id = ?`, decided, slug, itemID); err != nil {
		return "", err
	}
	s.report.Moves = append(s.report.Moves, Move{ItemID: itemID, From: current, To: decided})
	return decided, nil
}

func identitiesOn(ctx context.Context, q record.Queryer, itemID int64) ([]record.Identity, error) {
	rows, err := q.QueryContext(ctx, `SELECT source, external_id FROM item_identity WHERE item_id = ? ORDER BY id`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []record.Identity{}
	for rows.Next() {
		var id record.Identity
		if err := rows.Scan(&id.Source, &id.ExternalID); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (w *World) Seed(ctx context.Context, q record.Queryer, itemID int64, ids []record.Identity) error {
	for _, id := range ids {
		if _, err := q.ExecContext(ctx,
			`INSERT INTO item_identity (item_id, source, external_id) VALUES (?, ?, ?) ON CONFLICT (source, external_id) DO NOTHING`,
			itemID, id.Source, id.ExternalID); err != nil {
			return err
		}
		owner, err := itemOf(ctx, q, id)
		if err != nil {
			return err
		}
		if owner != itemID {
			return fmt.Errorf("seed item %d: %s belongs to item %d", itemID, id, owner)
		}
	}
	return nil
}

func FreeSlug(ctx context.Context, q record.Queryer, kindName, base string) (string, error) {
	for attempt := 1; ; attempt++ {
		slug := base
		if attempt > 1 {
			slug = fmt.Sprintf("%s-%d", base, attempt)
		}
		var taken int
		if err := q.QueryRowContext(ctx, `SELECT count(*) FROM item WHERE kind = ? AND slug = ?`, kindName, slug).Scan(&taken); err != nil {
			return "", err
		}
		if taken == 0 {
			return slug, nil
		}
	}
}

func Slugify(title string) string {
	withoutAccents, _, _ := transform.String(transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC), title)
	var b strings.Builder
	lastWasDash := true
	for _, r := range strings.ToLower(withoutAccents) {
		switch {
		case r < unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r)):
			b.WriteRune(r)
			lastWasDash = false
		case !lastWasDash:
			b.WriteRune('-')
			lastWasDash = true
		}
	}
	slug := strings.TrimSuffix(b.String(), "-")
	if slug == "" {
		return untitled
	}
	return slug
}
