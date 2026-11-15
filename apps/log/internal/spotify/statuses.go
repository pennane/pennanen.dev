package spotify

import (
	"context"
	"fmt"
	"slices"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/syncs"
)

func computeArtistStatuses(ctx context.Context, st *store.Store, p *syncs.Progress) error {
	shortTermPrefix := fmt.Sprintf("%s:%s:", SignalTopArtist, ShortTerm)
	listening, err := st.ItemsInNewestSnapshot(ctx, ObservationSource, SignalTopArtist, shortTermPrefix)
	if err != nil {
		return err
	}
	everListed, err := st.ItemsObserved(ctx, ObservationSource, SignalTopArtist)
	if err != nil {
		return err
	}
	for _, itemID := range everListed {
		status := kind.Listened
		if slices.Contains(listening, itemID) {
			status = kind.Listening
		}
		outcome, err := st.SetStatusFromSource(ctx, itemID, kind.SourceSpotify, status)
		if err != nil {
			p.Problem("artist status %d: %v", itemID, err)
			continue
		}
		p.Add(string(outcome), 1)
	}
	return nil
}
