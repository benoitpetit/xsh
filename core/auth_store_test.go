package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func testAuthStore(t *testing.T) (*FileAuthStore, Paths) {
	t.Helper()
	root := t.TempDir()
	paths := pathsForConfigDir(root)
	return NewFileAuthStore(paths), paths
}

func TestAuthStoreMigratesLegacySingleAccount(t *testing.T) {
	store, paths := testAuthStore(t)
	if err := os.MkdirAll(filepath.Dir(paths.AuthFile), 0700); err != nil {
		t.Fatal(err)
	}
	legacy := `{"auth_token":"legacy-token","ct0":"legacy-ct0"}`
	if err := os.WriteFile(paths.AuthFile, []byte(legacy), 0600); err != nil {
		t.Fatal(err)
	}
	creds, err := store.Load("default")
	if err != nil || creds == nil || creds.AuthToken != "legacy-token" {
		t.Fatalf("Load() = %#v, %v", creds, err)
	}
	data, err := os.ReadFile(paths.AuthFile)
	if err != nil {
		t.Fatal(err)
	}
	var migrated AuthData
	if err := json.Unmarshal(data, &migrated); err != nil || migrated.Accounts["default"] == nil {
		t.Fatalf("legacy auth was not migrated: %s; %v", data, err)
	}
}

func TestAuthStoreReportsMalformedJSONAndPreservesSource(t *testing.T) {
	store, paths := testAuthStore(t)
	if err := EnsurePrivateDir(filepath.Dir(paths.AuthFile)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paths.AuthFile, []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Load("default"); err == nil {
		t.Fatal("Load() accepted malformed auth JSON")
	}
	if _, err := os.Stat(paths.AuthFile + ".bak"); err != nil {
		t.Fatalf("auth backup missing: %v", err)
	}
}

func TestAuthStoreConcurrentSavesPreserveAccounts(t *testing.T) {
	store, _ := testAuthStore(t)
	var wg sync.WaitGroup
	for _, name := range []string{"one", "two", "three", "four"} {
		name := name
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := store.Save(&AuthCredentials{AuthToken: name + "-token", Ct0: name + "-ct0"}, name); err != nil {
				t.Errorf("Save(%q) error = %v", name, err)
			}
		}()
	}
	wg.Wait()
	accounts, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 4 {
		t.Fatalf("accounts = %#v, want all concurrent accounts", accounts)
	}
}

func TestResolveAccountPrecedence(t *testing.T) {
	stored := &AuthData{Default: "stored", Accounts: map[string]*AuthCredentials{"stored": {}, "config": {}, "explicit": {}}}
	cfg := &Config{DefaultAccount: "config"}
	if got := ResolveAccount("explicit", cfg, stored); got != "explicit" {
		t.Fatalf("explicit account = %q", got)
	}
	if got := ResolveAccount("", cfg, stored); got != "config" {
		t.Fatalf("config account = %q", got)
	}
	if got := ResolveAccount("", &Config{}, stored); got != "stored" {
		t.Fatalf("stored account = %q", got)
	}
}

func TestImportCookiesRequiresExactXDomains(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cookies.json")
	cookies := []map[string]string{
		{"name": "auth_token", "value": "valid-token", "domain": "evilx.com"},
		{"name": "ct0", "value": "valid-ct0", "domain": "sub.x.com"},
	}
	data, _ := json.Marshal(cookies)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportCookiesFromFile(path); err == nil {
		t.Fatal("ImportCookiesFromFile() accepted incomplete exact-domain cookies")
	}
}

func TestGetCredentialsUsesSelectedAccountBeforeEnvironment(t *testing.T) {
	t.Setenv("XSH_CONFIG_DIR", t.TempDir())
	t.Setenv("X_AUTH_TOKEN", "env-token")
	t.Setenv("X_CT0", "env-ct0")
	if err := SaveAuth(&AuthCredentials{AuthToken: "stored-token", Ct0: "stored-ct0"}, "selected"); err != nil {
		t.Fatal(err)
	}
	creds, err := GetCredentials("selected")
	if err != nil || creds == nil || creds.AuthToken != "stored-token" {
		t.Fatalf("GetCredentials() = %#v, %v; selected account lost precedence", creds, err)
	}
}

func TestGetCredentialsSavesBrowserCredentialsUnderSelectedAccount(t *testing.T) {
	t.Setenv("XSH_CONFIG_DIR", t.TempDir())
	t.Setenv("X_AUTH_TOKEN", "")
	t.Setenv("X_CT0", "")
	previous := tryBrowserExtraction
	tryBrowserExtraction = func() (*AuthCredentials, string, error) {
		return &AuthCredentials{AuthToken: "browser-token", Ct0: "browser-ct0"}, "fixture", nil
	}
	defer func() { tryBrowserExtraction = previous }()

	creds, err := GetCredentials("browser-account")
	if err != nil || creds == nil || creds.AccountName != "browser-account" {
		t.Fatalf("GetCredentials() = %#v, %v", creds, err)
	}
	stored, err := LoadStoredAuth("browser-account")
	if err != nil || stored == nil || stored.AuthToken != "browser-token" {
		t.Fatalf("browser credentials were not saved under selected account: %#v, %v", stored, err)
	}
}
