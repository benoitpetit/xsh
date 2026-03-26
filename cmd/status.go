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

		// Collect status information
		status := collectSystemStatus(checkNow)

		if showJSON || isJSONMode() || isYAMLMode() {
			output(status, func() {})
			return
		}

		// Display formatted status
		displayStatus(status)
	},
}

// systemStatus holds all status information
type systemStatus struct {
	Authenticated   bool                   `json:"authenticated"`
	Account         string                 `json:"account,omitempty"`
	EndpointHealth  endpointHealth         `json:"endpoint_health"`
	CacheStatus     cacheStatus            `json:"cache_status"`
	Connectivity    connectivityStatus     `json:"connectivity"`
	Timestamp       time.Time              `json:"timestamp"`
}

type endpointHealth struct {
	Healthy        bool     `json:"healthy"`
	TotalEndpoints int      `json:"total_endpoints"`
	FailedEndpoints []string `json:"failed_endpoints,omitempty"`
	Message        string   `json:"message"`
}

type cacheStatus struct {
	Valid        bool          `json:"valid"`
	Age          time.Duration `json:"age"`
	EndpointCount int          `json:"endpoint_count"`
	FeatureCount  int          `json:"feature_count"`
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

	// Check authentication
	creds, err := core.GetCredentials(account)
	if err == nil && creds != nil && creds.IsValid() {
		status.Authenticated = true
		status.Account = creds.AccountName
	}

	// Check endpoint health
	manager := core.GetEndpointManager()
	stats := manager.GetStats()
	
	status.CacheStatus = cacheStatus{
		Valid:         stats.CacheAge < 24*time.Hour,
		Age:           stats.CacheAge,
		EndpointCount: stats.TotalCount,
		FeatureCount:  stats.FeatureCount,
	}

	// Check connectivity
	status.Connectivity = checkConnectivity()

	// Perform endpoint check if requested
	if checkNow {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		healthy, issues := core.CheckEndpointHealth(ctx, nil)
		status.EndpointHealth = endpointHealth{
			Healthy:        healthy,
			TotalEndpoints: stats.TotalCount,
			FailedEndpoints: issues,
			Message:        "Checked just now",
		}
		if !healthy && len(issues) > 0 {
			status.EndpointHealth.Message = fmt.Sprintf("%d issues found", len(issues))
		}
	} else {
		// Use cached health status
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

	// Try to create endpoint discovery
	discovery, err := core.NewEndpointDiscovery(false)
	if err != nil {
		conn.DiscoveryWorks = false
		conn.Message = "Endpoint discovery unavailable"
		return conn
	}

	// Quick check if we can reach X.com
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
	// Header
	fmt.Println()
	fmt.Println(display.StyleBold.Render("XSH System Status"))
	fmt.Println(display.StyleMuted.Render("================="))
	fmt.Println()

	// Authentication
	authStatus := "✓ Authenticated"
	if !s.Authenticated {
		authStatus = "✗ Not authenticated"
	}
	fmt.Printf("%-20s %s\n", "Authentication:", authStatus)
	if s.Authenticated && s.Account != "" {
		fmt.Printf("%-20s %s\n", "  Account:", s.Account)
	}
	fmt.Println()

	// Endpoint Health
	healthStatus := "✓ Healthy"
	if !s.EndpointHealth.Healthy {
		healthStatus = "✗ Issues detected"
	}
	fmt.Printf("%-20s %s\n", "Endpoint Health:", healthStatus)
	fmt.Printf("%-20s %d endpoints\n", "  Total:", s.EndpointHealth.TotalEndpoints)
	if len(s.EndpointHealth.FailedEndpoints) > 0 {
		fmt.Printf("%-20s %v\n", "  Issues:", s.EndpointHealth.FailedEndpoints)
	}
	fmt.Printf("%-20s %s\n", "  Message:", s.EndpointHealth.Message)
	fmt.Println()

	// Cache Status
	cacheStatus := "✓ Valid"
	if !s.CacheStatus.Valid {
		cacheStatus = "✗ Expired"
	}
	fmt.Printf("%-20s %s\n", "Cache:", cacheStatus)
	fmt.Printf("%-20s %s\n", "  Age:", s.CacheStatus.Age.Round(time.Second))
	fmt.Printf("%-20s %d\n", "  Endpoints:", s.CacheStatus.EndpointCount)
	fmt.Printf("%-20s %d\n", "  Features:", s.CacheStatus.FeatureCount)
	fmt.Println()

	// Connectivity
	connStatus := "✓ OK"
	if !s.Connectivity.CanReachX || !s.Connectivity.DiscoveryWorks {
		connStatus = "✗ Issues"
	}
	fmt.Printf("%-20s %s\n", "Connectivity:", connStatus)
	if s.Connectivity.Message != "" {
		fmt.Printf("%-20s %s\n", "  Message:", s.Connectivity.Message)
	}
	fmt.Println()

	// Timestamp
	fmt.Printf("%-20s %s\n", "Last Updated:", s.Timestamp.Format("15:04:05"))
	fmt.Println()

	// Recommendations
	if !s.Authenticated {
		fmt.Println(display.PrintWarning("Run 'xsh auth login' to authenticate"))
	}
	if !s.CacheStatus.Valid {
		fmt.Println(display.PrintWarning("Run 'xsh endpoints refresh' to update endpoints"))
	}
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().Bool("check", false, "Perform fresh endpoint health check")
	statusCmd.Flags().Bool("json", false, "Output as JSON")
}
