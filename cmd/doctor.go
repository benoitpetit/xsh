// Package cmd provides the doctor command for xsh.
package cmd

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/spf13/cobra"
)

// doctorCmd runs diagnostics
type CheckResult struct {
	Name        string `json:"name"`
	Status      string `json:"status"` // "pass", "fail", "warn"
	Detail      string `json:"detail"`
	Remediation string `json:"remediation,omitempty"`
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run diagnostics",
	Long:  `Check the health of your xsh installation and configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		jsonOutput := isJSONMode()

		var checks []CheckResult

		checks = append(checks, checkAuth())
		checks = append(checks, checkEndpoints())
		checks = append(checks, checkNetwork())
		checks = append(checks, checkTLS())
		checks = append(checks, checkSystem())

		if jsonOutput || isYAMLMode() {
			output(checks, func() {
				printChecks(checks)
			})
		} else {
			printChecks(checks)
		}

		for _, c := range checks {
			if c.Status == "fail" && c.Name == "Auth" {
				os.Exit(core.ExitAuthError)
			}
			if c.Status == "fail" && c.Name == "Network" {
				os.Exit(core.ExitError)
			}
		}
	},
}

func checkAuth() CheckResult {
	creds, err := core.GetCredentials("")
	if err != nil {
		return CheckResult{
			Name:        "Auth",
			Status:      "fail",
			Detail:      err.Error(),
			Remediation: "xsh auth login",
		}
	}

	if !creds.IsValid() {
		return CheckResult{
			Name:        "Auth",
			Status:      "fail",
			Detail:      "Invalid credentials",
			Remediation: "xsh auth login",
		}
	}

	tokenHint := creds.AuthToken
	if len(tokenHint) > 8 {
		tokenHint = tokenHint[:8]
	}
	return CheckResult{
		Name:   "Auth",
		Status: "pass",
		Detail: fmt.Sprintf("Valid credentials (token: %s...)", tokenHint),
	}
}

func checkEndpoints() CheckResult {
	return endpointCheckResult(core.GetEndpointManager().GetStats())
}

func endpointCheckResult(stats core.EndpointStats) CheckResult {

	if stats.TotalCount == 0 {
		return CheckResult{
			Name:        "Endpoints",
			Status:      "warn",
			Detail:      "No endpoints available",
			Remediation: "xsh endpoints refresh",
		}
	}

	if stats.DynamicCount == 0 {
		return CheckResult{
			Name:        "Endpoints",
			Status:      "warn",
			Detail:      fmt.Sprintf("Using %d static fallback endpoints; dynamic discovery has not completed", stats.StaticCount),
			Remediation: "xsh endpoints refresh",
		}
	}

	if stats.DynamicCount < 10 {
		return CheckResult{
			Name:        "Endpoints",
			Status:      "warn",
			Detail:      fmt.Sprintf("Only %d dynamically discovered endpoints available", stats.DynamicCount),
			Remediation: "xsh endpoints refresh",
		}
	}

	return CheckResult{
		Name:   "Endpoints",
		Status: "pass",
		Detail: fmt.Sprintf("%d dynamic endpoints available (%d total with fallbacks)", stats.DynamicCount, stats.TotalCount),
	}
}

func checkNetwork() CheckResult {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://x.com")
	if err != nil {
		return CheckResult{
			Name:        "Network",
			Status:      "fail",
			Detail:      fmt.Sprintf("Cannot reach x.com: %v", err),
			Remediation: "verify DNS/network access, then run: xsh doctor",
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
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return CheckResult{
			Name:        "TLS",
			Status:      "warn",
			Detail:      fmt.Sprintf("TLS client initialization: %v", err),
			Remediation: "check proxy/TLS configuration, then run: xsh doctor",
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
	fmt.Println()
	fmt.Println(display.Title("Diagnostics"))
	fmt.Println(display.Separator(40))
	fmt.Println()

	headers := []string{"Check", "Status", "Detail"}
	var rows []display.TableRow
	for _, c := range checks {
		status := display.StatusBadge(c.Status)
		rows = append(rows, display.TableRow{c.Name, status, c.Detail})
	}
	fmt.Println(display.SimpleTable(headers, rows))

	fmt.Println()

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

	for _, c := range checks {
		if c.Remediation != "" && c.Status != "pass" {
			fmt.Println(display.Info(fmt.Sprintf("%s fix: %s", c.Name, c.Remediation)))
		}
	}
	if warn > 0 || fail > 0 {
		fmt.Println()
	}

	if fail > 0 {
		fmt.Println(display.Error(fmt.Sprintf("Found %d issue(s)", fail)))
	} else if warn > 0 {
		fmt.Println(display.Warning(fmt.Sprintf("Completed with %d warning(s)", warn)))
	} else {
		fmt.Println(display.Success("All checks passed!"))
	}
	fmt.Println()
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
