package store

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"slices"

	"github.com/pennane/pennanen.dev/apps/log/internal/rule"
)

func (s *Store) Propose(ctx context.Context, source string) error {
	i := slices.IndexFunc(s.proposers, func(p rule.Proposer) bool { return p.Source == source })
	if i < 0 {
		return fmt.Errorf("%w: no rules for source %q", ErrNotFound, source)
	}
	return s.reported(rule.Run(ctx, s.db, s.proposers[i], string(NewEntryVisibility)))
}

func (s *Store) ProposeAll(ctx context.Context) error {
	return s.reported(rule.RunAll(ctx, s.db, s.proposers, string(NewEntryVisibility)))
}

func (s *Store) reported(report rule.Report, err error) error {
	if err != nil {
		return err
	}
	if len(report.Logged) > 0 {
		slog.Info("rules logged items", "count", len(report.Logged))
	}
	return nil
}

func (s *Store) inTxThenPropose(ctx context.Context, work func(tx *sql.Tx) error) error {
	if err := s.inTx(ctx, work); err != nil {
		return err
	}
	if err := s.ProposeAll(ctx); err != nil {
		return fmt.Errorf("saved, but proposing again failed: %w", err)
	}
	return nil
}

func (s *Store) attachProposals(ctx context.Context, ids []int64, byID map[int64]*ItemView) error {
	for _, v := range byID {
		v.Suggestions = []Suggestion{}
	}
	return forChunks(ids, func(chunk []int64) error {
		rows, err := s.db.QueryContext(ctx, `
			SELECT item_id, rule, subject, value, strength, coalesce(changed_at, '') FROM proposal
			WHERE item_id IN (`+placeholders(len(chunk))+`) AND subject IN (?, ?, ?)
			ORDER BY item_id, rule, subject, value`,
			append(int64sAsArgs(chunk), rule.SubjectStatus, rule.SubjectTag, rule.SubjectDismissal)...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id int64
			var p rule.Proposal
			if err := rows.Scan(&id, &p.Rule, &p.Subject, &p.Value, &p.Strength, &p.ChangedAt); err != nil {
				return err
			}
			attachProposal(byID[id], p)
		}
		return rows.Err()
	})
}

func attachProposal(v *ItemView, p rule.Proposal) {
	switch {
	case p.Strength == rule.Suggesting:
		v.Suggestions = append(v.Suggestions, Suggestion{Rule: p.Rule, Subject: p.Subject, Value: p.Value})
	case p.Subject == rule.SubjectStatus && v.Entry != nil:
		v.Entry.StatusProposals = append(v.Entry.StatusProposals, StatusProposal{Rule: p.Rule, Value: p.Value, ChangedAt: p.ChangedAt})
	}
}
