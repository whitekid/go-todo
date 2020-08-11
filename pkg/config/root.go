package config

import (
	"time"

	"github.com/spf13/viper"
)

const (
	keyStorage              = "storage"
	keyClientID             = "client_id"
	keyClientSecret         = "client_secret"
	keyRootURL              = "root_url"
	keyCallbackURL          = "callback_url"
	teyTokenSigningKey      = "token_signkey"
	keyRefreshTokenDuration = "refresh_token_duration"
	keyAccessTokenDuration  = "access_token_duration"
)

func ClientID() string                    { return viper.GetString(keyClientID) }
func ClientSecret() string                { return viper.GetString(keyClientSecret) }
func RootURL() string                     { return viper.GetString(keyRootURL) }
func CallbackURL() string                 { return viper.GetString(keyCallbackURL) }
func Storage() string                     { return viper.GetString(keyStorage) }
func TokenSignKey() []byte                { return []byte(viper.GetString(teyTokenSigningKey)) }
func RefreshTokenDuration() time.Duration { return viper.GetDuration(keyRefreshTokenDuration) }
func AccessTokenDuration() time.Duration  { return viper.GetDuration(keyAccessTokenDuration) }
