package main

import (
	"math/rand"
	"runtime"

	"github.com/alexander-xyz/metrics/internal/repository"
)

func collectMetrics(store repository.Updater) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	store.UpdateGauge("Alloc", repository.Gauge(memStats.Alloc))
	store.UpdateGauge("BuckHashSys", repository.Gauge(memStats.BuckHashSys))
	store.UpdateGauge("Frees", repository.Gauge(memStats.Frees))
	store.UpdateGauge("GCCPUFraction", repository.Gauge(memStats.GCCPUFraction))
	store.UpdateGauge("GCSys", repository.Gauge(memStats.GCSys))
	store.UpdateGauge("HeapAlloc", repository.Gauge(memStats.HeapAlloc))
	store.UpdateGauge("HeapIdle", repository.Gauge(memStats.HeapIdle))
	store.UpdateGauge("HeapInuse", repository.Gauge(memStats.HeapInuse))
	store.UpdateGauge("HeapObjects", repository.Gauge(memStats.HeapObjects))
	store.UpdateGauge("HeapReleased", repository.Gauge(memStats.HeapReleased))
	store.UpdateGauge("HeapSys", repository.Gauge(memStats.HeapSys))
	store.UpdateGauge("LastGC", repository.Gauge(memStats.LastGC))
	store.UpdateGauge("Lookups", repository.Gauge(memStats.Lookups))
	store.UpdateGauge("MCacheInuse", repository.Gauge(memStats.MCacheInuse))
	store.UpdateGauge("MCacheSys", repository.Gauge(memStats.MCacheSys))
	store.UpdateGauge("MSpanInuse", repository.Gauge(memStats.MSpanInuse))
	store.UpdateGauge("MSpanSys", repository.Gauge(memStats.MSpanSys))
	store.UpdateGauge("Mallocs", repository.Gauge(memStats.Mallocs))
	store.UpdateGauge("NextGC", repository.Gauge(memStats.NextGC))
	store.UpdateGauge("NumForcedGC", repository.Gauge(memStats.NumForcedGC))
	store.UpdateGauge("NumGC", repository.Gauge(memStats.NumGC))
	store.UpdateGauge("OtherSys", repository.Gauge(memStats.OtherSys))
	store.UpdateGauge("PauseTotalNs", repository.Gauge(memStats.PauseTotalNs))
	store.UpdateGauge("StackInuse", repository.Gauge(memStats.StackInuse))
	store.UpdateGauge("StackSys", repository.Gauge(memStats.StackSys))
	store.UpdateGauge("Sys", repository.Gauge(memStats.Sys))
	store.UpdateGauge("TotalAlloc", repository.Gauge(memStats.TotalAlloc))
	store.UpdateGauge("RandomValue", repository.Gauge(rand.Float64()))
	store.UpdateCounter("PollCount", 1)
}
