// Package tests provides unit tests for transaction ID generation.
package tests

import (
	"testing"

	"github.com/benoitpetit/xsh/core"
)

func TestGetTransactionGenerator(t *testing.T) {
	// Test singleton pattern
	tg1 := core.GetTransactionGenerator()
	tg2 := core.GetTransactionGenerator()

	if tg1 != tg2 {
		t.Error("GetTransactionGenerator() should return singleton")
	}
}

func TestGenerateTransactionID(t *testing.T) {
	tg := core.GetTransactionGenerator()

	// Generate multiple IDs
	id1 := tg.Generate("GET", "/home")
	id2 := tg.Generate("POST", "/create")
	id3 := tg.Generate("GET", "/home")

	// IDs should not be empty
	if id1 == "" {
		t.Error("Generate() returned empty ID")
	}
	if id2 == "" {
		t.Error("Generate() returned empty ID")
	}

	// IDs should be at least 90 characters (observed length)
	if len(id1) < 90 {
		t.Errorf("Generate() ID too short: %d chars, expected at least 90", len(id1))
	}

	// Different operations should produce different IDs
	if id1 == id2 {
		t.Error("Generate() should produce different IDs for different operations")
	}

	// Same operation might produce different IDs due to timestamp
	_ = id3
	t.Logf("Generated ID length: %d", len(id1))
}

func TestNeedsTransactionID(t *testing.T) {
	tests := []struct {
		operation string
		want      bool
	}{
		{"SearchTimeline", true},
		{"HomeTimeline", true},
		{"UserTweets", true},
		{"CreateTweet", false},
		{"FavoriteTweet", false},
		{"DeleteTweet", false},
		{"Unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.operation, func(t *testing.T) {
			got := core.NeedsTransactionID(tt.operation)
			if got != tt.want {
				t.Errorf("NeedsTransactionID(%q) = %v, want %v", tt.operation, got, tt.want)
			}
		})
	}
}

func TestIsWriteOperation(t *testing.T) {
	tests := []struct {
		operation string
		want      bool
	}{
		{"CreateTweet", true},
		{"DeleteTweet", true},
		{"FavoriteTweet", true},
		{"CreateBookmark", true},
		{"SearchTimeline", false},
		{"HomeTimeline", false},
		{"Unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.operation, func(t *testing.T) {
			got := core.IsWriteOperation(tt.operation)
			if got != tt.want {
				t.Errorf("IsWriteOperation(%q) = %v, want %v", tt.operation, got, tt.want)
			}
		})
	}
}

func TestGenerateTransactionIDForRequest(t *testing.T) {
	// Write operations should return empty
	id := core.GenerateTransactionIDForRequest("POST", "/create", "CreateTweet")
	if id != "" {
		t.Error("GenerateTransactionIDForRequest() should return empty for write operations")
	}

	// Read operations should return ID
	id = core.GenerateTransactionIDForRequest("GET", "/search", "SearchTimeline")
	if id == "" {
		t.Error("GenerateTransactionIDForRequest() should return ID for read operations")
	}

	// Unknown operations should return empty
	id = core.GenerateTransactionIDForRequest("GET", "/unknown", "UnknownOp")
	_ = id
	// This might return ID or empty depending on implementation
}
