//go:build cgo

package browser

import (
	"os"
	"path/filepath"
	"testing"
)

func touchCookieDatabase(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte{}, 0o600); err != nil {
		t.Fatal(err)
	}
}

func candidateByName(t *testing.T, candidates []BrowserCandidate, name string) BrowserCandidate {
	t.Helper()
	for _, candidate := range candidates {
		if candidate.Name == name {
			return candidate
		}
	}
	t.Fatalf("candidate %q not found in %#v", name, candidates)
	return BrowserCandidate{}
}

func TestDiscoverBrowserCandidatesFindsChromeProfilesWithoutExecutable(t *testing.T) {
	home := t.TempDir()
	configDir := filepath.Join(home, ".config")
	cookiePath := filepath.Join(configDir, "google-chrome", "Default", "Network", "Cookies")
	touchCookieDatabase(t, cookiePath)

	candidates := discoverBrowserCandidates("linux", home, configDir, "", "")
	chrome := candidateByName(t, candidates, "chrome")
	if len(chrome.CookiePaths) != 1 || chrome.CookiePaths[0] != cookiePath {
		t.Fatalf("chrome candidate paths = %v, want [%s]", chrome.CookiePaths, cookiePath)
	}
}

func TestDiscoverBrowserCandidatesKeepsChromiumProfilesIsolated(t *testing.T) {
	home := t.TempDir()
	configDir := filepath.Join(home, ".config")
	chromeCookie := filepath.Join(configDir, "google-chrome", "Default", "Cookies")
	braveCookie := filepath.Join(configDir, "BraveSoftware", "Brave-Browser", "Default", "Network", "Cookies")
	touchCookieDatabase(t, chromeCookie)
	touchCookieDatabase(t, braveCookie)

	candidates := discoverBrowserCandidates("linux", home, configDir, "", "")
	chrome := candidateByName(t, candidates, "chrome")
	brave := candidateByName(t, candidates, "brave")
	if containsPath(chrome.CookiePaths, braveCookie) {
		t.Fatalf("chrome candidate contains Brave cookie path: %v", chrome.CookiePaths)
	}
	if containsPath(brave.CookiePaths, chromeCookie) {
		t.Fatalf("Brave candidate contains Chrome cookie path: %v", brave.CookiePaths)
	}
}

func TestDiscoverChromiumCookiePathsFindsAllProfilesAndLayouts(t *testing.T) {
	userDataDir := t.TempDir()
	profile3 := filepath.Join(userDataDir, "Profile 3", "Network", "Cookies")
	profile4 := filepath.Join(userDataDir, "Profile 4", "Cookies")
	touchCookieDatabase(t, profile3)
	touchCookieDatabase(t, profile4)

	paths := discoverChromiumCookiePaths(userDataDir)
	if !containsPath(paths, profile3) || !containsPath(paths, profile4) {
		t.Fatalf("discovered paths = %v, want both %s and %s", paths, profile3, profile4)
	}
}

func TestFindFirefoxProfilesAcceptsAnyProfileSuffix(t *testing.T) {
	profilesDir := t.TempDir()
	cookiePath := filepath.Join(profilesDir, "abc.default-esr", "cookies.sqlite")
	touchCookieDatabase(t, cookiePath)

	paths := findFirefoxProfiles(profilesDir)
	if len(paths) != 1 || paths[0] != cookiePath {
		t.Fatalf("Firefox paths = %v, want [%s]", paths, cookiePath)
	}
}

func containsPath(paths []string, wanted string) bool {
	for _, path := range paths {
		if path == wanted {
			return true
		}
	}
	return false
}
