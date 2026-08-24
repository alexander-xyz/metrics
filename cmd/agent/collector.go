package main

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"

	"github.com/alexander-xyz/metrics/internal/repository"
)

func collectMetrics(ctx context.Context, store repository.Updater) error {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	gauges := [...]struct {
		name  string
		value repository.Gauge
	}{
		{"Alloc", repository.Gauge(memStats.Alloc)},
		{"BuckHashSys", repository.Gauge(memStats.BuckHashSys)},
		{"Frees", repository.Gauge(memStats.Frees)},
		{"GCCPUFraction", repository.Gauge(memStats.GCCPUFraction)},
		{"GCSys", repository.Gauge(memStats.GCSys)},
		{"HeapAlloc", repository.Gauge(memStats.HeapAlloc)},
		{"HeapIdle", repository.Gauge(memStats.HeapIdle)},
		{"HeapInuse", repository.Gauge(memStats.HeapInuse)},
		{"HeapObjects", repository.Gauge(memStats.HeapObjects)},
		{"HeapReleased", repository.Gauge(memStats.HeapReleased)},
		{"HeapSys", repository.Gauge(memStats.HeapSys)},
		{"LastGC", repository.Gauge(memStats.LastGC)},
		{"Lookups", repository.Gauge(memStats.Lookups)},
		{"MCacheInuse", repository.Gauge(memStats.MCacheInuse)},
		{"MCacheSys", repository.Gauge(memStats.MCacheSys)},
		{"MSpanInuse", repository.Gauge(memStats.MSpanInuse)},
		{"MSpanSys", repository.Gauge(memStats.MSpanSys)},
		{"Mallocs", repository.Gauge(memStats.Mallocs)},
		{"NextGC", repository.Gauge(memStats.NextGC)},
		{"NumForcedGC", repository.Gauge(memStats.NumForcedGC)},
		{"NumGC", repository.Gauge(memStats.NumGC)},
		{"OtherSys", repository.Gauge(memStats.OtherSys)},
		{"PauseTotalNs", repository.Gauge(memStats.PauseTotalNs)},
		{"StackInuse", repository.Gauge(memStats.StackInuse)},
		{"StackSys", repository.Gauge(memStats.StackSys)},
		{"Sys", repository.Gauge(memStats.Sys)},
		{"TotalAlloc", repository.Gauge(memStats.TotalAlloc)},
		{"RandomValue", repository.Gauge(rand.Float64())},
	}

	for _, gauge := range gauges {
		if err := store.UpdateGauge(ctx, gauge.name, gauge.value); err != nil {
			return fmt.Errorf("collect %s: %w", gauge.name, err)
		}
	}

	if err := store.UpdateCounter(ctx, "PollCount", 1); err != nil {
		return fmt.Errorf("collect PollCount: %w", err)
	}

	return nil
}
