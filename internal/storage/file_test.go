package storage

import (
	"path/filepath"
	"testing"

	"github.com/alexander-xyz/metrics/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "metrics.json")

	source := repository.NewMemStorage()
	source.UpdateGauge("Alloc", repository.Gauge(1.5))
	source.UpdateCounter("PollCount", repository.Counter(7))

	require.NoError(t, Save(source, path))

	target := repository.NewMemStorage()
	require.NoError(t, Load(target, path))

	gauge, err := target.GetGauge("Alloc")
	require.NoError(t, err)
	assert.Equal(t, repository.Gauge(1.5), gauge)

	counter, err := target.GetCounter("PollCount")
	require.NoError(t, err)
	assert.Equal(t, repository.Counter(7), counter)
}

func TestLoadMissingFileIsNotAnError(t *testing.T) {
	store := repository.NewMemStorage()
	assert.NoError(t, Load(store, filepath.Join(t.TempDir(), "nope.json")))
}

func TestLoadDoesNotAccumulateCounters(t *testing.T) {
	path := filepath.Join(t.TempDir(), "metrics.json")

	source := repository.NewMemStorage()
	source.UpdateCounter("PollCount", repository.Counter(5))
	require.NoError(t, Save(source, path))

	target := repository.NewMemStorage()
	require.NoError(t, Load(target, path))
	require.NoError(t, Load(target, path))

	counter, err := target.GetCounter("PollCount")
	require.NoError(t, err)
	assert.Equal(t, repository.Counter(5), counter)
}

func TestSyncStorageWritesOnEveryUpdate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "metrics.json")

	sync := NewSyncStorage(repository.NewMemStorage(), path, func(err error) {
		t.Errorf("unexpected save error: %v", err)
	})
	sync.UpdateGauge("Alloc", repository.Gauge(3.25))

	target := repository.NewMemStorage()
	require.NoError(t, Load(target, path))

	gauge, err := target.GetGauge("Alloc")
	require.NoError(t, err)
	assert.Equal(t, repository.Gauge(3.25), gauge)
}
