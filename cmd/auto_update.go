package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/benoitpetit/xsh/core"
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
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#1DA1F2"))
		
		fmt.Println(titleStyle.Render("🔄 Automatic Endpoint Updater"))
		fmt.Println()
		fmt.Println("Method: Extract from X.com JavaScript bundles")
		fmt.Println()

		if autoUpdateDryRun {
			fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAD1F")).
				Render("📋 DRY RUN MODE - No changes will be made"))
			fmt.Println()
		}

		// Invalidate cache if force flag is set
		if autoUpdateForce {
			fmt.Println("🗑️  Clearing endpoint cache...")
			core.InvalidateCache()
			fmt.Println()
		}

		// Run auto-update using the JS-based discovery
		discovery := core.NewJSEndpointDiscovery()
		
		if autoUpdateDryRun {
			// Just check which endpoints are obsolete
			fmt.Println("🔍 Checking endpoints for obsolescence...")
			fmt.Println()
			
			obsolete, err := discovery.CheckObsoleteEndpoints()
			if err != nil {
				fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#E0245E")).
					Render(fmt.Sprintf("❌ Error checking endpoints: %v", err)))
				os.Exit(1)
			}
			
			if len(obsolete) == 0 {
				fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#00BA7C")).
					Render("✅ All endpoints are up to date!"))
				fmt.Println()
				fmt.Println("No updates needed.")
				return
			}
			
			fmt.Printf("⚠️  Found %d obsolete endpoint(s):\n\n", len(obsolete))
			
			for _, ep := range obsolete {
				fmt.Printf("  • %s\n", lipgloss.NewStyle().Bold(true).Render(ep.Name))
				fmt.Printf("    Current:  %s\n", ep.CurrentID)
				if ep.SuggestedID != "" {
					fmt.Printf("    Suggested: %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("#00BA7C")).Render(ep.SuggestedID))
				} else {
					fmt.Printf("    Suggested: %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAD1F")).Render("(not discovered - manual update needed)"))
				}
				fmt.Println()
			}
			
			fmt.Println("Run without --dry-run to apply these updates:")
			fmt.Println("  xsh auto-update")
			return
		} else {
			if err := discovery.UpdateObsoleteEndpoints(); err != nil {
				// Don't exit with error - manual instructions were shown
				fmt.Println()
				fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAD1F")).
					Render("⚠️  Some endpoints could not be auto-updated"))
				fmt.Println()
				fmt.Println("Please follow the manual instructions above.")
				os.Exit(0)
			}
		}

		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#00BA7C")).
			Render("✅ Auto-update complete!"))
		fmt.Println()
		fmt.Println("You can verify the updated endpoints with:")
		fmt.Println("  xsh endpoints list")
	},
}

func init() {
	rootCmd.AddCommand(autoUpdateCmd)

	autoUpdateCmd.Flags().BoolVar(&autoUpdateDryRun, "dry-run", false, "Check only, don't update")
	autoUpdateCmd.Flags().BoolVarP(&autoUpdateForce, "force", "f", false, "Force refresh (ignore cache)")
}
