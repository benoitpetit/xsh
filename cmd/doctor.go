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

	if fail > 0 {
		fmt.Println(display.Error(fmt.Sprintf("Found %d issue(s)", fail)))
	} else if warn > 0 {
		fmt.Println(display.Warning(fmt.Sprintf("All checks passed with %d warning(s)", warn)))
	} else {
		fmt.Println(display.Success("All checks passed!"))
	}
	fmt.Println()
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
