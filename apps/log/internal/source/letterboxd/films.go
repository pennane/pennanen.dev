package letterboxd

import (
	"cmp"
	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
	"slices"
)

type Shelf string

const (
	WatchedShelf   Shelf = "watched"
	WatchlistShelf Shelf = "watchlist"
)

type ListedFilm struct {
	Film     reading.LetterboxdFilm
	Rating   *float64
	Reaction reading.LetterboxdReaction
}

type Rewatch string

const (
	FirstWatch Rewatch = "first-watch"
	Rewatched  Rewatch = "rewatch"
)

type DiaryEntry struct {
	Film      reading.LetterboxdFilm
	WatchedOn string
	Rewatch   Rewatch
	Rating    *float64
	Reaction  reading.LetterboxdReaction
	Review    string
}

func oldestFirst(entries []DiaryEntry) []DiaryEntry {
	sorted := slices.Clone(entries)
	slices.SortStableFunc(sorted, func(a, b DiaryEntry) int { return cmp.Compare(a.WatchedOn, b.WatchedOn) })
	return sorted
}

func halfStars(n int) *float64 {
	stars := float64(n) / 2
	return &stars
}
