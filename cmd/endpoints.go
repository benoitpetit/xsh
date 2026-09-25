package cmd

import (
	"context"
	"fmt"
	"os"
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
	Run: func(cmd *cobra.Command, args []string) {
		manager := core.GetEndpointManager()
		endpoints := manager.ListEndpoints()
		stats := manager.GetStats()

		if isJSONMode() || isYAMLMode() {
			output(map[string]interface{}{
				"endpoints": endpoints,
				"stats":     stats,
			}, func() {})
			return
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

		for category, ops := range categories {
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
	},
}

// endpointsCheckCmd checks a specific endpoint
var endpointsCheckCmd = &cobra.Command{
	Use:   "check [operation]",
	Short: "Check status of a specific endpoint",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			checkAllEndpoints()
			return
		}

		operation := args[0]
		manager := core.GetEndpointManager()
		endpoint := manager.GetEndpoint(operation)
		isDynamic, status := manager.CheckEndpoint(operation)
		opFeatures := manager.GetOpFeatures(operation)
		quarantined := manager.IsQuarantined(operation)
		_, hasStaticFallback := core.GraphQLEndpoints[operation]
		if quarantined || (!isDynamic && !hasStaticFallback) {
			endpoint = ""
		}

		if isJSONMode() || isYAMLMode() {
			output(map[string]interface{}{
				"operation":   operation,
				"endpoint":    endpoint,
				"is_dynamic":  isDynamic,
				"status":      status,
				"quarantined": quarantined,
				"features":    opFeatures,
			}, func() {})
			return
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
		fmt.Println(display.KeyValue("Status:", display.StatusBadge(status)))

		if len(opFeatures) > 0 {
			fmt.Println()
			fmt.Println(display.Section(fmt.Sprintf("Features (%d)", len(opFeatures))))
			for feat, val := range opFeatures {
				if val {
					fmt.Println(display.Bullet(display.Success(feat)))
				} else {
					fmt.Println(display.Bullet(display.Error(feat)))
				}
			}
		}
	},
}

// endpointsRefreshCmd refreshes endpoints from X.com
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
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(display.Action("Refreshing endpoints from", "X.com"))
		fmt.Println(display.Muted("This may take a moment..."))

		start := time.Now()

		if err := core.RefreshEndpoints(); err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Refresh failed: %v", err)))
			os.Exit(core.ExitError)
			return
		}

		duration := time.Since(start).Round(time.Second)
		manager := core.GetEndpointManager()
		stats := manager.GetStats()

		if isJSONMode() || isYAMLMode() {
			output(map[string]interface{}{
				"success":   true,
				"duration":  duration.String(),
				"endpoints": stats.TotalCount,
				"features":  stats.FeatureCount,
			}, func() {})
			return
		}

		fmt.Println(display.Success(fmt.Sprintf("Refreshed %d endpoints in %s", stats.TotalCount, duration)))
	},
}

// endpointsStatusCmd shows endpoint system status
var endpointsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show endpoint system status",
	Run: func(cmd *cobra.Command, args []string) {
		manager := core.GetEndpointManager()
		stats := manager.GetStats()

		discovery, err := core.NewEndpointDiscovery(verbose)
		var healthStatus string
		var canDiscover bool

		if err != nil {
			healthStatus = fmt.Sprintf("Discovery unavailable: %v", err)
			canDiscover = false
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			cache, err := discovery.GetCachedEndpoints(ctx)
			if err != nil {
				healthStatus = fmt.Sprintf("Cache error: %v", err)
				canDiscover = false
			} else if cache.IsValid() {
				healthStatus = "OK"
				canDiscover = true
			} else {
				healthStatus = "Cache expired"
				canDiscover = true
			}
		}

		if isJSONMode() || isYAMLMode() {
			output(map[string]interface{}{
				"health":       healthStatus,
				"can_discover": canDiscover,
				"stats":        stats,
			}, func() {})
			return
		}

		fmt.Println(display.Title("Endpoint System Status"))
		fmt.Println(display.KeyValue("Health:", display.StatusBadge(healthStatus)))
		fmt.Println(display.KeyValue("Auto-discover:", fmt.Sprintf("%v", canDiscover)))
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
	},
}

// endpointsUpdateCmd manually updates an endpoint
var endpointsUpdateCmd = &cobra.Command{
	Use:   "update [operation] [endpoint-id/operation-name]",
	Short: "Manually update an endpoint",
	Args:  cobra.ExactArgs(2),
	Example: `  xsh endpoints update HomeTimeline abc123/HomeTimeline
  xsh endpoints update UserByScreenName xyz789/UserByScreenName`,
	Run: func(cmd *cobra.Command, args []string) {
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
			return
		}

		fmt.Println(display.Success(fmt.Sprintf("Updated %s → %s", operation, endpoint)))
	},
}

// endpointsResetCmd resets all endpoints
var endpointsResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset all endpoints to defaults",
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")

		if !force && !isJSONMode() {
			fmt.Print("Reset all endpoints to static defaults? [y/N] ")
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println(display.Warning("Aborted."))
				return
			}
		}

		core.InvalidateCache()

		if isJSONMode() || isYAMLMode() {
			output(map[string]string{"status": "reset"}, func() {})
			return
		}

		fmt.Println(display.Success("All endpoints reset to defaults"))
		fmt.Println(display.Info("Run 'xsh endpoints refresh' to discover fresh endpoints from X.com"))
	},
}

// checkAllEndpoints checks all critical endpoints
func checkAllEndpoints() {
	manager := core.GetEndpointManager()

	criticalOps := []string{
		"HomeTimeline",
		"HomeLatestTimeline",
		"UserByScreenName",
		"SearchTimeline",
		"TweetDetail",
		"UserTweets",
	}

	fmt.Println(display.Title("Checking critical endpoints"))
	fmt.Println()

	allOK := true
	for _, op := range criticalOps {
		endpoint := manager.GetEndpoint(op)
		isDynamic, _ := manager.CheckEndpoint(op)

		indicator := display.Muted("○")
		if isDynamic {
			indicator = display.Primary("◉")
		}

		status := display.Success("OK")
		if !isDynamic {
			status = display.Warning("STATIC")
			allOK = false
		}

		fmt.Println("  " + indicator + " " + lipgloss.NewStyle().Width(25).Render(op) + " " + display.Muted("→") + " " + endpoint + " [" + status + "]")
	}

	fmt.Println()
	if allOK {
		fmt.Println(display.Success("All critical endpoints are using dynamic discovery"))
	} else {
		fmt.Println(display.Warning("Some endpoints using static fallbacks"))
		fmt.Println(display.Info("Run 'xsh endpoints refresh' to enable dynamic discovery"))
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
