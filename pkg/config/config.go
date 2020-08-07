package config

import "github.com/spf13/viper"

const (
	KeyStorage     = "storage"
	DefaultStorage = "session"
)

func Storage() string {
	return viper.GetString(KeyStorage)
}
