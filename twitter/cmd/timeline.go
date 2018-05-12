package cmd

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
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

var timelineKey = TimelineKey{JSONKey{pinaf.New("ephemera", "timeline", "fetch")}}

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
	tl, err := api.GetUserTimeline(nil)
	if err != nil {
		glog.Exit(err)
	}
	b := new(leveldb.Batch)
	for _, status := range tl {
		if err := timelineKey.Put(b, status); err != nil {
			glog.Exit(err)
		}
	}
	if err := db.Write(b, nil); err != nil {
		glog.Exit(err)
	}
}

func TimelineDump(cmd *cobra.Command, args []string) {
	db, err := leveldb.OpenFile(viper.GetString("store"), nil)
	if err != nil {
		glog.Exit(err)
	}
	defer db.Close()

	i := timelineKey.Scan(db)
	if !i.First() {
		glog.Exit(`no timeline data retrieved; try "twitter timeline fetch" first`)
	}
	for i.Next() {
		e, err := i.Key()
		if err != nil {
			glog.Error(err)
		}
		fmt.Println(e.Time(), e.ID)
	}
}

type JSONKey struct {
	pinaf.Key
}

func (k JSONKey) Put(b *leveldb.Batch, subKey []byte, value interface{}) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(value)
	if err == nil {
		b.Put(k.Entry(subKey), buf.Bytes())
	}
	return err
}

func (k JSONKey) Get(db *leveldb.DB, subKey []byte, value interface{}) error {
	buf, err := db.Get(k.Entry(subKey), nil)
	if err == nil {
		err = json.NewDecoder(bytes.NewReader(buf)).Decode(value)
	}
	return err
}

type TimelineKey struct {
	key JSONKey
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
