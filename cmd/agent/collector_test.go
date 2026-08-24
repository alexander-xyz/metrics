package main

import (
	"context"
	"testing"

	"github.com/alexander-xyz/metrics/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectMetrics(t *testing.T) {
	ctx := context.Background()
	store := repository.NewMemStorage()
	require.NoError(t, collectMetrics(ctx, store))

	gauges, err := store.GetGauges(ctx)
	require.NoError(t, err)

	counters, err := store.GetCounters(ctx)
	require.NoError(t, err)

	assert.Equal(t, 28, len(gauges))
	assert.Equal(t, repository.Counter(1), counters["PollCount"])
	assert.Contains(t, gauges, "RandomValue")
}
