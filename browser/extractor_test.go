package browser

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/benoitpetit/xsh/core"
)

func TestExtractFromPathsFallsBackToLaterProfile(t *testing.T) {
	first := filepath.Join(t.TempDir(), "Default", "Cookies")
	second := filepath.Join(t.TempDir(), "Profile 3", "Network", "Cookies")
	paths := []string{first, second}
	called := make([]string, 0, len(paths))

	creds, err := extractFromPaths(paths, func(path string) (*core.AuthCredentials, error) {
		called = append(called, path)
		if path == first {
			return nil, errors.New("profile is not logged in")
		}
		return &core.AuthCredentials{AuthToken: "token", Ct0: "csrf"}, nil
	})
	if err != nil {
		t.Fatalf("extractFromPaths() error = %v", err)
	}
	if creds == nil || !creds.IsValid() {
		t.Fatalf("extractFromPaths() credentials = %#v, want valid credentials", creds)
	}
	if len(called) != 2 || called[0] != first || called[1] != second {
		t.Fatalf("profile extraction order = %v, want [%s %s]", called, first, second)
	}
}

func TestExtractFromPathsReportsAllProfileFailures(t *testing.T) {
	paths := []string{"first", "second"}
	_, err := extractFromPaths(paths, func(path string) (*core.AuthCredentials, error) {
		return nil, errors.New(path + " failed")
	})
	if err == nil {
		t.Fatal("extractFromPaths() error = nil, want combined failure")
	}
	for _, path := range paths {
		if !containsSubstring(err.Error(), path+" failed") {
			t.Fatalf("extractFromPaths() error = %v, missing %q", err, path+" failed")
		}
	}
}

func TestBrowserDiscoveryOrderIsStable(t *testing.T) {
	home := t.TempDir()
	configDir := filepath.Join(home, ".config")
	touchCookieDatabase(t, filepath.Join(configDir, "google-chrome", "Default", "Cookies"))
	touchCookieDatabase(t, filepath.Join(configDir, "BraveSoftware", "Brave-Browser", "Default", "Cookies"))
	touchCookieDatabase(t, filepath.Join(configDir, "chromium", "Default", "Cookies"))

	first := discoverBrowserCandidates("linux", home, configDir, "", "")
	second := discoverBrowserCandidates("linux", home, configDir, "", "")
	if len(first) != len(second) {
		t.Fatalf("candidate counts differ: %d and %d", len(first), len(second))
	}
	for i := range first {
		if first[i].Name != second[i].Name {
			t.Fatalf("candidate order changed: %v then %v", first, second)
		}
	}
}

func containsSubstring(value, wanted string) bool {
	for i := 0; i+len(wanted) <= len(value); i++ {
		if value[i:i+len(wanted)] == wanted {
			return true
		}
	}
	return false
}
