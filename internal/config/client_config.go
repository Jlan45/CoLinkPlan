package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ClientConfig struct {
	ClientToken string     `yaml:"client_token"`
	ServerURL   string     `yaml:"server_url"`
	MaxParallel int        `yaml:"max_parallel"`
	Providers   []Provider `yaml:"providers"`
}

type Provider struct {
	Type    string  `yaml:"type"`     // "openai" or "claude"
	APIKey  string  `yaml:"api_key"`  // Real API key
	BaseURL string  `yaml:"base_url"` // Optional base URL overrider
	Models  []Model `yaml:"models"`
}

type Model struct {
	Local         string `yaml:"local"`
	ServerMapping string `yaml:"server_mapping"`
}

func LoadClientConfig(path string) (*ClientConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg ClientConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}

	if cfg.MaxParallel <= 0 {
		cfg.MaxParallel = 1 // default
	}

	return &cfg, nil
}
