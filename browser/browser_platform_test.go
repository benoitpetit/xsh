package browser

import (
	"database/sql"
	"path/filepath"
	"runtime"
	"testing"

	_ "modernc.org/sqlite"
)

func TestFirefoxCookieExtractorReadsSQLiteWithoutCGOStub(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Firefox fixture path semantics are covered by the Windows build job")
	}
	databasePath := filepath.Join(t.TempDir(), "cookies.sqlite")
	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE moz_cookies (host TEXT, name TEXT, value TEXT, path TEXT);
INSERT INTO moz_cookies(host, name, value, path) VALUES
('x.com', 'auth_token', 'auth-value', '/'),
('x.com', 'ct0', 'csrf-value', '/');`)
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	credentials, err := (&FirefoxCookieExtractor{Path: databasePath}).ExtractCookies()
	if err != nil {
		t.Fatalf("ExtractCookies() error = %v", err)
	}
	if credentials.AuthToken != "auth-value" || credentials.Ct0 != "csrf-value" {
		t.Fatalf("credentials = %#v", credentials)
	}
}
