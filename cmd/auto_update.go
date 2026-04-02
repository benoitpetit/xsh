package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/spf13/cobra"
)

var (
	autoUpdateDryRun bool
	autoUpdateForce  bool
)

// autoUpdateCmd represents the auto-update command
var autoUpdateCmd = &cobra.Command{
	Use:   "auto-update",
	Short: "Automatically update obsolete GraphQL endpoints",
	Long: `Automatically discovers and updates obsolete GraphQL endpoints by extracting
the latest endpoint IDs from X.com's public JavaScript bundles.

This command will:
1. Check all configured endpoints for obsolescence (404 errors)
2. Fetch X.com JavaScript bundles
3. Extract the latest GraphQL operation IDs
4. Automatically update the endpoint configuration

This approach does NOT require authentication - it uses the same public
JS bundles that your browser downloads when visiting x.com.`,
	Example: `  # Auto-update all obsolete endpoints
  xsh auto-update

  # Force refresh (ignore cache)
  xsh auto-update --force

  # Dry run (check only, don't update)
  xsh auto-update --dry-run`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(display.Title("🔄 Automatic Endpoint Updater"))
		fmt.Println(display.Info("Method: Extract from X.com JavaScript bundles"))
		fmt.Println()

		if autoUpdateDryRun {
			fmt.Println(display.Warning("DRY RUN MODE - No changes will be made"))
			fmt.Println()
		}

		// Invalidate cache if force flag is set
		if autoUpdateForce {
			fmt.Println(display.Action("Clearing", "endpoint cache"))
			core.InvalidateCache()
			fmt.Println()
		}

		manager := core.GetEndpointManager()
		before := manager.ListEndpoints()

		discovery, err := core.NewEndpointDiscovery(verbose)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error initializing endpoint discovery: %v", err)))
			os.Exit(1)
		}

		var previousCache *core.EndpointCache
		if autoUpdateDryRun {
			previousCache, _ = discovery.LoadCache()
		}

		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		cache, err := discovery.DiscoverEndpoints(ctx)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Auto-update failed: %v", err)))
			os.Exit(1)
		}

		after := manager.ListEndpoints()
		changed := make([]string, 0)
		for op, newEndpoint := range after {
			if oldEndpoint, ok := before[op]; ok && oldEndpoint != newEndpoint {
				changed = append(changed, op)
			}
			if _, ok := before[op]; !ok {
				if _, isDynamic := cache.Endpoints[op]; isDynamic {
					changed = append(changed, op)
				}
			}
		}
		sort.Strings(changed)

		if autoUpdateDryRun {
			if previousCache != nil {
				_ = discovery.SaveCache(previousCache)
				discovery.UpdateMemoryCache(previousCache)
			} else {
				core.InvalidateCache()
			}

			if len(changed) == 0 {
				fmt.Println(display.Success("No endpoint changes detected"))
				return
			}

			fmt.Println(display.Warning(fmt.Sprintf("%d endpoint(s) would be updated", len(changed))))
			fmt.Println()
			for _, op := range changed {
				fmt.Println(display.Bullet(display.Bold(op)))
				fmt.Println(display.KeyValue("  Current", before[op]))
				fmt.Println(display.KeyValue("  New", after[op]))
				fmt.Println()
			}

			fmt.Println(display.Info("Run without --dry-run to apply these updates:"))
			fmt.Println(display.Code("  xsh auto-update"))
			return
		}

		fmt.Println()
		fmt.Println(display.Success("Auto-update complete!"))
		fmt.Println(display.Info(fmt.Sprintf("Discovered %d endpoints, updated %d entries.", len(cache.Endpoints), len(changed))))
		fmt.Println()
		fmt.Println(display.Info("You can verify the updated endpoints with:"))
		fmt.Println(display.Code("  xsh endpoints list"))
	},
}

func init() {
	rootCmd.AddCommand(autoUpdateCmd)

	autoUpdateCmd.Flags().BoolVar(&autoUpdateDryRun, "dry-run", false, "Check only, don't update")
	autoUpdateCmd.Flags().BoolVarP(&autoUpdateForce, "force", "f", false, "Force refresh (ignore cache)")
}
