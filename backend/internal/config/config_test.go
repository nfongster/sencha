package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaults_Values(t *testing.T) {
	cfg := Defaults()
	assert.Equal(t, "http://localhost:11434/v1", cfg.LLM.BaseURL)
	assert.Equal(t, "qwen3:8b", cfg.LLM.Model)
	assert.Equal(t, "", cfg.LLM.APIKey)
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/tmp/nonexistent-sencha-config-xxxxxxxx.json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading config file")
}

func TestLoad_InvalidJSON(t *testing.T) {
	f, err := os.CreateTemp("", "sencha-config-*.json")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	f.WriteString("{bad json}")
	f.Close()

	_, err = Load(f.Name())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing config file")
}

func TestLoad_FullConfig(t *testing.T) {
	f, err := os.CreateTemp("", "sencha-config-*.json")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	f.WriteString(`{"llm": {"base_url": "http://example.com:11434/v1", "model": "test-model", "api_key": "sk-abc"}}`)
	f.Close()

	cfg, err := Load(f.Name())
	require.NoError(t, err)
	assert.Equal(t, "http://example.com:11434/v1", cfg.LLM.BaseURL)
	assert.Equal(t, "test-model", cfg.LLM.Model)
	assert.Equal(t, "sk-abc", cfg.LLM.APIKey)
}

func TestLoad_PartialConfig_FillsDefaults(t *testing.T) {
	f, err := os.CreateTemp("", "sencha-config-*.json")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	f.WriteString(`{"llm": {"model": "custom-model"}}`)
	f.Close()

	cfg, err := Load(f.Name())
	require.NoError(t, err)
	assert.Equal(t, "custom-model", cfg.LLM.Model)
	assert.Equal(t, "http://localhost:11434/v1", cfg.LLM.BaseURL)
}

func TestLoad_EmptyFields_FallsBack(t *testing.T) {
	f, err := os.CreateTemp("", "sencha-config-*.json")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	f.WriteString(`{"llm": {}}`)
	f.Close()

	cfg, err := Load(f.Name())
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:11434/v1", cfg.LLM.BaseURL)
	assert.Equal(t, "qwen3:8b", cfg.LLM.Model)
	assert.Equal(t, "", cfg.LLM.APIKey)
}

func TestLoad_EmptyBaseURL_FallsBack(t *testing.T) {
	f, err := os.CreateTemp("", "sencha-config-*.json")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	f.WriteString(`{"llm": {"base_url": "", "model": "custom"}}`)
	f.Close()

	cfg, err := Load(f.Name())
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:11434/v1", cfg.LLM.BaseURL)
	assert.Equal(t, "custom", cfg.LLM.Model)
}
