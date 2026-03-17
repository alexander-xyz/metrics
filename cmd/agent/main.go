package main

import (
	"time"
)

const server = "http://localhost:8080"

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

func main() {
	const poolInterval = 2
	const reportInterval = 10

	store := MemStorage{
		gauges:   map[string]gauge{},
		counters: map[string]counter{},
	}

	currentPool := 0
	currentReport := 0

	for {
		time.Sleep(1 * time.Second)
		currentPool++
		currentReport++

		if currentPool == poolInterval {
			collectMetrics(&store)
			currentPool = 0
		}

		if currentReport == reportInterval {
			err := sendMetrics(store)
			if err != nil {
				panic(err)
			}
			currentReport = 0
		}
	}
}
