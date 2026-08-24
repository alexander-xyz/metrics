package main

import (
	"context"
	"log"
	"time"

	"github.com/alexander-xyz/metrics/internal/repository"
)

func main() {
	config, err := parseFlags()
	if err != nil {
		log.Fatal(err)
	}

	store := repository.NewMemStorage()
	ctx := context.Background()

	var currentPoll int64 = 0
	var currentReport int64 = 0

	for {
		time.Sleep(1 * time.Second)
		currentPoll++
		currentReport++

		if currentPoll == config.pollInterval {
			if err := collectMetrics(ctx, store); err != nil {
				log.Print(err)
			}

			currentPoll = 0
		}

		if currentReport == config.reportInterval {
			if err := sendMetrics(ctx, store, config); err != nil {
				log.Print(err)
			}

			if err := store.SetCounter(ctx, "PollCount", 0); err != nil {
				log.Print(err)
			}
			currentReport = 0
		}
	}
}
