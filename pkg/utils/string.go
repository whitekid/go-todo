package utils

import (
	"math/rand"
	"time"
)

const (
	Digits          = "0123456789"
	AsciiUpperCases = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	AsciiLowerCases = "abcdefghijklmnopqrstuvwxyz"
	AsciiLetters    = AsciiLowerCases + AsciiUpperCases
)

var seed = rand.New(rand.NewSource(time.Now().UnixNano()))

// RandomString generate random string
// TODO move to go-utils
func RandomString(l int) string {
	const charset = AsciiLetters + Digits

	b := make([]byte, l)

	for i := 0; i < l; i++ {
		b[i] = charset[seed.Intn(len(charset))]
	}
	return string(b)
}
