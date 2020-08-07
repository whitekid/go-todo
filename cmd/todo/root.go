package main

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	todo "github.com/whitekid/go-todo/pkg"
	"github.com/whitekid/go-todo/pkg/config"
)

var rootCmd = &cobra.Command{
	Use: "todo",
	RunE: func(cmd *cobra.Command, args []string) error {
		return todo.New().Serve(context.TODO(), args...)
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	fs := rootCmd.Flags()

	fs.StringP(config.KeyStorage, "s", config.DefaultStorage, "todo storage")
	viper.BindPFlag(config.KeyStorage, fs.Lookup(config.KeyStorage))
}

func initConfig() {
	viper.SetEnvPrefix("TODO")
	viper.AutomaticEnv()
}
