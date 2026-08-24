package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.json")
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	return path
}

func TestLoadServer(t *testing.T) {
	path := writeConfig(t, `{
		"address": "localhost:9911",
		"restore": true,
		"store_interval": "42s",
		"store_file": "/tmp/metrics.json",
		"database_dsn": "postgres://localhost/db",
		"crypto_key": "/tmp/key.pem"
	}`)

	cfg, err := LoadServer(path)
	require.NoError(t, err)

	assert.Equal(t, "localhost:9911", cfg.Address)
	assert.True(t, cfg.Restore)
	assert.Equal(t, 42*time.Second, cfg.StoreInterval.Duration)
	assert.Equal(t, "/tmp/metrics.json", cfg.StoreFile)
	assert.Equal(t, "postgres://localhost/db", cfg.DatabaseDSN)
	assert.Equal(t, "/tmp/key.pem", cfg.CryptoKey)
}

func TestLoadAgent(t *testing.T) {
	path := writeConfig(t, `{
		"address": "localhost:8080",
		"report_interval": "10s",
		"poll_interval": "1500ms",
		"crypto_key": "/tmp/public.pem"
	}`)

	cfg, err := LoadAgent(path)
	require.NoError(t, err)

	assert.Equal(t, "localhost:8080", cfg.Address)
	assert.Equal(t, 10*time.Second, cfg.ReportInterval.Duration)
	assert.Equal(t, 1500*time.Millisecond, cfg.PollInterval.Duration)
	assert.Equal(t, "/tmp/public.pem", cfg.CryptoKey)
}

func TestLoadRejectsBrokenFile(t *testing.T) {
	path := writeConfig(t, "не json")

	_, err := LoadServer(path)
	assert.Error(t, err)

	_, err = LoadAgent(path)
	assert.Error(t, err)
}

func TestLoadRejectsBrokenDuration(t *testing.T) {
	path := writeConfig(t, `{"store_interval": "полчаса"}`)

	_, err := LoadServer(path)
	assert.Error(t, err)
}

func TestLoadRejectsMissingFile(t *testing.T) {
	_, err := LoadServer(filepath.Join(t.TempDir(), "missing.json"))
	assert.Error(t, err)
}

func TestPathPrefersEnvironment(t *testing.T) {
	t.Setenv("CONFIG", "/from/env.json")
	assert.Equal(t, "/from/env.json", Path("/from/flag.json"))
}

func TestPathFallsBackToFlag(t *testing.T) {
	t.Setenv("CONFIG", "")
	assert.Equal(t, "/from/flag.json", Path("/from/flag.json"))
	assert.Empty(t, Path(""))
}
