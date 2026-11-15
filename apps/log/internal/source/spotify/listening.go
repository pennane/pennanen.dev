package spotify

import (
	"context"
	"math"
	"slices"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/source/syncs"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

const (
	pointsPerPlay   = 0.1
	CountWeighed    = "listening-weighed"
	wholePercentile = 100
)

var rangeFactors = map[TimeRange]float64{LongTerm: 3, MediumTerm: 2, ShortTerm: 1}

func weighed(points map[string]float64) float64 {
	weight := 0.0
	for listRange, sum := range points {
		weight += rangeFactors[TimeRange(listRange)] * sum
	}
	return weight
}

func listeningWeights(ctx context.Context, st *store.Store) (albums, artists map[int64]float64, err error) {
	trackPoints, err := st.TopListPoints(ctx, ObservationSource, record.SignalTopTrack)
	if err != nil {
		return nil, nil, err
	}
	artistPoints, err := st.TopListPoints(ctx, ObservationSource, record.SignalTopArtist)
	if err != nil {
		return nil, nil, err
	}
	plays, err := st.ObservationCounts(ctx, ObservationSource, record.SignalPlay)
	if err != nil {
		return nil, nil, err
	}
	mainArtists, err := st.RelationsOfType(ctx, kind.LinkBy)
	if err != nil {
		return nil, nil, err
	}
	albums = map[int64]float64{}
	for id, points := range trackPoints {
		albums[id] = weighed(points)
	}
	for id, count := range plays {
		albums[id] += float64(count) * pointsPerPlay
	}
	artists = map[int64]float64{}
	for id, points := range artistPoints {
		artists[id] = weighed(points)
	}
	for _, r := range mainArtists {
		artists[r.ToID] += albums[r.FromID]
	}
	return albums, artists, nil
}

func percentiles(ids []int64, weights map[int64]float64) map[int64]int64 {
	sorted := make([]float64, 0, len(ids))
	for _, id := range ids {
		sorted = append(sorted, weights[id])
	}
	slices.Sort(sorted)
	out := make(map[int64]int64, len(ids))
	for _, id := range ids {
		if len(ids) == 1 {
			out[id] = wholePercentile
			continue
		}
		below, _ := slices.BinarySearch(sorted, weights[id])
		out[id] = int64(math.Round(wholePercentile * float64(below) / float64(len(ids)-1)))
	}
	return out
}

func computeListening(ctx context.Context, st *store.Store, p *syncs.Progress) error {
	albums, artists, err := listeningWeights(ctx, st)
	if err != nil {
		return err
	}
	for kindName, weights := range map[string]map[int64]float64{kind.Album: albums, kind.Artist: artists} {
		logged, err := st.LoggedItems(ctx, kindName)
		if err != nil {
			return err
		}
		for id, listening := range percentiles(logged, weights) {
			if err := st.PutClaim(ctx, id, record.SourceSpotifyListening, "", listeningClaim(listening)); err != nil {
				p.Problem("listening %d: %v", id, err)
				continue
			}
			p.Add(CountWeighed, 1)
		}
	}
	return nil
}

func listeningClaim(listening int64) map[string]int64 {
	return map[string]int64{kind.FieldListening: listening}
}
