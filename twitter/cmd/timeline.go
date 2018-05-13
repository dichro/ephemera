package cmd

import (
	"bytes"
	"encoding/binary"
	"encoding/json"

	"github.com/ChimeraCoder/anaconda"
	"github.com/dichro/ephemera/pinaf"
	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb"
)

var tl = &cobra.Command{
	Use:   "timeline",
	Short: "timeline operations",
}

func init() {
	root.AddCommand(tl)
}

var timelineKey = TimelineKey{pinaf.JSONKey{pinaf.New("ephemera", "timeline", "fetch")}}

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
