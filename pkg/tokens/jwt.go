package tokens

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/whitekid/go-todo/pkg/config"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrInvalidClaimType = errors.New("invalid claim type")
	ErrClaimsInvalid    = errors.New("invalid claim")
)

type (
	ValidationError = jwt.ValidationError
)

// New create new jwt token using HMAC
func New(issuer string, duration time.Duration) (string, error) {
	claims := &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(duration).Unix(),
		Issuer:    issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(config.TokenKey())
	if err != nil {
		return "", err
	}
	return ss, nil

}

// Parse parse jwt token and return issuer
func Parse(s string) (string, error) {
	token, err := jwt.ParseWithClaims(s, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return config.TokenKey(), nil
	})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", ErrInvalidToken
	}

	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok {
		return "", ErrInvalidClaimType
	}

	return claims.Issuer, nil
}