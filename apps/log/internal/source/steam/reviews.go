package steam

import (
	"context"
	"fmt"
	"github.com/pennane/pennanen.dev/apps/log/internal/record/reading"
	"html"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const userAgent = "pennanen.dev log"

const reviewPagesBeyondAnyRealProfile = 100

type Review struct {
	AppID int64
	Take  reading.SteamTake
}

func (c *Client) Reviews(ctx context.Context) ([]Review, error) {
	thisYear := time.Now().UTC().Year()
	reviews := []Review{}
	for page := 1; page <= reviewPagesBeyondAnyRealProfile; page++ {
		body, err := c.getPage(ctx, fmt.Sprintf("%s/profiles/%s/recommended/?p=%d", c.CommunityBaseURL, c.SteamID, page))
		if err != nil {
			return nil, fmt.Errorf("reviews page %d: %w", page, err)
		}
		found, err := reviewsOn(body, thisYear)
		if err != nil {
			return nil, fmt.Errorf("reviews page %d: %w", page, err)
		}
		if len(found) == 0 {
			return reviews, nil
		}
		reviews = append(reviews, found...)
	}
	return nil, fmt.Errorf("reviews: more than %d pages", reviewPagesBeyondAnyRealProfile)
}

func (c *Client) getPage(ctx context.Context, address string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)
	res, err := c.HTTP.Do(req)
	if err != nil {
		return "", withoutURL(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", res.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	return string(body), err
}

const reviewBoxStart = `class="review_box"`

var (
	reviewedApp           = regexp.MustCompile(`steamcommunity\.com/app/(\d+)`)
	thumb                 = regexp.MustCompile(`icon_thumbs(Up|Down)`)
	hoursAtReview         = regexp.MustCompile(`([\d,.]+) hrs at review time`)
	hoursOnRecord         = regexp.MustCompile(`([\d,.]+) hrs on record`)
	reviewText            = regexp.MustCompile(`(?s)<div class="content\s*">(.*?)</div>`)
	postedOn              = regexp.MustCompile(`Posted (\d{1,2} [A-Z][a-z]+)(?:, (\d{4}))?\.`)
	lineBreak             = regexp.MustCompile(`(?i)<br\s*/?>`)
	anyTag                = regexp.MustCompile(`<[^>]*>`)
	recommendationOfThumb = map[string]reading.SteamRecommendation{"Up": reading.Recommended, "Down": reading.NotRecommended}
)

func reviewsOn(page string, thisYear int) ([]Review, error) {
	boxes := strings.Split(page, reviewBoxStart)[1:]
	reviews := make([]Review, 0, len(boxes))
	for _, box := range boxes {
		review, err := reviewIn(box, thisYear)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	return reviews, nil
}

func reviewIn(box string, thisYear int) (Review, error) {
	app := reviewedApp.FindStringSubmatch(box)
	if app == nil {
		return Review{}, fmt.Errorf("a review names no app")
	}
	appID, err := strconv.ParseInt(app[1], 10, 64)
	if err != nil {
		return Review{}, err
	}
	direction := thumb.FindStringSubmatch(box)
	if direction == nil {
		return Review{}, fmt.Errorf("review of %d has no thumb", appID)
	}
	take := reading.SteamTake{
		Recommendation: recommendationOfThumb[direction[1]],
		Review:         textOf(box),
		ReviewedOn:     postedDate(box, thisYear),
		HoursAtReview:  hoursOf(box),
	}
	return Review{AppID: appID, Take: take}, nil
}

func textOf(box string) string {
	content := reviewText.FindStringSubmatch(box)
	if content == nil {
		return ""
	}
	text := html.UnescapeString(anyTag.ReplaceAllString(lineBreak.ReplaceAllString(content[1], "\n"), ""))
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func postedDate(box string, thisYear int) string {
	posted := postedOn.FindStringSubmatch(box)
	if posted == nil {
		return ""
	}
	year := strconv.Itoa(thisYear)
	if posted[2] != "" {
		year = posted[2]
	}
	day, err := time.Parse("2 January 2006", posted[1]+" "+year)
	if err != nil {
		return ""
	}
	return day.Format(time.DateOnly)
}

func hoursOf(box string) *float64 {
	for _, pattern := range []*regexp.Regexp{hoursAtReview, hoursOnRecord} {
		found := pattern.FindStringSubmatch(box)
		if found == nil {
			continue
		}
		hours, err := strconv.ParseFloat(strings.ReplaceAll(found[1], ",", ""), 64)
		if err == nil {
			return &hours
		}
	}
	return nil
}
