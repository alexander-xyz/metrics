package handler

import (
	"crypto/rsa"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/alexander-xyz/metrics/internal/logger"
	"github.com/alexander-xyz/metrics/internal/repository"
	"github.com/go-chi/chi/v5"
)

// Storage объединяет операции хранилища, необходимые серверу метрик.
type Storage interface {
	repository.Updater
	repository.BatchUpdater
	repository.Getter
}

// GetRouter собирает маршрутизатор сервера метрик со всеми хендлерами
// и middleware: логированием, проверкой подписи, расшифровкой тела запроса
// и gzip-сжатием.
// Аргументы auditor и privateKey могут быть nil — тогда аудит запросов
// и расшифровка отключены.
func GetRouter(store Storage, db *sql.DB, key string, auditor Auditor, privateKey *rsa.PrivateKey) (*chi.Mux, error) {
	metricsHandler, err := GetMetricsHandler(store)
	if err != nil {
		return nil, fmt.Errorf("build metrics handler: %w", err)
	}

	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Use(logger.SignatureMiddleware(key))
	r.Use(logger.DecryptMiddleware(privateKey))
	r.Use(logger.GzipMiddleware)
	r.Get("/", metricsHandler)
	r.Get("/value/{type}/{id}", GetMetricHandler(store))
	r.Post("/update/{type}/{id}/{value}", UpdateMetricHandler(store, auditor))
	r.Post("/update", UpdateMetricJSONHandler(store, auditor))
	r.Post("/update/", UpdateMetricJSONHandler(store, auditor))
	r.Post("/value", GetMetricJSONHandler(store))
	r.Post("/value/", GetMetricJSONHandler(store))
	r.Post("/updates", UpdateMetricsJSONHandler(store, auditor))
	r.Post("/updates/", UpdateMetricsJSONHandler(store, auditor))
	r.Get("/ping", PingHandler(db))
	r.Mount("/debug/pprof", pprofRouter())

	return r, nil
}

func pprofRouter() *chi.Mux {
	r := chi.NewRouter()
	r.HandleFunc("/", pprof.Index)
	r.HandleFunc("/cmdline", pprof.Cmdline)
	r.HandleFunc("/profile", pprof.Profile)
	r.HandleFunc("/symbol", pprof.Symbol)
	r.HandleFunc("/trace", pprof.Trace)
	r.Method(http.MethodGet, "/{name}", http.HandlerFunc(pprof.Index))

	return r
}
