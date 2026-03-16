package main

import (
	"net/http"
	"strconv"
	"strings"
)

type gauge float64
type counter int64

type MemStorage struct {
	gauges   map[string]gauge
	counters map[string]counter
}

type Updater interface {
	UpdateGauge(string, gauge)
	UpdateCounter(string, counter)
}

func (store *MemStorage) UpdateGauge(name string, value gauge) {
	store.gauges[name] = value
}

func (store *MemStorage) UpdateCounter(name string, value counter) {
	store.counters[name] += value
}

var store = MemStorage{
	gauges:   map[string]gauge{},
	counters: map[string]counter{},
}

func updateMetric(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)

		return
	}

	path := req.URL.Path
	parts := strings.Split(path, "/")

	if len(parts) != 5 {
		res.WriteHeader(http.StatusNotFound)

		return
	}

	metricType := parts[2]  // "gauge"
	metricName := parts[3]  // "temperature"
	metricValue := parts[4] // "1.5"

	if metricType != "gauge" && metricType != "counter" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if metricName == "" || metricValue == "" {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	if metricType == "gauge" {
		value, err := strconv.ParseFloat(metricValue, 64)

		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		store.UpdateGauge(metricName, gauge(value))
		res.WriteHeader(http.StatusOK)
		return
	}

	value, err := strconv.ParseInt(metricValue, 10, 64)

	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	store.UpdateCounter(metricName, counter(value))

	res.WriteHeader(http.StatusOK)
	return
}

func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, updateMetric)

	return http.ListenAndServe(`:8080`, mux)
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
