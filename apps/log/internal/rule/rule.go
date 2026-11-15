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
	proposals, err := p.Propose(ctx, tx)
	if err != nil {
		return Report{}, err
	}
	var report Report
	var stored []Proposal
	for _, proposal := range proposals {
		if proposal.Subject != SubjectEntry {
			stored = append(stored, proposal)
			continue
		}
		newlyLogged, err := logUnlessDismissed(ctx, tx, proposal, visibility)
		if err != nil {
			return Report{}, err
		}
		if newlyLogged {
			report.Logged = append(report.Logged, proposal.ItemID)
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

func logUnlessDismissed(ctx context.Context, q record.Queryer, proposal Proposal, visibility string) (bool, error) {
	var dismissed int
	err := q.QueryRowContext(ctx, `SELECT count(*) FROM dismissal WHERE item_id = ?`, proposal.ItemID).Scan(&dismissed)
	if err != nil || dismissed > 0 {
		return false, err
	}
	logging, err := take.Log(ctx, q, proposal.ItemID, proposal.Rule, visibility)
	return logging == take.NewlyLogged, err
}

func insert(ctx context.Context, q record.Queryer, source string, proposal Proposal) error {
	var changedAt any
	if proposal.ChangedAt != "" {
		changedAt = Normalized(proposal.ChangedAt)
	}
	_, err := q.ExecContext(ctx,
		`INSERT INTO proposal (item_id, source, rule, subject, value, strength, changed_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT (item_id, rule, subject, value) DO NOTHING`,
		proposal.ItemID, source, proposal.Rule, proposal.Subject, proposal.Value, proposal.Strength, changedAt)
	return err
}
