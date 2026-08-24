package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	serverAddress  string
	pollInterval   int64
	reportInterval int64
}

func parseFlags() (*Config, error) {
	var flagRunAddr string

	config := Config{}

	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&config.reportInterval, "r", 10, "report interval in seconds")
	flag.Int64Var(&config.pollInterval, "p", 2, "update poll interval in seconds")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		flagRunAddr = envRunAddr
	}

	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		value, err := strconv.ParseInt(envReportInterval, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse REPORT_INTERVAL: %w", err)
		}

		config.reportInterval = value
	}

	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		value, err := strconv.ParseInt(envPollInterval, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse POLL_INTERVAL: %w", err)
		}

		config.pollInterval = value
	}

	if !strings.HasPrefix(flagRunAddr, "http") {
		config.serverAddress = "http://" + flagRunAddr
	} else {
		config.serverAddress = flagRunAddr
	}

	return &config, nil
}
