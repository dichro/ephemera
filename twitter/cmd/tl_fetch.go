package cmd

import (
	"fmt"
	"net/url"

	"github.com/ChimeraCoder/anaconda"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
)

var fetchTL = &cobra.Command{
	Use:   "fetch",
	Short: "retrieves timeline from twitter",
	Run:   TimelineFetch,
}

func init() {
	tl.AddCommand(fetchTL)
}

func twitterAPI() *anaconda.TwitterApi {
	const (
		id           = "twitter_id"
		secret       = "twitter_secret"
		accessToken  = "twitter_access_token"
		accessSecret = "twitter_access_secret"
	)

	anaconda.SetConsumerKey(viper.GetString(id))
	anaconda.SetConsumerSecret(viper.GetString(secret))
	return anaconda.NewTwitterApi(viper.GetString(accessToken), viper.GetString(accessSecret))
}

func TimelineFetch(cmd *cobra.Command, args []string) {
	db, err := leveldb.OpenFile(viper.GetString("store"), nil)
	if err != nil {
		glog.Exit(err)
	}
	defer db.Close()

	api := twitterAPI()

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
		id, err := i.Key()
		if err != nil {
			glog.Exit(err)
		}
		glog.Infof("earliest stored ID %d", id)
		v := make(url.Values)
		v.Set("max_id", fmt.Sprint(id-1))
		v.Set("count", "200")
		if n, err := fetch(api, b, v); err != nil {
			glog.Exit(err)
		} else {
			fmt.Println("retrieved", n, "older tweets")
		}

		i.Last()
		id, err = i.Key()
		if err != nil {
			glog.Exit(err)
		}
		glog.Infof("latest stored ID %d", id)
		v = make(url.Values)
		v.Set("since_id", fmt.Sprint(id))
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
