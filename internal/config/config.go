package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds application configuration loaded from file and environment.
type Config struct {
	// GitHub token for API authentication.
	GitHubToken string `yaml:"github_token"`

	// Default output format for analyze command.
	DefaultFormat string `yaml:"default_format"`

	// Default prompt mode.
	DefaultMode string `yaml:"default_mode"`

	// Default prompt language.
	DefaultLang string `yaml:"default_lang"`
}

// DefaultConfigPath returns the default config file path.
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "prism", "config.yaml")
}

// Load reads configuration from the given path.
// If the file does not exist, it returns a zero Config (not an error).
// Environment variables override file values.
func Load(path string) (Config, error) {
	var cfg Config

	if path == "" {
		path = os.Getenv("PRISM_CONFIG")
	}
	if path == "" {
		path = DefaultConfigPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return applyEnv(cfg), nil
		}
		return Config{}, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	return applyEnv(cfg), nil
}

// applyEnv overrides config values with environment variables.
func applyEnv(cfg Config) Config {
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		cfg.GitHubToken = token
	}
	return cfg
}
