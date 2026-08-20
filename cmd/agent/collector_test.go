package main

import (
	"testing"

	"github.com/alexander-xyz/metrics/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestCollectMetrics(t *testing.T) {
	store := repository.NewMemStorage()
	collectMetrics(store)

	assert.Equal(t, 28, len(store.GetGauges()))
	assert.Equal(t, repository.Counter(1), store.GetCounters()["PollCount"])
	assert.Contains(t, store.GetGauges(), "RandomValue")
}
