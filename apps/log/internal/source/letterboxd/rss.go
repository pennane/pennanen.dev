package letterboxd

import (
	"encoding/xml"
	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const reviewGUIDHead = "letterboxd-review-"

type rssItem struct {
	GUID        string `xml:"guid"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description"`
	WatchedDate string `xml:"https://letterboxd.com watchedDate"`
	Rewatch     string `xml:"https://letterboxd.com rewatch"`
	FilmTitle   string `xml:"https://letterboxd.com filmTitle"`
	FilmYear    string `xml:"https://letterboxd.com filmYear"`
	Rating      string `xml:"https://letterboxd.com memberRating"`
	Like        string `xml:"https://letterboxd.com memberLike"`
	MovieID     string `xml:"https://themoviedb.org movieId"`
}

var (
	filmSlugInLink   = regexp.MustCompile(`/film/([^/]+)/`)
	posterParagraph  = regexp.MustCompile(`<p>\s*<img[^>]*>\s*</p>`)
	spoilerParagraph = regexp.MustCompile(`<p>\s*<em>This review may contain spoilers\.</em>\s*</p>`)
	paragraphBreak   = regexp.MustCompile(`</p>\s*<p>`)
	lineBreak        = regexp.MustCompile(`<br\s*/?>`)
	anyTag           = regexp.MustCompile(`<[^>]+>`)
)

func parseDiary(feed []byte) ([]DiaryEntry, error) {
	var doc struct {
		Items []rssItem `xml:"channel>item"`
	}
	if err := xml.Unmarshal(feed, &doc); err != nil {
		return nil, err
	}
	entries := []DiaryEntry{}
	for _, it := range doc.Items {
		if it.FilmTitle == "" {
			continue
		}
		entries = append(entries, diaryEntryOf(it))
	}
	return entries, nil
}

func diaryEntryOf(it rssItem) DiaryEntry {
	year, _ := strconv.Atoi(it.FilmYear)
	tmdbID, _ := strconv.ParseInt(it.MovieID, 10, 64)
	e := DiaryEntry{
		Film:      reading.LetterboxdFilm{Slug: firstGroup(filmSlugInLink, it.Link), Title: it.FilmTitle, Year: year, TMDBID: tmdbID},
		WatchedOn: it.WatchedDate,
		Rewatch:   FirstWatch,
	}
	if e.WatchedOn == "" {
		e.WatchedOn = dayOf(it.PubDate)
	}
	if it.Rewatch == "Yes" {
		e.Rewatch = Rewatched
	}
	if rating, err := strconv.ParseFloat(it.Rating, 64); err == nil && rating > 0 {
		e.Rating = &rating
	}
	if it.Like == "Yes" {
		e.Reaction = reading.Liked
	}
	if strings.HasPrefix(it.GUID, reviewGUIDHead) {
		e.Review = reviewText(it.Description)
	}
	return e
}

func reviewText(description string) string {
	s := posterParagraph.ReplaceAllString(description, "")
	s = spoilerParagraph.ReplaceAllString(s, "")
	s = paragraphBreak.ReplaceAllString(s, "\n\n")
	s = lineBreak.ReplaceAllString(s, "\n")
	s = anyTag.ReplaceAllString(s, "")
	return strings.TrimSpace(html.UnescapeString(s))
}

func dayOf(pubDate string) string {
	t, err := time.Parse(time.RFC1123Z, strings.TrimSpace(pubDate))
	if err != nil {
		return ""
	}
	return t.Format(time.DateOnly)
}
