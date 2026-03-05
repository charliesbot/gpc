package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the CLI configuration values persisted to disk.
type Config struct {
	DefaultPackage string `json:"default_package,omitempty"`
	DefaultKeyFile string `json:"default_key_file,omitempty"`
	DefaultOutput  string `json:"default_output,omitempty"`
}

// DefaultConfigPath returns the default configuration file path
// at ~/.config/gpc/config.json.
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fall back to HOME env or current directory as a last resort.
		home = os.Getenv("HOME")
		if home == "" {
			home = "."
		}
	}
	return filepath.Join(home, ".config", "gpc", "config.json")
}

// Load reads the configuration from the default path. If the file does not
// exist, it returns a zero-value Config and a nil error.
func Load() (*Config, error) {
	return LoadFrom(DefaultConfigPath())
}

// LoadFrom reads the configuration from the given path. If the file does not
// exist, it returns a zero-value Config and a nil error.
func LoadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}
	return &cfg, nil
}

// Save writes the configuration to the default path, creating parent
// directories as needed.
func (c *Config) Save() error {
	return c.SaveTo(DefaultConfigPath())
}

// SaveTo writes the configuration to the given path, creating parent
// directories as needed.
func (c *Config) SaveTo(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}
	return nil
}
