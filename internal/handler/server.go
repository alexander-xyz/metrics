package handler

import (
	"fmt"

	"github.com/alexander-xyz/metrics/internal/repository"
	"github.com/go-chi/chi/v5"
)

type Storage interface {
	repository.Updater
	repository.Getter
}

func GetRouter(store Storage) (*chi.Mux, error) {
	metricsHandler, err := GetMetricsHandler(store)
	if err != nil {
		return nil, fmt.Errorf("build metrics handler: %w", err)
	}

	r := chi.NewRouter()
	r.Get("/", metricsHandler)
	r.Get("/value/{type}/{id}", GetMetricHandler(store))
	r.Post("/update/{type}/{id}/{value}", UpdateMetricHandler(store))

	return r, nil
}
