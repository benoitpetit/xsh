package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
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

		if isJSONMode() {
			outputJSON(map[string]interface{}{
				"endpoints": endpoints,
				"stats":     stats,
			})
			return
		}

		fmt.Println(display.PrintSuccess(fmt.Sprintf("Found %d endpoints", len(endpoints))))
		fmt.Printf("\nLast updated: %s ago\n\n", time.Since(stats.LastUpdated).Round(time.Second))

		// Group by category
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
			fmt.Printf("\n%s:\n", category)
			for _, op := range ops {
				if endpoint, ok := endpoints[op]; ok {
					isDynamic, status := manager.CheckEndpoint(op)
					indicator := "●"
					if isDynamic {
						indicator = "◉"
					}
					fmt.Printf("  %s %-25s -> %s (%s)\n", indicator, op, endpoint, status)
				}
			}
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
			// Check all critical endpoints
			checkAllEndpoints()
			return
		}

		operation := args[0]
		manager := core.GetEndpointManager()
		endpoint := manager.GetEndpoint(operation)
		isDynamic, status := manager.CheckEndpoint(operation)
		opFeatures := manager.GetOpFeatures(operation)

		if isJSONMode() {
			outputJSON(map[string]interface{}{
				"operation":   operation,
				"endpoint":    endpoint,
				"is_dynamic":  isDynamic,
				"status":      status,
				"features":    opFeatures,
			})
			return
		}

		fmt.Printf("Endpoint: %s\n", operation)
		fmt.Printf("URL: %s/%s\n", core.GraphQLBase, endpoint)
		fmt.Printf("Status: %s\n", status)
		
		if len(opFeatures) > 0 {
			fmt.Printf("\nFeatures (%d):\n", len(opFeatures))
			for feat, val := range opFeatures {
				indicator := "✗"
				if val {
					indicator = "✓"
				}
				fmt.Printf("  %s %s\n", indicator, feat)
			}
		}
	},
}

// endpointsRefreshCmd refreshes endpoints from X.com
var endpointsRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh endpoints from X.com",
	Long: `Fetches fresh GraphQL endpoints from X.com JavaScript bundles.

This will:
1. Fetch the X.com homepage
2. Extract JS bundle URLs
3. Download and parse bundles for GraphQL operations
4. Extract feature switches from __INITIAL_STATE__
5. Update the local cache

The process may take 10-30 seconds depending on network speed.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Refreshing endpoints from X.com...")
		fmt.Println("This may take a moment...")

		start := time.Now()
		
		// Invalidate cache first
		core.InvalidateCache()
		
		// Refresh
		if err := core.RefreshEndpoints(); err != nil {
			fmt.Println(display.PrintError(fmt.Sprintf("Refresh failed: %v", err)))
			os.Exit(core.ExitError)
			return
		}

		duration := time.Since(start).Round(time.Second)
		manager := core.GetEndpointManager()
		stats := manager.GetStats()

		if isJSONMode() {
			outputJSON(map[string]interface{}{
				"success":     true,
				"duration":    duration.String(),
				"endpoints":   stats.TotalCount,
				"features":    stats.FeatureCount,
			})
			return
		}

		fmt.Println(display.PrintSuccess(fmt.Sprintf("Refreshed %d endpoints in %s", stats.TotalCount, duration)))
	},
}

// endpointsStatusCmd shows endpoint system status
var endpointsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show endpoint system status",
	Run: func(cmd *cobra.Command, args []string) {
		manager := core.GetEndpointManager()
		stats := manager.GetStats()

		// Check health
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
				canDiscover = true
			} else if cache.IsValid() {
				healthStatus = "OK"
				canDiscover = true
			} else {
				healthStatus = "Cache expired"
				canDiscover = true
			}
		}

		if isJSONMode() {
			outputJSON(map[string]interface{}{
				"health":       healthStatus,
				"can_discover": canDiscover,
				"stats":        stats,
			})
			return
		}

		fmt.Println("Endpoint System Status")
		fmt.Println("======================")
		fmt.Printf("Health: %s\n", healthStatus)
		fmt.Printf("Can auto-discover: %v\n", canDiscover)
		fmt.Printf("Cached endpoints: %d\n", stats.TotalCount)
		fmt.Printf("Cached features: %d\n", stats.FeatureCount)
		fmt.Printf("Cache age: %s\n", stats.CacheAge.Round(time.Second))
		
		if stats.CacheAge > 24*time.Hour {
			fmt.Println()
			fmt.Println(display.PrintWarning("Cache is old. Consider running 'xsh endpoints refresh'"))
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

		if isJSONMode() {
			outputJSON(map[string]interface{}{
				"updated":   true,
				"operation": operation,
				"endpoint":  endpoint,
			})
			return
		}

		fmt.Println(display.PrintSuccess(fmt.Sprintf("Updated %s -> %s", operation, endpoint)))
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
				fmt.Println("Aborted.")
				return
			}
		}

		core.InvalidateCache()

		if isJSONMode() {
			outputJSON(map[string]string{"status": "reset"})
			return
		}

		fmt.Println(display.PrintSuccess("All endpoints reset to defaults"))
		fmt.Println("Run 'xsh endpoints refresh' to discover fresh endpoints from X.com")
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

	fmt.Println("Checking critical endpoints...")
	fmt.Println()

	allOK := true
	for _, op := range criticalOps {
		endpoint := manager.GetEndpoint(op)
		isDynamic, _ := manager.CheckEndpoint(op)
		
		indicator := "○"
		if isDynamic {
			indicator = "◉"
		}
		
		status := "OK"
		if !isDynamic {
			status = "STATIC"
			allOK = false
		}
		
		fmt.Printf("%s %-25s -> %s [%s]\n", indicator, op, endpoint, status)
	}

	fmt.Println()
	if allOK {
		fmt.Println(display.PrintSuccess("All critical endpoints are using dynamic discovery"))
	} else {
		fmt.Println(display.PrintWarning("Some endpoints using static fallbacks"))
		fmt.Println("Run 'xsh endpoints refresh' to enable dynamic discovery")
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
