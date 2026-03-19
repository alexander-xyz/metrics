package main

import (
	"time"

	"github.com/alexander-xyz/metrics/internal/repository"
)

func main() {
	parseFlags()
	store := repository.NewMemStorage()

	var currentPool int64 = 0
	var currentReport int64 = 0

	for {
		time.Sleep(1 * time.Second)
		currentPool++
		currentReport++

		if currentPool == poolInterval {
			collectMetrics(store)
			currentPool = 0
		}

		if currentReport == reportInterval {
			err := sendMetrics(store)
			if err != nil {
				panic(err)
			}
			currentReport = 0
		}
	}
}
