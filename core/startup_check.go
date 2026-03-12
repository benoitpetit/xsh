// Package core provides startup checks for endpoint validity.
package core

import (
	"fmt"
	"os"
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
	home, _ := os.UserHomeDir()
	return &StartupChecker{
		lastCheckFile: home + "/.config/xsh/.last_endpoint_check",
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
	os.MkdirAll(os.ExpandEnv("$HOME/.config/xsh"), 0755)
	data, _ := time.Now().MarshalText()
	os.WriteFile(sc.lastCheckFile, data, 0644)
}

// QuickCheckEndpoints performs a quick check of critical endpoints
func (sc *StartupChecker) QuickCheckEndpoints() ([]string, error) {
	fmt.Println("🔍 Checking API endpoints...")
	
	// Create a temporary client for checking
	client, err := NewXClient(nil, "", "")
	if err != nil {
		return nil, err
	}
	defer client.Close()

	obsolete := []string{}
	
	// Check critical endpoints
	endpointsToCheck := []string{
		"HomeTimeline",
		"UserByScreenName",
		"SearchTimeline",
	}

	for _, operation := range endpointsToCheck {
		fmt.Printf("   Checking %s... ", operation)
		
		// Try a minimal request
		if err := sc.checkEndpoint(client, operation); err != nil {
			fmt.Printf("❌\n")
			obsolete = append(obsolete, operation)
		} else {
			fmt.Printf("✓\n")
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
func (sc *StartupChecker) ShowUpdatePrompt(obsoleteEndpoints []string) {
	if len(obsoleteEndpoints) == 0 {
		return
	}

	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAD1F")).Bold(true)
	
	fmt.Println()
	fmt.Println(warningStyle.Render("⚠️  Some API endpoints are obsolete!"))
	fmt.Println()
	fmt.Printf("The following endpoints need to be updated: %v\n", obsoleteEndpoints)
	fmt.Println()
	fmt.Println("You can fix this by running:")
	fmt.Println()
	fmt.Println("  1. Automatic update (recommended):")
	fmt.Println("     xsh auto-update")
	fmt.Println()
	fmt.Println("  2. Manual update:")
	fmt.Println("     xsh endpoints update <operation> <new-id>")
	fmt.Println()
	fmt.Println("  3. Check endpoint status:")
	fmt.Println("     xsh endpoints check <operation>")
	fmt.Println()
}

// RunStartupCheck performs the full startup check workflow
func RunStartupCheck() {
	checker := NewStartupChecker()
	
	if !checker.ShouldCheck() {
		return
	}

	obsolete, err := checker.QuickCheckEndpoints()
	if err != nil {
		// Silently fail - don't block CLI usage
		return
	}

	if len(obsolete) > 0 {
		checker.ShowUpdatePrompt(obsolete)
	}
}
