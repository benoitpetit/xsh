// Package utils provides helper utilities.
package utils

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	// MaxTweetLength is the maximum character length for a tweet
	MaxTweetLength = 280
	// MaxHandleLength is the maximum length for a Twitter handle
	MaxHandleLength = 15
)

var (
	// Regex for tweet ID (numeric)
	tweetIDRegex = regexp.MustCompile(`^\d+$`)
	// Regex for Twitter handle (alphanumeric and underscore)
	handleRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

// ValidateTweetID checks if a string is a valid tweet ID (numeric)
func ValidateTweetID(id string) bool {
	if id == "" {
		return false
	}
	return tweetIDRegex.MatchString(id)
}

// ValidateTweetIDWithError validates a tweet ID and returns an error if invalid
func ValidateTweetIDWithError(id string) error {
	if !ValidateTweetID(id) {
		return fmt.Errorf("invalid tweet ID: %s", id)
	}
	return nil
}

// ValidateTweetText checks if text is valid for a tweet
func ValidateTweetText(text string) bool {
	if text == "" {
		return false
	}
	return CountTweetCharacters(text) <= MaxTweetLength
}

// ValidateTweetTextWithLimit validates text with a custom length limit
func ValidateTweetTextWithLimit(text string, maxLen int) (string, bool) {
	if text == "" {
		return "", false
	}
	if CountTweetCharacters(text) > maxLen {
		return "", false
	}
	return text, true
}

// ValidateTwitterHandle validates and normalizes a Twitter handle
func ValidateTwitterHandle(handle string) (string, bool) {
	// Remove @ prefix and whitespace
	handle = strings.TrimPrefix(strings.TrimSpace(handle), "@")

	if handle == "" || len(handle) > MaxHandleLength {
		return "", false
	}

	// Check if handle matches allowed characters
	if !handleRegex.MatchString(handle) {
		return "", false
	}

	return handle, true
}

// SanitizeInput removes control characters from input
func SanitizeInput(input string) string {
	input = strings.TrimSpace(input)
	var result strings.Builder
	for _, r := range input {
		if r != '\x00' && (r >= 32 || r == '\n' || r == '\t' || r == '\r') {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// SanitizeInputWithMaxLength sanitizes input and truncates if too long
func SanitizeInputWithMaxLength(input string, maxLen int) string {
	input = SanitizeInput(input)
	if CountTweetCharacters(input) <= maxLen {
		return input
	}
	return TruncateText(input, maxLen)
}

// SanitizeHandle normalizes a Twitter handle
func SanitizeHandle(handle string) string {
	handle = strings.TrimSpace(handle)
	handle = strings.TrimPrefix(handle, "@")
	handle = strings.ToLower(handle)
	return handle
}

// IsValidURL checks if a string is a valid URL
func IsValidURL(str string) bool {
	// Try parsing as URL with scheme
	u, err := url.Parse(str)
	if err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" {
		return true
	}
	// Try adding https:// prefix
	if !strings.Contains(str, "://") {
		u, err = url.Parse("https://" + str)
		if err == nil && u.Host != "" {
			return true
		}
	}
	return false
}

// TruncateText truncates text to max length with ellipsis
func TruncateText(text string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	length := CountTweetCharacters(text)
	if length <= maxLen {
		return text
	}
	// When maxLen is too small for ellipsis, just hard-truncate
	if maxLen <= 3 {
		runes := []rune(text)
		return string(runes[:maxLen])
	}
	// Account for "..." in count
	effectiveLen := maxLen - 3
	// Truncate rune by rune
	var result strings.Builder
	count := 0
	for _, r := range text {
		if count >= effectiveLen {
			break
		}
		result.WriteRune(r)
		count++
	}
	result.WriteString("...")
	return result.String()
}

// CountTweetCharacters counts characters in a string (handles Unicode properly)
func CountTweetCharacters(text string) int {
	return utf8.RuneCountInString(text)
}

// IsTweetTooLong checks if a tweet exceeds maximum length
func IsTweetTooLong(text string) bool {
	return CountTweetCharacters(text) > MaxTweetLength
}

// ExtractTweetIDFromURL extracts tweet ID from a Twitter/X URL
func ExtractTweetIDFromURL(urlStr string) (string, error) {
	// If it's just an ID
	if ValidateTweetID(urlStr) {
		return urlStr, nil
	}

	// Parse URL
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	// Check if it's a Twitter/X URL
	host := strings.ToLower(u.Host)
	if !strings.Contains(host, "twitter.com") && !strings.Contains(host, "x.com") {
		return "", fmt.Errorf("not a valid Twitter/X URL")
	}

	// Extract ID from path
	path := strings.Trim(u.Path, "/")
	parts := strings.Split(path, "/")

	// Handle /user/status/ID pattern
	if len(parts) >= 3 && (parts[1] == "status" || parts[2] == "status") {
		for _, part := range parts {
			if ValidateTweetID(part) {
				return part, nil
			}
		}
	}

	// Handle /i/web/status/ID pattern
	if len(parts) >= 4 && parts[1] == "i" && parts[2] == "web" && parts[3] == "status" {
		if len(parts) >= 5 && ValidateTweetID(parts[4]) {
			return parts[4], nil
		}
	}

	return "", fmt.Errorf("could not extract tweet ID from URL")
}

// NormalizeTweetID normalizes various forms of tweet IDs to standard numeric ID
func NormalizeTweetID(input string) (string, error) {
	// Try extracting from URL first
	id, err := ExtractTweetIDFromURL(input)
	if err == nil {
		return id, nil
	}

	// Try validating directly
	if ValidateTweetID(input) {
		return input, nil
	}

	return "", fmt.Errorf("invalid tweet ID format: %s", input)
}

// CleanSearchQuery normalizes a search query
func CleanSearchQuery(query string) string {
	// Replace multiple spaces/tabs with single space
	query = regexp.MustCompile(`[\s]+`).ReplaceAllString(query, " ")
	return strings.TrimSpace(query)
}

// IsEmptyOrWhitespace checks if string is empty or contains only whitespace
func IsEmptyOrWhitespace(s string) bool {
	return strings.TrimSpace(s) == ""
}

// MaxInt returns the maximum of two integers
func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MinInt returns the minimum of two integers
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ParseFloatFromInterface safely parses a float from interface{}
func ParseFloatFromInterface(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return 0
}

// ParseIntFromInterface safely parses an int from interface{}
func ParseIntFromInterface(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case string:
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return 0
}

// ParseStringFromInterface safely gets a string from interface{}
func ParseStringFromInterface(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
