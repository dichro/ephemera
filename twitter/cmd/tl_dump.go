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
	Run:   TimelineDump,
}

func init() {
	tl.AddCommand(dumpTL)
}

func TimelineDump(cmd *cobra.Command, args []string) {
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
			timelineDumpOne(db, id)
		}
	} else {
		timelineDumpAll(db)
	}
}

func timelineDumpOne(db *leveldb.DB, id int64) {
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

func timelineDumpAll(db *leveldb.DB) {
	i := timelineKey.Scan(db)
	defer i.Release()
	for i.Next() {
		id, err := i.Key()
		if err != nil {
			glog.Error(err)
		}
		fmt.Println(id)
	}
	if !i.Prev() {
		glog.Exit(`no timeline data retrieved; try "twitter timeline fetch" first`)
	}
}
