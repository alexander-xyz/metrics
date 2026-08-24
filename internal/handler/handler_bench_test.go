package handler

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/alexander-xyz/metrics/internal/repository"
)

func benchRouter(b *testing.B) http.Handler {
	router, err := GetRouter(repository.NewMemStorage(), nil, "", nil)
	if err != nil {
		b.Fatal(err)
	}

	return router
}

func benchBatch(size int) string {
	var sb strings.Builder

	sb.WriteByte('[')

	for i := 0; i < size; i++ {
		if i > 0 {
			sb.WriteByte(',')
		}

		sb.WriteString(`{"id":"metric`)
		sb.WriteString(strconv.Itoa(i))
		sb.WriteString(`","type":"gauge","value":1.5}`)
	}

	sb.WriteByte(']')

	return sb.String()
}

func BenchmarkUpdateMetricHandler(b *testing.B) {
	router := benchRouter(b)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/12.5", nil)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
	}
}

func BenchmarkUpdateMetricJSONHandler(b *testing.B) {
	router := benchRouter(b)
	body := `{"id":"Alloc","type":"gauge","value":12.5}`

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
	}
}

func BenchmarkUpdateMetricsJSONHandler(b *testing.B) {
	router := benchRouter(b)
	body := benchBatch(30)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/updates", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
	}
}

func BenchmarkUpdateMetricsJSONHandlerGzip(b *testing.B) {
	router := benchRouter(b)
	body := benchBatch(30)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/updates", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept-Encoding", "gzip")
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
	}
}

func BenchmarkGetMetricsHandler(b *testing.B) {
	store := repository.NewMemStorage()
	router, err := GetRouter(store, nil, "", nil)
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < 30; i++ {
		req := httptest.NewRequest(http.MethodPost, "/update/gauge/metric"+strconv.Itoa(i)+"/1.5", nil)
		router.ServeHTTP(httptest.NewRecorder(), req)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
	}
}
