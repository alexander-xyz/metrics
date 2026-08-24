package handler

import (
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
	router, err := GetRouter(store)
	require.NoError(t, err)

	srv := httptest.NewServer(router)
	defer srv.Close()

	testCases := []struct {
		method       string
		expectedCode int
		request      string
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
