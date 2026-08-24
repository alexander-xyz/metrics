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
	key            string
	cryptoKey      string
	pollInterval   int64
	reportInterval int64
	rateLimit      int64
}

func parseFlags() (*Config, error) {
	var flagRunAddr string

	config := Config{}

	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&config.reportInterval, "r", 10, "report interval in seconds")
	flag.Int64Var(&config.pollInterval, "p", 2, "update poll interval in seconds")
	flag.StringVar(&config.key, "k", "", "key for request signature")
	flag.Int64Var(&config.rateLimit, "l", 1, "limit of simultaneous outgoing requests")
	flag.StringVar(&config.cryptoKey, "crypto-key", "", "path to the public key file")
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

	if envKey := os.Getenv("KEY"); envKey != "" {
		config.key = envKey
	}

	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		config.cryptoKey = envCryptoKey
	}

	if envRateLimit := os.Getenv("RATE_LIMIT"); envRateLimit != "" {
		value, err := strconv.ParseInt(envRateLimit, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse RATE_LIMIT: %w", err)
		}

		config.rateLimit = value
	}

	if !strings.HasPrefix(flagRunAddr, "http") {
		config.serverAddress = "http://" + flagRunAddr
	} else {
		config.serverAddress = flagRunAddr
	}

	return &config, nil
}
