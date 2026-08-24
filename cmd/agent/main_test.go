package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	models "github.com/alexander-xyz/metrics/internal/model"
	"github.com/alexander-xyz/metrics/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectSystemMetrics(t *testing.T) {
	store := repository.NewMemStorage()
	ctx := context.Background()

	require.NoError(t, collectSystemMetrics(ctx, store))

	total, err := store.GetGauge(ctx, "TotalMemory")
	require.NoError(t, err)
	assert.Positive(t, float64(total))

	_, err = store.GetGauge(ctx, "FreeMemory")
	require.NoError(t, err)

	_, err = store.GetGauge(ctx, "CPUutilization1")
	require.NoError(t, err)
}

func TestPollRuntimeStopsOnContext(t *testing.T) {
	store := repository.NewMemStorage()
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})

	go func() {
		pollRuntime(ctx, store, 10*time.Millisecond)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("pollRuntime не завершился после отмены контекста")
	}

	value, err := store.GetCounter(context.Background(), "PollCount")
	require.NoError(t, err)
	assert.Positive(t, int64(value), "метрики успели собраться")
}

func TestReportSendsBatch(t *testing.T) {
	store := repository.NewMemStorage()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.NoError(t, store.UpdateGauge(ctx, "Alloc", 1.5))

	jobs := make(chan []models.Metrics, 1)

	go report(ctx, store, jobs, 10*time.Millisecond)

	select {
	case batch := <-jobs:
		require.Len(t, batch, 1)
		assert.Equal(t, "Alloc", batch[0].ID)
	case <-time.After(time.Second):
		t.Fatal("батч не отправлен в очередь")
	}
}

func TestWorkerResetsPollCount(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	store := repository.NewMemStorage()
	ctx := context.Background()

	require.NoError(t, store.UpdateCounter(ctx, "PollCount", 7))

	jobs := make(chan []models.Metrics, 1)
	value := 1.5
	jobs <- []models.Metrics{{ID: "Alloc", MType: models.Gauge, Value: &value}}
	close(jobs)

	var wg sync.WaitGroup
	wg.Add(1)

	config := &Config{serverAddress: srv.URL}
	worker(ctx, jobs, store, config, &wg)
	wg.Wait()

	counter, err := store.GetCounter(ctx, "PollCount")
	require.NoError(t, err)
	assert.Equal(t, repository.Counter(0), counter, "после отправки счётчик сбрасывается")
}
