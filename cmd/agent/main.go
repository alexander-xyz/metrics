package main

import (
	"context"
	"crypto/rsa"
	"log"
	"os/signal"
	"sync"
	"syscall"
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

func worker(ctx context.Context, jobs <-chan []models.Metrics, store repository.Updater, config *Config, publicKey *rsa.PublicKey, sender *grpcSender, wg *sync.WaitGroup) {
	defer wg.Done()

	for batch := range jobs {
		if err := send(ctx, batch, config, publicKey, sender); err != nil {
			log.Print(err)
			continue
		}

		if err := store.SetCounter(ctx, "PollCount", 0); err != nil {
			log.Print(err)
		}
	}
}

// sendTimeout ограничивает отправку батча, который дошёл до воркера
// после сигнала остановки.
const sendTimeout = 5 * time.Second

// send отправляет батч метрик выбранным транспортом: по gRPC, если он
// настроен, иначе по HTTP. Если контекст уже отменён сигналом остановки,
// отправка выполняется с собственным таймаутом: накопленные метрики
// нужно передать серверу до выхода.
func send(ctx context.Context, batch []models.Metrics, config *Config, publicKey *rsa.PublicKey, sender *grpcSender) error {
	if ctx.Err() != nil {
		var cancel context.CancelFunc

		ctx, cancel = context.WithTimeout(context.WithoutCancel(ctx), sendTimeout)
		defer cancel()
	}

	if sender != nil {
		return sender.Send(ctx, batch)
	}

	return sendBatch(ctx, batch, config, publicKey)
}

// flush ставит в очередь метрики, собранные к моменту остановки,
// чтобы воркеры успели отправить их на сервер.
func flush(store repository.Getter, jobs chan<- []models.Metrics) {
	batch, err := collectBatch(context.Background(), store)
	if err != nil {
		log.Print(err)
		return
	}

	if len(batch) == 0 {
		return
	}

	select {
	case jobs <- batch:
	default:
		log.Print("job queue is full, metrics are dropped")
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

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	var sender *grpcSender

	if config.grpcAddress != "" {
		sender, err = newGRPCSender(config.grpcAddress)
		if err != nil {
			log.Fatal(err)
		}

		defer sender.Close()

		log.Printf("sending metrics over grpc to %s", config.grpcAddress)
	}

	store := repository.NewMemStorage()
	jobs := make(chan []models.Metrics, config.rateLimit)

	var wg sync.WaitGroup

	for i := int64(0); i < config.rateLimit; i++ {
		wg.Add(1)

		go worker(ctx, jobs, store, config, publicKey, sender, &wg)
	}

	go pollRuntime(ctx, store, time.Duration(config.pollInterval)*time.Second)
	go pollSystem(ctx, store, time.Duration(config.pollInterval)*time.Second)

	report(ctx, store, jobs, time.Duration(config.reportInterval)*time.Second)

	log.Print("shutdown signal received")

	flush(store, jobs)

	close(jobs)
	wg.Wait()

	log.Print("all metrics sent")
}
