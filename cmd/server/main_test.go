package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexander-xyz/metrics/internal/audit"
	"github.com/alexander-xyz/metrics/internal/logger"
	"github.com/alexander-xyz/metrics/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenDatabaseWithoutDSN(t *testing.T) {
	db, err := openDatabase("")
	require.NoError(t, err)
	assert.Nil(t, db, "без DSN соединение не открывается")
}

func TestOpenDatabaseWithDSN(t *testing.T) {
	db, err := openDatabase("postgres://user:pass@localhost:5432/db?sslmode=disable")
	require.NoError(t, err)
	require.NotNil(t, db, "соединение создаётся лениво, без обращения к базе")

	require.NoError(t, db.Close())
}

func TestBuildAuditorWithoutParameters(t *testing.T) {
	require.NoError(t, logger.Initialize("info"))

	publisher := buildAuditor(&Config{})
	require.NotNil(t, publisher)

	publisher.Notify(audit.Event{Timestamp: 1})
}

func TestBuildAuditorWritesToFileAndURL(t *testing.T) {
	require.NoError(t, logger.Initialize("info"))

	received := make(chan struct{}, 1)

	srv := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		received <- struct{}{}
		res.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	path := filepath.Join(t.TempDir(), "audit.log")

	publisher := buildAuditor(&Config{auditFile: path, auditURL: srv.URL})
	publisher.Notify(audit.Event{Timestamp: 1, Metrics: []string{"Alloc"}, IPAddress: "127.0.0.1"})

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"metrics":["Alloc"]`)

	select {
	case <-received:
	case <-time.After(2 * time.Second):
		t.Fatal("событие не дошло до удалённого приёмника")
	}
}

func TestBuildStorageInMemory(t *testing.T) {
	require.NoError(t, logger.Initialize("info"))

	store, err := buildStorage(context.Background(), nil, &Config{fileStoragePath: ""})
	require.NoError(t, err)
	require.NotNil(t, store)

	require.NoError(t, store.UpdateGauge(context.Background(), "Alloc", 1.5))

	value, err := store.GetGauge(context.Background(), "Alloc")
	require.NoError(t, err)
	assert.Equal(t, repository.Gauge(1.5), value)
}

func TestBuildStorageSynchronousFileWrites(t *testing.T) {
	require.NoError(t, logger.Initialize("info"))

	path := filepath.Join(t.TempDir(), "metrics.json")
	ctx := context.Background()

	store, err := buildStorage(ctx, nil, &Config{fileStoragePath: path, storeInterval: 0})
	require.NoError(t, err)

	require.NoError(t, store.UpdateGauge(ctx, "Alloc", 1.5))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "Alloc", "метрика записана сразу после обновления")
}

func TestBuildStorageRestoresFromFile(t *testing.T) {
	require.NoError(t, logger.Initialize("info"))

	path := filepath.Join(t.TempDir(), "metrics.json")
	ctx := context.Background()

	first, err := buildStorage(ctx, nil, &Config{fileStoragePath: path, storeInterval: 0})
	require.NoError(t, err)
	require.NoError(t, first.UpdateCounter(ctx, "PollCount", 3))

	second, err := buildStorage(ctx, nil, &Config{fileStoragePath: path, storeInterval: 0, restore: true})
	require.NoError(t, err)

	value, err := second.GetCounter(ctx, "PollCount")
	require.NoError(t, err)
	assert.Equal(t, repository.Counter(3), value)
}

func TestBuildStorageRejectsBrokenFile(t *testing.T) {
	require.NoError(t, logger.Initialize("info"))

	path := filepath.Join(t.TempDir(), "metrics.json")
	require.NoError(t, os.WriteFile(path, []byte("не json"), 0666))

	_, err := buildStorage(context.Background(), nil, &Config{fileStoragePath: path, restore: true})
	assert.Error(t, err)
}

func TestShutdownSavesMetrics(t *testing.T) {
	require.NoError(t, logger.Initialize("info"))

	path := filepath.Join(t.TempDir(), "metrics.json")
	ctx := context.Background()

	store := repository.NewMemStorage()
	require.NoError(t, store.UpdateGauge(ctx, "Alloc", 777))

	srv := &http.Server{Addr: "localhost:0"}
	config := &Config{fileStoragePath: path, storeInterval: 3600}

	require.NoError(t, shutdown(srv, nil, store, config))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "Alloc", "несохранённые метрики записаны при остановке")
}

func TestShutdownSkipsSaveWithDatabase(t *testing.T) {
	require.NoError(t, logger.Initialize("info"))

	path := filepath.Join(t.TempDir(), "metrics.json")

	db, err := openDatabase("postgres://user:pass@localhost:5432/db?sslmode=disable")
	require.NoError(t, err)

	defer db.Close()

	srv := &http.Server{Addr: "localhost:0"}
	config := &Config{fileStoragePath: path}

	require.NoError(t, shutdown(srv, db, repository.NewMemStorage(), config))

	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err), "с базой данных файл не создаётся")
}

func TestRunServerStopsOnContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	config := &Config{
		serverAddress: "localhost:0",
		logLevel:      "info",
		storeInterval: 3600,
	}

	done := make(chan error, 1)

	go func() { done <- RunServer(ctx, config) }()

	time.Sleep(300 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("сервер не завершился после отмены контекста")
	}
}
