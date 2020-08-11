package main

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	todo "github.com/whitekid/go-todo/pkg"
	"github.com/whitekid/go-todo/pkg/config"
)

var rootCmd = &cobra.Command{
	Use:   "todo",
	Short: "run todo service",
	Long:  "run todo service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return todo.New().Serve(context.TODO(), args...)
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	config.InitFlagSet(rootCmd.Use, rootCmd.Flags())
}

func initConfig() {
	viper.SetEnvPrefix("TODO")
	viper.AutomaticEnv()
}
