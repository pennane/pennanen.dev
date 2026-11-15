package rule

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/take"
)

type Subject string

const (
	SubjectStatus    Subject = "status"
	SubjectTag       Subject = "tag"
	SubjectDismissal Subject = "dismissal"
	SubjectEntry     Subject = "entry"
)

const EntryLogged = "logged"

type Strength string

const (
	Applying   Strength = "applying"
	Suggesting Strength = "suggesting"
)

type Proposal struct {
	ItemID    int64
	Rule      string
	Subject   Subject
	Value     string
	Strength  Strength
	ChangedAt string
}

type Proposer struct {
	Source  string
	Propose func(ctx context.Context, q record.Queryer) ([]Proposal, error)
}

type Report struct {
	Logged    []int64
	Proposals int
}

const dateLayout = "2006-01-02"

func Normalized(at string) string {
	day, err := time.Parse(dateLayout, at)
	if err != nil {
		return at
	}
	return day.UTC().Format(time.RFC3339)
}

func RunAll(ctx context.Context, d *sql.DB, proposers []Proposer, visibility string) (Report, error) {
	var total Report
	for _, p := range proposers {
		report, err := Run(ctx, d, p, visibility)
		if err != nil {
			return total, fmt.Errorf("rule source %s: %w", p.Source, err)
		}
		total.Logged = append(total.Logged, report.Logged...)
		total.Proposals += report.Proposals
	}
	return total, nil
}

func Run(ctx context.Context, d *sql.DB, p Proposer, visibility string) (Report, error) {
	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		return Report{}, err
	}
	defer tx.Rollback()
	var report Report
	var stored []Proposal
	for {
		proposals, err := p.Propose(ctx, tx)
		if err != nil {
			return Report{}, err
		}
		var newlyLogged []int64
		stored, newlyLogged, err = logEntries(ctx, tx, proposals, visibility)
		if err != nil {
			return Report{}, err
		}
		report.Logged = append(report.Logged, newlyLogged...)
		if len(newlyLogged) == 0 {
			break
		}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM proposal WHERE source = ?`, p.Source); err != nil {
		return Report{}, err
	}
	for _, proposal := range stored {
		if err := insert(ctx, tx, p.Source, proposal); err != nil {
			return Report{}, err
		}
	}
	report.Proposals = len(stored)
	return report, tx.Commit()
}

func logEntries(ctx context.Context, q record.Queryer, proposals []Proposal, visibility string) ([]Proposal, []int64, error) {
	var stored []Proposal
	var newlyLogged []int64
	for _, proposal := range proposals {
		if proposal.Subject != SubjectEntry {
			stored = append(stored, proposal)
			continue
		}
		outcome, err := logUnlessDismissed(ctx, q, proposal, visibility)
		if err != nil {
			return nil, nil, err
		}
		if outcome == entryLogged {
			newlyLogged = append(newlyLogged, proposal.ItemID)
		}
	}
	return stored, newlyLogged, nil
}

type entryOutcome string

const (
	entryLogged    entryOutcome = "logged"
	entryKept      entryOutcome = "kept"
	entryDismissed entryOutcome = "dismissed"
)

func logUnlessDismissed(ctx context.Context, q record.Queryer, proposal Proposal, visibility string) (entryOutcome, error) {
	var dismissed int
	err := q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM dismissal WHERE item_id = ?)`, proposal.ItemID).Scan(&dismissed)
	if err != nil {
		return "", err
	}
	if dismissed == 1 {
		return entryDismissed, nil
	}
	logging, err := take.Log(ctx, q, proposal.ItemID, proposal.Rule, visibility)
	if err != nil {
		return "", err
	}
	if logging == take.NewlyLogged {
		return entryLogged, nil
	}
	return entryKept, nil
}

func insert(ctx context.Context, q record.Queryer, source string, proposal Proposal) error {
	var changedAt any
	if proposal.ChangedAt != "" {
		changedAt = Normalized(proposal.ChangedAt)
	}
	_, err := q.ExecContext(ctx,
		`INSERT INTO proposal (item_id, source, rule, subject, value, strength, changed_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		proposal.ItemID, source, proposal.Rule, proposal.Subject, proposal.Value, proposal.Strength, changedAt)
	return err
}
