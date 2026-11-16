package store_test

import (
	"context"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
)

func knowItem(ctx context.Context, s *store.Store, id record.Identity, data any) (int64, error) {
	if err := s.PutClaim(ctx, id, data); err != nil {
		return 0, err
	}
	return s.FindByIdentity(ctx, id.Source, id.ExternalID)
}

func putClaimOnItem(ctx context.Context, s *store.Store, itemID int64, id record.Identity, data any) error {
	key, err := store.KeyOf(ctx, s, itemID)
	if err != nil {
		return err
	}
	return s.PutClaimAbout(ctx, id, key, data)
}

func observeOnItem(ctx context.Context, s *store.Store, o store.Observation) (store.ObserveOutcome, error) {
	key, err := store.KeyOf(ctx, s, o.ItemID)
	if err != nil {
		return "", err
	}
	return s.Observe(ctx, record.Observation{About: key, Source: o.Source, Signal: o.Signal, Key: o.Key, ObservedAt: o.ObservedAt, Data: o.Data})
}

func claimOnItem(ctx context.Context, s *store.Store, itemID int64, source string) (reading.Claim, error) {
	key, err := store.KeyOf(ctx, s, itemID)
	if err != nil {
		return reading.Claim{Source: source}, err
	}
	return s.ClaimOf(ctx, key, source)
}
