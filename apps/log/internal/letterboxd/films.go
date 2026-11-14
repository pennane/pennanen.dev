package letterboxd

import (
	"cmp"
	"slices"

	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
)

type Shelf string

const (
	WatchedShelf   Shelf = "watched"
	WatchlistShelf Shelf = "watchlist"
)

type ListedFilm struct {
	Film     verdict.LetterboxdFilm
	Rating   *float64
	Reaction verdict.LetterboxdReaction
}

type Rewatch string

const (
	FirstWatch Rewatch = "first-watch"
	Rewatched  Rewatch = "rewatch"
)

type DiaryEntry struct {
	Film      verdict.LetterboxdFilm
	WatchedOn string
	Rewatch   Rewatch
	Rating    *float64
	Reaction  verdict.LetterboxdReaction
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
