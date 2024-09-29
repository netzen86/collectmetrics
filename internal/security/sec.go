package security

import (
	"crypto/hmac"
	"crypto/sha256"
)

func SignSendData(src, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(src)
	return h.Sum(nil)
}

func CompareSign(sign1, sign2 []byte) bool {
	return hmac.Equal(sign1, sign2)
}
