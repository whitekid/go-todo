package main

import (
	"context"

	"github.com/spf13/cobra"
	todo "github.com/whitekid/go-todo/pkg"
)

var rootCmd = &cobra.Command{
	Use: "playground",
	RunE: func(cmd *cobra.Command, args []string) error {
		return todo.New().Serve(context.TODO(), args...)
	},
}
