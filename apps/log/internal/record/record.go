package record

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

const SignalClaimRemoved = "claim-removed"

const timeLayout = "2006-01-02T15:04:05Z"

type Queryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type Readability string

const (
	Readable   Readability = "readable"
	Unreadable Readability = "unreadable"
)

type Reader func(source string, raw json.RawMessage) (json.RawMessage, Readability)

type Store struct {
	db             *sql.DB
	read           Reader
	readersVersion int
}

func New(db *sql.DB, read Reader, readersVersion int) *Store {
	return &Store{db: db, read: read, readersVersion: readersVersion}
}

type Change string

const (
	Added     Change = "added"
	Changed   Change = "changed"
	Unchanged Change = "unchanged"
)

type Removal string

const (
	Removed   Removal = "removed"
	WasAbsent Removal = "was-absent"
)

type Observation struct {
	About      Identity
	Source     string
	Signal     string
	Key        string
	ObservedAt string
	Data       json.RawMessage
}

type ObserveOutcome string

const (
	Observed     ObserveOutcome = "observed"
	AlreadyKnown ObserveOutcome = "already-known"
)

type Reading struct {
	Identity
	Values    map[string]any
	FetchedAt string
	ChangedAt string
}

func now() string {
	return time.Now().UTC().Format(timeLayout)
}

func (s *Store) PutClaim(ctx context.Context, q Queryer, id Identity, data json.RawMessage) (Change, error) {
	return s.PutClaimAt(ctx, q, id, data, now())
}

func (s *Store) PutClaimAt(ctx context.Context, q Queryer, id Identity, data json.RawMessage, at string) (Change, error) {
	reading, readability := s.read(id.Source, data)

	var storedData, changedAt string
	err := q.QueryRowContext(ctx,
		`SELECT data, changed_at FROM record_claim WHERE source = ? AND external_id = ?`,
		id.Source, id.ExternalID).Scan(&storedData, &changedAt)
	if errors.Is(err, sql.ErrNoRows) {
		if _, err := q.ExecContext(ctx,
			`INSERT INTO record_claim (source, external_id, data, fetched_at, changed_at) VALUES (?, ?, ?, ?, ?)`,
			id.Source, id.ExternalID, string(data), at, at); err != nil {
			return "", err
		}
		if err := s.writeReading(ctx, q, id, reading, readability); err != nil {
			return "", err
		}
		return Added, nil
	}
	if err != nil {
		return "", err
	}

	storedReading, err := storedReadingOf(ctx, q, id)
	if err != nil {
		return "", err
	}
	change := Unchanged
	if differs(storedReading, storedData, reading, readability, data) {
		change = Changed
		changedAt = at
	}
	if _, err := q.ExecContext(ctx,
		`UPDATE record_claim SET data = ?, fetched_at = ?, changed_at = ? WHERE source = ? AND external_id = ?`,
		string(data), at, changedAt, id.Source, id.ExternalID); err != nil {
		return "", err
	}
	if err := s.writeReading(ctx, q, id, reading, readability); err != nil {
		return "", err
	}
	return change, nil
}

func storedReadingOf(ctx context.Context, q Queryer, id Identity) (*string, error) {
	var stored string
	err := q.QueryRowContext(ctx,
		`SELECT reading FROM record_reading WHERE source = ? AND external_id = ?`,
		id.Source, id.ExternalID).Scan(&stored)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &stored, nil
}

func differs(storedReading *string, storedData string, reading json.RawMessage, readability Readability, data json.RawMessage) bool {
	switch {
	case storedReading != nil && readability == Readable:
		return *storedReading != string(reading)
	case storedReading == nil && readability == Unreadable:
		return storedData != string(data)
	default:
		return true
	}
}

func (s *Store) writeReading(ctx context.Context, q Queryer, id Identity, reading json.RawMessage, readability Readability) error {
	if readability == Unreadable {
		_, err := q.ExecContext(ctx,
			`DELETE FROM record_reading WHERE source = ? AND external_id = ?`,
			id.Source, id.ExternalID)
		return err
	}
	_, err := q.ExecContext(ctx,
		`INSERT INTO record_reading (source, external_id, version, reading) VALUES (?, ?, ?, ?)
		 ON CONFLICT (source, external_id) DO UPDATE SET version = excluded.version, reading = excluded.reading`,
		id.Source, id.ExternalID, s.readersVersion, string(reading))
	return err
}

func (s *Store) RemoveClaim(ctx context.Context, q Queryer, id Identity) (Removal, error) {
	return s.removeClaimAt(ctx, q, id, now())
}

func (s *Store) removeClaimAt(ctx context.Context, q Queryer, id Identity, at string) (Removal, error) {
	var data string
	err := q.QueryRowContext(ctx,
		`SELECT data FROM record_claim WHERE source = ? AND external_id = ?`,
		id.Source, id.ExternalID).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return WasAbsent, nil
	}
	if err != nil {
		return "", err
	}
	if _, err := s.Observe(ctx, q, Observation{
		About:      id,
		Source:     id.Source,
		Signal:     SignalClaimRemoved,
		Key:        SignalClaimRemoved + ":" + id.String() + ":" + at,
		ObservedAt: at,
		Data:       json.RawMessage(data),
	}); err != nil {
		return "", err
	}
	if _, err := q.ExecContext(ctx,
		`DELETE FROM record_claim WHERE source = ? AND external_id = ?`,
		id.Source, id.ExternalID); err != nil {
		return "", err
	}
	return Removed, nil
}

func (s *Store) Observe(ctx context.Context, q Queryer, o Observation) (ObserveOutcome, error) {
	result, err := q.ExecContext(ctx,
		`INSERT INTO record_observation (about_source, about_external_id, source, signal, key, observed_at, data)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT (source, key) DO NOTHING`,
		o.About.Source, o.About.ExternalID, o.Source, o.Signal, o.Key, o.ObservedAt, string(o.Data))
	if err != nil {
		return "", err
	}
	inserted, err := result.RowsAffected()
	if err != nil {
		return "", err
	}
	if inserted == 0 {
		return AlreadyKnown, nil
	}
	return Observed, nil
}

func (s *Store) Readings(ctx context.Context, q Queryer, ids []Identity) ([]Reading, error) {
	readings := make([]Reading, 0, len(ids))
	for _, id := range ids {
		var raw string
		var reading Reading
		err := q.QueryRowContext(ctx,
			`SELECT r.reading, c.fetched_at, c.changed_at
			 FROM record_reading r JOIN record_claim c ON c.source = r.source AND c.external_id = r.external_id
			 WHERE r.source = ? AND r.external_id = ?`,
			id.Source, id.ExternalID).Scan(&raw, &reading.FetchedAt, &reading.ChangedAt)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(raw), &reading.Values); err != nil {
			return nil, err
		}
		reading.Identity = id
		readings = append(readings, reading)
	}
	return readings, nil
}
