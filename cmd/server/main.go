package main

import (
	"log"
	"net/http"

	"github.com/alexander-xyz/metrics/internal/handler"
	"github.com/alexander-xyz/metrics/internal/repository"
)

func RunServer(store handler.Storage, config *Config) error {
	return http.ListenAndServe(config.serverAddress, handler.GetRouter(store))
}

func main() {
	config := parseFlags()
	store := repository.NewMemStorage()

	if err := RunServer(store, config); err != nil {
		log.Fatal(err)
	}
}
