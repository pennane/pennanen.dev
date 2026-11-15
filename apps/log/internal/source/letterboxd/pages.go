package letterboxd

import (
	"errors"
	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
	"html"
	"regexp"
	"strconv"
	"strings"
)

var (
	gridItem       = `<li class="griditem">`
	itemSlug       = regexp.MustCompile(`data-item-slug="([^"]+)"`)
	itemName       = regexp.MustCompile(`data-item-name="([^"]+)"`)
	itemLID        = regexp.MustCompile(`"lid":"([^"]+)"`)
	itemRating     = regexp.MustCompile(`\brated-(\d+)\b`)
	itemLiked      = "icon-liked"
	titleWithYear  = regexp.MustCompile(`^(.*) \((\d{4})\)$`)
	ErrNoFilmsSeen = errors.New("found no films on the page; its markup may have changed")
)

func parseFilmGrid(page string) ([]ListedFilm, error) {
	chunks := strings.Split(page, gridItem)
	films := []ListedFilm{}
	for _, chunk := range chunks[1:] {
		chunk = html.UnescapeString(chunk)
		lid, slug, name := firstGroup(itemLID, chunk), firstGroup(itemSlug, chunk), firstGroup(itemName, chunk)
		if lid == "" || name == "" {
			continue
		}
		title, year := splitTitleAndYear(name)
		listed := ListedFilm{Film: reading.LetterboxdFilm{ID: lid, Slug: slug, Title: title, Year: year}}
		if n, err := strconv.Atoi(firstGroup(itemRating, chunk)); err == nil && n > 0 {
			listed.Rating = halfStars(n)
		}
		if strings.Contains(chunk, itemLiked) {
			listed.Reaction = reading.Liked
		}
		films = append(films, listed)
	}
	if len(films) == 0 {
		return nil, ErrNoFilmsSeen
	}
	return films, nil
}

func splitTitleAndYear(name string) (string, int) {
	m := titleWithYear.FindStringSubmatch(name)
	if m == nil {
		return name, 0
	}
	year, _ := strconv.Atoi(m[2])
	return m[1], year
}

func firstGroup(re *regexp.Regexp, s string) string {
	if m := re.FindStringSubmatch(s); m != nil {
		return m[1]
	}
	return ""
}
