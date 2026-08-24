package main

import (
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

	var currentPoll int64 = 0
	var currentReport int64 = 0

	for {
		time.Sleep(1 * time.Second)
		currentPoll++
		currentReport++

		if currentPoll == config.pollInterval {
			collectMetrics(store)
			currentPoll = 0
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
