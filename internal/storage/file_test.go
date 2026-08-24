package storage

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/alexander-xyz/metrics/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveAndLoad(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metrics.json")

	source := repository.NewMemStorage()
	require.NoError(t, source.UpdateGauge(ctx, "Alloc", repository.Gauge(1.5)))
	require.NoError(t, source.UpdateCounter(ctx, "PollCount", repository.Counter(7)))

	require.NoError(t, Save(ctx, source, path))

	target := repository.NewMemStorage()
	require.NoError(t, Load(ctx, target, path))

	gauge, err := target.GetGauge(ctx, "Alloc")
	require.NoError(t, err)
	assert.Equal(t, repository.Gauge(1.5), gauge)

	counter, err := target.GetCounter(ctx, "PollCount")
	require.NoError(t, err)
	assert.Equal(t, repository.Counter(7), counter)
}

func TestLoadMissingFileIsNotAnError(t *testing.T) {
	store := repository.NewMemStorage()
	assert.NoError(t, Load(context.Background(), store, filepath.Join(t.TempDir(), "nope.json")))
}

func TestLoadDoesNotAccumulateCounters(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metrics.json")

	source := repository.NewMemStorage()
	require.NoError(t, source.UpdateCounter(ctx, "PollCount", repository.Counter(5)))
	require.NoError(t, Save(ctx, source, path))

	target := repository.NewMemStorage()
	require.NoError(t, Load(ctx, target, path))
	require.NoError(t, Load(ctx, target, path))

	counter, err := target.GetCounter(ctx, "PollCount")
	require.NoError(t, err)
	assert.Equal(t, repository.Counter(5), counter)
}

func TestSyncStorageWritesOnEveryUpdate(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metrics.json")

	sync := NewSyncStorage(repository.NewMemStorage(), path, func(err error) {
		t.Errorf("unexpected save error: %v", err)
	})
	require.NoError(t, sync.UpdateGauge(ctx, "Alloc", repository.Gauge(3.25)))

	target := repository.NewMemStorage()
	require.NoError(t, Load(ctx, target, path))

	gauge, err := target.GetGauge(ctx, "Alloc")
	require.NoError(t, err)
	assert.Equal(t, repository.Gauge(3.25), gauge)
}
