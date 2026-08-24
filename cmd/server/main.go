package main

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/alexander-xyz/metrics/internal/audit"
	"github.com/alexander-xyz/metrics/internal/crypt"
	"github.com/alexander-xyz/metrics/internal/handler"
	"github.com/alexander-xyz/metrics/internal/logger"
	"github.com/alexander-xyz/metrics/internal/repository"
	"github.com/alexander-xyz/metrics/internal/storage"
)

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

func saveWithInterval(ctx context.Context, store repository.Getter, config *Config) {
	ticker := time.NewTicker(time.Duration(config.storeInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := storage.Save(ctx, store, config.fileStoragePath); err != nil {
			logger.Log.Error(err.Error())
		}
	}
}

func buildStorage(ctx context.Context, db *sql.DB, config *Config) (handler.Storage, error) {
	if db != nil {
		if err := storage.Migrate(db); err != nil {
			return nil, err
		}

		logger.Log.Info("using database storage")

		return repository.NewPostgresStorage(db), nil
	}

	store := repository.NewMemStorage()

	if config.fileStoragePath == "" {
		logger.Log.Info("using in-memory storage")

		return store, nil
	}

	if config.restore {
		if err := storage.Load(ctx, store, config.fileStoragePath); err != nil {
			return nil, fmt.Errorf("restore metrics: %w", err)
		}

		logger.Log.Info("metrics restored", zap.String("file", config.fileStoragePath))
	}

	logger.Log.Info("using file storage", zap.String("file", config.fileStoragePath))

	if config.storeInterval == 0 {
		return storage.NewSyncStorage(store, config.fileStoragePath, func(err error) {
			logger.Log.Error(err.Error())
		}), nil
	}

	go saveWithInterval(ctx, store, config)

	return store, nil
}

func buildAuditor(config *Config) *audit.Publisher {
	publisher := audit.NewPublisher()

	onError := func(err error) {
		logger.Log.Error(err.Error())
	}

	if config.auditFile != "" {
		publisher.Register(audit.NewFileObserver(config.auditFile, onError))
		logger.Log.Info("audit to file enabled", zap.String("file", config.auditFile))
	}

	if config.auditURL != "" {
		publisher.Register(audit.NewHTTPObserver(config.auditURL, onError))
		logger.Log.Info("audit to url enabled", zap.String("url", config.auditURL))
	}

	return publisher
}

func RunServer(ctx context.Context, config *Config) error {
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
	}

	store, err := buildStorage(ctx, db, config)
	if err != nil {
		return err
	}

	var privateKey *rsa.PrivateKey

	if config.cryptoKey != "" {
		privateKey, err = crypt.LoadPrivateKey(config.cryptoKey)
		if err != nil {
			return fmt.Errorf("load private key: %w", err)
		}

		logger.Log.Info("request decryption enabled", zap.String("key", config.cryptoKey))
	}

	router, err := handler.GetRouter(store, db, config.key, buildAuditor(config), privateKey)
	if err != nil {
		return fmt.Errorf("build router: %w", err)
	}

	logger.Log.Info("starting server", zap.String("address", config.serverAddress))

	return http.ListenAndServe(config.serverAddress, router)
}

func main() {
	printBuildInfo()

	config, err := parseFlags()
	if err != nil {
		log.Fatal(err)
	}

	if err := RunServer(context.Background(), config); err != nil {
		log.Fatal(err)
	}
}
