// Package tests provides unit tests for media operations.
package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/benoitpetit/xsh/core"
)

func TestValidateMediaFile(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	
	tests := []struct {
		name        string
		content     []byte
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid small jpg",
			content: make([]byte, 100),
			wantErr: false,
		},
		{
			name:        "file too large",
			content:     make([]byte, 6*1024*1024), // 6MB
			wantErr:     true,
			errContains: "too large",
		},
		{
			name:        "empty file",
			content:     []byte{},
			wantErr:     true,
			errContains: "empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, "test.jpg")
			err := os.WriteFile(testFile, tt.content, 0644)
			if err != nil {
				t.Fatal(err)
			}

			size, mimeType, err := core.ValidateMediaFile(testFile)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateMediaFile() error = nil, wantErr %v", tt.wantErr)
				}
				if tt.errContains != "" && err != nil {
					if !contains(err.Error(), tt.errContains) {
						t.Errorf("ValidateMediaFile() error = %v, should contain %v", err, tt.errContains)
					}
				}
			} else {
				if err != nil {
					t.Errorf("ValidateMediaFile() error = %v, wantErr %v", err, tt.wantErr)
				}
				if size != int64(len(tt.content)) {
					t.Errorf("ValidateMediaFile() size = %v, want %v", size, len(tt.content))
				}
				if mimeType == "" {
					t.Error("ValidateMediaFile() mimeType should not be empty")
				}
			}
		})
	}
}

func TestValidateMediaFile_NotFound(t *testing.T) {
	_, _, err := core.ValidateMediaFile("/nonexistent/path/file.jpg")
	if err == nil {
		t.Error("ValidateMediaFile() should return error for non-existent file")
	}
}

func TestUploadMedia_NotAuthenticated(t *testing.T) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		t.Skip("Skipping - no credentials")
	}
	defer client.Close()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jpg")
	os.WriteFile(testFile, []byte("fake image data"), 0644)

	// This should fail due to auth
	_, err = core.UploadMediaFile(client, testFile)
	if err == nil {
		t.Error("UploadMediaFile() should fail without valid auth")
	}
}

func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
