package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
)

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store { return &Store{db: db} }

type queryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (s *Store) inTx(ctx context.Context, work func(tx *sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := work(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) CreateItem(ctx context.Context, kindName, title string, override map[string]any) (int64, error) {
	k, known := kind.Get(kindName)
	if !known {
		return 0, fmt.Errorf("%w: unknown kind %q", ErrInvalid, kindName)
	}
	if strings.TrimSpace(title) == "" {
		return 0, fmt.Errorf("%w: title is required", ErrInvalid)
	}
	if override == nil {
		override = map[string]any{}
	}
	override[kind.FieldTitle] = title
	for name, value := range override {
		if err := k.ValidateOverride(name, value); err != nil {
			return 0, fmt.Errorf("%w: %v", ErrInvalid, err)
		}
	}
	var id int64
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		var err error
		if id, err = insertItem(ctx, tx, kindName, title); err != nil {
			return err
		}
		if err := putClaim(ctx, tx, id, kind.SourceAP, "", override); err != nil {
			return err
		}
		return ensureEntry(ctx, tx, id)
	})
	return id, err
}

func (s *Store) CreateItemFromSource(ctx context.Context, kindName, title, source, externalID string, data any) (int64, error) {
	var id int64
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		var err error
		if id, err = insertItem(ctx, tx, kindName, title); err != nil {
			return err
		}
		if err := putClaim(ctx, tx, id, source, externalID, data); err != nil {
			return err
		}
		return ensureEntry(ctx, tx, id)
	})
	return id, err
}

func insertItem(ctx context.Context, q queryer, kindName, title string) (int64, error) {
	base := Slugify(title)
	for attempt := 1; ; attempt++ {
		slug := base
		if attempt > 1 {
			slug = fmt.Sprintf("%s-%d", base, attempt)
		}
		var taken int
		if err := q.QueryRowContext(ctx, `SELECT count(*) FROM item WHERE kind = ? AND slug = ?`, kindName, slug).Scan(&taken); err != nil {
			return 0, err
		}
		if taken > 0 {
			continue
		}
		result, err := q.ExecContext(ctx, `INSERT INTO item (kind, slug) VALUES (?, ?)`, kindName, slug)
		if err != nil {
			return 0, err
		}
		return result.LastInsertId()
	}
}

func ensureEntry(ctx context.Context, q queryer, itemID int64) error {
	_, err := q.ExecContext(ctx, `INSERT INTO entry (item_id) VALUES (?) ON CONFLICT DO NOTHING`, itemID)
	return err
}

func (s *Store) EnsureEntry(ctx context.Context, itemID int64) error {
	return ensureEntry(ctx, s.db, itemID)
}

const untitled = "untitled"

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

func (s *Store) FindByIdentity(ctx context.Context, source, externalID string) (int64, error) {
	var id int64
	err := s.db.QueryRowContext(ctx, `SELECT item_id FROM claim WHERE source = ? AND external_id = ?`, source, externalID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	return id, err
}

func (s *Store) PutClaim(ctx context.Context, itemID int64, source, externalID string, data any) error {
	return putClaim(ctx, s.db, itemID, source, externalID, data)
}

func putClaim(ctx context.Context, q queryer, itemID int64, source, externalID string, data any) error {
	raw, err := asJSON(data)
	if err != nil {
		return err
	}
	_, err = q.ExecContext(ctx, `
		INSERT INTO claim (item_id, source, external_id, data) VALUES (?, ?, NULLIF(?, ''), ?)
		ON CONFLICT (item_id, source) DO UPDATE SET
			external_id = excluded.external_id,
			data = excluded.data,
			fetched_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')`,
		itemID, source, externalID, string(raw))
	return err
}

func asJSON(data any) ([]byte, error) {
	switch d := data.(type) {
	case json.RawMessage:
		return d, nil
	case []byte:
		return d, nil
	}
	return json.Marshal(data)
}

func (s *Store) Claim(ctx context.Context, itemID int64, source string) (verdict.Claim, error) {
	c := verdict.Claim{Source: source}
	var externalID sql.NullString
	var data string
	err := s.db.QueryRowContext(ctx, `SELECT external_id, data, fetched_at FROM claim WHERE item_id = ? AND source = ?`, itemID, source).
		Scan(&externalID, &data, &c.FetchedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return c, ErrNotFound
	}
	c.ExternalID = externalID.String
	c.Data = json.RawMessage(data)
	return c, err
}

func (s *Store) ChangeOverride(ctx context.Context, itemID int64, changes map[string]any) error {
	k, err := lookupKind(ctx, s.db, itemID)
	if err != nil {
		return err
	}
	for name, value := range changes {
		if err := k.ValidateOverride(name, value); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalid, err)
		}
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		override := map[string]any{}
		var raw string
		err := tx.QueryRowContext(ctx, `SELECT data FROM claim WHERE item_id = ? AND source = ?`, itemID, kind.SourceAP).Scan(&raw)
		switch {
		case errors.Is(err, sql.ErrNoRows):
		case err != nil:
			return err
		default:
			if err := json.Unmarshal([]byte(raw), &override); err != nil {
				return err
			}
		}
		for name, value := range changes {
			if kind.IsRemoval(value) {
				delete(override, name)
			} else {
				override[name] = value
			}
		}
		return putClaim(ctx, tx, itemID, kind.SourceAP, "", override)
	})
}

func (s *Store) DeleteItem(ctx context.Context, itemID int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM item WHERE id = ?`, itemID)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func requireAffected(result sql.Result) error {
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
