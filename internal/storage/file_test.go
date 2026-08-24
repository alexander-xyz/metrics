package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	models "github.com/alexander-xyz/metrics/internal/model"
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

func TestSyncStorageWritesOnCounterAndBatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "metrics.json")
	store := NewSyncStorage(repository.NewMemStorage(), path, func(err error) {
		t.Errorf("unexpected error: %v", err)
	})

	ctx := context.Background()

	require.NoError(t, store.UpdateCounter(ctx, "PollCount", 2))
	require.NoError(t, store.SetCounter(ctx, "PollCount", 5))

	value := 1.5
	require.NoError(t, store.UpdateBatch(ctx, []models.Metrics{
		{ID: "Alloc", MType: models.Gauge, Value: &value},
	}))

	restored := repository.NewMemStorage()
	require.NoError(t, Load(ctx, restored, path))

	counter, err := restored.GetCounter(ctx, "PollCount")
	require.NoError(t, err)
	assert.Equal(t, repository.Counter(5), counter)

	gauge, err := restored.GetGauge(ctx, "Alloc")
	require.NoError(t, err)
	assert.Equal(t, repository.Gauge(1.5), gauge)
}

func TestSyncStorageReportsWriteError(t *testing.T) {
	var got error

	store := NewSyncStorage(repository.NewMemStorage(), filepath.Join(t.TempDir(), "missing", "metrics.json"), func(err error) {
		got = err
	})

	require.NoError(t, store.UpdateGauge(context.Background(), "Alloc", 1))
	assert.Error(t, got, "ошибка записи передаётся в onErr")
}

func TestLoadRejectsBrokenFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "metrics.json")
	require.NoError(t, os.WriteFile(path, []byte("не json"), 0666))

	assert.Error(t, Load(context.Background(), repository.NewMemStorage(), path))
}
