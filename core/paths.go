package core

import (
	"os"
	"path/filepath"
)

const (
	endpointCacheFileName    = "graphql_ops.json"
	transactionCacheFileName = "transaction_cache.json"
	startupMarkerFileName    = ".last_endpoint_check"
)

// Paths contains all persistent xsh paths under one portable configuration root.
type Paths struct {
	ConfigDir        string
	ConfigFile       string
	AuthFile         string
	EndpointCache    string
	TransactionCache string
	StartupMarker    string
}

func pathsForConfigDir(configDir string) Paths {
	return Paths{
		ConfigDir:        configDir,
		ConfigFile:       filepath.Join(configDir, ConfigFileName),
		AuthFile:         filepath.Join(configDir, AuthFileName),
		EndpointCache:    filepath.Join(configDir, endpointCacheFileName),
		TransactionCache: filepath.Join(configDir, transactionCacheFileName),
		StartupMarker:    filepath.Join(configDir, startupMarkerFileName),
	}
}

func pathsForBase(base string) Paths {
	return pathsForConfigDir(filepath.Join(base, ConfigDirName))
}

// GetPaths returns portable configuration and cache paths, creating the root.
// XSH_CONFIG_DIR is intentionally supported so tests and controlled deployments
// can isolate state without changing the user's real profile.
func GetPaths() (Paths, error) {
	configDir := os.Getenv("XSH_CONFIG_DIR")
	if configDir == "" {
		base, err := os.UserConfigDir()
		if err != nil {
			return Paths{}, err
		}
		configDir = filepath.Join(base, ConfigDirName)
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return Paths{}, err
	}
	return pathsForConfigDir(configDir), nil
}
