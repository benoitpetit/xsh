// Package cmd provides the doctor command for xsh.
package cmd

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/benoitpetit/xsh/core"
	"github.com/spf13/cobra"
)

// doctorCmd runs diagnostics
type CheckResult struct {
	Name   string `json:"name"`
	Status string `json:"status"` // "pass", "fail", "warn"
	Detail string `json:"detail"`
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run diagnostics",
	Long:  `Check the health of your xsh installation and configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		jsonOutput := isJSONMode()

		var checks []CheckResult

		// Check 1: Authentication
		checks = append(checks, checkAuth())

		// Check 2: Endpoints cache
		checks = append(checks, checkEndpoints())

		// Check 3: Network connectivity
		checks = append(checks, checkNetwork())

		// Check 4: TLS/HTTP client
		checks = append(checks, checkTLS())

		// Check 5: System info
		checks = append(checks, checkSystem())

		if jsonOutput || isYAMLMode() {
			output(checks, func() {
				printChecks(checks)
			})
		} else {
			printChecks(checks)
		}

		// Exit with error if any critical checks failed
		for _, c := range checks {
			if c.Status == "fail" && c.Name == "Auth" {
				os.Exit(core.ExitAuthError)
			}
			if c.Status == "fail" && c.Name == "Network" {
				os.Exit(1)
			}
		}
	},
}

func checkAuth() CheckResult {
	creds, err := core.GetCredentials("")
	if err != nil {
		return CheckResult{
			Name:   "Auth",
			Status: "fail",
			Detail: err.Error(),
		}
	}

	if !creds.IsValid() {
		return CheckResult{
			Name:   "Auth",
			Status: "fail",
			Detail: "Invalid credentials",
		}
	}

	return CheckResult{
		Name:   "Auth",
		Status: "pass",
		Detail: fmt.Sprintf("Valid credentials (token: %s...)", creds.AuthToken[:8]),
	}
}

func checkEndpoints() CheckResult {
	endpoints := core.GetGraphQLEndpoints()
	count := len(endpoints)

	if count == 0 {
		return CheckResult{
			Name:   "Endpoints",
			Status: "warn",
			Detail: "No cached endpoints found",
		}
	}

	if count < 10 {
		return CheckResult{
			Name:   "Endpoints",
			Status: "warn",
			Detail: fmt.Sprintf("Only %d endpoints cached", count),
		}
	}

	return CheckResult{
		Name:   "Endpoints",
		Status: "pass",
		Detail: fmt.Sprintf("%d endpoints available", count),
	}
}

func checkNetwork() CheckResult {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://x.com")
	if err != nil {
		return CheckResult{
			Name:   "Network",
			Status: "fail",
			Detail: fmt.Sprintf("Cannot reach x.com: %v", err),
		}
	}
	defer resp.Body.Close()

	return CheckResult{
		Name:   "Network",
		Status: "pass",
		Detail: "Can reach x.com",
	}
}

func checkTLS() CheckResult {
	// Try to create a uTLS client
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return CheckResult{
			Name:   "TLS",
			Status: "warn",
			Detail: fmt.Sprintf("TLS client initialization: %v", err),
		}
	}
	defer client.Close()

	return CheckResult{
		Name:   "TLS",
		Status: "pass",
		Detail: "TLS fingerprinting available",
	}
}

func checkSystem() CheckResult {
	return CheckResult{
		Name:   "System",
		Status: "pass",
		Detail: fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

func printChecks(checks []CheckResult) {
	fmt.Println("Running diagnostics...")
	fmt.Println()

	for _, c := range checks {
		symbol := "✓"
		if c.Status == "fail" {
			symbol = "✗"
		} else if c.Status == "warn" {
			symbol = "⚠"
		}

		fmt.Printf("%s %s: %s\n", symbol, c.Name, c.Detail)
	}

	fmt.Println()

	// Summary
	pass := 0
	fail := 0
	warn := 0
	for _, c := range checks {
		switch c.Status {
		case "pass":
			pass++
		case "fail":
			fail++
		case "warn":
			warn++
		}
	}

	if fail > 0 {
		fmt.Printf("Found %d issue(s)\n", fail)
	} else if warn > 0 {
		fmt.Printf("All checks passed with %d warning(s)\n", warn)
	} else {
		fmt.Println("All checks passed!")
	}
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
