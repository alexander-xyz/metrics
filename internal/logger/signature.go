package logger

import (
	"bytes"
	"io"
	"net/http"

	"github.com/alexander-xyz/metrics/internal/signature"
)

type signingWriter struct {
	http.ResponseWriter
	body   bytes.Buffer
	status int
}

func (w *signingWriter) WriteHeader(statusCode int) {
	w.status = statusCode
}

func (w *signingWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

// SignatureMiddleware — middleware, проверяющее подпись тела запроса
// и подписывающее тело ответа. При пустом ключе подпись не используется.
func SignatureMiddleware(key string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				h.ServeHTTP(w, r)
				return
			}

			if sum := r.Header.Get(signature.Header); sum != "" {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "cannot read request body", http.StatusBadRequest)
					return
				}

				r.Body.Close()

				if !signature.Valid(body, key, sum) {
					http.Error(w, "invalid signature", http.StatusBadRequest)
					return
				}

				r.Body = io.NopCloser(bytes.NewReader(body))
			}

			sw := &signingWriter{ResponseWriter: w, status: http.StatusOK}
			h.ServeHTTP(sw, r)

			w.Header().Set(signature.Header, signature.Sign(sw.body.Bytes(), key))
			w.WriteHeader(sw.status)

			if _, err := w.Write(sw.body.Bytes()); err != nil {
				Log.Error(err.Error())
			}
		})
	}
}
