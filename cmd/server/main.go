package main

import (
	"net/http"
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

func run() error {
	store := MemStorage{
		gauges:   map[string]gauge{},
		counters: map[string]counter{},
	}

	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, updateMetricHandler(&store))

	return http.ListenAndServe(`:8080`, mux)
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
