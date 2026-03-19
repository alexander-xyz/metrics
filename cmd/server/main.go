package main

import (
	"github.com/alexander-xyz/metrics/internal/handler"
	"github.com/alexander-xyz/metrics/internal/repository"
)

func main() {
	store := repository.NewMemStorage()

	if err := handler.RunServer(store); err != nil {
		panic(err)
	}
}
