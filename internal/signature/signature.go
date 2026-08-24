package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

const Header = "HashSHA256"

func Sign(data []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)

	return hex.EncodeToString(h.Sum(nil))
}

func Valid(data []byte, key, want string) bool {
	got, err := hex.DecodeString(want)
	if err != nil {
		return false
	}

	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)

	return hmac.Equal(h.Sum(nil), got)
}
