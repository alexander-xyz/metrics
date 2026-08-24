package main

import (
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
