package cmd

import (
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
)

var policyTL = &cobra.Command{
	Use:   "policy",
	Short: "measures policy against timeline",
	Run: func(cmd *cobra.Command, args []string) {
		TimelinePolicy(cmd, args, TweetType{})
	},
}

var dropsTL = &cobra.Command{
	Use:   "drops",
	Short: "lists tweets that policy wants to delete",
	Run: func(cmd *cobra.Command, args []string) {
		TimelinePolicyDrops(cmd, args, TweetType{})
	},
}

var keepsTL = &cobra.Command{
	Use:   "keeps",
	Short: "lists tweets that policy wants to keep",
	Run: func(cmd *cobra.Command, args []string) {
		TimelinePolicyKeeps(cmd, args, TweetType{})
	},
}

var policyFav = &cobra.Command{
	Use:   "policy",
	Short: "measures policy against timeline",
	Run: func(cmd *cobra.Command, args []string) {
		TimelinePolicy(cmd, args, FavType{})
	},
}

var dropsFav = &cobra.Command{
	Use:   "drops",
	Short: "lists favs that policy wants to delete",
	Run: func(cmd *cobra.Command, args []string) {
		TimelinePolicyDrops(cmd, args, FavType{})
	},
}

var keepsFav = &cobra.Command{
	Use:   "keeps",
	Short: "lists favs that policy wants to keep",
	Run: func(cmd *cobra.Command, args []string) {
		TimelinePolicyKeeps(cmd, args, FavType{})
	},
}

func init() {
	tl.AddCommand(policyTL)
	policyTL.AddCommand(dropsTL)
	policyTL.AddCommand(keepsTL)

	fv.AddCommand(policyFav)
	policyFav.AddCommand(dropsFav)
	policyFav.AddCommand(keepsFav)
}

func TimelinePolicyKeeps(cmd *cobra.Command, args []string, twitterType TwitterType) {
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
	for reason, tweets := range result.Kept {
		fmt.Println(reason)
		for _, tweet := range tweets {
			if _, err := twitterType.DeletesKey().Get(db, tweet.Id); err == nil {
				continue
			}
			tweetTmpl.Execute(os.Stdout, tweet)
			fmt.Println()
		}
	}
}

func TimelinePolicyDrops(cmd *cobra.Command, args []string, twitterType TwitterType) {
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
	for _, tweet := range result.Dropped {
		if _, err := twitterType.DeletesKey().Get(db, tweet.Id); err == nil {
			continue
		}
		tweetTmpl.Execute(os.Stdout, tweet)
		fmt.Println()
	}
}

func TimelinePolicy(cmd *cobra.Command, args []string, twitterType TwitterType) {
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

	fmt.Println("dropping", len(result.Dropped))
}

type Policy struct {
	MaxAge                time.Duration
	MinRetweets, MinStars int
	KeepMedia             bool
}

func LoadPolicyFromConfig() (Policy, error) {
	twitterPolicy := viper.Sub("twitter_policy")
	var p Policy
	err := twitterPolicy.Unmarshal(&p)
	return p, err
}

type Result struct {
	Kept    map[string][]anaconda.Tweet
	Dropped []anaconda.Tweet
}

func (p Policy) Apply(db *leveldb.DB, twitterType TwitterType) Result {
	r := Result{
		Kept: make(map[string][]anaconda.Tweet),
	}
	i := twitterType.Key().Scan(db)
	defer i.Release()
	for i.Next() {
		tweet, err := i.Value()
		if err != nil {
			errstr := fmt.Sprintf("zzy error %s", err.Error())
			r.Kept[errstr] = append(r.Kept[errstr], tweet)
			continue
		}
		if _, err := twitterType.DeletesKey().Get(db, tweet.Id); err == nil {
			//r.Kept["already deleted"]++
			// TODO(dichro): make a .Has method
			continue
		}
		if keep, reason := twitterType.Keep(p, tweet, time.Now()); keep {
			r.Kept[reason] = append(r.Kept[reason], tweet)
		} else {
			r.Dropped = append(r.Dropped, tweet)
		}
	}
	return r
}

const tweetTmplStr = `{{ .Id }} {{ .CreatedAt }} {{ .FavoriteCount }}S {{ .RetweetCount }}RT
{{ .FullText }}
`

var tweetTmpl *template.Template

func init() {
	tweetTmpl = template.Must(template.New("tweet").Parse(tweetTmplStr))
}
