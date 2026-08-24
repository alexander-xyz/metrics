package main

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeAgentConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "agent.json")
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	return path
}

func TestApplyConfigFile(t *testing.T) {
	path := writeAgentConfig(t, `{
		"address": "localhost:9911",
		"report_interval": "30s",
		"poll_interval": "5s",
		"crypto_key": "/tmp/public.pem"
	}`)

	config := Config{configFile: path}
	runAddr := "localhost:8080"

	require.NoError(t, applyConfigFile(&config, &runAddr))

	assert.Equal(t, "localhost:9911", runAddr)
	assert.Equal(t, int64(30), config.reportInterval)
	assert.Equal(t, int64(5), config.pollInterval)
	assert.Equal(t, "/tmp/public.pem", config.cryptoKey)
}

func TestEnvironmentBeatsConfigFile(t *testing.T) {
	path := writeAgentConfig(t, `{"address": "localhost:9911", "poll_interval": "5s"}`)

	t.Setenv("ADDRESS", "localhost:7777")
	t.Setenv("POLL_INTERVAL", "9")

	config := Config{configFile: path, pollInterval: 9}
	runAddr := "localhost:7777"

	require.NoError(t, applyConfigFile(&config, &runAddr))

	assert.Equal(t, "localhost:7777", runAddr, "переменная окружения важнее файла")
	assert.Equal(t, int64(9), config.pollInterval)
}

func TestConfigFileIgnoredWithoutPath(t *testing.T) {
	t.Setenv("CONFIG", "")

	config := Config{pollInterval: 2}
	runAddr := "localhost:8080"

	require.NoError(t, applyConfigFile(&config, &runAddr))

	assert.Equal(t, int64(2), config.pollInterval)
}

func TestApplyConfigFileRejectsBrokenFile(t *testing.T) {
	path := writeAgentConfig(t, "{")

	config := Config{configFile: path}
	runAddr := ""

	assert.Error(t, applyConfigFile(&config, &runAddr))
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

	assert.Equal(t, "http://localhost:8080", config.serverAddress)
	assert.Equal(t, int64(10), config.reportInterval)
	assert.Equal(t, int64(2), config.pollInterval)
	assert.Equal(t, int64(1), config.rateLimit)
}

func TestParseFlagsReadsCommandLine(t *testing.T) {
	resetFlags(t, "-a", "localhost:9999", "-r", "30", "-p", "5", "-l", "4",
		"-k", "secret", "-g", "localhost:3200", "-crypto-key", "/tmp/pub.pem")

	config, err := parseFlags()
	require.NoError(t, err)

	assert.Equal(t, "http://localhost:9999", config.serverAddress)
	assert.Equal(t, int64(30), config.reportInterval)
	assert.Equal(t, int64(5), config.pollInterval)
	assert.Equal(t, int64(4), config.rateLimit)
	assert.Equal(t, "secret", config.key)
	assert.Equal(t, "localhost:3200", config.grpcAddress)
	assert.Equal(t, "/tmp/pub.pem", config.cryptoKey)
}

func TestParseFlagsKeepsSchemeFromFlag(t *testing.T) {
	resetFlags(t, "-a", "https://metrics.example.com")

	config, err := parseFlags()
	require.NoError(t, err)

	assert.Equal(t, "https://metrics.example.com", config.serverAddress)
}

func TestParseFlagsReadsEnvironment(t *testing.T) {
	resetFlags(t)

	t.Setenv("ADDRESS", "localhost:7070")
	t.Setenv("REPORT_INTERVAL", "25")
	t.Setenv("POLL_INTERVAL", "3")
	t.Setenv("RATE_LIMIT", "8")
	t.Setenv("KEY", "env-key")
	t.Setenv("GRPC_ADDRESS", "localhost:3300")
	t.Setenv("CRYPTO_KEY", "/tmp/env.pem")

	config, err := parseFlags()
	require.NoError(t, err)

	assert.Equal(t, "http://localhost:7070", config.serverAddress)
	assert.Equal(t, int64(25), config.reportInterval)
	assert.Equal(t, int64(3), config.pollInterval)
	assert.Equal(t, int64(8), config.rateLimit)
	assert.Equal(t, "env-key", config.key)
	assert.Equal(t, "localhost:3300", config.grpcAddress)
	assert.Equal(t, "/tmp/env.pem", config.cryptoKey)
}

func TestParseFlagsRejectsBrokenEnvironment(t *testing.T) {
	testCases := []struct {
		name  string
		key   string
		value string
	}{
		{name: "интервал отчёта", key: "REPORT_INTERVAL", value: "часто"},
		{name: "интервал опроса", key: "POLL_INTERVAL", value: "редко"},
		{name: "лимит", key: "RATE_LIMIT", value: "много"},
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
