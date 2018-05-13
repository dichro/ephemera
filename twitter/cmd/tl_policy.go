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

func init() {
	tl.AddCommand(policyTL)
	policyTL.AddCommand(dropsTL)
}

var p = Policy{
	MaxAge:      52 * 7 * 24 * time.Hour,
	MinRetweets: 1,
	MinStars:    1,
}

func TimelinePolicyDrops(cmd *cobra.Command, args []string) {
	db, err := leveldb.OpenFile(viper.GetString("store"), nil)
	if err != nil {
		glog.Exit(err)
	}
	defer db.Close()
	result := p.Apply(db)
	for _, tweet := range result.Dropped {
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
	result := p.Apply(db)
	for r, n := range result.Kept {
		fmt.Println("kept", n, "because", r)
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
	if tweet.RetweetCount >= p.MinRetweets || tweet.FavoriteCount >= p.MinStars {
		return true, "too popular"
	}
	if len(tweet.Entities.Media) > 0 {
		return true, "has media"
	}
	if tweet.InReplyToStatusID != 0 {
		return true, "replies"
	}
	return false, "no rule match"
}

type Result struct {
	Kept    map[string]int
	Dropped []anaconda.Tweet
}

func (p Policy) Apply(db *leveldb.DB) Result {
	r := Result{
		Kept: make(map[string]int),
	}
	i := timelineKey.Scan(db)
	defer i.Release()
	for i.Next() {
		tweet, err := i.Value()
		if err != nil {
			r.Kept[fmt.Sprintf("zzy error %s", err.Error())]++
			continue
		}
		if keep, reason := p.Keep(tweet); keep {
			r.Kept[reason]++
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
