package config

import "github.com/spf13/viper"

func World() string { return viper.GetString("world") }
