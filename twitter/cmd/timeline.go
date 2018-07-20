package cmd

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"github.com/ChimeraCoder/anaconda"
	"github.com/dichro/ephemera/pinaf"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb"
	"net/url"
	"strings"
	"time"
)

var tl = &cobra.Command{
	Use:   "timeline",
	Short: "timeline operations",
}

var fv = &cobra.Command{
	Use:   "favorites",
	Short: "favorites operations",
}

func init() {
	root.AddCommand(tl)
	root.AddCommand(fv)
}

type TwitterType interface {
	Name() string
	Key() TimelineKey
	ApiCall(api *anaconda.TwitterApi, v url.Values) ([]anaconda.Tweet, error)
	DeletesKey() TimelineKey
	DeleteCall(api *anaconda.TwitterApi, id int64) (anaconda.Tweet, error)
	Keep(Policy, anaconda.Tweet, time.Time) (bool, string)
}

type FavType struct{}

func (FavType) Name() string {
	return "fav"
}

func (FavType) Key() TimelineKey {
	return TimelineKey{pinaf.JSONKey{pinaf.New("ephemera", "favorites", "fetch")}}
}
func (FavType) ApiCall(api *anaconda.TwitterApi, v url.Values) ([]anaconda.Tweet, error) {
	return api.GetFavorites(v)
}
func (FavType) DeletesKey() TimelineKey {
	return TimelineKey{pinaf.JSONKey{pinaf.New("ephemera", "favorites", "drops")}}
}

func (FavType) DeleteCall(api *anaconda.TwitterApi, id int64) (anaconda.Tweet, error) {
	return api.Unfavorite(id)
}

func (FavType) Keep(p Policy, tweet anaconda.Tweet, now time.Time) (keep bool, reason string) {
	if t, err := tweet.CreatedAtTime(); err != nil {
		return true, "unparseable creation time"
	} else {
		if now.Sub(t) < p.MaxAge {
			return true, "fav too recent"
		}
	}
	return false, "no rule match"
}

type TweetType struct{}

func (TweetType) Name() string {
	return "tweets"
}

func (TweetType) Key() TimelineKey {
	return TimelineKey{pinaf.JSONKey{pinaf.New("ephemera", "timeline", "fetch")}}
}

func (TweetType) ApiCall(api *anaconda.TwitterApi, v url.Values) ([]anaconda.Tweet, error) {
	return api.GetUserTimeline(v)
}

func (TweetType) DeletesKey() TimelineKey {
	return TimelineKey{pinaf.JSONKey{pinaf.New("ephemera", "timeline", "drops")}}
}

func (TweetType) DeleteCall(api *anaconda.TwitterApi, id int64) (anaconda.Tweet, error) {
	return api.DeleteTweet(id, true)
}

func (TweetType) Keep(p Policy, tweet anaconda.Tweet, now time.Time) (keep bool, reason string) {
	if t, err := tweet.CreatedAtTime(); err != nil {
		return true, "unparseable creation time"
	} else {
		if now.Sub(t) < p.MaxAge {
			return true, "too recent"
		}
	}
	if strings.HasPrefix(tweet.Text, "RT @") || tweet.Retweeted {
		return false, "retweet"
	}
	if tweet.RetweetCount >= p.MinRetweets || tweet.FavoriteCount >= p.MinStars {
		return true, "too popular"
	}
	if len(tweet.Entities.Media) > 0 && p.KeepMedia {
		return true, "has media"
	}
	/*
		                if tweet.InReplyToStatusID != 0 {
				                        return true, "replies"
							                }
	*/
	return false, "no rule match"
}

type TimelineKey struct {
	key pinaf.JSONKey
}

func (k TimelineKey) Get(db *leveldb.DB, tweetID int64) (tweet anaconda.Tweet, err error) {
	var key bytes.Buffer
	if err = binary.Write(&key, binary.BigEndian, tweetID); err == nil {
		err = k.key.Get(db, key.Bytes(), &tweet)
	}
	return
}

func (k TimelineKey) Put(b *leveldb.Batch, tweet anaconda.Tweet) error {
	var key bytes.Buffer
	if err := binary.Write(&key, binary.BigEndian, tweet.Id); err != nil {
		return err
	}
	return k.key.Put(b, key.Bytes(), tweet)
}

func (k TimelineKey) Scan(db *leveldb.DB) TimelineIterator {
	return TimelineIterator{k.key.Scan(db)}
}

func (k TimelineKey) IdRange(db *leveldb.DB) (low, high int64, err error) {
	i := k.Scan(db)
	defer i.Release()
	if !i.First() {
		return
	}
	if low, err = i.Key(); err == nil {
		i.Last()
		high, err = i.Key()
	}
	glog.Infof("idRange yielded %d, %d error %v", low, high, err)
	return
}

type TimelineIterator struct {
	pinaf.Iterator
}

func (i TimelineIterator) Key() (tweetID int64, err error) {
	err = binary.Read(bytes.NewReader(i.Iterator.Key()), binary.BigEndian, &tweetID)
	return
}

func (i TimelineIterator) Value() (tweet anaconda.Tweet, err error) {
	err = json.Unmarshal(i.Iterator.Value(), &tweet)
	return
}
