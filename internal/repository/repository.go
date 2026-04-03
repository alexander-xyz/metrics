package repository

import "errors"

type Gauge float64
type Counter int64

type MemStorage struct {
	gauges   map[string]Gauge
	counters map[string]Counter
}

type Updater interface {
	UpdateGauge(string, Gauge)
	UpdateCounter(string, Counter)
	SetCounter(string, Counter)
}

type Getter interface {
	GetGauges() map[string]Gauge
	GetGauge(string) (Gauge, error)
	GetCounters() map[string]Counter
	GetCounter(string) (Counter, error)
}

func (store *MemStorage) UpdateGauge(name string, value Gauge) {
	store.gauges[name] = value
}

func (store *MemStorage) UpdateCounter(name string, value Counter) {
	store.counters[name] += value
}

func (store *MemStorage) SetCounter(name string, value Counter) {
	store.counters[name] = value
}

func (store *MemStorage) GetGauge(gauge string) (Gauge, error) {
	val, exist := store.gauges[gauge]

	if !exist {
		return Gauge(0), errors.New("gauge not found")
	}

	return val, nil
}

func (store *MemStorage) GetGauges() map[string]Gauge {
	return store.gauges
}

func (store *MemStorage) GetCounters() map[string]Counter {
	return store.counters
}

func (store *MemStorage) GetCounter(counter string) (Counter, error) {
	val, exist := store.counters[counter]

	if !exist {
		return Counter(0), errors.New("counter not found")
	}

	return val, nil
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   map[string]Gauge{},
		counters: map[string]Counter{},
	}
}
