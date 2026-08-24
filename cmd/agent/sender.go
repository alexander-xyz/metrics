package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"

	models "github.com/alexander-xyz/metrics/internal/model"
	"github.com/alexander-xyz/metrics/internal/repository"
)

func compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	zw := gzip.NewWriter(&buf)

	if _, err := zw.Write(data); err != nil {
		return nil, fmt.Errorf("write gzip data: %w", err)
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

func postMetric(url string, metric models.Metrics) error {
	body, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("marshal metric %s: %w", metric.ID, err)
	}

	compressed, err := compress(body)
	if err != nil {
		return fmt.Errorf("compress metric %s: %w", metric.ID, err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(compressed))
	if err != nil {
		return fmt.Errorf("build request for metric %s: %w", metric.ID, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(req)
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
