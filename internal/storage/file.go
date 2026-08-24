// Package storage сохраняет метрики в файл и применяет миграции базы данных.
package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"

	models "github.com/alexander-xyz/metrics/internal/model"
	"github.com/alexander-xyz/metrics/internal/repository"
)

// Storage объединяет операции хранилища, необходимые для синхронной записи в файл.
type Storage interface {
	repository.Updater
	repository.BatchUpdater
	repository.Getter
}

// Save записывает все метрики хранилища в файл по указанному пути.
func Save(ctx context.Context, store repository.Getter, path string) error {
	gauges, err := store.GetGauges(ctx)
	if err != nil {
		return fmt.Errorf("read gauges: %w", err)
	}

	counters, err := store.GetCounters(ctx)
	if err != nil {
		return fmt.Errorf("read counters: %w", err)
	}

	metrics := make([]models.Metrics, 0, len(gauges)+len(counters))

	for name, value := range gauges {
		v := float64(value)
		metrics = append(metrics, models.Metrics{ID: name, MType: models.Gauge, Value: &v})
	}

	for name, value := range counters {
		d := int64(value)
		metrics = append(metrics, models.Metrics{ID: name, MType: models.Counter, Delta: &d})
	}

	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal metrics: %w", err)
	}

	if err := os.WriteFile(path, data, 0666); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	return nil
}

// Load восстанавливает метрики из файла по указанному пути.
// Отсутствие файла не считается ошибкой.
func Load(ctx context.Context, store repository.Updater, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("read %s: %w", path, err)
	}

	var metrics []models.Metrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return fmt.Errorf("unmarshal %s: %w", path, err)
	}

	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value != nil {
				if err := store.UpdateGauge(ctx, metric.ID, repository.Gauge(*metric.Value)); err != nil {
					return fmt.Errorf("restore gauge %s: %w", metric.ID, err)
				}
			}
		case models.Counter:
			if metric.Delta != nil {
				if err := store.SetCounter(ctx, metric.ID, repository.Counter(*metric.Delta)); err != nil {
					return fmt.Errorf("restore counter %s: %w", metric.ID, err)
				}
			}
		}
	}

	return nil
}

// SyncStorage — обёртка над хранилищем, сохраняющая метрики в файл
// сразу после каждого изменения.
type SyncStorage struct {
	Storage
	path  string
	onErr func(error)
}

// NewSyncStorage создаёт хранилище с синхронной записью в файл.
// Ошибки записи передаются в onErr.
func NewSyncStorage(store Storage, path string, onErr func(error)) *SyncStorage {
	return &SyncStorage{Storage: store, path: path, onErr: onErr}
}

func (s *SyncStorage) save(ctx context.Context) {
	if err := Save(ctx, s.Storage, s.path); err != nil && s.onErr != nil {
		s.onErr(err)
	}
}

func (s *SyncStorage) UpdateGauge(ctx context.Context, name string, value repository.Gauge) error {
	if err := s.Storage.UpdateGauge(ctx, name, value); err != nil {
		return err
	}

	s.save(ctx)

	return nil
}

func (s *SyncStorage) UpdateCounter(ctx context.Context, name string, value repository.Counter) error {
	if err := s.Storage.UpdateCounter(ctx, name, value); err != nil {
		return err
	}

	s.save(ctx)

	return nil
}

func (s *SyncStorage) SetCounter(ctx context.Context, name string, value repository.Counter) error {
	if err := s.Storage.SetCounter(ctx, name, value); err != nil {
		return err
	}

	s.save(ctx)

	return nil
}

func (s *SyncStorage) UpdateBatch(ctx context.Context, metrics []models.Metrics) error {
	if err := s.Storage.UpdateBatch(ctx, metrics); err != nil {
		return err
	}

	s.save(ctx)

	return nil
}
