package main

import (
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

func TestCollectMetrics(t *testing.T) {
	store := &mockStorage{
		gauges:   make(map[string]repository.Gauge),
		counters: make(map[string]repository.Counter),
	}
	collectMetrics(store)

	assert.Equal(t, 28, len(store.gauges))
	assert.Equal(t, repository.Counter(1), store.counters["PollCount"])
	assert.Contains(t, store.gauges, "RandomValue")
}
