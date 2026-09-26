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
