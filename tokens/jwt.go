package tokens

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/whitekid/go-todo/config"
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
// duration jwt token expiration time from now
func New(issuer string, duration time.Duration) (string, error) {
	claims := &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(duration).Unix(),
		Issuer:    issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(config.TokenSignKey())
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
		return config.TokenSignKey(), nil
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

// IsExpired return true if err is expired error
func IsExpired(err error) bool {
	if e, ok := err.(*jwt.ValidationError); ok {
		return e.Errors == jwt.ValidationErrorExpired
	}

	return false

}
