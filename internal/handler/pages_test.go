package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alexander-xyz/metrics/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsPageListsStoredMetrics(t *testing.T) {
	store := repository.NewMemStorage()
	ctx := t.Context()

	require.NoError(t, store.UpdateGauge(ctx, "Alloc", 1.5))
	require.NoError(t, store.UpdateCounter(ctx, "PollCount", 3))

	router, err := GetRouter(store, nil, "", nil)
	require.NoError(t, err)

	res := httptest.NewRecorder()
	router.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Header().Get("Content-type"), "text/html")

	body := res.Body.String()
	assert.Contains(t, body, "Alloc")
	assert.Contains(t, body, "PollCount")
}

func TestPingWithoutDatabase(t *testing.T) {
	router, err := GetRouter(repository.NewMemStorage(), nil, "", nil)
	require.NoError(t, err)

	res := httptest.NewRecorder()
	router.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/ping", nil))

	assert.Equal(t, http.StatusInternalServerError, res.Code, "без базы ping не проходит")
}

func TestGetMetricRejectsUnknownType(t *testing.T) {
	router, err := GetRouter(repository.NewMemStorage(), nil, "", nil)
	require.NoError(t, err)

	res := httptest.NewRecorder()
	router.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/value/unknown/Alloc", nil))

	assert.Equal(t, http.StatusBadRequest, res.Code)
}

func TestGetMetricNotFound(t *testing.T) {
	router, err := GetRouter(repository.NewMemStorage(), nil, "", nil)
	require.NoError(t, err)

	res := httptest.NewRecorder()
	router.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/value/counter/Missing", nil))

	assert.Equal(t, http.StatusNotFound, res.Code)
}

func TestUpdateMetricRejectsBrokenValues(t *testing.T) {
	router, err := GetRouter(repository.NewMemStorage(), nil, "", nil)
	require.NoError(t, err)

	testCases := []struct {
		request      string
		expectedCode int
	}{
		{request: "/update/gauge/Alloc/nonsense", expectedCode: http.StatusBadRequest},
		{request: "/update/counter/PollCount/1.5", expectedCode: http.StatusBadRequest},
		{request: "/update/unknown/Alloc/1", expectedCode: http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.request, func(t *testing.T) {
			res := httptest.NewRecorder()
			router.ServeHTTP(res, httptest.NewRequest(http.MethodPost, tc.request, nil))

			assert.Equal(t, tc.expectedCode, res.Code)
		})
	}
}

func TestUpdateMetricJSONRejectsBrokenBody(t *testing.T) {
	router, err := GetRouter(repository.NewMemStorage(), nil, "", nil)
	require.NoError(t, err)

	testCases := []struct {
		name         string
		body         string
		expectedCode int
	}{
		{name: "не json", body: "{", expectedCode: http.StatusBadRequest},
		{name: "неизвестный тип", body: `{"id":"Alloc","type":"summary","value":1}`, expectedCode: http.StatusBadRequest},
		{name: "пустое имя", body: `{"id":"","type":"gauge","value":1}`, expectedCode: http.StatusNotFound},
		{name: "gauge без значения", body: `{"id":"Alloc","type":"gauge"}`, expectedCode: http.StatusBadRequest},
		{name: "counter без значения", body: `{"id":"PollCount","type":"counter"}`, expectedCode: http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")

			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)

			assert.Equal(t, tc.expectedCode, res.Code)
		})
	}
}

func TestUpdateMetricsJSONRejectsBrokenBatch(t *testing.T) {
	router, err := GetRouter(repository.NewMemStorage(), nil, "", nil)
	require.NoError(t, err)

	testCases := []struct {
		name         string
		body         string
		expectedCode int
	}{
		{name: "не json", body: "[", expectedCode: http.StatusBadRequest},
		{name: "неизвестный тип", body: `[{"id":"Alloc","type":"summary"}]`, expectedCode: http.StatusBadRequest},
		{name: "пустое имя", body: `[{"id":"","type":"gauge","value":1}]`, expectedCode: http.StatusNotFound},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/updates", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")

			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)

			assert.Equal(t, tc.expectedCode, res.Code)
		})
	}
}
