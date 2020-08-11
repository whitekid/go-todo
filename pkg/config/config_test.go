package config

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefault(t *testing.T) {
	// clear TODO_XXX to prevent viper.AutomaticEnv()
	for _, v := range os.Environ() {
		if !strings.HasPrefix(v, "TODO_") {
			continue
		}

		values := strings.SplitAfterN(v, "=", 2)
		if len(values) > 1 {
			os.Unsetenv(values[0])
		}
	}

	InitDefaults()

	type args struct {
		key      string
		configFn func() interface{}
	}
	tests := [...]struct {
		name string
		args args
	}{
		{"ClientID", args{keyClientID, func() interface{} { return ClientID() }}},
		{"ClientSecret", args{keyClientSecret, func() interface{} { return ClientSecret() }}},
		{"RootURL", args{keyRootURL, func() interface{} { return RootURL() }}},
		{"CallbackURL", args{keyCallbackURL, func() interface{} { return CallbackURL() }}},
		{"Storage", args{keyStorage, func() interface{} { return Storage() }}},
		{"TokenSignKey", args{teyTokenSigningKey, func() interface{} { return TokenSignKey() }}},
		{"RefreshTokenDuration", args{keyRefreshTokenDuration, func() interface{} { return RefreshTokenDuration() }}},
		{"AccessTokenDuration", args{keyAccessTokenDuration, func() interface{} { return AccessTokenDuration() }}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// find default value from configuration
			var defaultValue interface{}
		exit:
			for use := range configs {
				for _, config := range configs[use] {
					if config.key == tt.args.key {
						defaultValue = config.defaultValue
						break exit
					}
				}
			}
			require.NotNil(t, defaultValue)

			got := tt.args.configFn()
			require.Equal(t, defaultValue, got, "got = %s, want = %s", got, defaultValue)
		})
	}
}
