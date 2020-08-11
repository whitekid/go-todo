package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/whitekid/go-todo/config"
)

var helloCmd = &cobra.Command{
	Use:   "hello",
	Short: "hello world example",
	Long:  "hello world example",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("hello %s\n", viper.GetString("world"))
	},
}

func init() {
	rootCmd.AddCommand(helloCmd)
	config.InitFlagSet(helloCmd.Use, helloCmd.Flags())
}
