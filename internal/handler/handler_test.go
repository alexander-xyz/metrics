package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexander-xyz/metrics/internal/repository"
	"github.com/stretchr/testify/assert"
)

type mockStorage struct {
	gauges   map[string]repository.Gauge
	counters map[string]repository.Counter
}

func (m *mockStorage) UpdateGauge(name string, value repository.Gauge) {
	m.gauges[name] = value
}

func (m *mockStorage) UpdateCounter(name string, value repository.Counter) {
	m.counters[name] += value
}

func TestUpdateMetricHandler(t *testing.T) {
	store := &mockStorage{
		gauges:   make(map[string]repository.Gauge),
		counters: make(map[string]repository.Counter),
	}

	type want struct {
		contentType string
		statusCode  int
	}
	tests := []struct {
		name    string
		request string
		store   repository.Updater
		want    want
	}{
		{
			name: "simple test #1",
			want: want{
				contentType: "text/plain",
				statusCode:  200,
			},
			store:   store,
			request: "/update/gauge/LastGC/1.25",
		},
		{
			name: "simple test #2",
			want: want{
				contentType: "text/plain",
				statusCode:  404,
			},
			store:   store,
			request: "/update/gauge/",
		},
		{
			name: "simple test #3",
			want: want{
				contentType: "text/plain",
				statusCode:  400,
			},
			store:   store,
			request: "/update/wrong-type/test/3.5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(UpdateMetricHandler(tt.store))
			h(w, request)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
		})
	}
}
