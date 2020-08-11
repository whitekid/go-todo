package tokens

import (
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestToken(t *testing.T) {
	type args struct {
		issuer   string
		duration time.Duration
	}
	tests := [...]struct {
		name           string
		args           args
		wantErr        bool
		wantParseError bool
		wantErrType    interface{}
	}{
		{"", args{"issuer", time.Minute}, false, false, nil},
		{"expired", args{"issuer", -time.Minute}, false, true, &jwt.ValidationError{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.issuer, tt.args.duration)
			if (err != nil) != tt.wantErr {
				require.Failf(t, `New() failed`, `error = %v, wantErr = %v`, err, tt.wantErr)
			}

			issuer, err := Parse(got)
			if (err != nil) != tt.wantParseError {
				require.Failf(t, `Parse() failed`, `error = %v, wantErr = %v`, err, tt.wantErr)
			}
			if tt.wantParseError {
				require.IsType(t, tt.wantErrType, err)
				return
			}

			require.Equal(t, tt.args.issuer, issuer)
		})
	}
}

func TestJWT(t *testing.T) {
	type args struct {
		signKey  []byte
		issuer   string
		duration time.Duration
	}
	tests := [...]struct {
		name    string
		args    args
		wantErr bool
	}{
		{"", args{[]byte("hello"), "issuer", time.Minute}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clames := &jwt.StandardClaims{
				ExpiresAt: time.Now().Add(tt.args.duration).Unix(),
				Issuer:    tt.args.issuer,
			}

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, clames)
			ss, err := token.SignedString(tt.args.signKey)
			require.NoError(t, err)

			got, err := jwt.ParseWithClaims(ss, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.Errorf("Unexpected signing method: %v", token.Header["alg"])
				}
				return tt.args.signKey, nil
			})
			if (err != nil) != tt.wantErr {
				require.Failf(t, `New() failed`, `error = %v, wantErr = %v`, err, tt.wantErr)
			}

			claims, ok := got.Claims.(*jwt.StandardClaims)
			require.True(t, ok, "%+v", got.Claims)
			require.True(t, got.Valid)

			require.Equal(t, tt.args.issuer, claims.Issuer)
			require.Equal(t, time.Now().Add(tt.args.duration).Unix(), claims.ExpiresAt)
		})
	}
}
