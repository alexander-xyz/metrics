package main

import (
	"time"

	"github.com/alexander-xyz/metrics/internal/repository"
)

const server = "http://localhost:8080"

func main() {
	const poolInterval = 2
	const reportInterval = 10

	store := repository.NewMemStorage()

	currentPool := 0
	currentReport := 0

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
