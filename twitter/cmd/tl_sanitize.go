package cmd

import (
	"fmt"

	"github.com/dichro/ephemera/pinaf"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
)

var sanitizeTL = &cobra.Command{
	Use:   "sanitize",
	Short: "sanitizes timeline according to policy",
	Run:   TimelineSanitize,
}

func init() {
	tl.AddCommand(sanitizeTL)
}

var deletesKey = TimelineKey{pinaf.JSONKey{pinaf.New("ephemera", "timeline", "drops")}}

func TimelineSanitize(cmd *cobra.Command, args []string) {
	db, err := leveldb.OpenFile(viper.GetString("store"), nil)
	if err != nil {
		glog.Exit(err)
	}
	defer db.Close()
	policy, err := LoadPolicyFromConfig()
	if err != nil {
		glog.Exit(err)
	}
	result := policy.Apply(db)
	for r, n := range result.Kept {
		fmt.Println("kept", len(n), "because", r)
	}

	fmt.Println("dropping", len(result.Dropped), "in total")

	api := twitterAPI()
	for _, tweet := range result.Dropped {
		if _, err := deletesKey.Get(db, tweet.Id); err == nil {
			// TODO(dichro): make a .Has method
			continue
		}
		b := new(leveldb.Batch)
		fmt.Println("dropping", tweet.Id)
		t, err := api.DeleteTweet(tweet.Id, true)
		if err != nil {
			glog.Error(err)
			continue
		}
		deletesKey.Put(b, t)
		if err := db.Write(b, nil); err != nil {
			glog.Exit(err)
		}
	}
}
