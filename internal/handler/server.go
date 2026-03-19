package handler

import (
	"net/http"

	"github.com/alexander-xyz/metrics/internal/repository"
	"github.com/go-chi/chi/v5"
)

type Storage interface {
	repository.Updater
	repository.Getter
}

func GetRouter(store Storage) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", GetMetricsHandler(store))
	r.Get("/value/{type}/{id}", GetMetricHandler(store))
	r.Post("/update/{type}/{id}/{value}", UpdateMetricHandler(store))

	return r
}

func RunServer(store Storage) error {
	return http.ListenAndServe(`:8080`, GetRouter(store))
}
