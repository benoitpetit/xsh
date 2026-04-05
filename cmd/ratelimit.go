// Package cmd provides rate limit dashboard command for xsh.
package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/spf13/cobra"
)

// ratelimitCmd shows the current rate limit status
var ratelimitCmd = &cobra.Command{
	Use:     "ratelimit",
	Aliases: []string{"rl", "limits"},
	Short:   "Show API rate limit status",
	Long: `Display rate limit information for recently used API endpoints.
Rate limits are tracked automatically from response headers during normal usage.

To populate rate limit data, first run some commands (feed, search, user, etc.)
then run this command to see the current state.

Examples:
  xsh feed && xsh ratelimit             # View after fetching feed
  xsh ratelimit --probe                 # Make a lightweight request to check limits
  xsh ratelimit --json                  # Machine-readable output`,
	Run: func(cmd *cobra.Command, args []string) {
		probe, _ := cmd.Flags().GetBool("probe")

		if probe {
			// Make a lightweight request to populate rate limit headers
			client, err := getClient("")
			if err != nil {
				fmt.Println(display.Error(err.Error()))
				os.Exit(core.ExitAuthError)
				return
			}
			defer client.Close()

			fmt.Println(display.Muted("Probing rate limits..."))

			// Probe a few common endpoints
			probeEndpoints(client)
		}

		limits := core.GetRateLimits()

		if len(limits) == 0 {
			fmt.Println(display.Muted("No rate limit data yet. Run some commands first, or use --probe."))
			return
		}

		// Sort by endpoint name
		sort.Slice(limits, func(i, j int) bool {
			return limits[i].Endpoint < limits[j].Endpoint
		})

		output(limits, func() {
			fmt.Println(display.FormatRateLimits(limits))
		})
	},
}

func probeEndpoints(client *core.XClient) {
	// HomeTimeline — lightweight probe
	_, _ = core.GetHomeTimeline(client, "ForYou", 1, "")

	// Search — lightweight probe
	_, _ = core.SearchTweets(client, "test", "Top", 1, "")
}

func init() {
	rootCmd.AddCommand(ratelimitCmd)
	ratelimitCmd.Flags().Bool("probe", false, "Make test requests to populate rate limit data")
}
