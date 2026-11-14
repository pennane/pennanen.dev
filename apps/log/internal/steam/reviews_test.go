package steam

import (
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/verdict"
)

const reviewsPage = `<div class="review_box">
	<div class="leftcol"><a href="https://steamcommunity.com/app/1091500"></a></div>
	<div class="thumb"><img src="https://community.fastly.steamstatic.com/public/shared/images/userreviews/icon_thumbsUp.png"></div>
	<div class="title"><a href="#">Recommended</a></div>
	<div class="hours">
		1,096.7 hrs on record			(23.1 hrs at review time)		</div>
	<div class="content ">crashes alot &quot;but&quot; very niice. <br><br>10/10			</div>
	<div class="posted">
		Posted 11 December, 2020.			Last edited 13 December, 2020.	</div>
</div>
<div class="review_box">
	<div class="leftcol"><a href="https://steamcommunity.com/app/620"></a></div>
	<div class="thumb"><img src="https://community.fastly.steamstatic.com/public/shared/images/userreviews/icon_thumbsDown.png"></div>
	<div class="hours">
		6.2 hrs on record			</div>
	<div class="content ">ass gameplay			</div>
	<div class="posted">
		Posted 16 August.	</div>
</div>`

func TestAReviewsPageGivesEachReviewsThumbTextDateAndHours(t *testing.T) {
	reviews, err := reviewsOn(reviewsPage, 2026)
	if err != nil {
		t.Fatal(err)
	}
	if len(reviews) != 2 {
		t.Fatalf("reviews %+v", reviews)
	}
	first, second := reviews[0], reviews[1]
	if first.AppID != 1091500 || first.Take.Recommendation != verdict.Recommended ||
		first.Take.Review != "crashes alot \"but\" very niice.\n\n10/10" ||
		first.Take.ReviewedOn != "2020-12-11" || *first.Take.HoursAtReview != 23.1 {
		t.Errorf("first %+v", first.Take)
	}
	if second.AppID != 620 || second.Take.Recommendation != verdict.NotRecommended ||
		second.Take.ReviewedOn != "2026-08-16" || *second.Take.HoursAtReview != 6.2 {
		t.Errorf("second %+v", second.Take)
	}
}

func TestAPageWithoutReviewsEndsTheList(t *testing.T) {
	reviews, err := reviewsOn(`<div>Showing 21-14 of 14 entries</div>`, 2026)
	if err != nil || len(reviews) != 0 {
		t.Errorf("reviews %+v err %v", reviews, err)
	}
}
