package main

import (
	"net/http"

	"github.com/alexander-xyz/metrics/internal/handler"
	"github.com/alexander-xyz/metrics/internal/repository"
)

func RunServer(store handler.Storage) error {
	return http.ListenAndServe(flagRunAddr, handler.GetRouter(store))
}

func main() {
	parseFlags()
	store := repository.NewMemStorage()

	if err := RunServer(store); err != nil {
		panic(err)
	}
}
