package cmd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/spf13/cobra"
)

// statusCmd shows system status including endpoint monitoring
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show system status and endpoint health",
	Long: `Display the current status of xsh including:
- Authentication status
- Endpoint system health
- API connectivity
- Cache status`,
	RunE: func(cmd *cobra.Command, args []string) error {
		showJSON, _ := cmd.Flags().GetBool("json")
		checkNow, _ := cmd.Flags().GetBool("check")
		localOnly, _ := cmd.Flags().GetBool("local")

		status := collectSystemStatus(checkNow, !localOnly)
		return outputSystemStatus(status, showJSON)
	},
}

// systemStatus holds all status information
type systemStatus struct {
	Authenticated  bool               `json:"authenticated"`
	Account        string             `json:"account,omitempty"`
	EndpointHealth endpointHealth     `json:"endpoint_health"`
	CacheStatus    cacheStatus        `json:"cache_status"`
	Connectivity   connectivityStatus `json:"connectivity"`
	Timestamp      time.Time          `json:"timestamp"`
}

type endpointHealth struct {
	Healthy         bool     `json:"healthy"`
	TotalEndpoints  int      `json:"total_endpoints"`
	FailedEndpoints []string `json:"failed_endpoints,omitempty"`
	Message         string   `json:"message"`
}

type cacheStatus struct {
	Valid         bool          `json:"valid"`
	Age           time.Duration `json:"age"`
	EndpointCount int           `json:"endpoint_count"`
	FeatureCount  int           `json:"feature_count"`
}

type connectivityStatus struct {
	Checked          bool   `json:"checked"`
	CanReachX        bool   `json:"can_reach_x"`
	DiscoveryChecked bool   `json:"discovery_checked"`
	DiscoveryWorks   bool   `json:"discovery_works"`
	Message          string `json:"message,omitempty"`
}

type statusChecks struct {
	connectivity   func(context.Context) connectivityStatus
	endpointHealth func(context.Context) (bool, []string)
}

func collectSystemStatus(checkNow bool, allowNetwork bool) *systemStatus {
	return collectSystemStatusWithChecks(checkNow, allowNetwork, statusChecks{
		connectivity: checkConnectivity,
		endpointHealth: func(ctx context.Context) (bool, []string) {
			return core.CheckEndpointHealth(ctx, nil)
		},
	})
}

func collectSystemStatusWithChecks(checkNow bool, allowNetwork bool, checks statusChecks) *systemStatus {
	status := &systemStatus{
		Timestamp: time.Now(),
	}

	creds, err := core.GetCredentials(account)
	if err == nil && creds != nil && creds.IsValid() {
		status.Authenticated = true
		status.Account = creds.AccountName
	}

	manager := core.GetEndpointManager()
	stats := manager.GetStats()

	status.CacheStatus = cacheStatus{
		Valid:         stats.DynamicCount > 0 && stats.CacheAge < 24*time.Hour,
		Age:           stats.CacheAge,
		EndpointCount: stats.DynamicCount,
		FeatureCount:  stats.FeatureCount,
	}

	if allowNetwork {
		ctx, cancel := context.WithTimeout(runtimeContext(), 10*time.Second)
		status.Connectivity = checks.connectivity(ctx)
		cancel()
	} else {
		status.Connectivity.Message = "Skipped (--local)"
	}

	if checkNow && allowNetwork {
		ctx, cancel := context.WithTimeout(runtimeContext(), 30*time.Second)
		defer cancel()

		healthy, issues := checks.endpointHealth(ctx)
		status.EndpointHealth = endpointHealth{
			Healthy:         healthy,
			TotalEndpoints:  stats.TotalCount,
			FailedEndpoints: issues,
			Message:         "Endpoint metadata validated",
		}
		if !healthy && len(issues) > 0 {
			status.EndpointHealth.Message = fmt.Sprintf("%d endpoint metadata issues found", len(issues))
		}
	} else if checkNow {
		status.EndpointHealth = endpointHealth{
			Healthy:        status.CacheStatus.Valid,
			TotalEndpoints: stats.TotalCount,
			Message:        "Endpoint check skipped (--local); showing local cache status",
		}
	} else {
		status.EndpointHealth = endpointHealth{
			Healthy:        status.CacheStatus.Valid,
			TotalEndpoints: stats.TotalCount,
			Message:        "Using cached endpoint metadata (use --check to validate it)",
		}
	}

	return status
}

func checkConnectivity(ctx context.Context) connectivityStatus {
	return probeConnectivity(ctx, http.DefaultClient, "https://x.com/")
}

func probeConnectivity(ctx context.Context, client *http.Client, target string) connectivityStatus {
	status := connectivityStatus{Checked: true}
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, target, nil)
	if err != nil {
		status.Message = fmt.Sprintf("Could not create X.com connectivity request: %v", err)
		return status
	}
	response, err := client.Do(req)
	if err != nil {
		if isNetworkProbeError(err) {
			status.Message = fmt.Sprintf("Cannot reach X.com: %v", err)
		} else {
			status.Message = fmt.Sprintf("X.com connectivity probe failed: %v", err)
		}
		return status
	}
	defer response.Body.Close()

	status.CanReachX = true
	status.Message = "X.com reachable; endpoint discovery was not checked"
	if response.StatusCode >= http.StatusInternalServerError {
		status.Message = fmt.Sprintf("X.com reachable but returned HTTP %d; endpoint discovery was not checked", response.StatusCode)
	}
	return status
}

func isNetworkProbeError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var operationErr *net.OpError
	var dnsErr *net.DNSError
	return errors.As(err, &operationErr) || errors.As(err, &dnsErr)
}

func connectivityStatusLabel(status connectivityStatus) string {
	if !status.Checked {
		return "Not checked"
	}
	if !status.CanReachX || (status.DiscoveryChecked && !status.DiscoveryWorks) {
		return "Issues"
	}
	if status.DiscoveryChecked {
		return "OK"
	}
	return "Reachable"
}

func outputSystemStatus(status *systemStatus, showJSON bool) error {
	if showJSON {
		return outputJSON(status)
	}
	if isJSONMode() || isYAMLMode() || isCompactMode() {
		return output(status, func() {})
	}
	displayStatus(status)
	return nil
}

func displayStatus(s *systemStatus) {
	fmt.Println()
	fmt.Println(display.Title("XSH System Status"))
	fmt.Println(display.Separator(40))
	fmt.Println()

	// Authentication
	fmt.Println(display.Section("Authentication"))
	if s.Authenticated {
		fmt.Println(display.KeyValue("Status:", display.Success("Authenticated")))
		if s.Account != "" {
			fmt.Println(display.KeyValue("Account:", s.Account))
		}
	} else {
		fmt.Println(display.KeyValue("Status:", display.Error("Not authenticated")))
	}
	fmt.Println()

	// Endpoint Health
	fmt.Println(display.Section("Endpoint Health"))
	if s.EndpointHealth.Healthy {
		fmt.Println(display.KeyValue("Status:", display.Success("Healthy")))
	} else {
		fmt.Println(display.KeyValue("Status:", display.Error("Issues detected")))
	}
	fmt.Println(display.KeyValue("Total:", fmt.Sprintf("%d endpoints", s.EndpointHealth.TotalEndpoints)))
	if len(s.EndpointHealth.FailedEndpoints) > 0 {
		fmt.Println(display.KeyValue("Issues:", fmt.Sprintf("%v", s.EndpointHealth.FailedEndpoints)))
	}
	fmt.Println(display.KeyValue("Message:", s.EndpointHealth.Message))
	fmt.Println()

	// Cache Status
	fmt.Println(display.Section("Cache Status"))
	if s.CacheStatus.Valid {
		fmt.Println(display.KeyValue("Status:", display.Success("Valid")))
	} else {
		fmt.Println(display.KeyValue("Status:", display.Error("Expired")))
	}
	if s.CacheStatus.Age > 0 {
		fmt.Println(display.KeyValue("Age:", s.CacheStatus.Age.Round(time.Second).String()))
	} else {
		fmt.Println(display.KeyValue("Age:", "never"))
	}
	fmt.Println(display.KeyValue("Endpoints:", fmt.Sprintf("%d", s.CacheStatus.EndpointCount)))
	fmt.Println(display.KeyValue("Features:", fmt.Sprintf("%d", s.CacheStatus.FeatureCount)))
	fmt.Println()

	// Connectivity
	fmt.Println(display.Section("Connectivity"))
	switch label := connectivityStatusLabel(s.Connectivity); label {
	case "Not checked":
		fmt.Println(display.KeyValue("Status:", display.Muted(label)))
	case "OK", "Reachable":
		fmt.Println(display.KeyValue("Status:", display.Success(label)))
	default:
		fmt.Println(display.KeyValue("Status:", display.Error(label)))
	}
	if s.Connectivity.Message != "" {
		fmt.Println(display.KeyValue("Message:", s.Connectivity.Message))
	}
	fmt.Println()

	// Timestamp
	fmt.Println(display.KeyValue("Last Updated:", s.Timestamp.Format("15:04:05")))
	fmt.Println()

	// Recommendations
	if !s.Authenticated {
		fmt.Println(display.Warning("Run 'xsh auth login' to authenticate"))
	}
	if !s.CacheStatus.Valid {
		fmt.Println(display.Warning("Run 'xsh endpoints refresh' to update endpoints"))
	}
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().Bool("check", false, "Validate cached endpoint freshness and required operations (may fetch discovery if cache is unavailable)")
	statusCmd.Flags().Bool("local", false, "Use local cache only; skip connectivity and endpoint checks")
	statusCmd.Flags().Bool("json", false, "Output as JSON")
}
