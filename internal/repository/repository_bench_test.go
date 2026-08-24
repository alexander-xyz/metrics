package repository

import (
	"context"
	"strconv"
	"testing"

	models "github.com/alexander-xyz/metrics/internal/model"
)

func benchStore(size int) *MemStorage {
	store := NewMemStorage()
	ctx := context.Background()

	for i := 0; i < size; i++ {
		name := "metric" + strconv.Itoa(i)
		store.UpdateGauge(ctx, name, Gauge(i))
		store.UpdateCounter(ctx, name, Counter(i))
	}

	return store
}

func BenchmarkMemStorageUpdateGauge(b *testing.B) {
	store := NewMemStorage()
	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		store.UpdateGauge(ctx, "Alloc", 42)
	}
}

func BenchmarkMemStorageUpdateCounter(b *testing.B) {
	store := NewMemStorage()
	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		store.UpdateCounter(ctx, "PollCount", 1)
	}
}

func BenchmarkMemStorageGetGauges(b *testing.B) {
	store := benchStore(30)
	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		store.GetGauges(ctx)
	}
}

func BenchmarkMemStorageUpdateBatch(b *testing.B) {
	store := NewMemStorage()
	ctx := context.Background()

	value := 1.5
	delta := int64(1)
	metrics := make([]models.Metrics, 0, 30)

	for i := 0; i < 30; i++ {
		metrics = append(metrics, models.Metrics{
			ID:    "metric" + strconv.Itoa(i),
			MType: models.Gauge,
			Value: &value,
		})
	}

	metrics = append(metrics, models.Metrics{ID: "PollCount", MType: models.Counter, Delta: &delta})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		store.UpdateBatch(ctx, metrics)
	}
}
