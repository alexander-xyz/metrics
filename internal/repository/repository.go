package repository

import (
	"context"
	"errors"
	"sync"

	models "github.com/alexander-xyz/metrics/internal/model"
)

var ErrNotFound = errors.New("metric not found")

type Gauge float64
type Counter int64

type Updater interface {
	UpdateGauge(ctx context.Context, name string, value Gauge) error
	UpdateCounter(ctx context.Context, name string, value Counter) error
	SetCounter(ctx context.Context, name string, value Counter) error
}

type BatchUpdater interface {
	UpdateBatch(ctx context.Context, metrics []models.Metrics) error
}

type Getter interface {
	GetGauges(ctx context.Context) (map[string]Gauge, error)
	GetGauge(ctx context.Context, name string) (Gauge, error)
	GetCounters(ctx context.Context) (map[string]Counter, error)
	GetCounter(ctx context.Context, name string) (Counter, error)
}

type MemStorage struct {
	mu       sync.RWMutex
	gauges   map[string]Gauge
	counters map[string]Counter
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   map[string]Gauge{},
		counters: map[string]Counter{},
	}
}

func (store *MemStorage) UpdateGauge(_ context.Context, name string, value Gauge) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.gauges[name] = value

	return nil
}

func (store *MemStorage) UpdateCounter(_ context.Context, name string, value Counter) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.counters[name] += value

	return nil
}

func (store *MemStorage) SetCounter(_ context.Context, name string, value Counter) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.counters[name] = value

	return nil
}

func (store *MemStorage) GetGauge(_ context.Context, name string) (Gauge, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	value, exist := store.gauges[name]
	if !exist {
		return 0, ErrNotFound
	}

	return value, nil
}

func (store *MemStorage) GetGauges(_ context.Context) (map[string]Gauge, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	gauges := make(map[string]Gauge, len(store.gauges))
	for name, value := range store.gauges {
		gauges[name] = value
	}

	return gauges, nil
}

func (store *MemStorage) GetCounter(_ context.Context, name string) (Counter, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	value, exist := store.counters[name]
	if !exist {
		return 0, ErrNotFound
	}

	return value, nil
}

func (store *MemStorage) GetCounters(_ context.Context) (map[string]Counter, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	counters := make(map[string]Counter, len(store.counters))
	for name, value := range store.counters {
		counters[name] = value
	}

	return counters, nil
}

func (store *MemStorage) UpdateBatch(_ context.Context, metrics []models.Metrics) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value != nil {
				store.gauges[metric.ID] = Gauge(*metric.Value)
			}
		case models.Counter:
			if metric.Delta != nil {
				store.counters[metric.ID] += Counter(*metric.Delta)
			}
		}
	}

	return nil
}
