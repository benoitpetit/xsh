// Package core provides startup checks for endpoint validity.
package core

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	client, err := NewEndpointProbeClient("")
	if err != nil {
		return nil, err
	}
	defer client.Close()

	obsolete := []string{}

	// Check critical endpoints
	manager := GetEndpointManager()
	endpointsToCheck := criticalEndpointsForCache(manager.getCache())
	confirmed := 0

	for _, operation := range endpointsToCheck {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		fmt.Fprintf(errOut, "   Checking %s... ", operation)

		probe := ProbeGraphQLEndpoint(ctx, client, operation, manager.GetEndpoint(operation), manager.GetOpFeatures(operation))
		if probe.State == "obsolete" {
			fmt.Fprintln(errOut, "❌")
			obsolete = append(obsolete, operation)
		} else if probe.State == "healthy" {
			fmt.Fprintln(errOut, "✓")
			confirmed++
		} else {
			fmt.Fprintf(errOut, "? (%s)\n", probe.State)
		}
	}

	if confirmed > 0 {
		sc.MarkChecked()
	}
	return obsolete, nil
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
