package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

const defaultStoragePath = "/tmp/metrics-db.json"

type Config struct {
	serverAddress   string
	logLevel        string
	storeInterval   int64
	fileStoragePath string
	restore         bool
}

func parseFlags() (*Config, error) {
	config := Config{}

	flag.StringVar(&config.serverAddress, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&config.logLevel, "l", "info", "log level")
	flag.Int64Var(&config.storeInterval, "i", 300, "store interval in seconds, 0 makes writes synchronous")
	flag.StringVar(&config.fileStoragePath, "f", defaultStoragePath, "path to the metrics storage file")
	flag.BoolVar(&config.restore, "r", true, "restore metrics from the storage file on start")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		config.serverAddress = envRunAddr
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		config.logLevel = envLogLevel
	}

	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		value, err := strconv.ParseInt(envStoreInterval, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse STORE_INTERVAL: %w", err)
		}

		config.storeInterval = value
	}

	if envStoragePath := os.Getenv("FILE_STORAGE_PATH"); envStoragePath != "" {
		config.fileStoragePath = envStoragePath
	}

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		value, err := strconv.ParseBool(envRestore)
		if err != nil {
			return nil, fmt.Errorf("parse RESTORE: %w", err)
		}

		config.restore = value
	}

	return &config, nil
}
