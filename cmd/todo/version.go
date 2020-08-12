package main

import (
	"github.com/spf13/cobra"
	"github.com/whitekid/go-todo/config"
	"github.com/whitekid/go-todo/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show version information",
	Long:  "show version information",
	Run: func(cmd *cobra.Command, args []string) {
		version.Print()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	config.InitFlagSet(versionCmd.Use, versionCmd.Flags())
}
