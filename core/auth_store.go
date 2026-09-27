package core

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
)

// AuthStore isolates account persistence from command and client selection.
type AuthStore interface {
	Load(account string) (*AuthCredentials, error)
	Save(*AuthCredentials, string) error
	List() ([]string, error)
	SetDefault(string) error
	Remove(string) error
}

type FileAuthStore struct {
	paths Paths
	mu    sync.Mutex
}

func NewFileAuthStore(paths Paths) *FileAuthStore {
	return &FileAuthStore{paths: paths}
}

func (s *FileAuthStore) loadLocked() (*AuthData, error) {
	data, err := os.ReadFile(s.paths.AuthFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &AuthData{Accounts: make(map[string]*AuthCredentials)}, nil
		}
		return nil, err
	}

	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, s.corruptError(err)
	}
	if raw, ok := envelope["accounts"]; ok {
		var authData AuthData
		if err := json.Unmarshal(data, &authData); err != nil {
			return nil, s.corruptError(err)
		}
		if authData.Accounts == nil {
			authData.Accounts = make(map[string]*AuthCredentials)
		}
		if raw == nil {
			return nil, s.corruptError(fmt.Errorf("accounts field is null"))
		}
		return &authData, nil
	}

	var creds AuthCredentials
	if err := json.Unmarshal(data, &creds); err != nil || !creds.IsValid() {
		if err == nil {
			err = fmt.Errorf("missing valid auth_token and ct0")
		}
		return nil, s.corruptError(err)
	}
	name := creds.AccountName
	if name == "" {
		name = "default"
	}
	creds.AccountName = name
	authData := &AuthData{Default: name, Accounts: map[string]*AuthCredentials{name: &creds}}
	if err := s.writeLocked(authData); err != nil {
		return nil, fmt.Errorf("migrate legacy auth: %w", err)
	}
	return authData, nil
}

func (s *FileAuthStore) corruptError(err error) error {
	if backupErr := preserveCorruptFile(s.paths.AuthFile); backupErr != nil {
		return fmt.Errorf("invalid auth file: %v; preserving corrupt auth failed: %w", err, backupErr)
	}
	return fmt.Errorf("invalid auth file; corrupt source preserved at %s.bak: %w", s.paths.AuthFile, err)
}

func (s *FileAuthStore) writeLocked(data *AuthData) error {
	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return WriteFileAtomic(s.paths.AuthFile, encoded, 0600)
}

func cloneAuthCredentials(creds *AuthCredentials) *AuthCredentials {
	if creds == nil {
		return nil
	}
	copy := *creds
	if creds.Cookies != nil {
		copy.Cookies = make(map[string]string, len(creds.Cookies))
		for key, value := range creds.Cookies {
			copy.Cookies[key] = value
		}
	}
	return &copy
}

func (s *FileAuthStore) Load(account string) (*AuthCredentials, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := s.loadLocked()
	if err != nil {
		return nil, err
	}
	if account == "" {
		account = data.Default
		if account == "" {
			names := sortedAccountNames(data.Accounts)
			if len(names) > 0 {
				account = names[0]
			}
		}
	}
	return cloneAuthCredentials(data.Accounts[account]), nil
}

func (s *FileAuthStore) Save(creds *AuthCredentials, account string) error {
	if creds == nil || !creds.IsValid() {
		return fmt.Errorf("valid auth credentials are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := s.loadLocked()
	if err != nil {
		return err
	}
	if account == "" {
		account = creds.AccountName
	}
	if account == "" {
		account = "default"
	}
	copy := cloneAuthCredentials(creds)
	copy.AccountName = account
	data.Accounts[account] = copy
	if data.Default == "" {
		data.Default = account
	}
	return s.writeLocked(data)
}

func (s *FileAuthStore) List() ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := s.loadLocked()
	if err != nil {
		return nil, err
	}
	return sortedAccountNames(data.Accounts), nil
}

func (s *FileAuthStore) SetDefault(account string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := s.loadLocked()
	if err != nil {
		return err
	}
	if _, ok := data.Accounts[account]; !ok {
		return fmt.Errorf("account '%s' not found", account)
	}
	data.Default = account
	return s.writeLocked(data)
}

func (s *FileAuthStore) Remove(account string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := s.loadLocked()
	if err != nil {
		return err
	}
	if _, ok := data.Accounts[account]; !ok {
		return fmt.Errorf("account '%s' not found", account)
	}
	delete(data.Accounts, account)
	if data.Default == account {
		data.Default = ""
		names := sortedAccountNames(data.Accounts)
		if len(names) > 0 {
			data.Default = names[0]
		}
	}
	return s.writeLocked(data)
}

func sortedAccountNames(accounts map[string]*AuthCredentials) []string {
	names := make([]string, 0, len(accounts))
	for name := range accounts {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func defaultAuthStore() (*FileAuthStore, error) {
	paths, err := GetPaths()
	if err != nil {
		return nil, err
	}
	return NewFileAuthStore(paths), nil
}

// ResolveAccount applies explicit command, configured, then stored defaults.
func ResolveAccount(explicit string, cfg *Config, stored *AuthData) string {
	if explicit != "" {
		return explicit
	}
	if cfg != nil && cfg.DefaultAccount != "" {
		return cfg.DefaultAccount
	}
	if stored != nil {
		return stored.Default
	}
	return ""
}
