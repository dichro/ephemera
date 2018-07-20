package cmd

import (
	"github.com/ChimeraCoder/anaconda"
	"testing"
	"time"
)

func TestKeepTweets(t *testing.T) {
	p := Policy{MaxAge: 5 * time.Hour,
		MinStars:    4,
		MinRetweets: 4,
		KeepMedia:   true}
	now, _ := time.Parse(time.RubyDate, "Tue Jul 17 20:15:43 -0700 2018")

	var tests = []struct {
		tweet anaconda.Tweet
		want  bool
	}{
		{anaconda.Tweet{CreatedAt: "fullofweasels"}, true},
		{anaconda.Tweet{CreatedAt: "Tue Jul 17 20:15:43 -0700 2018"}, true},
		{anaconda.Tweet{CreatedAt: "Tue Jul 17 14:15:43 -0700 2018"}, false},
		{anaconda.Tweet{CreatedAt: "Tue Jul 17 20:15:43 -0700 2016"}, false},
		{anaconda.Tweet{RetweetCount: 3,
			CreatedAt: "Tue Jul 17 20:15:43 -0700 2016"}, false},
		{anaconda.Tweet{FavoriteCount: 3,
			CreatedAt: "Tue Jul 17 20:15:43 -0700 2016"}, false},
		{anaconda.Tweet{RetweetCount: 5,
			CreatedAt: "Tue Jul 17 20:15:43 -0700 2016"}, true},
		{anaconda.Tweet{FavoriteCount: 5,
			CreatedAt: "Tue Jul 17 20:15:43 -0700 2016"}, true},
		{anaconda.Tweet{Retweeted: true, RetweetCount: 5,
			CreatedAt: "Tue Jul 17 20:15:43 -0700 2016"}, false},
		{anaconda.Tweet{Retweeted: true, FavoriteCount: 5,
			CreatedAt: "Tue Jul 17 20:15:43 -0700 2016"}, false},
	}

	for _, test := range tests {
		got, reason := (TweetType{}).Keep(p, test.tweet, now)
		if got != test.want {
			t.Errorf("Keep(...) was %v %v should be %v for %v",
				got, reason, test.want, test.tweet)
		}
	}

}

func TestKeepFavs(t *testing.T) {
	p := Policy{MaxAge: 5 * time.Hour,
		MinStars:    4,
		MinRetweets: 4,
		KeepMedia:   true}
	now, _ := time.Parse(time.RubyDate, "Tue Jul 17 20:15:43 -0700 2018")

	var tests = []struct {
		tweet anaconda.Tweet
		want  bool
	}{
		{anaconda.Tweet{CreatedAt: "fullofweasels"}, true},
		{anaconda.Tweet{CreatedAt: "Tue Jul 17 20:15:43 -0700 2018"}, true},
		{anaconda.Tweet{CreatedAt: "Tue Jul 17 14:15:43 -0700 2018"}, false},
		{anaconda.Tweet{CreatedAt: "Tue Jul 17 20:15:43 -0700 2016"}, false},
		{anaconda.Tweet{RetweetCount: 3,
			CreatedAt: "Tue Jul 17 20:15:43 -0700 2016"}, false},
		{anaconda.Tweet{FavoriteCount: 3,
			CreatedAt: "Tue Jul 17 20:15:43 -0700 2016"}, false},
		{anaconda.Tweet{RetweetCount: 5,
			CreatedAt: "Tue Jul 17 20:15:43 -0700 2016"}, false},
		{anaconda.Tweet{FavoriteCount: 5,
			CreatedAt: "Tue Jul 17 20:15:43 -0700 2016"}, false},
		{anaconda.Tweet{Retweeted: true, RetweetCount: 5,
			CreatedAt: "Tue Jul 17 20:15:43 -0700 2016"}, false},
		{anaconda.Tweet{Retweeted: true, FavoriteCount: 5,
			CreatedAt: "Tue Jul 17 20:15:43 -0700 2016"}, false},
	}

	for _, test := range tests {
		got, reason := (FavType{}).Keep(p, test.tweet, now)
		if got != test.want {
			t.Errorf("Keep(...) was %v %v should be %v for %v",
				got, reason, test.want, test.tweet)
		}
	}

}
