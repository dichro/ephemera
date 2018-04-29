package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/huandu/facebook"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
	"golang.org/x/oauth2"

	ofb "golang.org/x/oauth2/facebook"
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
		id       = "facebook_id"
		secret   = "facebook_secret"
		redirect = "facebook_redirect"
	)
	conf := oauth2.Config{
		ClientID:     viper.GetString(id),
		ClientSecret: viper.GetString(secret),
		RedirectURL:  viper.GetString(redirect),
		Scopes:       []string{"publish_actions", "user_posts"},
		Endpoint:     ofb.Endpoint,
	}
	if len(conf.ClientID) == 0 {
		glog.Exitf(`no "%s" in config file`, id)
	}
	if len(conf.ClientSecret) == 0 {
		glog.Exitf(`no "%s" in config file`, secret)
	}
	if len(conf.RedirectURL) == 0 {
		glog.Exitf(`no "%s" in config file`, redirect)
	}

	author := "default"
	// if user exists in state, use it
	var u User
	if val, err := db.Get([]byte(author), nil); err == nil {
		if err := gob.NewDecoder(bytes.NewReader(val)).Decode(&u); err == nil {
			fmt.Printf("restored %#v, %s\n", u, u.LastEntry)
		} else {
			glog.Exit(err)
		}
	} else {
		if len(*code) > 0 {
			if *dryRun {
				glog.Exit("--code requires --dry_run=false, since it consumes the code")
			}
			// if code exists in args, exchange it for a token, store it in state, use it
			if token, err := conf.Exchange(oauth2.NoContext, *code); err == nil {
				fmt.Println("exchanged token")
				u.Token = token
				u.dirty = true
			} else {
				glog.Exit(err)
			}
		} else {
			// else print auth url
			fmt.Println("url", conf.AuthCodeURL(author, oauth2.AccessTypeOffline))
		}
	}

	client := conf.Client(oauth2.NoContext, u.Token)
	session := facebook.Session{
		Version:    "v2.12",
		HttpClient: client,
	}

	res, err := session.Get("/me/feed", map[string]interface{}{
		"until": 1493485263,
	})
	if err != nil {
		glog.Error(err)
	} else {
		if err := parseFeedResult(res); err != nil {
			glog.Error(err)
		}
	}

	if u.dirty {
		glog.Infof("writing out %s: %#v", author, u)
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(u); err != nil {
			glog.Exit(err)
		}
		if !*dryRun {
			if err := db.Put([]byte(author), buf.Bytes(), nil); err != nil {
				glog.Exit(err)
			}
		}
	}
}

type Feed struct {
	Posts []interface{} `json:"data"`
}

func parseFeedResult(result facebook.Result) error {
	var feed Feed
	if err := result.Decode(&feed); err != nil {
		return err
	}
	glog.Infof("retrieved %d posts", len(feed.Posts))
	for _, post := range feed.Posts {
		fmt.Println("post", post)
	}
	return nil
}
