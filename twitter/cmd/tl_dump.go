package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
)

var dumpTL = &cobra.Command{
	Use:   "dump",
	Short: "dumps timeline retrieved from twitter",
	Run: func(cmd *cobra.Command, args []string) {
		TimelineDump(cmd, args, TweetType{})
	},
}

var dumpFavs = &cobra.Command{
	Use:   "dump",
	Short: "dumps favorites retrieved from twitter",
	Run: func(cmd *cobra.Command, args []string) {
		TimelineDump(cmd, args, FavType{})
	},
}

func init() {
	tl.AddCommand(dumpTL)
	fv.AddCommand(dumpFavs)
}

func TimelineDump(cmd *cobra.Command, args []string, twitterType TwitterType) {
	db, err := leveldb.OpenFile(viper.GetString("store"), nil)
	if err != nil {
		glog.Exit(err)
	}
	defer db.Close()

	if len(args) > 0 {
		for _, a := range args {
			id, err := strconv.ParseInt(a, 0, 64)
			if err != nil {
				fmt.Println(err)
				continue
			}
			timelineDumpOne(db, id, twitterType.Key())
		}
	} else {
		timelineDumpAll(db, twitterType.Key())
	}
}

func timelineDumpOne(db *leveldb.DB, id int64,
	timelineKey TimelineKey) {
	tweet, err := timelineKey.Get(db, id)
	if err != nil {
		glog.Error(err)
		return
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent(" ", " ")
	if err := enc.Encode(&tweet); err != nil {
		glog.Error(err)
	}
}

func timelineDumpAll(db *leveldb.DB, timelineKey TimelineKey) {
	i := timelineKey.Scan(db)
	defer i.Release()
	for i.Next() {
		t, err := i.Value()
		if err != nil {
			glog.Error(err)
		}
		if err := tweetTmpl.Execute(os.Stdout, t); err != nil {
			glog.Error(err)
		}
	}
	if !i.Prev() {
		glog.Exit(`no timeline data retrieved; try "twitter timeline fetch" first`)
	}
}
