package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
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
	Run: func(cmd *cobra.Command, args []string) {
		showJSON, _ := cmd.Flags().GetBool("json")
		checkNow, _ := cmd.Flags().GetBool("check")

		status := collectSystemStatus(checkNow)

		if showJSON || isJSONMode() || isYAMLMode() {
			output(status, func() {})
			return
		}

		displayStatus(status)
	},
}

// systemStatus holds all status information
type systemStatus struct {
	Authenticated   bool               `json:"authenticated"`
	Account         string             `json:"account,omitempty"`
	EndpointHealth  endpointHealth     `json:"endpoint_health"`
	CacheStatus     cacheStatus        `json:"cache_status"`
	Connectivity    connectivityStatus `json:"connectivity"`
	Timestamp       time.Time          `json:"timestamp"`
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
	CanReachX      bool   `json:"can_reach_x"`
	DiscoveryWorks bool   `json:"discovery_works"`
	Message        string `json:"message,omitempty"`
}

func collectSystemStatus(checkNow bool) *systemStatus {
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
		Valid:         stats.CacheAge < 24*time.Hour,
		Age:           stats.CacheAge,
		EndpointCount: stats.TotalCount,
		FeatureCount:  stats.FeatureCount,
	}

	status.Connectivity = checkConnectivity()

	if checkNow {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		healthy, issues := core.CheckEndpointHealth(ctx, nil)
		status.EndpointHealth = endpointHealth{
			Healthy:         healthy,
			TotalEndpoints:  stats.TotalCount,
			FailedEndpoints: issues,
			Message:         "Checked just now",
		}
		if !healthy && len(issues) > 0 {
			status.EndpointHealth.Message = fmt.Sprintf("%d issues found", len(issues))
		}
	} else {
		status.EndpointHealth = endpointHealth{
			Healthy:        status.CacheStatus.Valid,
			TotalEndpoints: stats.TotalCount,
			Message:        "Using cached status (use --check for fresh check)",
		}
	}

	return status
}

func checkConnectivity() connectivityStatus {
	conn := connectivityStatus{
		CanReachX:      true,
		DiscoveryWorks: true,
	}

	discovery, err := core.NewEndpointDiscovery(false)
	if err != nil {
		conn.DiscoveryWorks = false
		conn.Message = "Endpoint discovery unavailable"
		return conn
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = discovery.GetCachedEndpoints(ctx)
	if err != nil {
		conn.DiscoveryWorks = false
		conn.Message = fmt.Sprintf("Cannot reach X.com: %v", err)
	}

	return conn
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
	fmt.Println(display.KeyValue("Age:", s.CacheStatus.Age.Round(time.Second).String()))
	fmt.Println(display.KeyValue("Endpoints:", fmt.Sprintf("%d", s.CacheStatus.EndpointCount)))
	fmt.Println(display.KeyValue("Features:", fmt.Sprintf("%d", s.CacheStatus.FeatureCount)))
	fmt.Println()

	// Connectivity
	fmt.Println(display.Section("Connectivity"))
	if s.Connectivity.CanReachX && s.Connectivity.DiscoveryWorks {
		fmt.Println(display.KeyValue("Status:", display.Success("OK")))
	} else {
		fmt.Println(display.KeyValue("Status:", display.Error("Issues")))
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
	statusCmd.Flags().Bool("check", false, "Perform fresh endpoint health check")
	statusCmd.Flags().Bool("json", false, "Output as JSON")
}
