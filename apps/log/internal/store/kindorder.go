package store

import (
	"context"
	"fmt"
	"slices"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
)

const kindOrderKey = "kind-order"

type KindOrder []string

func (s *Store) KindOrder(ctx context.Context) (KindOrder, error) {
	stored := KindOrder{}
	if _, err := s.readSetting(ctx, kindOrderKey, &stored); err != nil {
		return nil, err
	}
	return stored.completed(), nil
}

func (s *Store) SetKindOrder(ctx context.Context, order KindOrder) (KindOrder, error) {
	for i, name := range order {
		if _, known := kind.Get(name); !known {
			return nil, fmt.Errorf("%w: no kind %q", ErrInvalid, name)
		}
		if slices.Contains(order[:i], name) {
			return nil, fmt.Errorf("%w: %q is ordered twice", ErrInvalid, name)
		}
	}
	return order.completed(), s.writeSetting(ctx, kindOrderKey, order)
}

func (o KindOrder) completed() KindOrder {
	known := KindOrder{}
	for _, name := range o {
		if _, exists := kind.Get(name); exists {
			known = append(known, name)
		}
	}
	for _, k := range kind.All() {
		if !slices.Contains(known, k.Name) {
			known = append(known, k.Name)
		}
	}
	return known
}
