package main

import (
	"bytes"
	"compress/gzip"
	"context"
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

func sendMetrics(ctx context.Context, store repository.Getter, config *Config) error {
	url := config.serverAddress + "/update"

	gauges, err := store.GetGauges(ctx)
	if err != nil {
		return fmt.Errorf("read gauges: %w", err)
	}

	counters, err := store.GetCounters(ctx)
	if err != nil {
		return fmt.Errorf("read counters: %w", err)
	}

	for name, value := range gauges {
		v := float64(value)

		if err := postMetric(url, models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &v,
		}); err != nil {
			return err
		}
	}

	for name, value := range counters {
		d := int64(value)

		if err := postMetric(url, models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: &d,
		}); err != nil {
			return err
		}
	}

	return nil
}
