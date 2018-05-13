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

func idRange(db *leveldb.DB) (low, high int64, err error) {
	i := timelineKey.Scan(db)
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

func TimelineFetch(cmd *cobra.Command, args []string) {
	db, err := leveldb.OpenFile(viper.GetString("store"), nil)
	if err != nil {
		glog.Exit(err)
	}
	defer db.Close()

	api := twitterAPI()

	low, high, err := idRange(db)
	retrieved := 0
	for high != 0 {
		if err != nil {
			glog.Exit(err)
		}
		v := make(url.Values)
		v.Set("since_id", fmt.Sprint(high))
		v.Set("count", "200")
		b := new(leveldb.Batch)
		n, err := fetch(api, b, v)
		glog.Infof("fetching ids above %d: %d error %v", high, n, err)
		if err != nil {
			glog.Exit(err)
		}
		if n == 0 {
			break
		}
		if err := db.Write(b, nil); err != nil {
			glog.Exit(err)
		}
		retrieved += n
		low, high, err = idRange(db)
	}
	if retrieved > 0 {
		fmt.Println("retrieved", retrieved, "recent tweets")
		retrieved = 0
	}

	for {
		if err != nil {
			glog.Exit(err)
		}
		v := make(url.Values)
		if low != 0 {
			v.Set("max_id", fmt.Sprint(low-1))
		}
		v.Set("count", "200")
		b := new(leveldb.Batch)
		n, err := fetch(api, b, v)
		glog.Infof("fetching ids below %d: %d error %v", low, n, err)
		if err != nil {
			glog.Exit(err)
		}
		if n == 0 {
			break
		}
		if err := db.Write(b, nil); err != nil {
			glog.Exit(err)
		}
		retrieved += n
		low, high, err = idRange(db)
	}
	if retrieved > 0 {
		fmt.Println("retrieved", retrieved, "older tweets")
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
