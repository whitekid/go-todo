package config

import (
	"github.com/spf13/viper"
	"github.com/whitekid/go-todo/pkg/storage/badger"
)

const (
	KeyStorage     = "storage"
	DefaultStorage = "badger"
)

func init() {
	viper.SetDefault(KeyStorage, badger.Name)
}

func Storage() string {
	return viper.GetString(KeyStorage)
}
