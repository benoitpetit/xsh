package cmd

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// endpointsCmd represents the endpoints command
var endpointsCmd = &cobra.Command{
	Use:   "endpoints",
	Short: "Manage GraphQL endpoints",
	Long: `View and manage Twitter/X GraphQL API endpoints.

The CLI automatically discovers and caches endpoints from X.com.
Use these commands to check status, refresh, or troubleshoot endpoints.`,
}

// endpointsListCmd lists all endpoints
var endpointsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all GraphQL endpoints",
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := core.GetEndpointManager()
		endpoints := manager.ListEndpoints()
		stats := manager.GetStats()

		if isJSONMode() || isYAMLMode() {
			output(map[string]interface{}{
				"endpoints": endpoints,
				"stats":     stats,
			}, func() {})
			return nil
		}

		fmt.Println(display.Success(fmt.Sprintf("Found %d endpoints", len(endpoints))))
		fmt.Println(display.Muted(fmt.Sprintf("Last updated: %s ago", time.Since(stats.LastUpdated).Round(time.Second))))
		fmt.Println()

		categories := map[string][]string{
			"Timeline":  {"HomeTimeline", "HomeLatestTimeline"},
			"Search":    {"SearchTimeline"},
			"Tweets":    {"TweetDetail", "TweetResultByRestId", "TweetResultsByRestIds"},
			"Users":     {"UserByScreenName", "UserByRestId", "UserTweets", "UserTweetsAndReplies", "UserMedia"},
			"Social":    {"Followers", "Following", "Likes"},
			"Bookmarks": {"Bookmarks", "BookmarkSearchTimeline"},
			"Write":     {"CreateTweet", "DeleteTweet", "FavoriteTweet", "UnfavoriteTweet", "CreateRetweet", "DeleteRetweet", "CreateBookmark", "DeleteBookmark"},
		}
		shown := make(map[string]bool, len(endpoints))
		for _, ops := range categories {
			for _, op := range ops {
				shown[op] = true
			}
		}
		for op := range endpoints {
			if !shown[op] {
				categories["Other"] = append(categories["Other"], op)
			}
		}
		sort.Strings(categories["Other"])
		categoryNames := make([]string, 0, len(categories))
		for name := range categories {
			categoryNames = append(categoryNames, name)
		}
		sort.Strings(categoryNames)

		for _, category := range categoryNames {
			ops := categories[category]
			fmt.Println(display.Subtitle(category))
			for _, op := range ops {
				if endpoint, ok := endpoints[op]; ok {
					isDynamic, status := manager.CheckEndpoint(op)
					indicator := display.Muted("●")
					if isDynamic {
						indicator = display.Primary("◉")
					}
					fmt.Println("  " + indicator + " " + lipgloss.NewStyle().Width(25).Render(op) + " " + display.Muted("→") + " " + endpoint + " (" + display.StatusBadge(status) + ")")
				}
			}
			fmt.Println()
		}

		return nil
	},
}

// endpointsCheckCmd checks a specific endpoint
var endpointsCheckCmd = &cobra.Command{
	Use:   "check [operation]",
	Short: "Check status of a specific endpoint",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			checkAllEndpoints(cmd.Context())
			return nil
		}

		operation := args[0]
		manager := core.GetEndpointManager()
		endpoint := manager.GetEndpoint(operation)
		isDynamic, status := manager.CheckEndpoint(operation)
		opFeatures, featureSource := manager.GetOpFeaturesWithSource(operation)
		quarantined := manager.IsQuarantined(operation)
		_, hasStaticFallback := core.GraphQLEndpoints[operation]
		if quarantined || (!isDynamic && !hasStaticFallback) {
			endpoint = ""
		}
		discoveryState := endpointDiscoveryState(isDynamic, hasStaticFallback, quarantined)
		probe := core.EndpointProbe{State: "unavailable", Message: "No endpoint is available"}
		if endpoint != "" {
			client, err := core.NewEndpointProbeClient(account)
			if err != nil {
				probe = core.EndpointProbe{State: "auth_error", Message: err.Error()}
			} else {
				defer client.Close()
				probe = core.ProbeGraphQLEndpoint(cmd.Context(), client, operation, endpoint, opFeatures)
			}
		}
		if manager.RecordProbeResult(operation, probe) {
			quarantined = true
			endpoint = ""
			isDynamic, status = manager.CheckEndpoint(operation)
			discoveryState = endpointDiscoveryState(isDynamic, hasStaticFallback, quarantined)
		}

		if isJSONMode() || isYAMLMode() {
			output(map[string]interface{}{
				"operation":       operation,
				"endpoint":        endpoint,
				"is_dynamic":      isDynamic,
				"discovery_state": discoveryState,
				"status":          status,
				"quarantined":     quarantined,
				"verified":        probe.State == "healthy",
				"features":        opFeatures,
				"feature_source":  featureSource,
				"probe":           probe,
			}, func() {})
			return nil
		}

		fmt.Println(display.Title("Endpoint Check"))
		fmt.Println(display.KeyValue("Operation:", operation))
		if quarantined {
			fmt.Println(display.KeyValue("URL:", "not selected (quarantined after 404)"))
		} else if endpoint == "" {
			fmt.Println(display.KeyValue("URL:", "not available (no verified endpoint)"))
		} else {
			fmt.Println(display.KeyValue("URL:", fmt.Sprintf("%s/%s", core.GraphQLBase, endpoint)))
		}
		fmt.Println(display.KeyValue("Discovery:", discoveryState))
		fmt.Println(display.KeyValue("Endpoint source:", display.StatusBadge(status)))
		fmt.Println(display.KeyValue("Verification:", probe.State+" — "+probe.Message))

		if len(opFeatures) > 0 {
			fmt.Println()
			fmt.Println(display.Section(fmt.Sprintf("Features from %s (%d)", featureSource, len(opFeatures))))
			for feat, val := range opFeatures {
				if val {
					fmt.Println(display.Bullet(display.Success(feat)))
				} else {
					fmt.Println(display.Bullet(display.Error(feat)))
				}
			}
		}

		return nil
	},
}

func endpointDiscoveryState(dynamic, hasStaticFallback, quarantined bool) string {
	switch {
	case quarantined:
		return "quarantined"
	case dynamic:
		return "discovered"
	case hasStaticFallback:
		return "static_fallback"
	default:
		return "unavailable"
	}
}

// endpointsRefreshCmd refreshes endpoints from X.com
func showEndpointRefreshProgress() bool {
	return !isJSONMode() && !isYAMLMode() && !isCompactMode()
}

var endpointsRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh endpoints from X.com",
	Long: `Fetches fresh GraphQL endpoints from the authenticated X.com web client.

This will:
1. Fetch the authenticated X.com homepage using the stored session cookies
2. Extract JS bundle and nested chunk URLs recursively
3. Download and parse bundles for GraphQL operations and generated URLs
4. Extract available feature switches
5. Validate the authenticated shell and update the local cache

A logged-out shell or a missing session is rejected. No public shell is used
as an authenticated endpoint source.

The process may take 10-30 seconds depending on network speed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if showEndpointRefreshProgress() {
			fmt.Println(display.Action("Refreshing endpoints from", "X.com"))
			fmt.Println(display.Muted("This may take a moment..."))
		}

		start := time.Now()

		ctx, cancel := context.WithTimeout(cmd.Context(), 120*time.Second)
		defer cancel()
		if err := core.GetEndpointManager().RefreshEndpointsForAccount(ctx, account); err != nil {
			if isJSONMode() || isYAMLMode() {
				_ = output(map[string]interface{}{"success": false, "error": err.Error()}, func() {})
			} else {
				fmt.Println(display.Error(fmt.Sprintf("Refresh failed: %v", err)))
			}
			abortCommand(core.ExitError)
			return nil
		}

		duration := time.Since(start).Round(time.Second)
		manager := core.GetEndpointManager()
		stats := manager.GetStats()

		if isJSONMode() || isYAMLMode() {
			output(map[string]interface{}{
				"success":         true,
				"duration":        duration.String(),
				"endpoints":       stats.DynamicCount,
				"total_endpoints": stats.TotalCount,
				"features":        stats.FeatureCount,
			}, func() {})
			return nil
		}

		fmt.Println(display.Success(fmt.Sprintf("Refreshed %d dynamic endpoints in %s", stats.DynamicCount, duration)))

		return nil
	},
}

// endpointsStatusCmd shows endpoint system status
var endpointsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show endpoint system status",
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := core.GetEndpointManager()
		stats := manager.GetStats()

		discovery, err := core.NewEndpointDiscovery(verbose)
		var healthStatus string
		var cacheAvailable bool

		if err != nil {
			healthStatus = fmt.Sprintf("Discovery unavailable: %v", err)
		} else {
			cache, err := discovery.LoadCache()
			if err != nil {
				healthStatus = "No dynamic endpoint cache"
			} else if len(cache.Endpoints) == 0 {
				healthStatus = "No dynamic endpoints (quarantine state retained)"
			} else if cache.IsValid() {
				healthStatus = "Dynamic cache current"
				cacheAvailable = true
			} else if cache.IsUsable() {
				healthStatus = "Dynamic cache stale but usable"
				cacheAvailable = true
			} else {
				healthStatus = "Dynamic cache expired"
			}
		}

		if isJSONMode() || isYAMLMode() {
			output(map[string]interface{}{
				"health":            healthStatus,
				"cache_available":   cacheAvailable,
				"discovery_checked": false,
				"stats":             stats,
			}, func() {})
			return nil
		}

		fmt.Println(display.Title("Endpoint System Status"))
		fmt.Println(display.KeyValue("Health:", display.StatusBadge(healthStatus)))
		fmt.Println(display.KeyValue("Dynamic cache:", fmt.Sprintf("%v", cacheAvailable)))
		fmt.Println(display.KeyValue("Discovery checked:", "no (run xsh endpoints refresh)"))
		fmt.Println(display.KeyValue("Available endpoints:", fmt.Sprintf("%d", stats.TotalCount)))
		fmt.Println(display.KeyValue("Dynamic endpoints:", fmt.Sprintf("%d", stats.DynamicCount)))
		fmt.Println(display.KeyValue("Static fallbacks:", fmt.Sprintf("%d", stats.StaticCount)))
		fmt.Println(display.KeyValue("Quarantined (404):", fmt.Sprintf("%d", stats.QuarantinedCount)))
		fmt.Println(display.KeyValue("Cached features:", fmt.Sprintf("%d", stats.FeatureCount)))
		if stats.CacheAge > 0 {
			fmt.Println(display.KeyValue("Cache age:", stats.CacheAge.Round(time.Second).String()))
		} else {
			fmt.Println(display.KeyValue("Cache age:", "never"))
		}

		if stats.DynamicCount == 0 {
			fmt.Println()
			fmt.Println(display.Warning("No dynamically discovered endpoints are available; static fallbacks may be in use"))
			fmt.Println(display.Info("Log in to X and run 'xsh endpoints refresh' when network access is available"))
		} else if stats.CacheAge > 24*time.Hour {
			fmt.Println()
			fmt.Println(display.Warning("Cache is old. Consider running 'xsh endpoints refresh'"))
		}

		return nil
	},
}

// endpointsUpdateCmd manually updates an endpoint
var endpointsUpdateCmd = &cobra.Command{
	Use:   "update [operation] [endpoint-id/operation-name]",
	Short: "Manually update an endpoint",
	Args:  cobra.ExactArgs(2),
	Example: `  xsh endpoints update HomeTimeline abc123/HomeTimeline
  xsh endpoints update UserByScreenName xyz789/UserByScreenName`,
	RunE: func(cmd *cobra.Command, args []string) error {
		operation := args[0]
		endpoint := args[1]

		manager := core.GetEndpointManager()
		manager.UpdateEndpoint(operation, endpoint)

		if isJSONMode() || isYAMLMode() {
			output(map[string]interface{}{
				"updated":   true,
				"operation": operation,
				"endpoint":  endpoint,
			}, func() {})
			return nil
		}

		fmt.Println(display.Success(fmt.Sprintf("Updated %s → %s", operation, endpoint)))

		return nil
	},
}

// endpointsResetCmd resets all endpoints
var endpointsResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset all endpoints to defaults",
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")

		if !force && !isJSONMode() {
			fmt.Print("Reset all endpoints to static defaults? [y/N] ")
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println(display.Warning("Aborted."))
				return nil
			}
		}

		core.InvalidateCache()

		if isJSONMode() || isYAMLMode() {
			output(map[string]string{"status": "reset"}, func() {})
			return nil
		}

		fmt.Println(display.Success("All endpoints reset to defaults"))
		fmt.Println(display.Info("Run 'xsh endpoints refresh' to discover fresh endpoints from X.com"))

		return nil
	},
}

// checkAllEndpoints checks all critical endpoints
func checkAllEndpoints(ctx context.Context) {
	manager := core.GetEndpointManager()

	criticalOps := []string{"Viewer", "UserByScreenName", "SearchTimeline"}
	for _, op := range []string{"HomeTimeline", "HomeLatestTimeline"} {
		if dynamic, _ := manager.CheckEndpoint(op); dynamic {
			criticalOps = append([]string{op}, criticalOps...)
			break
		}
	}

	fmt.Println(display.Title("Checking critical endpoints"))
	fmt.Println()

	client, clientErr := core.NewEndpointProbeClient(account)
	if client != nil {
		defer client.Close()
	}
	allOK := true
	for _, op := range criticalOps {
		endpoint := manager.GetEndpoint(op)
		isDynamic, _ := manager.CheckEndpoint(op)
		if manager.IsQuarantined(op) {
			allOK = false
			fmt.Println("  " + display.Warning("!") + " " + lipgloss.NewStyle().Width(25).Render(op) + " [" + display.Warning("quarantined") + "] " + display.Muted("endpoint excluded after confirmed 404"))
			continue
		}
		if _, supported := core.GraphQLEndpoints[op]; !isDynamic && !supported {
			continue
		}

		indicator := display.Muted("○")
		if isDynamic {
			indicator = display.Primary("◉")
		}

		probe := core.EndpointProbe{State: "auth_error", Message: "No stored authentication available"}
		if clientErr == nil {
			probe = core.ProbeGraphQLEndpoint(ctx, client, op, endpoint, manager.GetOpFeatures(op))
		}
		if manager.RecordProbeResult(op, probe) {
			allOK = false
			fmt.Println("  " + display.Warning("!") + " " + lipgloss.NewStyle().Width(25).Render(op) + " [" + display.Warning("quarantined") + "] " + display.Muted(probe.Message))
			continue
		}
		status := display.Success(probe.State)
		if probe.State != "healthy" {
			status = display.Warning(probe.State)
			allOK = false
		}

		fmt.Println("  " + indicator + " " + lipgloss.NewStyle().Width(25).Render(op) + " " + display.Muted("→") + " " + endpoint + " [" + status + "] " + display.Muted(probe.Message))
	}

	fmt.Println()
	if allOK {
		fmt.Println(display.Success("All checked endpoints returned GraphQL data"))
	} else {
		fmt.Println(display.Warning("Some endpoints could not be confirmed by a live request"))
	}
}

func init() {
	rootCmd.AddCommand(endpointsCmd)
	endpointsCmd.AddCommand(endpointsListCmd)
	endpointsCmd.AddCommand(endpointsCheckCmd)
	endpointsCmd.AddCommand(endpointsRefreshCmd)
	endpointsCmd.AddCommand(endpointsStatusCmd)
	endpointsCmd.AddCommand(endpointsUpdateCmd)
	endpointsCmd.AddCommand(endpointsResetCmd)

	endpointsResetCmd.Flags().BoolP("force", "f", false, "Skip confirmation")
}
