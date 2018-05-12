package main

import (
	"flag"

	"github.com/dichro/ephemera/twitter/cmd"
	"github.com/golang/glog"
	"github.com/spf13/viper"
)

func main() {
	flag.Parse()

	viper.SetConfigName("ephemera")
	viper.AddConfigPath("$HOME/.config")
	if err := viper.ReadInConfig(); err != nil {
		glog.Exit(err)
	}
	cmd.Execute()
}

// login: unimplemented. do the thing on the site.
// tl fetch: get tweets; applies policy if it exists
// tl policy: configure policy; scans TL to inform impact
// tl sanitize: applies policy
