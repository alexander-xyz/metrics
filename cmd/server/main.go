package main

import (
	"net/http"

	"github.com/alexander-xyz/metrics/internal/handler"
	"github.com/alexander-xyz/metrics/internal/repository"
)

func run() error {
	store := repository.NewMemStorage()

	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, handler.UpdateMetricHandler(store))

	return http.ListenAndServe(`:8080`, mux)
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
