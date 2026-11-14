package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/store"
)

func TestSecretIsMissingUntilPutThenReplaced(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	if _, err := s.Secret(ctx, "spotify-refresh-token"); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("unset secret err = %v", err)
	}
	for _, value := range []string{"a", "b"} {
		if err := s.PutSecret(ctx, "spotify-refresh-token", value); err != nil {
			t.Fatal(err)
		}
		got, err := s.Secret(ctx, "spotify-refresh-token")
		if err != nil || got != value {
			t.Fatalf("secret = %q, %v; want %q", got, err, value)
		}
	}
}

func TestDeletedSecretIsMissing(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	if err := s.PutSecret(ctx, "k", "v"); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteSecret(ctx, "k"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Secret(ctx, "k"); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}
