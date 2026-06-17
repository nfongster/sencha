package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type LLMConfig struct {
	BaseURL string `json:"base_url"`
	Model   string `json:"model"`
	APIKey  string `json:"api_key"`
}

type Config struct {
	LLM LLMConfig `json:"llm"`
}

func Defaults() *Config {
	return &Config{
		LLM: LLMConfig{
			BaseURL: "http://localhost:11434/v1",
			Model:   "qwen3:8b",
			APIKey:  "",
		},
	}
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	cfg := Defaults()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if cfg.LLM.BaseURL == "" {
		cfg.LLM.BaseURL = "http://localhost:11434/v1"
	}
	if cfg.LLM.Model == "" {
		cfg.LLM.Model = "qwen3:8b"
	}

	return cfg, nil
}
