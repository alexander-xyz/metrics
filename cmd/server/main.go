package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/alexander-xyz/metrics/internal/handler"
	"github.com/alexander-xyz/metrics/internal/repository"
)

func RunServer(store handler.Storage, config *Config) error {
	router, err := handler.GetRouter(store)
	if err != nil {
		return fmt.Errorf("build router: %w", err)
	}

	return http.ListenAndServe(config.serverAddress, router)
}

func main() {
	config := parseFlags()
	store := repository.NewMemStorage()

	if err := RunServer(store, config); err != nil {
		log.Fatal(err)
	}
}
