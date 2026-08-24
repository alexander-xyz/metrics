package logger

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexander-xyz/metrics/internal/crypt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func echoHandler(got *[]byte) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		body, _ := io.ReadAll(req.Body)
		*got = body
		res.WriteHeader(http.StatusOK)
	})
}

func TestDecryptMiddlewareDecryptsBody(t *testing.T) {
	require.NoError(t, Initialize("info"))

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	var got []byte

	srv := httptest.NewServer(DecryptMiddleware(key)(echoHandler(&got)))
	defer srv.Close()

	message := []byte(`[{"id":"Alloc","type":"gauge","value":1.5}]`)

	encrypted, err := crypt.Encrypt(&key.PublicKey, message)
	require.NoError(t, err)

	res, err := http.Post(srv.URL, "application/json", bytes.NewReader(encrypted))
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, message, got)
}

func TestDecryptMiddlewareSkippedWithoutKey(t *testing.T) {
	var got []byte

	srv := httptest.NewServer(DecryptMiddleware(nil)(echoHandler(&got)))
	defer srv.Close()

	res, err := http.Post(srv.URL, "application/json", bytes.NewReader([]byte("открытый текст")))
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, []byte("открытый текст"), got, "без ключа тело не трогается")
}

func TestDecryptMiddlewareRejectsUnencryptedBody(t *testing.T) {
	require.NoError(t, Initialize("info"))

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	var got []byte

	srv := httptest.NewServer(DecryptMiddleware(key)(echoHandler(&got)))
	defer srv.Close()

	res, err := http.Post(srv.URL, "application/json", bytes.NewReader([]byte("не шифровано")))
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestDecryptMiddlewarePassesEmptyBody(t *testing.T) {
	require.NoError(t, Initialize("info"))

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	var got []byte

	srv := httptest.NewServer(DecryptMiddleware(key)(echoHandler(&got)))
	defer srv.Close()

	res, err := http.Get(srv.URL)
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}
