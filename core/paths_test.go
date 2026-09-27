package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathsForBaseUsesPortableLayout(t *testing.T) {
	baseCases := []struct {
		name string
		base string
	}{
		{name: "linux xdg", base: filepath.Join("/tmp", "linux-config")},
		{name: "macos library", base: filepath.Join("/tmp", "macos-library", "Application Support")},
		{name: "windows appdata", base: filepath.Join("C:", "Users", "tester", "AppData", "Roaming")},
	}

	for _, tc := range baseCases {
		t.Run(tc.name, func(t *testing.T) {
			paths := pathsForBase(tc.base)
			wantDir := filepath.Join(tc.base, ConfigDirName)
			if paths.ConfigDir != wantDir {
				t.Fatalf("ConfigDir = %q, want %q", paths.ConfigDir, wantDir)
			}
			if paths.ConfigFile != filepath.Join(wantDir, ConfigFileName) {
				t.Fatalf("ConfigFile = %q", paths.ConfigFile)
			}
			if paths.AuthFile != filepath.Join(wantDir, AuthFileName) {
				t.Fatalf("AuthFile = %q", paths.AuthFile)
			}
			if paths.EndpointCache != filepath.Join(wantDir, "graphql_ops.json") {
				t.Fatalf("EndpointCache = %q", paths.EndpointCache)
			}
			if paths.TransactionCache != filepath.Join(wantDir, "transaction_cache.json") {
				t.Fatalf("TransactionCache = %q", paths.TransactionCache)
			}
			if paths.StartupMarker != filepath.Join(wantDir, ".last_endpoint_check") {
				t.Fatalf("StartupMarker = %q", paths.StartupMarker)
			}
		})
	}
}

func TestGetPathsUsesExplicitConfigRoot(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XSH_CONFIG_DIR", root)

	paths, err := GetPaths()
	if err != nil {
		t.Fatalf("GetPaths() error = %v", err)
	}
	if paths.ConfigDir != root {
		t.Fatalf("ConfigDir = %q, want override %q", paths.ConfigDir, root)
	}
	if _, err := os.Stat(root); err != nil {
		t.Fatalf("config root was not created: %v", err)
	}
}

func TestGetPathsUsesUserConfigDirWhenNoOverride(t *testing.T) {
	t.Setenv("XSH_CONFIG_DIR", "")
	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)

	paths, err := GetPaths()
	if err != nil {
		t.Fatalf("GetPaths() error = %v", err)
	}
	want := filepath.Join(base, ConfigDirName)
	if paths.ConfigDir != want {
		t.Fatalf("ConfigDir = %q, want %q", paths.ConfigDir, want)
	}
}

func TestExplicitConfigRootNeverTouchesHome(t *testing.T) {
	configRoot := t.TempDir()
	home := t.TempDir()
	t.Setenv("XSH_CONFIG_DIR", configRoot)
	t.Setenv("HOME", home)

	if _, err := LoadConfig(); err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if err := DefaultConfig().Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	checker := NewStartupChecker()
	checker.MarkChecked()

	legacyHomePath := filepath.Join(home, ".config", ConfigDirName)
	if _, err := os.Stat(legacyHomePath); !os.IsNotExist(err) {
		t.Fatalf("legacy HOME path was touched: stat error = %v", err)
	}
}

func TestCompatibilityPathWrappersUseGetPaths(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XSH_CONFIG_DIR", root)

	paths, err := GetPaths()
	if err != nil {
		t.Fatal(err)
	}
	configDir, err := GetConfigDir()
	if err != nil || configDir != paths.ConfigDir {
		t.Fatalf("GetConfigDir() = %q, %v; want %q", configDir, err, paths.ConfigDir)
	}
	configPath, err := GetConfigPath()
	if err != nil || configPath != paths.ConfigFile {
		t.Fatalf("GetConfigPath() = %q, %v; want %q", configPath, err, paths.ConfigFile)
	}
	authPath, err := GetAuthFile()
	if err != nil || authPath != paths.AuthFile {
		t.Fatalf("GetAuthFile() = %q, %v; want %q", authPath, err, paths.AuthFile)
	}
}
