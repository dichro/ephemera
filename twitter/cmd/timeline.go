package cmd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net/url"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/dichro/ephemera/pinaf"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
)

var tl = &cobra.Command{
	Use:   "timeline",
	Short: "timeline operations",
}

var fetchTL = &cobra.Command{
	Use:   "fetch",
	Short: "retrieves timeline from twitter",
	Run:   TimelineFetch,
}

var dumpTL = &cobra.Command{
	Use:   "dump",
	Short: "dumps timeline retrieved from twitter",
	Run:   TimelineDump,
}

func init() {
	root.AddCommand(tl)
	tl.AddCommand(fetchTL)
	tl.AddCommand(dumpTL)
}

var timelineKey = TimelineKey{pinaf.JSONKey{pinaf.New("ephemera", "timeline", "fetch")}}

func TimelineFetch(cmd *cobra.Command, args []string) {
	db, err := leveldb.OpenFile(viper.GetString("store"), nil)
	if err != nil {
		glog.Exit(err)
	}
	defer db.Close()

	const (
		id           = "twitter_id"
		secret       = "twitter_secret"
		accessToken  = "twitter_access_token"
		accessSecret = "twitter_access_secret"
	)

	anaconda.SetConsumerKey(viper.GetString(id))
	anaconda.SetConsumerSecret(viper.GetString(secret))
	api := anaconda.NewTwitterApi(viper.GetString(accessToken), viper.GetString(accessSecret))

	i := timelineKey.Scan(db)
	defer i.Release()
	b := new(leveldb.Batch)
	if !i.First() {
		if n, err := fetch(api, b, nil); err != nil {
			glog.Exit(err)
		} else {
			fmt.Println("retrieved", n, "tweets")
		}
	} else {
		entry, err := i.Key()
		if err != nil {
			glog.Exit(err)
		}
		glog.Infof("earliest stored ID %d from %s", entry.ID, entry.Time())
		v := make(url.Values)
		v.Set("max_id", fmt.Sprint(entry.ID-1))
		v.Set("count", "200")
		if n, err := fetch(api, b, v); err != nil {
			glog.Exit(err)
		} else {
			fmt.Println("retrieved", n, "older tweets")
		}

		i.Last()
		entry, err = i.Key()
		if err != nil {
			glog.Exit(err)
		}
		glog.Infof("latest stored ID %d from %s", entry.ID, entry.Time())
		v = make(url.Values)
		v.Set("since_id", fmt.Sprint(entry.ID))
		v.Set("count", "200")
		if n, err := fetch(api, b, v); err != nil {
			glog.Exit(err)
		} else {
			fmt.Println("retrieved", n, "newer tweets")
		}
	}
	if err := db.Write(b, nil); err != nil {
		glog.Exit(err)
	}
}

func fetch(api *anaconda.TwitterApi, b *leveldb.Batch, v url.Values) (int, error) {
	tl, err := api.GetUserTimeline(v)
	if err != nil {
		return 0, err
	}
	for i, status := range tl {
		if err := timelineKey.Put(b, status); err != nil {
			return i, err
		}
	}
	return len(tl), nil
}

func TimelineDump(cmd *cobra.Command, args []string) {
	db, err := leveldb.OpenFile(viper.GetString("store"), nil)
	if err != nil {
		glog.Exit(err)
	}
	defer db.Close()

	i := timelineKey.Scan(db)
	defer i.Release()
	for i.Next() {
		e, err := i.Key()
		if err != nil {
			glog.Error(err)
		}
		fmt.Println(e.Time(), e.ID)
	}
	if !i.Prev() {
		glog.Exit(`no timeline data retrieved; try "twitter timeline fetch" first`)
	}
}

type TimelineKey struct {
	key pinaf.JSONKey
}

type TimelineEntry struct {
	ID               int64
	TimestampSeconds int64
}

func (e TimelineEntry) Time() time.Time { return time.Unix(e.TimestampSeconds, 0) }

func (k TimelineKey) Put(b *leveldb.Batch, tweet anaconda.Tweet) error {
	entry := TimelineEntry{ID: tweet.Id}
	t, err := tweet.CreatedAtTime()
	if err != nil {
		return err
	}
	entry.TimestampSeconds = t.Unix()
	var key bytes.Buffer
	if err := binary.Write(&key, binary.BigEndian, &entry); err != nil {
		return err
	}
	return k.key.Put(b, key.Bytes(), tweet)
}

func (k TimelineKey) Scan(db *leveldb.DB) TimelineIterator {
	return TimelineIterator{k.key.Scan(db)}
}

type TimelineIterator struct {
	pinaf.Iterator
}

func (i TimelineIterator) Key() (entry TimelineEntry, err error) {
	err = binary.Read(bytes.NewReader(i.Iterator.Key()), binary.BigEndian, &entry)
	return
}
