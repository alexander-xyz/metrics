package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexander-xyz/metrics/internal/logger"
	models "github.com/alexander-xyz/metrics/internal/model"
	"github.com/alexander-xyz/metrics/internal/repository"
	"github.com/alexander-xyz/metrics/internal/signature"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompressRoundTrip(t *testing.T) {
	data := []byte(`{"id":"Alloc","type":"gauge","value":1.5}`)

	compressed, err := compress(data)
	require.NoError(t, err)
	assert.NotEqual(t, data, compressed)

	zr, err := gzip.NewReader(bytes.NewReader(compressed))
	require.NoError(t, err)

	got, err := io.ReadAll(zr)
	require.NoError(t, err)
	assert.Equal(t, data, got)
}

func TestPostBatch(t *testing.T) {
	var (
		gotEncoding string
		gotSum      string
		gotMetrics  []models.Metrics
	)

	srv := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		gotEncoding = req.Header.Get("Content-Encoding")
		gotSum = req.Header.Get(signature.Header)

		zr, err := gzip.NewReader(req.Body)
		require.NoError(t, err)

		require.NoError(t, json.NewDecoder(zr).Decode(&gotMetrics))
		res.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	value := 1.5
	metrics := []models.Metrics{{ID: "Alloc", MType: models.Gauge, Value: &value}}

	require.NoError(t, postBatch(srv.URL, "key", nil, metrics))

	assert.Equal(t, "gzip", gotEncoding)
	assert.NotEmpty(t, gotSum, "подпись передана")
	assert.Equal(t, "Alloc", gotMetrics[0].ID)
}

func TestPostBatchFailsOnBadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	assert.Error(t, postBatch(srv.URL, "", nil, nil))
}

func TestPostBatchWithRetryStopsOnCancelledContext(t *testing.T) {
	calls := 0

	srv := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		calls++
		res.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	assert.Error(t, postBatchWithRetry(ctx, srv.URL, "", nil, nil))
	assert.Equal(t, 1, calls, "отменённый контекст не повторяет отправку")
}

func TestCollectBatch(t *testing.T) {
	store := repository.NewMemStorage()
	ctx := context.Background()

	require.NoError(t, store.UpdateGauge(ctx, "Alloc", 1.5))
	require.NoError(t, store.UpdateCounter(ctx, "PollCount", 3))

	metrics, err := collectBatch(ctx, store)
	require.NoError(t, err)
	require.Len(t, metrics, 2)

	byID := map[string]models.Metrics{}
	for _, m := range metrics {
		byID[m.ID] = m
	}

	require.NotNil(t, byID["Alloc"].Value)
	assert.Equal(t, 1.5, *byID["Alloc"].Value)
	assert.Equal(t, models.Gauge, byID["Alloc"].MType)

	require.NotNil(t, byID["PollCount"].Delta)
	assert.Equal(t, int64(3), *byID["PollCount"].Delta)
	assert.Equal(t, models.Counter, byID["PollCount"].MType)
}

func TestPostBatchSendsRealIP(t *testing.T) {
	var gotIP string

	srv := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		gotIP = req.Header.Get(logger.RealIPHeader)
		res.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	require.NoError(t, postBatch(srv.URL, "", nil, nil))

	assert.Equal(t, localIP(), gotIP, "агент передаёт свой адрес в X-Real-IP")
}

func TestLocalIPIsRoutable(t *testing.T) {
	ip := localIP()
	if ip == "" {
		t.Skip("у хоста нет внешних адресов")
	}

	parsed := net.ParseIP(ip)
	require.NotNil(t, parsed)
	assert.False(t, parsed.IsLoopback(), "loopback не подходит для X-Real-IP")
}
