package logger

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexander-xyz/metrics/internal/signature"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func jsonHandler(body string) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")
		res.Write([]byte(body))
	})
}

func TestGzipMiddlewareCompressesJSON(t *testing.T) {
	srv := httptest.NewServer(GzipMiddleware(jsonHandler(`{"status":"ok"}`)))
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)
	req.Header.Set("Accept-Encoding", "gzip")

	res, err := http.DefaultTransport.RoundTrip(req)
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, "gzip", res.Header.Get("Content-Encoding"))

	zr, err := gzip.NewReader(res.Body)
	require.NoError(t, err)

	body, err := io.ReadAll(zr)
	require.NoError(t, err)
	assert.Equal(t, `{"status":"ok"}`, string(body))
}

func TestGzipMiddlewareSkipsPlainText(t *testing.T) {
	handler := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/plain")
		res.Write([]byte("plain"))
	})

	srv := httptest.NewServer(GzipMiddleware(handler))
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)
	req.Header.Set("Accept-Encoding", "gzip")

	res, err := http.DefaultTransport.RoundTrip(req)
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Empty(t, res.Header.Get("Content-Encoding"))

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, "plain", string(body))
}

func TestGzipMiddlewareDecompressesRequest(t *testing.T) {
	var got []byte

	handler := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		got, _ = io.ReadAll(req.Body)
		res.WriteHeader(http.StatusOK)
	})

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	zw.Write([]byte("compressed body"))
	require.NoError(t, zw.Close())

	srv := httptest.NewServer(GzipMiddleware(handler))
	defer srv.Close()

	req, err := http.NewRequest(http.MethodPost, srv.URL, &buf)
	require.NoError(t, err)
	req.Header.Set("Content-Encoding", "gzip")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, "compressed body", string(got))
}

func TestGzipMiddlewareRejectsBrokenBody(t *testing.T) {
	srv := httptest.NewServer(GzipMiddleware(jsonHandler("{}")))
	defer srv.Close()

	req, err := http.NewRequest(http.MethodPost, srv.URL, bytes.NewReader([]byte("не gzip")))
	require.NoError(t, err)
	req.Header.Set("Content-Encoding", "gzip")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestSignatureMiddlewareSignsResponse(t *testing.T) {
	require.NoError(t, Initialize("info"))

	srv := httptest.NewServer(SignatureMiddleware("key")(jsonHandler(`{"status":"ok"}`)))
	defer srv.Close()

	body := []byte("request")

	req, err := http.NewRequest(http.MethodPost, srv.URL, bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set(signature.Header, signature.Sign(body, "key"))

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, signature.Sign([]byte(`{"status":"ok"}`), "key"), res.Header.Get(signature.Header))
}

func TestSignatureMiddlewareRejectsWrongSignature(t *testing.T) {
	require.NoError(t, Initialize("info"))

	srv := httptest.NewServer(SignatureMiddleware("key")(jsonHandler("{}")))
	defer srv.Close()

	req, err := http.NewRequest(http.MethodPost, srv.URL, bytes.NewReader([]byte("request")))
	require.NoError(t, err)
	req.Header.Set(signature.Header, signature.Sign([]byte("request"), "other"))

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestSignatureMiddlewareSkippedWithoutKey(t *testing.T) {
	srv := httptest.NewServer(SignatureMiddleware("")(jsonHandler("{}")))
	defer srv.Close()

	res, err := http.Get(srv.URL)
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Empty(t, res.Header.Get(signature.Header))
}

func TestRequestLogger(t *testing.T) {
	require.NoError(t, Initialize("info"))

	srv := httptest.NewServer(RequestLogger(jsonHandler(`{"status":"ok"}`)))
	defer srv.Close()

	res, err := http.Get(srv.URL)
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}
