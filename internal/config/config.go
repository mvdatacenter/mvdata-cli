package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the resolved API URL and token.
type Config struct {
	APIURL   string `json:"api_url"`
	APIToken string `json:"api_token"`
}

// Resolve builds a Config by merging values from flags, environment variables,
// and the config file (~/.mvdata/config.json). Precedence: flags > env > file.
func Resolve(flagURL, flagToken string) (*Config, error) {
	cfg := &Config{}

	// Load config file first (lowest precedence).
	if err := cfg.loadFile(); err != nil {
		return nil, err
	}

	// Environment variables override file.
	if v := os.Getenv("MVDATA_API_URL"); v != "" {
		cfg.APIURL = v
	}
	if v := os.Getenv("MVDATA_API_TOKEN"); v != "" {
		cfg.APIToken = v
	}

	// Flags override everything.
	if flagURL != "" {
		cfg.APIURL = flagURL
	}
	if flagToken != "" {
		cfg.APIToken = flagToken
	}

	if cfg.APIURL == "" {
		return nil, fmt.Errorf("API URL is required (set --api-url, MVDATA_API_URL, or api_url in ~/.mvdata/config.json)")
	}
	if cfg.APIToken == "" {
		return nil, fmt.Errorf("API token is required (set --api-token, MVDATA_API_TOKEN, or api_token in ~/.mvdata/config.json)")
	}

	return cfg, nil
}

func (c *Config) loadFile() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil // Not fatal — other sources may provide values.
	}

	path := filepath.Join(home, ".mvdata", "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading config file: %w", err)
	}

	if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf("parsing config file %s: %w", path, err)
	}
	return nil
}
