package main

import (
	"math/rand"
	"runtime"
)

func collectMetrics(store Updater) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	store.UpdateGauge("Alloc", gauge(memStats.Alloc))
	store.UpdateGauge("BuckHashSys", gauge(memStats.BuckHashSys))
	store.UpdateGauge("Frees", gauge(memStats.Frees))
	store.UpdateGauge("GCCPUFraction", gauge(memStats.GCCPUFraction))
	store.UpdateGauge("GCSys", gauge(memStats.GCSys))
	store.UpdateGauge("HeapAlloc", gauge(memStats.HeapAlloc))
	store.UpdateGauge("HeapIdle", gauge(memStats.HeapIdle))
	store.UpdateGauge("HeapInuse", gauge(memStats.HeapInuse))
	store.UpdateGauge("HeapObjects", gauge(memStats.HeapObjects))
	store.UpdateGauge("HeapReleased", gauge(memStats.HeapReleased))
	store.UpdateGauge("HeapSys", gauge(memStats.HeapSys))
	store.UpdateGauge("LastGC", gauge(memStats.LastGC))
	store.UpdateGauge("Lookups", gauge(memStats.Lookups))
	store.UpdateGauge("MCacheInuse", gauge(memStats.MCacheInuse))
	store.UpdateGauge("MCacheSys", gauge(memStats.MCacheSys))
	store.UpdateGauge("MSpanInuse", gauge(memStats.MSpanInuse))
	store.UpdateGauge("MSpanSys", gauge(memStats.MSpanSys))
	store.UpdateGauge("Mallocs", gauge(memStats.Mallocs))
	store.UpdateGauge("NextGC", gauge(memStats.NextGC))
	store.UpdateGauge("NumForcedGC", gauge(memStats.NumForcedGC))
	store.UpdateGauge("NumGC", gauge(memStats.NumGC))
	store.UpdateGauge("OtherSys", gauge(memStats.OtherSys))
	store.UpdateGauge("PauseTotalNs", gauge(memStats.PauseTotalNs))
	store.UpdateGauge("StackInuse", gauge(memStats.StackInuse))
	store.UpdateGauge("StackSys", gauge(memStats.StackSys))
	store.UpdateGauge("Sys", gauge(memStats.Sys))
	store.UpdateGauge("TotalAlloc", gauge(memStats.TotalAlloc))
	store.UpdateGauge("RandomValue", gauge(rand.Float64()))
	store.UpdateCounter("PollCount", 1)
}
