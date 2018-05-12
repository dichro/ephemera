package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/golang/glog"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
	"golang.org/x/oauth2"
)

var (
	code   = flag.String("code", "", "oauth code for user")
	dryRun = flag.Bool("dry_run", true, "don't do anything for real")
)

type User struct {
	Token     *oauth2.Token
	LastEntry time.Time

	dirty bool
}

func main() {
	flag.Parse()

	viper.SetConfigName("ephemera")
	viper.AddConfigPath("$HOME/.config")
	if err := viper.ReadInConfig(); err != nil {
		glog.Exit(err)
	}

	db, err := leveldb.OpenFile(viper.GetString("store"), nil)
	if err != nil {
		glog.Exit(err)
	}
	defer db.Close()

	const (
		id            = "twitter_id"
		secret        = "twitter_secret"
		access_token  = "twitter_access_token"
		access_secret = "twitter_access_secret"
	)

	anaconda.SetConsumerKey(viper.GetString(id))
	anaconda.SetConsumerSecret(viper.GetString(secret))
	api := anaconda.NewTwitterApi(viper.GetString(access_token), viper.GetString(access_secret))
	tl, err := api.GetUserTimeline(nil)
	if err != nil {
		glog.Exit(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent(" ", " ")
	for i, status := range tl {
		fmt.Println(i)
		if err := enc.Encode(status); err != nil {
			glog.Exit(err)
		}
	}
}
