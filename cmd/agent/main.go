package main

import (
	"log"
	"time"

	"github.com/alexander-xyz/metrics/internal/repository"
)

func main() {
	config := parseFlags()
	store := repository.NewMemStorage()

	var currentPool int64 = 0
	var currentReport int64 = 0

	for {
		time.Sleep(1 * time.Second)
		currentPool++
		currentReport++

		if currentPool == config.poolInterval {
			collectMetrics(store)
			currentPool = 0
		}

		if currentReport == config.reportInterval {
			err := sendMetrics(store, config)
			if err != nil {
				log.Print(err)
			}

			store.SetCounter("PollCount", 0)
			currentReport = 0
		}
	}
}
