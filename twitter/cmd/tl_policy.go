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
	Run:   TimelinePolicy,
}

var dropsTL = &cobra.Command{
	Use:   "drops",
	Short: "lists tweets that policy wants to delete",
	Run:   TimelinePolicyDrops,
}

var keepsTL = &cobra.Command{
	Use:   "keeps",
	Short: "lists tweets that policy wants to keep",
	Run:   TimelinePolicyKeeps,
}

func init() {
	tl.AddCommand(policyTL)
	policyTL.AddCommand(dropsTL)
	policyTL.AddCommand(keepsTL)
}

var defaultPolicy = Policy{
	MaxAge:      26 * 7 * 24 * time.Hour,
	MinRetweets: 3,
	MinStars:    3,
}

func TimelinePolicyKeeps(cmd *cobra.Command, args []string) {
	db, err := leveldb.OpenFile(viper.GetString("store"), nil)
	if err != nil {
		glog.Exit(err)
	}
	defer db.Close()
	result := defaultPolicy.Apply(db)
	for reason, tweets := range result.Kept {
		fmt.Println(reason)
		for _, tweet := range tweets {
			if _, err := deletesKey.Get(db, tweet.Id); err == nil {
				continue
			}
			tweetTmpl.Execute(os.Stdout, tweet)
			fmt.Println()
		}
	}
}

func TimelinePolicyDrops(cmd *cobra.Command, args []string) {
	db, err := leveldb.OpenFile(viper.GetString("store"), nil)
	if err != nil {
		glog.Exit(err)
	}
	defer db.Close()
	result := defaultPolicy.Apply(db)
	for _, tweet := range result.Dropped {
		if _, err := deletesKey.Get(db, tweet.Id); err == nil {
			continue
		}
		tweetTmpl.Execute(os.Stdout, tweet)
		fmt.Println()
	}
}

func TimelinePolicy(cmd *cobra.Command, args []string) {
	db, err := leveldb.OpenFile(viper.GetString("store"), nil)
	if err != nil {
		glog.Exit(err)
	}
	defer db.Close()
	result := defaultPolicy.Apply(db)
	for r, n := range result.Kept {
		fmt.Println("kept", len(n), "because", r)
	}

	fmt.Println("dropping", len(result.Dropped))
}

type Policy struct {
	MaxAge                time.Duration
	MinRetweets, MinStars int
}

func (p Policy) Keep(tweet anaconda.Tweet) (keep bool, reason string) {
	if t, err := tweet.CreatedAtTime(); err != nil {
		return true, "unparseable creation time"
	} else {
		if time.Now().Sub(t) < p.MaxAge {
			return true, "too recent"
		}
	}
	if !tweet.Retweeted && (tweet.RetweetCount >= p.MinRetweets || tweet.FavoriteCount >= p.MinStars) {
		return true, "too popular"
	}
	if !tweet.Retweeted && len(tweet.Entities.Media) > 0 {
		return true, "has media"
	}
	/*
		if tweet.InReplyToStatusID != 0 {
			return true, "replies"
		}
	*/
	return false, "no rule match"
}

type Result struct {
	Kept    map[string][]anaconda.Tweet
	Dropped []anaconda.Tweet
}

func (p Policy) Apply(db *leveldb.DB) Result {
	r := Result{
		Kept: make(map[string][]anaconda.Tweet),
	}
	i := timelineKey.Scan(db)
	defer i.Release()
	for i.Next() {
		tweet, err := i.Value()
		if err != nil {
			errstr := fmt.Sprintf("zzy error %s", err.Error())
			r.Kept[errstr] = append(r.Kept[errstr], tweet)
			continue
		}
		if _, err := deletesKey.Get(db, tweet.Id); err == nil {
			//r.Kept["already deleted"]++
			// TODO(dichro): make a .Has method
			continue
		}
		if keep, reason := p.Keep(tweet); keep {
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
