// Package core provides startup checks for endpoint validity.
package core

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// StartupChecker handles automatic endpoint checks on CLI startup
type StartupChecker struct {
	lastCheckFile string
	checkInterval time.Duration
}

// NewStartupChecker creates a new startup checker
func NewStartupChecker() *StartupChecker {
	paths, err := GetPaths()
	lastCheckFile := ""
	if err == nil {
		lastCheckFile = paths.StartupMarker
	}
	return &StartupChecker{
		lastCheckFile: lastCheckFile,
		checkInterval: 24 * time.Hour, // Check once per day
	}
}

// ShouldCheck returns true if we should perform an endpoint check
func (sc *StartupChecker) ShouldCheck() bool {
	data, err := os.ReadFile(sc.lastCheckFile)
	if err != nil {
		// No last check file, should check
		return true
	}

	var lastCheck time.Time
	if err := lastCheck.UnmarshalText(data); err != nil {
		return true
	}

	return time.Since(lastCheck) > sc.checkInterval
}

// MarkChecked marks that we've done a check
func (sc *StartupChecker) MarkChecked() {
	if sc.lastCheckFile == "" {
		return
	}
	_ = os.MkdirAll(filepath.Dir(sc.lastCheckFile), 0700)
	data, _ := time.Now().MarshalText()
	os.WriteFile(sc.lastCheckFile, data, 0644)
}

// QuickCheckEndpoints performs a quick check of critical endpoints
func (sc *StartupChecker) QuickCheckEndpoints(ctx context.Context, errOut io.Writer) ([]string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if errOut == nil {
		errOut = io.Discard
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	fmt.Fprintln(errOut, "🔍 Checking API endpoints...")

	// Create a temporary client for checking
	client, err := NewXClient(nil, "", "")
	if err != nil {
		return nil, err
	}
	client.authRefreshAttempted = true // Don't refresh credentials during startup check
	defer client.Close()

	obsolete := []string{}

	// Check critical endpoints
	endpointsToCheck := []string{
		"HomeTimeline",
		"UserByScreenName",
		"SearchTimeline",
	}

	for _, operation := range endpointsToCheck {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		fmt.Fprintf(errOut, "   Checking %s... ", operation)

		// Try a minimal request
		if err := sc.checkEndpoint(client, operation); err != nil {
			fmt.Fprintln(errOut, "❌")
			obsolete = append(obsolete, operation)
		} else {
			fmt.Fprintln(errOut, "✓")
		}
	}

	sc.MarkChecked()
	return obsolete, nil
}

// checkEndpoint tries to make a minimal request to check if endpoint works
func (sc *StartupChecker) checkEndpoint(client *XClient, operation string) error {
	// Use GraphQLGet with minimal variables
	variables := map[string]interface{}{
		"count": 1,
	}

	_, err := client.GraphQLGet(operation, variables)

	if err != nil {
		// Check if it's a 404 "Query not found" error
		if apiErr, ok := err.(*APIError); ok && apiErr.StatusCode == 404 {
			// Check the response data for "Query not found"
			if strings.Contains(apiErr.ResponseData, "Query not found") ||
				strings.Contains(apiErr.ResponseData, "Not Found") ||
				strings.Contains(apiErr.Message, "Query not found") {
				return fmt.Errorf("endpoint obsolete: Query not found")
			}
		}
		// For UserByScreenName, we need to check with actual username
		if operation == "UserByScreenName" {
			vars := map[string]interface{}{
				"screen_name": "twitter",
			}
			_, err2 := client.GraphQLGet(operation, vars)
			if err2 != nil {
				if apiErr, ok := err2.(*APIError); ok && apiErr.StatusCode == 404 {
					if strings.Contains(apiErr.ResponseData, "Query not found") ||
						strings.Contains(apiErr.ResponseData, "Not Found") ||
						strings.Contains(apiErr.Message, "Query not found") {
						return fmt.Errorf("endpoint obsolete: Query not found")
					}
				}
			}
		}
		// Other errors (401, 403, 400, etc.) mean the endpoint exists
		// The endpoint is valid even if we get auth errors
		return nil
	}

	return nil
}

// ShowUpdatePrompt shows a prompt to update endpoints if needed
func (sc *StartupChecker) ShowUpdatePrompt(obsoleteEndpoints []string, errOut io.Writer) {
	if len(obsoleteEndpoints) == 0 {
		return
	}
	if errOut == nil {
		errOut = io.Discard
	}

	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAD1F")).Bold(true)

	fmt.Fprintln(errOut)
	fmt.Fprintln(errOut, warningStyle.Render("⚠️  Some API endpoints are obsolete!"))
	fmt.Fprintln(errOut)
	fmt.Fprintf(errOut, "The following endpoints need to be updated: %v\n", obsoleteEndpoints)
	fmt.Fprintln(errOut)
	fmt.Fprintln(errOut, "You can fix this by running:")
	fmt.Fprintln(errOut)
	fmt.Fprintln(errOut, "  1. Automatic update (recommended):")
	fmt.Fprintln(errOut, "     xsh auto-update")
	fmt.Fprintln(errOut)
	fmt.Fprintln(errOut, "  2. Manual update:")
	fmt.Fprintln(errOut, "     xsh endpoints update <operation> <new-id>")
	fmt.Fprintln(errOut)
	fmt.Fprintln(errOut, "  3. Check endpoint status:")
	fmt.Fprintln(errOut, "     xsh endpoints check <operation>")
	fmt.Fprintln(errOut)
}

// RunStartupCheck performs the full startup check workflow
func RunStartupCheck(ctx context.Context, errOut io.Writer) error {
	checker := NewStartupChecker()

	if !checker.ShouldCheck() {
		return nil
	}

	obsolete, err := checker.QuickCheckEndpoints(ctx, errOut)
	if err != nil {
		return err
	}

	if len(obsolete) > 0 {
		checker.ShowUpdatePrompt(obsolete, errOut)
	}
	return nil
}
