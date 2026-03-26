package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/benoitpetit/xsh/core"
)

func TestImportCookiesFromFile(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Test with valid cookie JSON
	validCookies := `[
		{"name": "auth_token", "value": "test_token_12345", "domain": ".x.com"},
		{"name": "ct0", "value": "test_ct0_67890", "domain": ".x.com"}
	]`
	validFile := filepath.Join(tmpDir, "valid_cookies.json")
	if err := os.WriteFile(validFile, []byte(validCookies), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	creds, err := core.ImportCookiesFromFile(validFile)
	if err != nil {
		t.Errorf("ImportCookiesFromFile() error = %v", err)
	}
	if creds == nil {
		t.Fatal("Expected credentials, got nil")
	}
	if creds.AuthToken != "test_token_12345" {
		t.Errorf("Expected auth_token 'test_token_12345', got '%s'", creds.AuthToken)
	}
	if creds.Ct0 != "test_ct0_67890" {
		t.Errorf("Expected ct0 'test_ct0_67890', got '%s'", creds.Ct0)
	}

	// Test with missing file
	_, err = core.ImportCookiesFromFile(filepath.Join(tmpDir, "nonexistent.json"))
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Test with missing required cookies
	invalidCookies := `[
		{"name": "some_cookie", "value": "some_value", "domain": ".x.com"}
	]`
	invalidFile := filepath.Join(tmpDir, "invalid_cookies.json")
	if err := os.WriteFile(invalidFile, []byte(invalidCookies), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = core.ImportCookiesFromFile(invalidFile)
	if err == nil {
		t.Error("Expected error for missing auth_token and ct0")
	}
}

func TestListAccounts_NoAuthFile(t *testing.T) {
	// When no auth file exists, should return empty list without error
	accounts, err := core.ListAccounts()
	if err != nil {
		t.Errorf("ListAccounts() error = %v", err)
	}
	if accounts == nil {
		t.Error("Expected empty slice, got nil")
	}
}

func TestSetDefaultAccount_NoAuthFile(t *testing.T) {
	// When no auth file exists, should return error
	err := core.SetDefaultAccount("nonexistent")
	if err == nil {
		t.Error("Expected error when auth file doesn't exist")
	}
}
