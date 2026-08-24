package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeServerConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "server.json")
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	return path
}

func TestApplyConfigFile(t *testing.T) {
	path := writeServerConfig(t, `{
		"address": "localhost:9911",
		"restore": true,
		"store_interval": "42s",
		"store_file": "/tmp/from-config.json",
		"database_dsn": "postgres://localhost/db",
		"crypto_key": "/tmp/key.pem"
	}`)

	config := Config{configFile: path}
	require.NoError(t, applyConfigFile(&config))

	assert.Equal(t, "localhost:9911", config.serverAddress)
	assert.Equal(t, "/tmp/from-config.json", config.fileStoragePath)
	assert.Equal(t, "postgres://localhost/db", config.databaseDSN)
	assert.Equal(t, "/tmp/key.pem", config.cryptoKey)
	assert.Equal(t, int64(42), config.storeInterval)
	assert.True(t, config.restore)
}

func TestEnvironmentBeatsConfigFile(t *testing.T) {
	path := writeServerConfig(t, `{"address": "localhost:9911", "database_dsn": "from-file"}`)

	t.Setenv("ADDRESS", "localhost:7777")
	t.Setenv("DATABASE_DSN", "from-env")

	config := Config{configFile: path, serverAddress: "localhost:7777", databaseDSN: "from-env"}
	require.NoError(t, applyConfigFile(&config))

	assert.Equal(t, "localhost:7777", config.serverAddress, "переменная окружения важнее файла")
	assert.Equal(t, "from-env", config.databaseDSN)
}

func TestConfigFileIgnoredWithoutPath(t *testing.T) {
	t.Setenv("CONFIG", "")

	config := Config{serverAddress: "localhost:8080"}
	require.NoError(t, applyConfigFile(&config))

	assert.Equal(t, "localhost:8080", config.serverAddress)
}

func TestConfigPathFromEnvironment(t *testing.T) {
	path := writeServerConfig(t, `{"address": "localhost:9911"}`)
	t.Setenv("CONFIG", path)

	config := Config{}
	require.NoError(t, applyConfigFile(&config))

	assert.Equal(t, "localhost:9911", config.serverAddress, "путь к файлу берётся из CONFIG")
}

func TestApplyConfigFileRejectsBrokenFile(t *testing.T) {
	path := writeServerConfig(t, "не json")

	config := Config{configFile: path}
	assert.Error(t, applyConfigFile(&config))
}
