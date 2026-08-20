package main

import (
	"flag"
)

type Config struct {
	serverAddress string
}

func parseFlags() *Config {
	config := Config{}

	flag.StringVar(&config.serverAddress, "a", "localhost:8080", "address and port to run server")
	flag.Parse()

	return &config
}
