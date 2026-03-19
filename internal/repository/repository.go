package repository

type Gauge float64
type Counter int64

type MemStorage struct {
	gauges   map[string]Gauge
	counters map[string]Counter
}

type Updater interface {
	UpdateGauge(string, Gauge)
	UpdateCounter(string, Counter)
}

type Getter interface {
	GetGauges() map[string]Gauge
	GetCounters() map[string]Counter
}

func (store *MemStorage) UpdateGauge(name string, value Gauge) {
	store.gauges[name] = value
}

func (store *MemStorage) UpdateCounter(name string, value Counter) {
	store.counters[name] += value
}

func (store *MemStorage) GetGauges() map[string]Gauge {
	return store.gauges
}

func (store *MemStorage) GetCounters() map[string]Counter {
	return store.counters
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   map[string]Gauge{},
		counters: map[string]Counter{},
	}
}
