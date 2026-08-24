package main

import (
	"flag"
	"os"
)

type Config struct {
	serverAddress string
	logLevel      string
}

func parseFlags() *Config {
	config := Config{}

	flag.StringVar(&config.serverAddress, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&config.logLevel, "l", "info", "log level")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		config.serverAddress = envRunAddr
	}

	return &config
}
