package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/benoitpetit/xsh/core"
)

// TestAuthCredentials teste la validation des credentials
func TestAuthCredentials(t *testing.T) {
	tests := []struct {
		name      string
		authToken string
		ct0       string
		wantValid bool
	}{
		{
			name:      "valid credentials",
			authToken: "abc123",
			ct0:       "xyz789",
			wantValid: true,
		},
		{
			name:      "empty credentials",
			authToken: "",
			ct0:       "",
			wantValid: false,
		},
		{
			name:      "partial credentials - no ct0",
			authToken: "abc",
			ct0:       "",
			wantValid: false,
		},
		{
			name:      "partial credentials - no auth_token",
			authToken: "",
			ct0:       "xyz",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creds := &core.AuthCredentials{
				AuthToken: tt.authToken,
				Ct0:       tt.ct0,
			}
			if got := creds.IsValid(); got != tt.wantValid {
				t.Errorf("IsValid() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

// TestEnvAuth teste la récupération des credentials depuis les variables d'environnement
func TestEnvAuth(t *testing.T) {
	tests := []struct {
		name          string
		envVars       map[string]string
		wantToken     string
		wantCt0       string
		wantNil       bool
	}{
		{
			name: "X_ env vars",
			envVars: map[string]string{
				"X_AUTH_TOKEN": "token1",
				"X_CT0":        "csrf1",
			},
			wantToken: "token1",
			wantCt0:   "csrf1",
			wantNil:   false,
		},
		{
			name: "TWITTER_ env vars",
			envVars: map[string]string{
				"TWITTER_AUTH_TOKEN": "t2",
				"TWITTER_CT0":        "c2",
			},
			wantToken: "t2",
			wantCt0:   "c2",
			wantNil:   false,
		},
		{
			name:          "no env vars",
			envVars:       map[string]string{},
			wantNil:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Nettoyer les variables d'environnement
			os.Unsetenv("X_AUTH_TOKEN")
			os.Unsetenv("X_CT0")
			os.Unsetenv("TWITTER_AUTH_TOKEN")
			os.Unsetenv("TWITTER_CT0")

			// Définir les variables de test
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			got := core.GetAuthFromEnv()
			if tt.wantNil {
				if got != nil {
					t.Errorf("GetAuthFromEnv() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Errorf("GetAuthFromEnv() = nil, want non-nil")
				return
			}

			if got.AuthToken != tt.wantToken {
				t.Errorf("AuthToken = %v, want %v", got.AuthToken, tt.wantToken)
			}
			if got.Ct0 != tt.wantCt0 {
				t.Errorf("Ct0 = %v, want %v", got.Ct0, tt.wantCt0)
			}
		})
	}
}

// TestStoredAuth teste la sauvegarde et le chargement des credentials
func TestStoredAuth(t *testing.T) {
	// Créer un répertoire temporaire
	tmpDir := t.TempDir()

	// Utiliser une variable d'environnement pour changer le chemin
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Créer le répertoire .config/xsh
	configDir := filepath.Join(tmpDir, ".config", "xsh")
	os.MkdirAll(configDir, 0755)

	t.Run("save and load", func(t *testing.T) {
		creds := &core.AuthCredentials{
			AuthToken: "saved_token",
			Ct0:       "saved_ct0",
		}

		err := core.SaveAuth(creds, "test_account")
		if err != nil {
			t.Fatalf("SaveAuth() error = %v", err)
		}

		loaded, err := core.LoadStoredAuth("test_account")
		if err != nil {
			t.Fatalf("LoadStoredAuth() error = %v", err)
		}

		if loaded == nil {
			t.Fatal("LoadStoredAuth() returned nil")
		}

		if loaded.AuthToken != "saved_token" {
			t.Errorf("AuthToken = %v, want %v", loaded.AuthToken, "saved_token")
		}
	})

	t.Run("load nonexistent", func(t *testing.T) {
		loaded, err := core.LoadStoredAuth("nonexistent")
		if err != nil {
			t.Fatalf("LoadStoredAuth() error = %v", err)
		}
		if loaded != nil {
			t.Errorf("LoadStoredAuth() = %v, want nil", loaded)
		}
	})

	t.Run("multi account", func(t *testing.T) {
		creds1 := &core.AuthCredentials{
			AuthToken: "t1",
			Ct0:       "c1",
		}
		creds2 := &core.AuthCredentials{
			AuthToken: "t2",
			Ct0:       "c2",
		}

		err := core.SaveAuth(creds1, "account1")
		if err != nil {
			t.Fatalf("SaveAuth(account1) error = %v", err)
		}

		err = core.SaveAuth(creds2, "account2")
		if err != nil {
			t.Fatalf("SaveAuth(account2) error = %v", err)
		}

		loaded1, err := core.LoadStoredAuth("account1")
		if err != nil {
			t.Fatalf("LoadStoredAuth(account1) error = %v", err)
		}

		loaded2, err := core.LoadStoredAuth("account2")
		if err != nil {
			t.Fatalf("LoadStoredAuth(account2) error = %v", err)
		}

		if loaded1.AuthToken != "t1" {
			t.Errorf("account1 AuthToken = %v, want %v", loaded1.AuthToken, "t1")
		}
		if loaded2.AuthToken != "t2" {
			t.Errorf("account2 AuthToken = %v, want %v", loaded2.AuthToken, "t2")
		}
	})
}

// TestListAccounts teste la liste des comptes
func TestListAccounts(t *testing.T) {
	tmpDir := t.TempDir()

	// Utiliser une variable d'environnement pour changer le chemin
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Créer le répertoire .config/xsh
	configDir := filepath.Join(tmpDir, ".config", "xsh")
	os.MkdirAll(configDir, 0755)

	// Créer plusieurs comptes
	creds1 := &core.AuthCredentials{AuthToken: "t1", Ct0: "c1"}
	creds2 := &core.AuthCredentials{AuthToken: "t2", Ct0: "c2"}

	core.SaveAuth(creds1, "work")
	core.SaveAuth(creds2, "personal")

	accounts, err := core.ListAccounts()
	if err != nil {
		t.Fatalf("ListAccounts() error = %v", err)
	}

	if len(accounts) != 2 {
		t.Errorf("ListAccounts() returned %d accounts, want 2", len(accounts))
	}

	// Vérifier que les deux comptes sont présents
	hasWork := false
	hasPersonal := false
	for _, acc := range accounts {
		if acc == "work" {
			hasWork = true
		}
		if acc == "personal" {
			hasPersonal = true
		}
	}

	if !hasWork {
		t.Error("ListAccounts() missing 'work' account")
	}
	if !hasPersonal {
		t.Error("ListAccounts() missing 'personal' account")
	}
}

// TestSetDefaultAccount teste le changement de compte par défaut
func TestSetDefaultAccount(t *testing.T) {
	tmpDir := t.TempDir()

	// Utiliser une variable d'environnement pour changer le chemin
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Créer le répertoire .config/xsh
	configDir := filepath.Join(tmpDir, ".config", "xsh")
	os.MkdirAll(configDir, 0755)

	// Créer deux comptes
	creds1 := &core.AuthCredentials{AuthToken: "t1", Ct0: "c1"}
	core.SaveAuth(creds1, "account1")

	creds2 := &core.AuthCredentials{AuthToken: "t2", Ct0: "c2"}
	core.SaveAuth(creds2, "account2")

	// Changer le compte par défaut
	err := core.SetDefaultAccount("account2")
	if err != nil {
		t.Fatalf("SetDefaultAccount() error = %v", err)
	}

	// Vérifier que le compte par défaut a changé
	loaded, err := core.LoadStoredAuth("")
	if err != nil {
		t.Fatalf("LoadStoredAuth() error = %v", err)
	}

	if loaded == nil {
		t.Fatal("LoadStoredAuth() returned nil")
	}

	if loaded.AuthToken != "t2" {
		t.Errorf("Default account AuthToken = %v, want %v", loaded.AuthToken, "t2")
	}

	// Tester avec un compte inexistant
	err = core.SetDefaultAccount("nonexistent")
	if err == nil {
		t.Error("SetDefaultAccount(nonexistent) should return error")
	}
}

// TestRemoveAuth teste la suppression d'un compte
func TestRemoveAuth(t *testing.T) {
	tmpDir := t.TempDir()

	// Utiliser une variable d'environnement pour changer le chemin
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Créer le répertoire .config/xsh
	configDir := filepath.Join(tmpDir, ".config", "xsh")
	os.MkdirAll(configDir, 0755)

	// Créer un compte
	creds := &core.AuthCredentials{AuthToken: "t1", Ct0: "c1"}
	core.SaveAuth(creds, "to_delete")

	// Supprimer le compte
	err := core.RemoveAuth("to_delete")
	if err != nil {
		t.Fatalf("RemoveAuth() error = %v", err)
	}

	// Vérifier que le compte n'existe plus
	loaded, err := core.LoadStoredAuth("to_delete")
	if err != nil {
		t.Fatalf("LoadStoredAuth() error = %v", err)
	}
	if loaded != nil {
		t.Error("LoadStoredAuth() should return nil after removal")
	}

	// Tester la suppression d'un compte inexistant
	err = core.RemoveAuth("nonexistent")
	if err == nil {
		t.Error("RemoveAuth(nonexistent) should return error")
	}
}

// TestAuthError teste le type d'erreur d'authentification
func TestAuthError(t *testing.T) {
	err := &core.AuthError{Message: "test error"}
	if err.Error() != "test error" {
		t.Errorf("Error() = %v, want %v", err.Error(), "test error")
	}
}

// TestSanitizeCookieValue teste la sanitization des valeurs de cookies
func TestSanitizeCookieValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no quotes",
			input:    "abc123",
			expected: "abc123",
		},
		{
			name:     "with surrounding quotes",
			input:    `"abc123"`,
			expected: "abc123",
		},
		{
			name:     "with internal quotes",
			input:    `ab"c123`,
			expected: "abc123",
		},
		{
			name:     "with spaces",
			input:    "  abc123  ",
			expected: "abc123",
		},
		{
			name:     "with commas",
			input:    "abc,123",
			expected: "abc123",
		},
		{
			name:     "with semicolons",
			input:    "abc;123",
			expected: "abc123",
		},
		{
			name:     "with backslashes",
			input:    `abc\123`,
			expected: "abc123",
		},
		{
			name:     "complex case",
			input:    `"abc;,"def\\"`,
			expected: "abcdef",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only invalid chars",
			input:    `";,\\`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := core.SanitizeCookieValue(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizeCookieValue(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestAuthCredentialsSanitization teste la sanitization des credentials
func TestAuthCredentialsSanitization(t *testing.T) {
	creds := &core.AuthCredentials{
		AuthToken: `"token"with"quotes"`,
		Ct0:       `"csrf\"value"`,
		Cookies: map[string]string{
			"auth_token": `"token"with"quotes"`,
			"ct0":        `"csrf\"value"`,
			"other":      `"normal"`,
		},
	}

	// Test GetSanitizedAuthToken
	sanitizedToken := creds.GetSanitizedAuthToken()
	expectedToken := `tokenwithquotes`
	if sanitizedToken != expectedToken {
		t.Errorf("GetSanitizedAuthToken() = %q, want %q", sanitizedToken, expectedToken)
	}

	// Test GetSanitizedCt0
	sanitizedCt0 := creds.GetSanitizedCt0()
	expectedCt0 := `csrfvalue`
	if sanitizedCt0 != expectedCt0 {
		t.Errorf("GetSanitizedCt0() = %q, want %q", sanitizedCt0, expectedCt0)
	}

	// Test GetSanitizedCookies
	sanitizedCookies := creds.GetSanitizedCookies()
	if sanitizedCookies["auth_token"] != expectedToken {
		t.Errorf("GetSanitizedCookies()[auth_token] = %q, want %q", sanitizedCookies["auth_token"], expectedToken)
	}
	if sanitizedCookies["ct0"] != expectedCt0 {
		t.Errorf("GetSanitizedCookies()[ct0] = %q, want %q", sanitizedCookies["ct0"], expectedCt0)
	}
	if sanitizedCookies["other"] != "normal" {
		t.Errorf("GetSanitizedCookies()[other] = %q, want %q", sanitizedCookies["other"], "normal")
	}
}
