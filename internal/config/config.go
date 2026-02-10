package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configDir = "esq"

// Config holds all CLI configuration.
type Config struct {
	CurrentEnv   string                  `json:"current_env"`
	Environments map[string]*Environment `json:"environments"`
}

// Environment represents a single Elasticsearch cluster.
type Environment struct {
	URL string `json:"url"`
}

// configPath returns the path to the config file, respecting XDG_CONFIG_HOME.
func configPath() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, configDir, "config.json")
}

// Load reads the config from disk. Returns a default config if the file doesn't exist.
func Load() (*Config, error) {
	path := configPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				Environments: make(map[string]*Environment),
			}, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if cfg.Environments == nil {
		cfg.Environments = make(map[string]*Environment)
	}
	return &cfg, nil
}

// Save writes the config to disk.
func Save(cfg *Config) error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// GetCurrentEnv returns the active environment or an error if none is set.
func (c *Config) GetCurrentEnv() (*Environment, string, error) {
	if c.CurrentEnv == "" {
		return nil, "", fmt.Errorf("no active environment. Run 'esq config add <name> --url <url>' and 'esq config use <name>'")
	}
	env, ok := c.Environments[c.CurrentEnv]
	if !ok {
		return nil, "", fmt.Errorf("environment %q not found in config", c.CurrentEnv)
	}
	return env, c.CurrentEnv, nil
}

// EnvNames returns sorted environment names.
func (c *Config) EnvNames() []string {
	names := make([]string, 0, len(c.Environments))
	for name := range c.Environments {
		names = append(names, name)
	}
	return names
}
