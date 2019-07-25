package utils

import (
	"crypto/rand"
	"math/big"
)

func RandStringRunes(n int) string {
	var letterRunes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890")
	b := make([]rune, n)
	for i := range b {
		index, _ := rand.Int(rand.Reader, big.NewInt(51))
		b[i] = letterRunes[index.Int64()]
	}
	return string(b)
}
