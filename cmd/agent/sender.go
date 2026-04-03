package main

import (
	"fmt"
	"net/http"

	"github.com/alexander-xyz/metrics/internal/repository"
)

func sendMetrics(store repository.Getter, config *Config) error {
	for t, v := range store.GetGauges() {
		url := fmt.Sprintf("%s/update/gauge/%s/%g", config.serverAddress, t, v)
		resp, err := http.Post(url, "text/plain", nil)
		if err != nil {
			return fmt.Errorf("error sending gauge metrics: %w", err)
		}
		resp.Body.Close()
	}

	for t, v := range store.GetCounters() {
		url := fmt.Sprintf("%s/update/counter/%s/%d", config.serverAddress, t, v)
		resp, err := http.Post(url, "text/plain", nil)
		if err != nil {
			return fmt.Errorf("error sending counter metrics: %w", err)
		}
		resp.Body.Close()
	}

	return nil
}
