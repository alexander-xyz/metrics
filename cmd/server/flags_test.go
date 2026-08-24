package main

import (
	"flag"
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
		"crypto_key": "/tmp/key.pem",
		"trusted_subnet": "10.0.0.0/8"
	}`)

	config := Config{configFile: path}
	require.NoError(t, applyConfigFile(&config))

	assert.Equal(t, "localhost:9911", config.serverAddress)
	assert.Equal(t, "/tmp/from-config.json", config.fileStoragePath)
	assert.Equal(t, "postgres://localhost/db", config.databaseDSN)
	assert.Equal(t, "/tmp/key.pem", config.cryptoKey)
	assert.Equal(t, int64(42), config.storeInterval)
	assert.True(t, config.restore)
	assert.Equal(t, "10.0.0.0/8", config.trustedSubnet)
}

func TestTrustedSubnetFromEnvironmentBeatsFile(t *testing.T) {
	path := writeServerConfig(t, `{"trusted_subnet": "10.0.0.0/8"}`)

	t.Setenv("TRUSTED_SUBNET", "192.168.0.0/16")

	config := Config{configFile: path, trustedSubnet: "192.168.0.0/16"}
	require.NoError(t, applyConfigFile(&config))

	assert.Equal(t, "192.168.0.0/16", config.trustedSubnet)
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

// resetFlags возвращает набор флагов в исходное состояние: parseFlags
// работает с глобальным flag.CommandLine.
func resetFlags(t *testing.T, args ...string) {
	t.Helper()

	oldArgs := os.Args
	oldFlags := flag.CommandLine

	t.Cleanup(func() {
		os.Args = oldArgs
		flag.CommandLine = oldFlags
	})

	flag.CommandLine = flag.NewFlagSet(oldArgs[0], flag.ContinueOnError)
	os.Args = append([]string{oldArgs[0]}, args...)
}

func TestParseFlagsDefaults(t *testing.T) {
	resetFlags(t)

	config, err := parseFlags()
	require.NoError(t, err)

	assert.Equal(t, "localhost:8080", config.serverAddress)
	assert.Equal(t, "info", config.logLevel)
	assert.Equal(t, int64(300), config.storeInterval)
	assert.Equal(t, defaultStoragePath, config.fileStoragePath)
	assert.True(t, config.restore)
	assert.Empty(t, config.databaseDSN)
}

func TestParseFlagsReadsCommandLine(t *testing.T) {
	resetFlags(t, "-a", "localhost:9999", "-i", "7", "-f", "/tmp/m.json",
		"-d", "postgres://localhost/db", "-k", "secret", "-t", "10.0.0.0/8",
		"-g", "localhost:3200", "-crypto-key", "/tmp/key.pem", "-r=false")

	config, err := parseFlags()
	require.NoError(t, err)

	assert.Equal(t, "localhost:9999", config.serverAddress)
	assert.Equal(t, int64(7), config.storeInterval)
	assert.Equal(t, "/tmp/m.json", config.fileStoragePath)
	assert.Equal(t, "postgres://localhost/db", config.databaseDSN)
	assert.Equal(t, "secret", config.key)
	assert.Equal(t, "10.0.0.0/8", config.trustedSubnet)
	assert.Equal(t, "localhost:3200", config.grpcAddress)
	assert.Equal(t, "/tmp/key.pem", config.cryptoKey)
	assert.False(t, config.restore)
}

func TestParseFlagsReadsEnvironment(t *testing.T) {
	resetFlags(t)

	t.Setenv("ADDRESS", "localhost:7070")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("STORE_INTERVAL", "15")
	t.Setenv("FILE_STORAGE_PATH", "/tmp/env.json")
	t.Setenv("RESTORE", "false")
	t.Setenv("DATABASE_DSN", "postgres://env/db")
	t.Setenv("KEY", "env-key")
	t.Setenv("TRUSTED_SUBNET", "172.16.0.0/12")
	t.Setenv("GRPC_ADDRESS", "localhost:3300")
	t.Setenv("CRYPTO_KEY", "/tmp/env.pem")

	config, err := parseFlags()
	require.NoError(t, err)

	assert.Equal(t, "localhost:7070", config.serverAddress)
	assert.Equal(t, "debug", config.logLevel)
	assert.Equal(t, int64(15), config.storeInterval)
	assert.Equal(t, "/tmp/env.json", config.fileStoragePath)
	assert.False(t, config.restore)
	assert.Equal(t, "postgres://env/db", config.databaseDSN)
	assert.Equal(t, "env-key", config.key)
	assert.Equal(t, "172.16.0.0/12", config.trustedSubnet)
	assert.Equal(t, "localhost:3300", config.grpcAddress)
	assert.Equal(t, "/tmp/env.pem", config.cryptoKey)
}

func TestParseFlagsRejectsBrokenEnvironment(t *testing.T) {
	testCases := []struct {
		name  string
		key   string
		value string
	}{
		{name: "интервал", key: "STORE_INTERVAL", value: "полчаса"},
		{name: "восстановление", key: "RESTORE", value: "может быть"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetFlags(t)
			t.Setenv(tc.key, tc.value)

			_, err := parseFlags()
			assert.Error(t, err)
		})
	}
}
