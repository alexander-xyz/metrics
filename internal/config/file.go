// Package config читает конфигурацию приложения из JSON-файла.
// Значения из файла имеют меньший приоритет, чем флаги командной строки
// и переменные окружения.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Duration разбирает интервалы, записанные в файле строкой: "1s", "300ms".
type Duration struct {
	time.Duration
}

// UnmarshalJSON разбирает значение интервала из строки формата time.Duration.
func (d *Duration) UnmarshalJSON(data []byte) error {
	var value string

	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("parse duration: %w", err)
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fmt.Errorf("parse duration %q: %w", value, err)
	}

	d.Duration = parsed

	return nil
}

// Server описывает конфигурацию сервера метрик в JSON-файле.
type Server struct {
	Address       string   `json:"address"`
	StoreFile     string   `json:"store_file"`
	DatabaseDSN   string   `json:"database_dsn"`
	CryptoKey     string   `json:"crypto_key"`
	StoreInterval Duration `json:"store_interval"`
	Restore       bool     `json:"restore"`
}

// Agent описывает конфигурацию агента в JSON-файле.
type Agent struct {
	Address        string   `json:"address"`
	CryptoKey      string   `json:"crypto_key"`
	ReportInterval Duration `json:"report_interval"`
	PollInterval   Duration `json:"poll_interval"`
}

// LoadServer читает конфигурацию сервера из файла.
func LoadServer(path string) (*Server, error) {
	var config Server

	if err := load(path, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadAgent читает конфигурацию агента из файла.
func LoadAgent(path string) (*Agent, error) {
	var config Agent

	if err := load(path, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// load читает файл и разбирает его как JSON.
func load(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file: %w", err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("parse config file %s: %w", path, err)
	}

	return nil
}

// Path возвращает путь к файлу конфигурации: значение флага или
// переменной окружения CONFIG, которая имеет приоритет.
func Path(flagValue string) string {
	if env := os.Getenv("CONFIG"); env != "" {
		return env
	}

	return flagValue
}
