package store_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/store"
)

func TestKindsKeepTheOrderApGaveThemAndNewOnesFollow(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	order, err := s.KindOrder(ctx)
	if err != nil || !slices.Equal(order, store.KindOrder{"game", "movie", "series"}) {
		t.Fatalf("default order %v %v", order, err)
	}
	if _, err := s.SetKindOrder(ctx, store.KindOrder{"movie", "series"}); err != nil {
		t.Fatal(err)
	}
	if order, _ := s.KindOrder(ctx); !slices.Equal(order, store.KindOrder{"movie", "series", "game"}) {
		t.Errorf("order %v", order)
	}
	if _, err := s.SetKindOrder(ctx, store.KindOrder{"movie", "movie"}); !errors.Is(err, store.ErrInvalid) {
		t.Errorf("twice: %v", err)
	}
	if _, err := s.SetKindOrder(ctx, store.KindOrder{"wine"}); !errors.Is(err, store.ErrInvalid) {
		t.Errorf("unknown: %v", err)
	}
}
