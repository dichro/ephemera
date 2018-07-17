package main

import (
	"flag"

	"github.com/dichro/ephemera/twitter/cmd"
	"github.com/golang/glog"
	"github.com/spf13/viper"
)

func main() {
	configPath := flag.String("config_path", "$HOME/.config", "where to find the config file")
	flag.Parse()

	viper.SetConfigName("ephemera")
	viper.AddConfigPath(*configPath)
	if err := viper.ReadInConfig(); err != nil {
		glog.Exit(err)
	}
	cmd.Execute()
}

// login: unimplemented. do the thing on the site.
// tl fetch: get tweets; applies policy if it exists
// tl policy: configure policy; scans TL to inform impact
// tl sanitize: applies policy
