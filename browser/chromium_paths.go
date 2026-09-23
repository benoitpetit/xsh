package browser

import (
	"os"
	"path/filepath"
)

// discoverChromiumCookiePaths returns cookie database candidates for every
// profile under a Chromium user-data directory. Newer Chromium builds store
// the database under Profile/Network/Cookies, while older builds used
// Profile/Cookies.
func discoverChromiumCookiePaths(userDataDir string) []string {
	profiles := []string{"Default", "Profile 1", "Profile 2"}
	if entries, err := os.ReadDir(userDataDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() && entry.Name() != "Default" {
				profiles = append(profiles, entry.Name())
			}
		}
	}

	paths := make([]string, 0, len(profiles)*2)
	seen := make(map[string]struct{})
	for _, profile := range profiles {
		for _, path := range []string{
			filepath.Join(userDataDir, profile, "Network", "Cookies"),
			filepath.Join(userDataDir, profile, "Cookies"),
		} {
			if _, ok := seen[path]; ok {
				continue
			}
			seen[path] = struct{}{}
			paths = append(paths, path)
		}
	}
	return paths
}

func linuxChromiumCookiePath(browserName string) string {
	home, _ := os.UserHomeDir()
	configDir, err := os.UserConfigDir()
	if err != nil || configDir == "" {
		configDir = filepath.Join(home, ".config")
	}

	userDataDirs := map[string]string{
		"brave":    filepath.Join(configDir, "BraveSoftware/Brave-Browser"),
		"edge":     filepath.Join(configDir, "microsoft-edge"),
		"chromium": filepath.Join(configDir, "chromium"),
	}
	userDataDir, ok := userDataDirs[browserName]
	if !ok {
		return ""
	}

	paths := discoverChromiumCookiePaths(userDataDir)
	for _, path := range paths {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}
	if len(paths) > 0 {
		return paths[0]
	}
	return ""
}

func hasExistingCookieDatabase(paths []string) bool {
	for _, path := range paths {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return true
		}
	}
	return false
}
