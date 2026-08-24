package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/alexander-xyz/metrics/internal/config"
)

type Config struct {
	serverAddress  string
	key            string
	cryptoKey      string
	configFile     string
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
	flag.StringVar(&config.configFile, "c", "", "path to the JSON configuration file")
	flag.StringVar(&config.configFile, "config", "", "path to the JSON configuration file")
	flag.Parse()

	if err := applyConfigFile(&config, &flagRunAddr); err != nil {
		return nil, err
	}

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

// applyConfigFile дополняет конфигурацию агента значениями из JSON-файла.
// Опции, заданные флагом или переменной окружения, остаются как есть.
func applyConfigFile(target *Config, runAddr *string) error {
	path := config.Path(target.configFile)
	if path == "" {
		return nil
	}

	fromFile, err := config.LoadAgent(path)
	if err != nil {
		return err
	}

	set := setFlags()

	if !set["a"] && os.Getenv("ADDRESS") == "" && fromFile.Address != "" {
		*runAddr = fromFile.Address
	}

	if !set["crypto-key"] && os.Getenv("CRYPTO_KEY") == "" && fromFile.CryptoKey != "" {
		target.cryptoKey = fromFile.CryptoKey
	}

	if !set["r"] && os.Getenv("REPORT_INTERVAL") == "" && fromFile.ReportInterval.Duration != 0 {
		target.reportInterval = int64(fromFile.ReportInterval.Seconds())
	}

	if !set["p"] && os.Getenv("POLL_INTERVAL") == "" && fromFile.PollInterval.Duration != 0 {
		target.pollInterval = int64(fromFile.PollInterval.Seconds())
	}

	return nil
}

// setFlags возвращает имена флагов, заданных в командной строке явно.
func setFlags() map[string]bool {
	set := map[string]bool{}

	flag.Visit(func(f *flag.Flag) {
		set[f.Name] = true
	})

	return set
}
