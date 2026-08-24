package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"sync"

	models "github.com/alexander-xyz/metrics/internal/model"
	"github.com/alexander-xyz/metrics/internal/repository"
	"github.com/alexander-xyz/metrics/internal/signature"
)

var gzipWriterPool = sync.Pool{
	New: func() any {
		return gzip.NewWriter(nil)
	},
}

func compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	zw := gzipWriterPool.Get().(*gzip.Writer)
	zw.Reset(&buf)

	defer gzipWriterPool.Put(zw)

	if _, err := zw.Write(data); err != nil {
		return nil, fmt.Errorf("write gzip data: %w", err)
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

var retryDelays = []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}

func postBatchWithRetry(ctx context.Context, url, key string, metrics []models.Metrics) error {
	err := postBatch(url, key, metrics)
	if err == nil {
		return nil
	}

	for _, delay := range retryDelays {
		select {
		case <-ctx.Done():
			return err
		case <-time.After(delay):
		}

		if err = postBatch(url, key, metrics); err == nil {
			return nil
		}
	}

	return err
}

func postBatch(url, key string, metrics []models.Metrics) error {
	body, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("marshal metrics: %w", err)
	}

	compressed, err := compress(body)
	if err != nil {
		return fmt.Errorf("compress metrics: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(compressed))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	if key != "" {
		req.Header.Set(signature.Header, signature.Sign(compressed, key))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("send metrics: unexpected status %d", resp.StatusCode)
	}

	return nil
}

func collectBatch(ctx context.Context, store repository.Getter) ([]models.Metrics, error) {
	gauges, err := store.GetGauges(ctx)
	if err != nil {
		return nil, fmt.Errorf("read gauges: %w", err)
	}

	counters, err := store.GetCounters(ctx)
	if err != nil {
		return nil, fmt.Errorf("read counters: %w", err)
	}

	metrics := make([]models.Metrics, 0, len(gauges)+len(counters))

	for name, value := range gauges {
		v := float64(value)
		metrics = append(metrics, models.Metrics{ID: name, MType: models.Gauge, Value: &v})
	}

	for name, value := range counters {
		d := int64(value)
		metrics = append(metrics, models.Metrics{ID: name, MType: models.Counter, Delta: &d})
	}

	return metrics, nil
}

func sendBatch(ctx context.Context, metrics []models.Metrics, config *Config) error {
	return postBatchWithRetry(ctx, config.serverAddress+"/updates/", config.key, metrics)
}
