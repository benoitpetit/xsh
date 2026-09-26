package browser

import (
	"os"
	"path/filepath"
	"runtime"
)

// BrowserCandidate describes one supported browser with its own cookie
// database candidates. Paths are ordered from the conventional profile to
// less common profiles and packaging layouts.
type BrowserCandidate struct {
	Name        string
	CookiePaths []string
}

var browserDiscoveryOrder = []string{"chrome", "brave", "edge", "chromium", "firefox"}

// DiscoverBrowserCandidates finds browsers that have at least one cookie
// database on the current host. The browser executable does not need to be in
// PATH because extraction operates on the profile database and OS key store.
func DiscoverBrowserCandidates() []BrowserCandidate {
	home, _ := os.UserHomeDir()
	configDir, _ := os.UserConfigDir()
	if configDir == "" {
		configDir = filepath.Join(home, ".config")
	}

	return discoverBrowserCandidates(
		runtime.GOOS,
		home,
		configDir,
		os.Getenv("LOCALAPPDATA"),
		os.Getenv("APPDATA"),
	)
}

func discoverBrowserCandidates(goos, home, configDir, localAppData, appData string) []BrowserCandidate {
	candidates := make([]BrowserCandidate, 0, len(browserDiscoveryOrder))
	for _, name := range browserDiscoveryOrder {
		paths := cookiePathsForBrowser(goos, home, configDir, localAppData, appData, name)
		existing := existingCookiePaths(paths)
		if len(existing) > 0 {
			candidates = append(candidates, BrowserCandidate{Name: name, CookiePaths: existing})
		}
	}
	return candidates
}

func cookiePathsForBrowser(goos, home, configDir, localAppData, appData, browserName string) []string {
	if browserName == "firefox" {
		return firefoxCookiePaths(goos, home, configDir, appData)
	}

	var paths []string
	for _, userDataDir := range chromiumUserDataDirs(goos, home, configDir, localAppData, browserName) {
		paths = append(paths, discoverChromiumCookiePaths(userDataDir)...)
	}
	return uniqueStrings(paths)
}

func existingCookiePaths(paths []string) []string {
	existing := make([]string, 0, len(paths))
	for _, path := range paths {
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			existing = append(existing, path)
		}
	}
	return existing
}

func chromiumUserDataDirs(goos, home, configDir, localAppData, browserName string) []string {
	var relative []string
	switch browserName {
	case "chrome":
		relative = []string{"google-chrome", "google-chrome-beta", "google-chrome-unstable"}
	case "brave":
		relative = []string{"BraveSoftware/Brave-Browser", "BraveSoftware/Brave-Browser-Beta", "BraveSoftware/Brave-Browser-Nightly"}
	case "edge":
		relative = []string{"microsoft-edge", "microsoft-edge-beta", "microsoft-edge-dev"}
	case "chromium":
		relative = []string{"chromium", "chromium-browser"}
	default:
		return nil
	}

	dirs := make([]string, 0, len(relative)+2)
	switch goos {
	case "linux":
		for _, path := range relative {
			dirs = append(dirs, filepath.Join(configDir, path))
		}
		switch browserName {
		case "chrome":
			dirs = append(dirs, filepath.Join(home, ".var/app/com.google.Chrome/config/google-chrome"))
		case "chromium":
			dirs = append(dirs,
				filepath.Join(home, "snap/chromium/common/chromium"),
				filepath.Join(home, ".var/app/org.chromium.Chromium/config/chromium"),
			)
		}
	case "darwin":
		for _, path := range relative {
			dirs = append(dirs, filepath.Join(home, "Library/Application Support", path))
		}
	case "windows":
		for _, path := range relative {
			dirs = append(dirs, filepath.Join(localAppData, path, "User Data"))
		}
	}
	return uniqueStrings(dirs)
}

func firefoxCookiePaths(goos, home, configDir, appData string) []string {
	var roots []string
	switch goos {
	case "linux":
		roots = []string{
			filepath.Join(configDir, "mozilla/firefox"),
			filepath.Join(home, ".mozilla/firefox"),
			filepath.Join(home, "snap/firefox/common/.mozilla/firefox"),
			filepath.Join(home, ".var/app/org.mozilla.firefox/.mozilla/firefox"),
		}
	case "darwin":
		roots = []string{filepath.Join(home, "Library/Application Support/Firefox/Profiles")}
	case "windows":
		roots = []string{filepath.Join(appData, "Mozilla/Firefox/Profiles")}
	}

	paths := make([]string, 0)
	for _, root := range uniqueStrings(roots) {
		paths = append(paths, findFirefoxProfiles(root)...)
	}
	return uniqueStrings(paths)
}
