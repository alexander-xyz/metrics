package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockStorage struct {
	gauges   map[string]gauge
	counters map[string]counter
}

func (m *mockStorage) UpdateGauge(name string, value gauge) {
	m.gauges[name] = value
}

func (m *mockStorage) UpdateCounter(name string, value counter) {
	m.counters[name] += value
}

func TestCollectMetrics(t *testing.T) {
	store := &mockStorage{
		gauges:   make(map[string]gauge),
		counters: make(map[string]counter),
	}
	collectMetrics(store)

	assert.Equal(t, 28, len(store.gauges))
	assert.Equal(t, counter(1), store.counters["PollCount"])
	assert.Contains(t, store.gauges, "RandomValue")
}
