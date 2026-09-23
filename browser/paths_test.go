//go:build cgo

package browser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetDefaultFirefoxPathsUsesXDGConfigDir(t *testing.T) {
	home := t.TempDir()
	configDir := t.TempDir()
	profileDir := filepath.Join(configDir, "mozilla", "firefox", "abc.default-release")
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cookiePath := filepath.Join(profileDir, "cookies.sqlite")
	if err := os.WriteFile(cookiePath, []byte{}, 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", configDir)

	paths := GetDefaultFirefoxPaths()
	if len(paths) != 1 || paths[0] != cookiePath {
		t.Fatalf("GetDefaultFirefoxPaths() = %v, want [%s]", paths, cookiePath)
	}
}

func TestGetDefaultChromePathsIncludesProfileNetworkCookies(t *testing.T) {
	home := t.TempDir()
	cookiePath := filepath.Join(home, ".config", "google-chrome", "Profile 3", "Network", "Cookies")
	if err := os.MkdirAll(filepath.Dir(cookiePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cookiePath, []byte{}, 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", home)

	paths := GetDefaultChromePaths()
	for _, path := range paths {
		if path == cookiePath {
			return
		}
	}
	t.Fatalf("GetDefaultChromePaths() = %v, want it to include %s", paths, cookiePath)
}
