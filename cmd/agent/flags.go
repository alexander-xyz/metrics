package main

import (
	"flag"
	"strings"
)

type Config struct {
	serverAddress  string
	poolInterval   int64
	reportInterval int64
}

func parseFlags() *Config {
	var flagRunAddr string

	config := Config{}

	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&config.reportInterval, "r", 10, "report interval in seconds")
	flag.Int64Var(&config.poolInterval, "p", 2, "update pool interval in seconds")
	flag.Parse()

	if !strings.HasPrefix(flagRunAddr, "http") {
		config.serverAddress = "http://" + flagRunAddr
	} else {
		config.serverAddress = flagRunAddr
	}

	return &config
}
