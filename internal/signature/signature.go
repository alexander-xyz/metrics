// Package signature подписывает тело запросов и ответов по алгоритму HMAC-SHA256.
package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// Header — имя HTTP-заголовка, в котором передаётся подпись.
const Header = "HashSHA256"

// Sign возвращает HMAC-SHA256 подпись данных в шестнадцатеричном виде.
func Sign(data []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)

	return hex.EncodeToString(h.Sum(nil))
}

// Valid сообщает, соответствует ли подпись want данным и ключу.
func Valid(data []byte, key, want string) bool {
	got, err := hex.DecodeString(want)
	if err != nil {
		return false
	}

	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)

	return hmac.Equal(h.Sum(nil), got)
}
