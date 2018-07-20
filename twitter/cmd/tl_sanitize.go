package cmd

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
)

var sanitizeTL = &cobra.Command{
	Use:   "sanitize",
	Short: "sanitizes timeline according to policy",
	Run: func(cmd *cobra.Command, args []string) {
		TimelineSanitize(cmd, args, TweetType{})
	},
}

var sanitizeFavs = &cobra.Command{
	Use:   "sanitize",
	Short: "sanitizes favorites according to policy",
	Run: func(cmd *cobra.Command, args []string) {
		TimelineSanitize(cmd, args, FavType{})
	},
}

func init() {
	tl.AddCommand(sanitizeTL)
	fv.AddCommand(sanitizeFavs)
}

func TimelineSanitize(cmd *cobra.Command, args []string, twitterType TwitterType) {
	db, err := leveldb.OpenFile(viper.GetString("store"), nil)
	if err != nil {
		glog.Exit(err)
	}
	defer db.Close()
	policy, err := LoadPolicyFromConfig()
	if err != nil {
		glog.Exit(err)
	}
	result := policy.Apply(db, twitterType)
	for r, n := range result.Kept {
		fmt.Println("kept", len(n), "because", r)
	}

	fmt.Println("dropping", len(result.Dropped), "in total")

	api := twitterAPI()
	for _, tweet := range result.Dropped {
		if _, err := twitterType.DeletesKey().Get(db, tweet.Id); err == nil {
			// TODO(dichro): make a .Has method
			continue
		}
		b := new(leveldb.Batch)
		fmt.Println("dropping", tweet.Id)
		t, err := twitterType.DeleteCall(api, tweet.Id)
		if err != nil {
			glog.Error(err)
			continue
		}
		twitterType.DeletesKey().Put(b, t)
		if err := db.Write(b, nil); err != nil {
			glog.Exit(err)
		}
	}
}
