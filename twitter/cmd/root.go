package cmd

import (
	"flag"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var root = &cobra.Command{
	Use:   "twitter",
	Short: "Content management for Twitter",
}

func Execute() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	if err := root.Execute(); err != nil {
		glog.Exit(err)
	}
}
