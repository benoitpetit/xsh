// Package tests provides unit tests for validation utilities.
package tests

import (
	"strings"
	"testing"

	"github.com/benoitpetit/xsh/utils"
)

// TestValidateTweetID tests tweet ID validation
func TestValidateTweetID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected bool
	}{
		{"valid numeric", "123456789", true},
		{"valid long", "1234567890123456789", true},
		{"empty", "", false},
		{"non-numeric", "abc123", false},
		{"with spaces", "123 456", false},
		{"special chars", "123-456", false},
		{"decimal", "123.456", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ValidateTweetID(tt.id)
			if result != tt.expected {
				t.Errorf("ValidateTweetID(%q) = %v, want %v", tt.id, result, tt.expected)
			}
		})
	}
}

// TestValidateTweetIDWithError tests tweet ID validation with error return
func TestValidateTweetIDWithError(t *testing.T) {
	// Valid ID
	err := utils.ValidateTweetIDWithError("123456789")
	if err != nil {
		t.Errorf("ValidateTweetIDWithError(valid) = %v, want nil", err)
	}

	// Invalid ID
	err = utils.ValidateTweetIDWithError("abc")
	if err == nil {
		t.Error("ValidateTweetIDWithError(invalid) should return error")
	}
}

// TestValidateTweetText tests tweet text validation
func TestValidateTweetText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{"valid short", "Hello world", true},
		{"valid exact", strings.Repeat("a", 280), true},
		{"too long", strings.Repeat("a", 281), false},
		{"empty", "", false},
		{"unicode", "Hello 世界 🌍", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ValidateTweetText(tt.text)
			if result != tt.expected {
				t.Errorf("ValidateTweetText(%q) = %v, want %v", tt.text, result, tt.expected)
			}
		})
	}
}

// TestValidateTweetTextWithLimit tests text validation with custom limit
func TestValidateTweetTextWithLimit(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		maxLen     int
		wantValid  bool
		wantText   string
	}{
		{"valid", "Hello", 10, true, "Hello"},
		{"exact limit", "Hello", 5, true, "Hello"},
		{"too long", "Hello world", 5, false, ""},
		{"empty", "", 10, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, valid := utils.ValidateTweetTextWithLimit(tt.text, tt.maxLen)
			if valid != tt.wantValid {
				t.Errorf("ValidateTweetTextWithLimit valid = %v, want %v", valid, tt.wantValid)
			}
			if valid && text != tt.wantText {
				t.Errorf("ValidateTweetTextWithLimit text = %q, want %q", text, tt.wantText)
			}
		})
	}
}

// TestValidateTwitterHandle tests handle validation
func TestValidateTwitterHandle(t *testing.T) {
	tests := []struct {
		name          string
		handle        string
		wantValid     bool
		wantNormalized string
	}{
		{"valid simple", "testuser", true, "testuser"},
		{"with @ prefix", "@testuser", true, "testuser"},
		{"with numbers", "user123", true, "user123"},
		{"with underscore", "test_user", true, "test_user"},
		{"max length", "abcdefghijklmno", true, "abcdefghijklmno"}, // 15 chars
		{"too long", "abcdefghijklmnop", false, ""},               // 16 chars
		{"empty", "", false, ""},
		{"with dash", "test-user", false, ""},
		{"with space", "test user", false, ""},
		{"special chars", "test@user", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalized, valid := utils.ValidateTwitterHandle(tt.handle)
			if valid != tt.wantValid {
				t.Errorf("ValidateTwitterHandle(%q) valid = %v, want %v", tt.handle, valid, tt.wantValid)
			}
			if valid && normalized != tt.wantNormalized {
				t.Errorf("ValidateTwitterHandle(%q) normalized = %q, want %q", tt.handle, normalized, tt.wantNormalized)
			}
		})
	}
}

// TestSanitizeInput tests input sanitization
func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"clean text", "Hello world", "Hello world"},
		{"with whitespace", "  Hello world  ", "Hello world"},
		{"with newlines", "Hello\nworld", "Hello\nworld"},
		{"with tabs", "Hello\tworld", "Hello\tworld"},
		{"null bytes", "Hello\x00world", "Helloworld"},
		{"control chars", "Hello\x01\x02world", "Helloworld"},
		{"mixed", "  Hello\n\x00world\t  ", "Hello\nworld"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.SanitizeInput(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeInput(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestSanitizeInputWithMaxLength tests sanitization with max length
func TestSanitizeInputWithMaxLength(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"under limit", "Hello", 10, "Hello"},
		{"exact limit", "Hello", 5, "Hello"},
		{"over limit", "Hello world", 5, "He..."},
		{"unicode", "Hello 世界", 5, "He..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.SanitizeInputWithMaxLength(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("SanitizeInputWithMaxLength(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestSanitizeHandle tests handle sanitization
func TestSanitizeHandle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "TestUser", "testuser"},
		{"with @", "@TestUser", "testuser"},
		{"with spaces", "  TestUser  ", "testuser"},
		{"mixed case", "TeStUsEr", "testuser"},
		{"already lower", "testuser", "testuser"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.SanitizeHandle(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeHandle(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestIsValidURL tests URL validation
func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"http", "http://example.com", true},
		{"https", "https://example.com", true},
		{"with path", "https://example.com/path", true},
		{"no protocol", "example.com", true},
		{"empty", "", false},
		{"just text", "not a url", false},
		{"with spaces", "https://example .com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.IsValidURL(tt.url)
			if got != tt.want {
				t.Errorf("IsValidURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

// TestTruncateText tests text truncation
func TestTruncateText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxLen   int
		expected string
	}{
		{"under limit", "Hello", 10, "Hello"},
		{"exact limit", "Hello", 5, "Hello"},
		{"over limit", "Hello world", 8, "Hello..."},
		{"max 3", "Hello", 3, "Hel"},
		{"empty", "", 10, ""},
		{"unicode", "Hello 世界", 8, "Hello 世界"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.TruncateText(tt.text, tt.maxLen)
			if result != tt.expected {
				t.Errorf("TruncateText(%q, %d) = %q, want %q", tt.text, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestCountTweetCharacters tests character counting
func TestCountTweetCharacters(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{"ascii", "Hello world", 11},
		{"unicode", "Hello 世界", 8}, // 5 + 1 space + 2 chars
		{"emoji", "Hello 🌍", 7},    // 5 + 1 space + 1 emoji
		{"empty", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.CountTweetCharacters(tt.text)
			if result != tt.expected {
				t.Errorf("CountTweetCharacters(%q) = %d, want %d", tt.text, result, tt.expected)
			}
		})
	}
}

// TestIsTweetTooLong tests tweet length check
func TestIsTweetTooLong(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{"short", "Hello", false},
		{"exact 280", strings.Repeat("a", 280), false},
		{"too long", strings.Repeat("a", 281), true},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.IsTweetTooLong(tt.text)
			if result != tt.expected {
				t.Errorf("IsTweetTooLong(%q) = %v, want %v", tt.text, result, tt.expected)
			}
		})
	}
}

// TestExtractTweetIDFromURL tests URL parsing for tweet IDs
func TestExtractTweetIDFromURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		wantID   string
		wantErr  bool
	}{
		{"x.com status", "https://x.com/user/status/123456789", "123456789", false},
		{"twitter.com status", "https://twitter.com/user/status/987654321", "987654321", false},
		{"web status", "https://twitter.com/i/web/status/555555555", "555555555", false},
		{"just ID", "123456789", "123456789", false},
		{"invalid URL", "https://example.com/something", "", true},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := utils.ExtractTweetIDFromURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractTweetIDFromURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
				return
			}
			if id != tt.wantID {
				t.Errorf("ExtractTweetIDFromURL(%q) = %q, want %q", tt.url, id, tt.wantID)
			}
		})
	}
}

// TestNormalizeTweetID tests tweet ID normalization
func TestNormalizeTweetID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantID   string
		wantErr  bool
	}{
		{"valid ID", "123456789", "123456789", false},
		{"URL to ID", "https://x.com/user/status/123456789", "123456789", false},
		{"invalid", "abc", "", true},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := utils.NormalizeTweetID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeTweetID(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if id != tt.wantID {
				t.Errorf("NormalizeTweetID(%q) = %q, want %q", tt.input, id, tt.wantID)
			}
		})
	}
}

// TestCleanSearchQuery tests query cleaning
func TestCleanSearchQuery(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "hello world", "hello world"},
		{"extra spaces", "hello    world", "hello world"},
		{"leading trailing", "  hello world  ", "hello world"},
		{"mixed whitespace", "hello\t\tworld", "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.CleanSearchQuery(tt.input)
			if result != tt.expected {
				t.Errorf("CleanSearchQuery(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestIsEmptyOrWhitespace tests empty check
func TestIsEmptyOrWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"empty", "", true},
		{"spaces", "   ", true},
		{"tabs", "\t\t", true},
		{"newlines", "\n\n", true},
		{"mixed", "  \t\n  ", true},
		{"text", "hello", false},
		{"text with spaces", "  hello  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.IsEmptyOrWhitespace(tt.input)
			if result != tt.expected {
				t.Errorf("IsEmptyOrWhitespace(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestMaxInt tests max function
func TestMaxInt(t *testing.T) {
	tests := []struct {
		a, b     int
		expected int
	}{
		{1, 2, 2},
		{5, 3, 5},
		{0, 0, 0},
		{-1, 1, 1},
		{-5, -10, -5},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := utils.MaxInt(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("MaxInt(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// TestMinInt tests min function
func TestMinInt(t *testing.T) {
	tests := []struct {
		a, b     int
		expected int
	}{
		{1, 2, 1},
		{5, 3, 3},
		{0, 0, 0},
		{-1, 1, -1},
		{-5, -10, -10},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := utils.MinInt(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("MinInt(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
