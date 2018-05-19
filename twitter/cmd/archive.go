package cmd

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
)

var archive = &cobra.Command{
	Use:   "archive",
	Short: "load Twitter backup archive",
	Run:   ArchiveLoad,
}

func init() {
	root.AddCommand(archive)
}

func ArchiveLoad(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		glog.Exit("need archive filename as argument")
	}
	db, err := leveldb.OpenFile(viper.GetString("store"), nil)
	if err != nil {
		glog.Exit(err)
	}
	defer db.Close()

	f, err := os.Open(args[0])
	if err != nil {
		glog.Exit(err)
	}
	st, err := f.Stat()
	if err != nil {
		glog.Exit(err)
	}
	r, err := zip.NewReader(f, st.Size())
	if err != nil {
		glog.Exit(err)
	}

	var ids []int64
	for _, f := range r.File {
		if f.Name != "tweets.csv" {
			continue
		}
		glog.Info("found tweets.csv in archive")
		rr, err := f.Open()
		if err != nil {
			glog.Exit(err)
		}
		recs, err := csv.NewReader(rr).ReadAll()
		if err != nil {
			glog.Exit(err)
		}
		if recs[0][0] != "tweet_id" {
			glog.Exit("unexpected CSV format")
		}

		for _, rec := range recs[1:] {
			id, err := strconv.ParseInt(rec[0], 10, 64)
			if err != nil {
				glog.Exit(err)
			}
			if _, err := timelineKey.Get(db, id); err == nil {
				continue // we already have this tweet
			}
			ids = append(ids, id)
		}
		fmt.Println("loading", len(ids), "of", len(recs))
	}

	if len(ids) == 0 {
		return
	}

	api := twitterAPI()
	const maxIDs = 100
	for i := 0; i < len(ids); i += maxIDs {
		end := i + maxIDs
		if end > len(ids) {
			end = len(ids)
		}
		fmt.Println("fetch", i, end)
		tweets, err := api.GetTweetsLookupByIds(ids[i:end], nil)
		if err != nil {
			glog.Exit(err)
		}
		b := new(leveldb.Batch)
		for _, t := range tweets {
			if err := timelineKey.Put(b, t); err != nil {
				glog.Exit(err)
			}
		}
		if err := db.Write(b, nil); err != nil {
			glog.Exit(err)
		}
	}
}
