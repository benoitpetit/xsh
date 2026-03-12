package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/benoitpetit/xsh/utils"
)

// DisplayConfig contains display settings
type DisplayConfig struct {
	Theme          string `toml:"theme"`
	ShowEngagement bool   `toml:"show_engagement"`
	ShowTimestamps bool   `toml:"show_timestamps"`
	MaxWidth       int    `toml:"max_width"`
}

// RequestConfig contains request settings
type RequestConfig struct {
	Delay      float64 `toml:"delay"`
	Proxy      string  `toml:"proxy,omitempty"`
	Timeout    int     `toml:"timeout"`
	MaxRetries int     `toml:"max_retries"`
}

// Config is the root configuration
type Config struct {
	DefaultCount   int              `toml:"default_count"`
	DefaultAccount string           `toml:"default_account,omitempty"`
	Display        DisplayConfig    `toml:"display"`
	Request        RequestConfig    `toml:"request"`
	Filter         utils.FilterConfig `toml:"filter"`
}

// DefaultConfig returns a Config with default values
func DefaultConfig() *Config {
	return &Config{
		DefaultCount: DefaultCount,
		Display: DisplayConfig{
			Theme:          "default",
			ShowEngagement: true,
			ShowTimestamps: true,
			MaxWidth:       100,
		},
		Request: RequestConfig{
			Delay:      DefaultDelaySec,
			Timeout:    30,
			MaxRetries: 3,
		},
		Filter: *utils.DefaultFilterConfig(),
	}
}

// GetConfigDir returns the config directory, creating it if needed
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(home, ".config", ConfigDirName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}
	return configDir, nil
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigFileName), nil
}

// LoadConfig loads config from TOML file, with defaults for missing values
func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// Start with defaults
	cfg := DefaultConfig()

	// If file doesn't exist, return defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return cfg, nil
	}

	// Load from file
	if _, err := toml.DecodeFile(configPath, cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	return cfg, nil
}

// Save saves config to TOML file
func (c *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	return encoder.Encode(c)
}
