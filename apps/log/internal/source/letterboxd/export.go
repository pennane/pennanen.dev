package letterboxd

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
	"io"
	"strconv"
	"strings"
)

type Export struct {
	Watched   []ListedFilm
	Watchlist []reading.LetterboxdFilm
	Diary     []DiaryEntry
}

const (
	watchedFile   = "watched.csv"
	ratingsFile   = "ratings.csv"
	likedFile     = "likes/films.csv"
	watchlistFile = "watchlist.csv"
	diaryFile     = "diary.csv"
	reviewsFile   = "reviews.csv"
	shortLinkHead = "https://boxd.it/"
)

var ErrNotAnExport = errors.New("this zip has no watched.csv; is it a Letterboxd export?")

type row map[string]string

type nameAndYear struct {
	name string
	year int
}

func ParseExport(archive []byte) (Export, error) {
	z, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return Export{}, fmt.Errorf("reading the zip: %w", err)
	}
	files := map[string][]row{}
	for _, name := range []string{watchedFile, ratingsFile, likedFile, watchlistFile, diaryFile, reviewsFile} {
		rows, err := readCSV(z, name)
		if err != nil {
			return Export{}, err
		}
		files[name] = rows
	}
	if files[watchedFile] == nil {
		return Export{}, ErrNotAnExport
	}

	watched := map[string]*ListedFilm{}
	order := []string{}
	lidOf := map[nameAndYear]string{}
	see := func(r row) *ListedFilm {
		film := filmOf(r)
		if film.ID == "" {
			return nil
		}
		lidOf[nameAndYear{film.Title, film.Year}] = film.ID
		if listed, seen := watched[film.ID]; seen {
			return listed
		}
		watched[film.ID] = &ListedFilm{Film: film}
		order = append(order, film.ID)
		return watched[film.ID]
	}
	for _, r := range files[watchedFile] {
		see(r)
	}
	for _, r := range files[ratingsFile] {
		if listed := see(r); listed != nil {
			listed.Rating = ratingOf(r)
		}
	}
	for _, r := range files[likedFile] {
		if listed := see(r); listed != nil {
			listed.Reaction = reading.Liked
		}
	}

	export := Export{Watched: []ListedFilm{}, Watchlist: []reading.LetterboxdFilm{}}
	for _, id := range order {
		export.Watched = append(export.Watched, *watched[id])
	}
	for _, r := range files[watchlistFile] {
		if film := filmOf(r); film.ID != "" {
			export.Watchlist = append(export.Watchlist, film)
		}
	}
	export.Diary = diaryOf(files[diaryFile], files[reviewsFile], lidOf)
	return export, nil
}

func diaryOf(diary, reviews []row, lidOf map[nameAndYear]string) []DiaryEntry {
	type entryKey struct {
		film      nameAndYear
		watchedOn string
	}
	keyOf := func(r row) entryKey {
		year, _ := strconv.Atoi(r["Year"])
		return entryKey{nameAndYear{r["Name"], year}, watchedOn(r)}
	}
	reviewOf := map[entryKey]row{}
	for _, r := range reviews {
		reviewOf[keyOf(r)] = r
	}
	entries := []DiaryEntry{}
	add := func(r row, review string) {
		key := keyOf(r)
		e := DiaryEntry{
			Film:      reading.LetterboxdFilm{ID: lidOf[key.film], Title: key.film.name, Year: key.film.year},
			WatchedOn: key.watchedOn,
			Rewatch:   FirstWatch,
			Rating:    ratingOf(r),
			Review:    strings.TrimSpace(review),
		}
		if r["Rewatch"] == "Yes" {
			e.Rewatch = Rewatched
		}
		entries = append(entries, e)
	}
	for _, r := range diary {
		review := reviewOf[keyOf(r)]
		delete(reviewOf, keyOf(r))
		add(r, review["Review"])
	}
	for _, r := range reviews {
		if _, unpaired := reviewOf[keyOf(r)]; unpaired {
			add(r, r["Review"])
		}
	}
	return entries
}

func readCSV(z *zip.Reader, name string) ([]row, error) {
	f, err := z.Open(name)
	if err != nil {
		return nil, nil
	}
	defer f.Close()
	records, err := csv.NewReader(io.LimitReader(f, 16<<20)).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	if len(records) == 0 {
		return []row{}, nil
	}
	header, rows := records[0], []row{}
	for _, record := range records[1:] {
		r := row{}
		for i, column := range header {
			if i < len(record) {
				r[column] = record[i]
			}
		}
		rows = append(rows, r)
	}
	return rows, nil
}

func filmOf(r row) reading.LetterboxdFilm {
	year, _ := strconv.Atoi(r["Year"])
	return reading.LetterboxdFilm{
		ID:    strings.Trim(strings.TrimPrefix(r["Letterboxd URI"], shortLinkHead), "/"),
		Title: r["Name"],
		Year:  year,
	}
}

func ratingOf(r row) *float64 {
	rating, err := strconv.ParseFloat(r["Rating"], 64)
	if err != nil || rating <= 0 {
		return nil
	}
	return &rating
}

func watchedOn(r row) string {
	if day := r["Watched Date"]; day != "" {
		return day
	}
	return r["Date"]
}
