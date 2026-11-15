package spotify

import (
	"context"
	"fmt"
	"slices"

	"github.com/pennane/pennanen.dev/apps/log/internal/kind"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/syncs"
)

const (
	artistCountPrefix = "artist-"
	albumCountPrefix  = "album-"
)

func distinct(groups ...[]int64) []int64 {
	return slices.Compact(slices.Sorted(slices.Values(slices.Concat(groups...))))
}

func setComputed(ctx context.Context, st *store.Store, p *syncs.Progress, countPrefix string, itemID int64, status string) {
	outcome, err := st.SetStatusFromSource(ctx, itemID, kind.SourceSpotify, status)
	if err != nil {
		p.Problem("%sstatus %d: %v", countPrefix, itemID, err)
		return
	}
	p.Add(countPrefix+string(outcome), 1)
}

func computeAlbumStatuses(ctx context.Context, st *store.Store, p *syncs.Progress) error {
	saved, err := st.ItemsObserved(ctx, ObservationSource, SignalSaved)
	if err != nil {
		return err
	}
	inTopTracks, err := st.ItemsObserved(ctx, ObservationSource, SignalTopTrack)
	if err != nil {
		return err
	}
	played, err := st.ItemsObserved(ctx, ObservationSource, SignalPlay)
	if err != nil {
		return err
	}
	heard := distinct(inTopTracks, played)
	for _, itemID := range saved {
		status := kind.Want
		if slices.Contains(heard, itemID) {
			status = kind.Listened
		}
		setComputed(ctx, st, p, albumCountPrefix, itemID, status)
	}
	return nil
}

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
	handPicked, err := st.ItemsWithStatusSetBy(ctx, kind.SourceSpotify, kind.Artist)
	if err != nil {
		return err
	}
	saved, err := st.ItemsObserved(ctx, ObservationSource, SignalSaved)
	if err != nil {
		return err
	}
	mainArtistsOfSaved, err := st.RelatedTo(ctx, saved, kind.RelationBy)
	if err != nil {
		return err
	}
	for _, itemID := range distinct(everListed, handPicked, mainArtistsOfSaved) {
		status := kind.Listened
		if slices.Contains(listening, itemID) {
			status = kind.Listening
		}
		setComputed(ctx, st, p, artistCountPrefix, itemID, status)
	}
	return nil
}
