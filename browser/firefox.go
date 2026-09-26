//go:build cgo
// +build cgo

package browser

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/benoitpetit/xsh/core"
	_ "github.com/mattn/go-sqlite3"
)

// FirefoxCookieExtractor extracts cookies from Firefox
type FirefoxCookieExtractor struct {
	Path string
}

// FirefoxCookie represents a Firefox cookie
type FirefoxCookie struct {
	Host  string `json:"host"`
	Name  string `json:"name"`
	Value string `json:"value"`
	Path  string `json:"path"`
}

// GetDefaultFirefoxPaths returns default Firefox cookie paths by OS
func GetDefaultFirefoxPaths() []string {
	home, _ := os.UserHomeDir()
	configDir, _ := os.UserConfigDir()
	if configDir == "" {
		configDir = filepath.Join(home, ".config")
	}
	return firefoxCookiePaths(runtime.GOOS, home, configDir, os.Getenv("APPDATA"))
}

// findFirefoxProfiles finds Firefox profile directories containing cookies.sqlite
func findFirefoxProfiles(profilesDir string) []string {
	var paths []string

	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return paths
	}

	for _, entry := range entries {
		if entry.IsDir() {
			cookiePath := filepath.Join(profilesDir, entry.Name(), "cookies.sqlite")
			if _, err := os.Stat(cookiePath); err == nil {
				paths = append(paths, cookiePath)
			}
		}
	}

	return paths
}

// ExtractCookies extracts Twitter/X cookies from Firefox
func (f *FirefoxCookieExtractor) ExtractCookies() (*core.AuthCredentials, error) {
	if f.Path == "" {
		paths := GetDefaultFirefoxPaths()
		if len(paths) == 0 {
			return nil, fmt.Errorf("no Firefox cookie database found")
		}
		f.Path = paths[0] // Use first profile
	}

	// Copy database to temp location
	// Use secure temp file with random name to prevent symlink attacks
	tempFile, err := os.CreateTemp("", "firefox_cookies_*.db")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tempDB := tempFile.Name()
	tempFile.Close()

	if err := copyFile(f.Path, tempDB); err != nil {
		os.Remove(tempDB)
		return nil, fmt.Errorf("failed to copy cookie database: %w", err)
	}
	defer os.Remove(tempDB)

	db, err := sql.Open("sqlite3", tempDB)
	if err != nil {
		return nil, fmt.Errorf("failed to open cookie database: %w", err)
	}
	defer db.Close()

	// Query for Twitter/X cookies
	query := `
		SELECT host, name, value, path
		FROM moz_cookies
		WHERE host = 'twitter.com'
		   OR host LIKE '%.twitter.com'
		   OR host = 'x.com'
		   OR host LIKE '%.x.com'
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query cookies: %w", err)
	}
	defer rows.Close()

	acc := newCookieAccumulator()

	for rows.Next() {
		var cookie FirefoxCookie
		err := rows.Scan(&cookie.Host, &cookie.Name, &cookie.Value, &cookie.Path)
		if err != nil {
			continue
		}

		acc.add(cookie.Host, cookie.Name, cookie.Value, nil)
	}

	return acc.credentials("Firefox")
}

// IsFirefoxAvailable checks if Firefox cookies are available
func IsFirefoxAvailable() bool {
	paths := GetDefaultFirefoxPaths()
	return len(paths) > 0
}
