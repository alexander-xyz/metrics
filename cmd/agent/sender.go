package main

import (
	"fmt"
	"net/http"

	"github.com/alexander-xyz/metrics/internal/repository"
)

func sendMetrics(store repository.Getter) error {
	for t, v := range store.GetGauges() {
		url := fmt.Sprintf("%s/update/gauge/%s/%g", serverAddr, t, v)
		_, err := http.Post(url, "text/plain", nil)
		if err != nil {
			return err
		}
	}

	for t, v := range store.GetCounters() {
		url := fmt.Sprintf("%s/update/counter/%s/%d", serverAddr, t, v)
		_, err := http.Post(url, "text/plain", nil)
		if err != nil {
			return err
		}
	}

	return nil
}
