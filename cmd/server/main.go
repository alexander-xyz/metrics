package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/alexander-xyz/metrics/internal/handler"
	"github.com/alexander-xyz/metrics/internal/logger"
	"github.com/alexander-xyz/metrics/internal/repository"
	"github.com/alexander-xyz/metrics/internal/storage"
)

func saveWithInterval(store repository.Getter, config *Config) {
	ticker := time.NewTicker(time.Duration(config.storeInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := storage.Save(store, config.fileStoragePath); err != nil {
			logger.Log.Error(err.Error())
		}
	}
}

func openDatabase(dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, nil
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	return db, nil
}

func RunServer(store *repository.MemStorage, config *Config) error {
	if err := logger.Initialize(config.logLevel); err != nil {
		return fmt.Errorf("initialize logger: %w", err)
	}

	defer logger.Log.Sync()

	db, err := openDatabase(config.databaseDSN)
	if err != nil {
		return err
	}

	if db != nil {
		defer db.Close()

		logger.Log.Info("database configured", zap.String("dsn", config.databaseDSN))
	}

	if config.restore {
		if err := storage.Load(store, config.fileStoragePath); err != nil {
			return fmt.Errorf("restore metrics: %w", err)
		}

		logger.Log.Info("metrics restored", zap.String("file", config.fileStoragePath))
	}

	var served handler.Storage = store

	if config.storeInterval == 0 {
		served = storage.NewSyncStorage(store, config.fileStoragePath, func(err error) {
			logger.Log.Error(err.Error())
		})
	} else {
		go saveWithInterval(store, config)
	}

	router, err := handler.GetRouter(served, db)
	if err != nil {
		return fmt.Errorf("build router: %w", err)
	}

	logger.Log.Info("starting server",
		zap.String("address", config.serverAddress),
		zap.Int64("store_interval", config.storeInterval),
		zap.String("file", config.fileStoragePath),
	)

	return http.ListenAndServe(config.serverAddress, router)
}

func main() {
	config, err := parseFlags()
	if err != nil {
		log.Fatal(err)
	}

	store := repository.NewMemStorage()

	if err := RunServer(store, config); err != nil {
		log.Fatal(err)
	}
}
