package config

import (
	"encoding/json"
	"fmt"
	"log"
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

func maskKey(key string) string {
	if key == "" {
		return "(empty)"
	}
	if len(key) <= 4 {
		return "****"
	}
	return key[:4] + "****"
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if cfg.LLM.BaseURL == "" {
		return nil, fmt.Errorf("config file: llm.base_url is required")
	}
	if cfg.LLM.Model == "" {
		return nil, fmt.Errorf("config file: llm.model is required")
	}

	log.Printf("[config] loaded from %q — base_url=%q model=%q api_key=%q", path, cfg.LLM.BaseURL, cfg.LLM.Model, maskKey(cfg.LLM.APIKey))

	return &cfg, nil
}
