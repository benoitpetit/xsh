package core

import (
	"bytes"
	"fmt"
	"os"

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

// NetworkConfig contains network settings
type NetworkConfig struct {
	Proxy string `toml:"proxy,omitempty"`
}

// RequestConfig contains request settings
type RequestConfig struct {
	Delay            float64 `toml:"delay"`
	Timeout          int     `toml:"timeout"`
	MaxRetries       int     `toml:"max_retries"`
	MaxResponseBytes int64   `toml:"max_response_bytes"`
}

// Config is the root configuration
type Config struct {
	DefaultCount   int                `toml:"default_count"`
	DefaultAccount string             `toml:"default_account,omitempty"`
	Display        DisplayConfig      `toml:"display"`
	Request        RequestConfig      `toml:"request"`
	Network        NetworkConfig      `toml:"network"`
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
			Delay:            DefaultDelaySec,
			Timeout:          30,
			MaxRetries:       3,
			MaxResponseBytes: defaultMaxResponseBytes,
		},
		Network: NetworkConfig{
			Proxy: "",
		},
		Filter: utils.DefaultFilterConfig(),
	}
}

// GetConfigDir returns the config directory, creating it if needed
func GetConfigDir() (string, error) {
	paths, err := GetPaths()
	if err != nil {
		return "", err
	}
	return paths.ConfigDir, nil
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	paths, err := GetPaths()
	if err != nil {
		return "", err
	}
	return paths.ConfigFile, nil
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
		decodeErr := fmt.Errorf("failed to decode config: %w", err)
		if backupErr := preserveCorruptFile(configPath); backupErr != nil {
			return nil, fmt.Errorf("%v; preserving corrupt config failed: %w", decodeErr, backupErr)
		}
		return nil, fmt.Errorf("%v; corrupt source preserved at %s.bak", decodeErr, configPath)
	}

	return cfg, nil
}

// Save saves config to TOML file
func (c *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	var encoded bytes.Buffer
	if err := toml.NewEncoder(&encoded).Encode(c); err != nil {
		return err
	}
	return WriteFileAtomic(configPath, encoded.Bytes(), 0600)
}
