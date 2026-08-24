package main

import (
	"fmt"
	"log"
	"net/http"

	"go.uber.org/zap"

	"github.com/alexander-xyz/metrics/internal/handler"
	"github.com/alexander-xyz/metrics/internal/logger"
	"github.com/alexander-xyz/metrics/internal/repository"
)

func RunServer(store handler.Storage, config *Config) error {
	if err := logger.Initialize(config.logLevel); err != nil {
		return fmt.Errorf("initialize logger: %w", err)
	}

	defer logger.Log.Sync()

	router, err := handler.GetRouter(store)
	if err != nil {
		return fmt.Errorf("build router: %w", err)
	}

	logger.Log.Info("starting server", zap.String("address", config.serverAddress))

	return http.ListenAndServe(config.serverAddress, router)
}

func main() {
	config := parseFlags()
	store := repository.NewMemStorage()

	if err := RunServer(store, config); err != nil {
		log.Fatal(err)
	}
}
