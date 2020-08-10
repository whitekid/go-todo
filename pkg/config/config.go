package config

import (
	"time"

	"github.com/spf13/viper"
)

const (
	DefaultStorage = "badger"

	KeyStorage              = "storage"
	KeyTokenKey             = "signkey"
	KeyRefreshTokenDuration = "refresh_token_duration"
	KeyAccessTokenDuration  = "access_token_duration"
)

func init() {
	viper.SetDefault(KeyStorage, DefaultStorage)
	viper.SetDefault(KeyTokenKey, "9b768518-f780-44fd-9784-7481b7be2a4e")
	viper.SetDefault(KeyRefreshTokenDuration, time.Hour*24*14) // refresh token expires in 2 weeks
	viper.SetDefault(KeyAccessTokenDuration, time.Minute*30)   // access token expires in 30 mins
}

// Storage default storage
func Storage() string {
	return viper.GetString(KeyStorage)
}

// TokenKey key used in jwt key
func TokenKey() []byte {
	return []byte(viper.GetString(KeyTokenKey))
}

func RefreshTokenDuration() time.Duration {
	return viper.GetDuration(KeyRefreshTokenDuration)
}

func AccessTokenDuratin() time.Duration {
	return viper.GetDuration(KeyAccessTokenDuration)
}
