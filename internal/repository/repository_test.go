package repository

import (
	"context"
	"testing"

	models "github.com/alexander-xyz/metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemStorageGauges(t *testing.T) {
	store := NewMemStorage()
	ctx := context.Background()

	require.NoError(t, store.UpdateGauge(ctx, "Alloc", 12.5))

	value, err := store.GetGauge(ctx, "Alloc")
	require.NoError(t, err)
	assert.Equal(t, Gauge(12.5), value)

	require.NoError(t, store.UpdateGauge(ctx, "Alloc", 3.5))

	value, err = store.GetGauge(ctx, "Alloc")
	require.NoError(t, err)
	assert.Equal(t, Gauge(3.5), value, "gauge замещает значение")

	_, err = store.GetGauge(ctx, "Missing")
	assert.ErrorIs(t, err, ErrNotFound)

	gauges, err := store.GetGauges(ctx)
	require.NoError(t, err)
	assert.Equal(t, map[string]Gauge{"Alloc": 3.5}, gauges)

	gauges["Alloc"] = 100
	stored, err := store.GetGauge(ctx, "Alloc")
	require.NoError(t, err)
	assert.Equal(t, Gauge(3.5), stored, "GetGauges возвращает копию")
}

func TestMemStorageCounters(t *testing.T) {
	store := NewMemStorage()
	ctx := context.Background()

	require.NoError(t, store.UpdateCounter(ctx, "PollCount", 2))
	require.NoError(t, store.UpdateCounter(ctx, "PollCount", 3))

	value, err := store.GetCounter(ctx, "PollCount")
	require.NoError(t, err)
	assert.Equal(t, Counter(5), value, "counter суммируется")

	require.NoError(t, store.SetCounter(ctx, "PollCount", 1))

	value, err = store.GetCounter(ctx, "PollCount")
	require.NoError(t, err)
	assert.Equal(t, Counter(1), value)

	_, err = store.GetCounter(ctx, "Missing")
	assert.ErrorIs(t, err, ErrNotFound)

	counters, err := store.GetCounters(ctx)
	require.NoError(t, err)
	assert.Equal(t, map[string]Counter{"PollCount": 1}, counters)

	counters["PollCount"] = 100
	stored, err := store.GetCounter(ctx, "PollCount")
	require.NoError(t, err)
	assert.Equal(t, Counter(1), stored, "GetCounters возвращает копию")
}

func TestMemStorageUpdateBatch(t *testing.T) {
	store := NewMemStorage()
	ctx := context.Background()

	value := 1.5
	delta := int64(4)

	metrics := []models.Metrics{
		{ID: "Alloc", MType: models.Gauge, Value: &value},
		{ID: "PollCount", MType: models.Counter, Delta: &delta},
		{ID: "PollCount", MType: models.Counter, Delta: &delta},
		{ID: "NoValue", MType: models.Gauge},
		{ID: "NoDelta", MType: models.Counter},
	}

	require.NoError(t, store.UpdateBatch(ctx, metrics))

	gauge, err := store.GetGauge(ctx, "Alloc")
	require.NoError(t, err)
	assert.Equal(t, Gauge(1.5), gauge)

	counter, err := store.GetCounter(ctx, "PollCount")
	require.NoError(t, err)
	assert.Equal(t, Counter(8), counter, "counter в батче суммируется")

	_, err = store.GetGauge(ctx, "NoValue")
	assert.ErrorIs(t, err, ErrNotFound, "метрика без значения не сохраняется")

	_, err = store.GetCounter(ctx, "NoDelta")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMemStorageConcurrentAccess(t *testing.T) {
	store := NewMemStorage()
	ctx := context.Background()

	done := make(chan struct{})

	for i := 0; i < 4; i++ {
		go func() {
			defer func() { done <- struct{}{} }()

			for j := 0; j < 100; j++ {
				store.UpdateCounter(ctx, "PollCount", 1)
				store.UpdateGauge(ctx, "Alloc", Gauge(j))
				store.GetCounters(ctx)
				store.GetGauges(ctx)
			}
		}()
	}

	for i := 0; i < 4; i++ {
		<-done
	}

	value, err := store.GetCounter(ctx, "PollCount")
	require.NoError(t, err)
	assert.Equal(t, Counter(400), value)
}
