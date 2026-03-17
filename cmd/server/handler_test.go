package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_updateMetricHandler(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}
	tests := []struct {
		name    string
		request string
		store   Updater
		want    want
	}{
		{
			name: "simple test #1",
			want: want{
				contentType: "text/plain",
				statusCode:  200,
			},
			store: &MemStorage{
				gauges:   map[string]gauge{},
				counters: map[string]counter{},
			},
			request: "/update/gauge/LastGC/1.25",
		},
		{
			name: "simple test #2",
			want: want{
				contentType: "text/plain",
				statusCode:  404,
			},
			store: &MemStorage{
				gauges:   map[string]gauge{},
				counters: map[string]counter{},
			},
			request: "/update/gauge/",
		},
		{
			name: "simple test #3",
			want: want{
				contentType: "text/plain",
				statusCode:  400,
			},
			store: &MemStorage{
				gauges:   map[string]gauge{},
				counters: map[string]counter{},
			},
			request: "/update/wrong-type/test/3.5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(updateMetricHandler(tt.store))
			h(w, request)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
		})
	}
}
