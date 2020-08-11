package main

import (
	"context"

	"github.com/spf13/cobra"
	todo "github.com/whitekid/go-todo"
	"github.com/whitekid/go-todo/config"
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
	cobra.OnInitialize(config.InitConfig)

	config.InitFlagSet(rootCmd.Use, rootCmd.Flags())
}
