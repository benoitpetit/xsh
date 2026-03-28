package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/utils"
)

// TestDefaultConfig tests default configuration
func TestDefaultConfig(t *testing.T) {
	cfg := core.DefaultConfig()

	if cfg.DefaultCount != 20 {
		t.Errorf("DefaultCount = %d, want 20", cfg.DefaultCount)
	}

	if cfg.Display.Theme != "default" {
		t.Errorf("Display.Theme = %s, want default", cfg.Display.Theme)
	}

	if !cfg.Display.ShowEngagement {
		t.Error("Display.ShowEngagement should be true")
	}

	if cfg.Request.Timeout != 30 {
		t.Errorf("Request.Timeout = %d, want 30", cfg.Request.Timeout)
	}

	if cfg.Request.MaxRetries != 3 {
		t.Errorf("Request.MaxRetries = %d, want 3", cfg.Request.MaxRetries)
	}

	// Verify filter weights
	if cfg.Filter.LikesWeight != 1.0 {
		t.Errorf("Filter.LikesWeight = %f, want 1.0", cfg.Filter.LikesWeight)
	}
}

// TestConfigLoadSave tests configuration loading and saving
func TestConfigLoadSave(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Use environment variable to change path
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create the .config/xsh directory
	configDir := filepath.Join(tmpDir, ".config", "xsh")
	os.MkdirAll(configDir, 0755)

	t.Run("load default when file doesn't exist", func(t *testing.T) {
		cfg, err := core.LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}

		if cfg.DefaultCount != 20 {
			t.Errorf("DefaultCount = %d, want 20", cfg.DefaultCount)
		}
	})

	t.Run("save and load custom config", func(t *testing.T) {
		cfg := &core.Config{
			DefaultCount: 50,
			Display: core.DisplayConfig{
				Theme:          "dark",
				ShowEngagement: false,
				MaxWidth:       120,
			},
			Request: core.RequestConfig{
				Delay:      2.0,
				Timeout:    60,
				MaxRetries: 5,
			},
		}

		err := cfg.Save()
		if err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		// Reload configuration
		loaded, err := core.LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}

		if loaded.DefaultCount != 50 {
			t.Errorf("DefaultCount = %d, want 50", loaded.DefaultCount)
		}

		if loaded.Display.Theme != "dark" {
			t.Errorf("Display.Theme = %s, want dark", loaded.Display.Theme)
		}

		if loaded.Display.ShowEngagement {
			t.Error("Display.ShowEngagement should be false")
		}

		if loaded.Request.Timeout != 60 {
			t.Errorf("Request.Timeout = %d, want 60", loaded.Request.Timeout)
		}
	})
}

// TestFilterConfig tests filter configuration
func TestFilterConfig(t *testing.T) {
	cfg := core.DefaultConfig()

	// Verify default weights
	expectedWeights := map[string]float64{
		"Likes":     1.0,
		"Retweets":  1.5,
		"Replies":   0.5,
		"Bookmarks": 2.0,
		"ViewsLog":  0.3,
	}

	if cfg.Filter.LikesWeight != expectedWeights["Likes"] {
		t.Errorf("LikesWeight = %f, want %f", cfg.Filter.LikesWeight, expectedWeights["Likes"])
	}

	if cfg.Filter.RetweetsWeight != expectedWeights["Retweets"] {
		t.Errorf("RetweetsWeight = %f, want %f", cfg.Filter.RetweetsWeight, expectedWeights["Retweets"])
	}

	if cfg.Filter.RepliesWeight != expectedWeights["Replies"] {
		t.Errorf("RepliesWeight = %f, want %f", cfg.Filter.RepliesWeight, expectedWeights["Replies"])
	}

	if cfg.Filter.BookmarksWeight != expectedWeights["Bookmarks"] {
		t.Errorf("BookmarksWeight = %f, want %f", cfg.Filter.BookmarksWeight, expectedWeights["Bookmarks"])
	}

	if cfg.Filter.ViewsLogWeight != expectedWeights["ViewsLog"] {
		t.Errorf("ViewsLogWeight = %f, want %f", cfg.Filter.ViewsLogWeight, expectedWeights["ViewsLog"])
	}
}

// TestDefaultFilterConfig tests the DefaultFilterConfig function
func TestDefaultFilterConfig(t *testing.T) {
	cfg := utils.DefaultFilterConfig()

	if cfg.LikesWeight != 1.0 {
		t.Errorf("LikesWeight = %f, want 1.0", cfg.LikesWeight)
	}
}

// TestGetConfigDir tests configuration directory retrieval
func TestGetConfigDir(t *testing.T) {
	dir, err := core.GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir() error = %v", err)
	}

	if dir == "" {
		t.Error("GetConfigDir() returned empty string")
	}

	// Verify directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("Config directory %s does not exist", dir)
	}
}

// TestGetConfigPath tests configuration path retrieval
func TestGetConfigPath(t *testing.T) {
	path, err := core.GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() error = %v", err)
	}

	if path == "" {
		t.Error("GetConfigPath() returned empty string")
	}

	// Verify path ends with config.toml
	if filepath.Base(path) != "config.toml" {
		t.Errorf("Config path %s should end with config.toml", path)
	}
}

// TestConfigSetValues tests configuration value modification
func TestConfigSetValues(t *testing.T) {
	cfg := core.DefaultConfig()

	// Modify values
	cfg.DefaultCount = 100
	cfg.Display.Theme = "light"
	cfg.Display.ShowEngagement = false
	cfg.Request.Delay = 3.5
	cfg.Filter.LikesWeight = 2.0

	if cfg.DefaultCount != 100 {
		t.Errorf("DefaultCount = %d, want 100", cfg.DefaultCount)
	}

	if cfg.Display.Theme != "light" {
		t.Errorf("Display.Theme = %s, want light", cfg.Display.Theme)
	}

	if cfg.Display.ShowEngagement {
		t.Error("Display.ShowEngagement should be false")
	}

	if cfg.Request.Delay != 3.5 {
		t.Errorf("Request.Delay = %f, want 3.5", cfg.Request.Delay)
	}

	if cfg.Filter.LikesWeight != 2.0 {
		t.Errorf("Filter.LikesWeight = %f, want 2.0", cfg.Filter.LikesWeight)
	}
}
