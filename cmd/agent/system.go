package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/alexander-xyz/metrics/internal/repository"
)

func collectSystemMetrics(ctx context.Context, store repository.Updater) error {
	memory, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return fmt.Errorf("read memory stats: %w", err)
	}

	if err = store.UpdateGauge(ctx, "TotalMemory", repository.Gauge(memory.Total)); err != nil {
		return fmt.Errorf("collect TotalMemory: %w", err)
	}

	if err = store.UpdateGauge(ctx, "FreeMemory", repository.Gauge(memory.Free)); err != nil {
		return fmt.Errorf("collect FreeMemory: %w", err)
	}

	utilization, err := cpu.PercentWithContext(ctx, 0, true)
	if err != nil {
		return fmt.Errorf("read cpu stats: %w", err)
	}

	for i, value := range utilization {
		name := "CPUutilization" + strconv.Itoa(i+1)

		if err := store.UpdateGauge(ctx, name, repository.Gauge(value)); err != nil {
			return fmt.Errorf("collect %s: %w", name, err)
		}
	}

	return nil
}
