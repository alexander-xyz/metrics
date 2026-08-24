package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	models "github.com/alexander-xyz/metrics/internal/model"
	"github.com/alexander-xyz/metrics/internal/repository"
)

func postMetric(url string, metric models.Metrics) error {
	body, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("marshal metric %s: %w", metric.ID, err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("send metric %s: %w", metric.ID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("send metric %s: unexpected status %d", metric.ID, resp.StatusCode)
	}

	return nil
}

func sendMetrics(store repository.Getter, config *Config) error {
	url := config.serverAddress + "/update"

	for name, value := range store.GetGauges() {
		v := float64(value)

		err := postMetric(url, models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &v,
		})
		if err != nil {
			return err
		}
	}

	for name, value := range store.GetCounters() {
		d := int64(value)

		err := postMetric(url, models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: &d,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
