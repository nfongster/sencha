package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestLoad_PartialConfig_MissingBaseURL(t *testing.T) {
	f, err := os.CreateTemp("", "sencha-config-*.json")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	f.WriteString(`{"llm": {"model": "custom-model"}}`)
	f.Close()

	_, err = Load(f.Name())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "base_url")
	assert.Contains(t, err.Error(), "required")
}

func TestLoad_EmptyFields_Errors(t *testing.T) {
	f, err := os.CreateTemp("", "sencha-config-*.json")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	f.WriteString(`{"llm": {}}`)
	f.Close()

	_, err = Load(f.Name())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "base_url")
}

func TestLoad_EmptyBaseURL_Errors(t *testing.T) {
	f, err := os.CreateTemp("", "sencha-config-*.json")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	f.WriteString(`{"llm": {"base_url": "", "model": "custom"}}`)
	f.Close()

	_, err = Load(f.Name())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "base_url")
}
