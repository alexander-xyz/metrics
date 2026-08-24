package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexander-xyz/metrics/internal/repository"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateMetricHandler(t *testing.T) {
	store := repository.NewMemStorage()
	router, err := GetRouter(store, nil, "", nil, nil, nil)
	require.NoError(t, err)

	srv := httptest.NewServer(router)
	defer srv.Close()

	testCases := []struct {
		method       string
		request      string
		expectedCode int
	}{
		{method: http.MethodGet, expectedCode: http.StatusMethodNotAllowed, request: "/update/gauge/LastGC/1.25"},
		{method: http.MethodPut, expectedCode: http.StatusMethodNotAllowed, request: "/update/gauge/LastGC/1.25"},
		{method: http.MethodDelete, expectedCode: http.StatusMethodNotAllowed, request: "/update/gauge/LastGC/1.25"},
		{method: http.MethodPost, expectedCode: http.StatusNotFound, request: "/update/gauge/"},
		{method: http.MethodPost, expectedCode: http.StatusOK, request: "/update/gauge/LastGC/1.25"},
		{method: http.MethodPost, expectedCode: http.StatusBadRequest, request: "/update/wrong-type/test/3.5"},
	}
	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			req := resty.New().R()
			req.Method = tc.method
			req.URL = srv.URL + tc.request

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")

			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Response code didn't match expected")
		})
	}
}

func TestUpdateMetricJSONHandler(t *testing.T) {
	store := repository.NewMemStorage()
	router, err := GetRouter(store, nil, "", nil, nil, nil)
	require.NoError(t, err)

	srv := httptest.NewServer(router)
	defer srv.Close()

	testCases := []struct {
		name         string
		body         string
		expectedBody string
		expectedCode int
	}{
		{
			name:         "gauge",
			body:         `{"id":"LastGC","type":"gauge","value":1.25}`,
			expectedCode: http.StatusOK,
			expectedBody: `{"id":"LastGC","type":"gauge","value":1.25}`,
		},
		{
			name:         "counter accumulates",
			body:         `{"id":"PollCount","type":"counter","delta":5}`,
			expectedCode: http.StatusOK,
			expectedBody: `{"id":"PollCount","type":"counter","delta":5}`,
		},
		{
			name:         "counter accumulates twice",
			body:         `{"id":"PollCount","type":"counter","delta":3}`,
			expectedCode: http.StatusOK,
			expectedBody: `{"id":"PollCount","type":"counter","delta":8}`,
		},
		{
			name:         "unknown type",
			body:         `{"id":"Some","type":"histogram","value":1}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "empty name",
			body:         `{"id":"","type":"gauge","value":1}`,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "broken json",
			body:         `{"id":`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "gauge without value",
			body:         `{"id":"Alloc","type":"gauge"}`,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := resty.New().R().
				SetHeader("Content-Type", "application/json").
				SetBody(tc.body).
				Post(srv.URL + "/update")

			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, tc.expectedCode, resp.StatusCode())

			if tc.expectedBody != "" {
				assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
				assert.JSONEq(t, tc.expectedBody, string(resp.Body()))
			}
		})
	}
}

func TestGetMetricJSONHandler(t *testing.T) {
	store := repository.NewMemStorage()
	require.NoError(t, store.UpdateGauge(context.Background(), "Alloc", repository.Gauge(42.5)))
	require.NoError(t, store.UpdateCounter(context.Background(), "PollCount", repository.Counter(7)))

	router, err := GetRouter(store, nil, "", nil, nil, nil)
	require.NoError(t, err)

	srv := httptest.NewServer(router)
	defer srv.Close()

	testCases := []struct {
		name         string
		body         string
		expectedBody string
		expectedCode int
	}{
		{
			name:         "known gauge",
			body:         `{"id":"Alloc","type":"gauge"}`,
			expectedCode: http.StatusOK,
			expectedBody: `{"id":"Alloc","type":"gauge","value":42.5}`,
		},
		{
			name:         "known counter",
			body:         `{"id":"PollCount","type":"counter"}`,
			expectedCode: http.StatusOK,
			expectedBody: `{"id":"PollCount","type":"counter","delta":7}`,
		},
		{
			name:         "unknown metric",
			body:         `{"id":"Nope","type":"gauge"}`,
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := resty.New().R().
				SetHeader("Content-Type", "application/json").
				SetBody(tc.body).
				Post(srv.URL + "/value")

			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, tc.expectedCode, resp.StatusCode())

			if tc.expectedBody != "" {
				assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
				assert.JSONEq(t, tc.expectedBody, string(resp.Body()))
			}
		})
	}
}

func TestUpdateMetricsJSONHandler(t *testing.T) {
	store := repository.NewMemStorage()
	router, err := GetRouter(store, nil, "", nil, nil, nil)
	require.NoError(t, err)

	srv := httptest.NewServer(router)
	defer srv.Close()

	testCases := []struct {
		name         string
		body         string
		expectedCode int
	}{
		{
			name:         "batch of two",
			body:         `[{"id":"Alloc","type":"gauge","value":1.5},{"id":"PollCount","type":"counter","delta":4}]`,
			expectedCode: http.StatusOK,
		},
		{
			name:         "counter accumulates across batches",
			body:         `[{"id":"PollCount","type":"counter","delta":6}]`,
			expectedCode: http.StatusOK,
		},
		{
			name:         "empty batch rejected",
			body:         `[]`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "unknown type rejected",
			body:         `[{"id":"Alloc","type":"histogram","value":1}]`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "empty name rejected",
			body:         `[{"id":"","type":"gauge","value":1}]`,
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, reqErr := resty.New().R().
				SetHeader("Content-Type", "application/json").
				SetBody(tc.body).
				Post(srv.URL + "/updates/")

			assert.NoError(t, reqErr)
			assert.Equal(t, tc.expectedCode, resp.StatusCode())
		})
	}

	gauge, err := store.GetGauge(context.Background(), "Alloc")
	require.NoError(t, err)
	assert.Equal(t, repository.Gauge(1.5), gauge)

	counter, err := store.GetCounter(context.Background(), "PollCount")
	require.NoError(t, err)
	assert.Equal(t, repository.Counter(10), counter)
}
