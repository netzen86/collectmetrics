package security

import (
	"crypto/hmac"
	"crypto/sha256"
)

func SingSendData(src, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(src)
	return h.Sum(nil)
}

func CompareSing(sing1, sing2 []byte) bool {
	return hmac.Equal(sing1, sing2)
}
