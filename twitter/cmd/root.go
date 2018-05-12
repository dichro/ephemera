package cmd

import (
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use:   "twitter",
	Short: "Content management for Twitter",
}

func Execute() {
	if err := root.Execute(); err != nil {
		glog.Exit(err)
	}
}
