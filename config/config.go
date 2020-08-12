package config

import (
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/whitekid/go-utils/log"
)

// NOTE 각 파일에 별도로 분리하면 더 깔끔하겠지만, init()의 호출 순서 때문에 문제가 발행함
var configs = map[string][]struct {
	key          string
	short        string
	defaultValue interface{}
	description  string
}{
	"todo": {
		{keyStorage, "s", "badger", "todo storage type"},
		{keyClientID, "", "your-client-id", "google auth client id"},
		{keyClientSecret, "", "your-client-secret", "google auth client secret"},
		{keyRootURL, "u", "http://127.0.0.1", "application root url"},
		{keyCallbackURL, "", "/oauth/callback", "oauth callback url"},
		{teyTokenSigningKey, "", []byte("signing-key"), "jwt token signing key"},
		{keyRefreshTokenDuration, "", time.Hour * 24 * 14, "refresh token duration"}, // refresh token expires in 2 weeks
		{keyAccessTokenDuration, "", time.Minute * 30, "access token duration"},      // access token expires in 30 mins
	},
	"hello": {
		{"world", "w", "world", "saying hello world"},
	},
}

func init() {
	initDefaults()
}

func InitConfig() {
	viper.SetEnvPrefix("TODO")
	viper.AutomaticEnv()
}

// TODO move to go-utils
func initDefaults() {
	// InitDefaults initialize config
	for use := range configs {
		for _, config := range configs[use] {
			if config.defaultValue != nil {
				viper.SetDefault(config.key, config.defaultValue)
			}
		}
	}
}

// InitFlagSet cobra.Command와 연결
func InitFlagSet(use string, fs *pflag.FlagSet) {
	for _, config := range configs[use] {
		switch v := config.defaultValue.(type) {
		case string:
			fs.StringP(config.key, config.short, v, config.description)
		case time.Duration:
			fs.DurationP(config.key, config.short, v, config.description)
		case []byte:
			fs.BytesHexP(config.key, config.short, v, config.description)
		default:
			log.Errorf("unsupported type %T", config.defaultValue)
		}
		viper.BindPFlag(config.key, fs.Lookup(config.key))
	}
}
