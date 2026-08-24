package logger

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"

	"github.com/alexander-xyz/metrics/internal/crypt"
)

// DecryptMiddleware расшифровывает тело запроса приватным ключом.
// При пустом ключе запросы проходят без изменений.
func DecryptMiddleware(key *rsa.PrivateKey) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == nil || r.Body == nil {
				h.ServeHTTP(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "cannot read request body", http.StatusBadRequest)
				return
			}

			r.Body.Close()

			if len(body) == 0 {
				r.Body = io.NopCloser(bytes.NewReader(body))
				h.ServeHTTP(w, r)

				return
			}

			decrypted, err := crypt.Decrypt(key, body)
			if err != nil {
				Log.Error(err.Error())
				http.Error(w, "cannot decrypt request body", http.StatusBadRequest)

				return
			}

			r.Body = io.NopCloser(bytes.NewReader(decrypted))
			r.ContentLength = int64(len(decrypted))

			h.ServeHTTP(w, r)
		})
	}
}
