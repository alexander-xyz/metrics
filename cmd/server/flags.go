package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/alexander-xyz/metrics/internal/config"
)

const defaultStoragePath = "/tmp/metrics-db.json"

type Config struct {
	serverAddress   string
	logLevel        string
	fileStoragePath string
	databaseDSN     string
	key             string
	auditFile       string
	auditURL        string
	cryptoKey       string
	configFile      string
	trustedSubnet   string
	storeInterval   int64
	restore         bool
}

func parseFlags() (*Config, error) {
	config := Config{}

	flag.StringVar(&config.serverAddress, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&config.logLevel, "l", "info", "log level")
	flag.Int64Var(&config.storeInterval, "i", 300, "store interval in seconds, 0 makes writes synchronous")
	flag.StringVar(&config.fileStoragePath, "f", defaultStoragePath, "path to the metrics storage file")
	flag.BoolVar(&config.restore, "r", true, "restore metrics from the storage file on start")
	flag.StringVar(&config.databaseDSN, "d", "", "database connection string")
	flag.StringVar(&config.key, "k", "", "key for request signature")
	flag.StringVar(&config.auditFile, "audit-file", "", "path to the audit log file")
	flag.StringVar(&config.auditURL, "audit-url", "", "url of the remote audit receiver")
	flag.StringVar(&config.cryptoKey, "crypto-key", "", "path to the private key file")
	flag.StringVar(&config.configFile, "c", "", "path to the JSON configuration file")
	flag.StringVar(&config.configFile, "config", "", "path to the JSON configuration file")
	flag.StringVar(&config.trustedSubnet, "t", "", "trusted subnet in CIDR notation")
	flag.Parse()

	if err := applyConfigFile(&config); err != nil {
		return nil, err
	}

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

	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		config.databaseDSN = envDatabaseDSN
	}

	if envKey := os.Getenv("KEY"); envKey != "" {
		config.key = envKey
	}

	if envAuditFile := os.Getenv("AUDIT_FILE"); envAuditFile != "" {
		config.auditFile = envAuditFile
	}

	if envAuditURL := os.Getenv("AUDIT_URL"); envAuditURL != "" {
		config.auditURL = envAuditURL
	}

	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		config.cryptoKey = envCryptoKey
	}

	if envSubnet := os.Getenv("TRUSTED_SUBNET"); envSubnet != "" {
		config.trustedSubnet = envSubnet
	}

	return &config, nil
}

// applyConfigFile дополняет конфигурацию значениями из JSON-файла.
// Опции, заданные флагом или переменной окружения, остаются как есть.
func applyConfigFile(target *Config) error {
	path := config.Path(target.configFile)
	if path == "" {
		return nil
	}

	fromFile, err := config.LoadServer(path)
	if err != nil {
		return err
	}

	set := setFlags()

	if !set["a"] && os.Getenv("ADDRESS") == "" && fromFile.Address != "" {
		target.serverAddress = fromFile.Address
	}

	if !set["f"] && os.Getenv("FILE_STORAGE_PATH") == "" && fromFile.StoreFile != "" {
		target.fileStoragePath = fromFile.StoreFile
	}

	if !set["d"] && os.Getenv("DATABASE_DSN") == "" && fromFile.DatabaseDSN != "" {
		target.databaseDSN = fromFile.DatabaseDSN
	}

	if !set["crypto-key"] && os.Getenv("CRYPTO_KEY") == "" && fromFile.CryptoKey != "" {
		target.cryptoKey = fromFile.CryptoKey
	}

	if !set["i"] && os.Getenv("STORE_INTERVAL") == "" && fromFile.StoreInterval.Duration != 0 {
		target.storeInterval = int64(fromFile.StoreInterval.Seconds())
	}

	if !set["r"] && os.Getenv("RESTORE") == "" {
		target.restore = fromFile.Restore
	}

	if !set["t"] && os.Getenv("TRUSTED_SUBNET") == "" && fromFile.TrustedSubnet != "" {
		target.trustedSubnet = fromFile.TrustedSubnet
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
