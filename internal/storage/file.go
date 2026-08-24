package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"

	models "github.com/alexander-xyz/metrics/internal/model"
	"github.com/alexander-xyz/metrics/internal/repository"
)

type Storage interface {
	repository.Updater
	repository.Getter
}

func Save(store repository.Getter, path string) error {
	metrics := make([]models.Metrics, 0, len(store.GetGauges())+len(store.GetCounters()))

	for name, value := range store.GetGauges() {
		v := float64(value)
		metrics = append(metrics, models.Metrics{ID: name, MType: models.Gauge, Value: &v})
	}

	for name, value := range store.GetCounters() {
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

func Load(store repository.Updater, path string) error {
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
				store.UpdateGauge(metric.ID, repository.Gauge(*metric.Value))
			}
		case models.Counter:
			if metric.Delta != nil {
				store.SetCounter(metric.ID, repository.Counter(*metric.Delta))
			}
		}
	}

	return nil
}

type SyncStorage struct {
	Storage
	path  string
	onErr func(error)
}

func NewSyncStorage(store Storage, path string, onErr func(error)) *SyncStorage {
	return &SyncStorage{Storage: store, path: path, onErr: onErr}
}

func (s *SyncStorage) save() {
	if err := Save(s.Storage, s.path); err != nil && s.onErr != nil {
		s.onErr(err)
	}
}

func (s *SyncStorage) UpdateGauge(name string, value repository.Gauge) {
	s.Storage.UpdateGauge(name, value)
	s.save()
}

func (s *SyncStorage) UpdateCounter(name string, value repository.Counter) {
	s.Storage.UpdateCounter(name, value)
	s.save()
}

func (s *SyncStorage) SetCounter(name string, value repository.Counter) {
	s.Storage.SetCounter(name, value)
	s.save()
}
