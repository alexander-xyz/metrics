package main

import (
	"context"
	"crypto/rsa"
	"log"
	"sync"
	"time"

	"github.com/alexander-xyz/metrics/internal/crypt"
	models "github.com/alexander-xyz/metrics/internal/model"
	"github.com/alexander-xyz/metrics/internal/repository"
)

func pollRuntime(ctx context.Context, store repository.Updater, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := collectMetrics(ctx, store); err != nil {
				log.Print(err)
			}
		}
	}
}

func pollSystem(ctx context.Context, store repository.Updater, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := collectSystemMetrics(ctx, store); err != nil {
				log.Print(err)
			}
		}
	}
}

func report(ctx context.Context, store repository.Getter, jobs chan<- []models.Metrics, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			batch, err := collectBatch(ctx, store)
			if err != nil {
				log.Print(err)
				continue
			}

			if len(batch) == 0 {
				continue
			}

			select {
			case <-ctx.Done():
				return
			case jobs <- batch:
			}
		}
	}
}

func worker(ctx context.Context, jobs <-chan []models.Metrics, store repository.Updater, config *Config, publicKey *rsa.PublicKey, wg *sync.WaitGroup) {
	defer wg.Done()

	for batch := range jobs {
		if err := sendBatch(ctx, batch, config, publicKey); err != nil {
			log.Print(err)
			continue
		}

		if err := store.SetCounter(ctx, "PollCount", 0); err != nil {
			log.Print(err)
		}
	}
}

func main() {
	printBuildInfo()

	config, err := parseFlags()
	if err != nil {
		log.Fatal(err)
	}

	var publicKey *rsa.PublicKey

	if config.cryptoKey != "" {
		publicKey, err = crypt.LoadPublicKey(config.cryptoKey)
		if err != nil {
			log.Fatal(err)
		}
	}

	ctx := context.Background()
	store := repository.NewMemStorage()
	jobs := make(chan []models.Metrics, config.rateLimit)

	var wg sync.WaitGroup

	for i := int64(0); i < config.rateLimit; i++ {
		wg.Add(1)

		go worker(ctx, jobs, store, config, publicKey, &wg)
	}

	go pollRuntime(ctx, store, time.Duration(config.pollInterval)*time.Second)
	go pollSystem(ctx, store, time.Duration(config.pollInterval)*time.Second)

	report(ctx, store, jobs, time.Duration(config.reportInterval)*time.Second)

	close(jobs)
	wg.Wait()
}
