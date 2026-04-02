package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
)

var (
	trendsLocation string
	trendsWOEID    int
)

// trendsCmd represents the trends command
var trendsCmd = &cobra.Command{
	Use:   "trends",
	Short: "View trending topics",
	Long: `View trending topics on Twitter/X by location.

Without flags, shows worldwide trends.
Use --location for city/country or --woeid for specific location ID.`,
	Example: `  xsh trends                          # Worldwide
  xsh trends --location "Paris"       # Paris trends
  xsh trends --location "France"      # France trends
  xsh trends --woeid 1                # Worldwide (WOEID 1)
  xsh trends --woeid 615702           # Paris (WOEID 615702)`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			return
		}
		defer client.Close()

		// Get trends
		trends, err := core.GetTrends(client, trendsWOEID)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to fetch trends: %v", err)))
			return
		}

		if isJSONMode() || isYAMLMode() {
			output(trends, func() {})
			return
		}

		// Display trends
		fmt.Println(display.Title("🔥 Trending Topics"))
		if trendsLocation != "" {
			fmt.Println(display.Muted(fmt.Sprintf("   Location: %s", trendsLocation)))
		} else if trendsWOEID != 0 {
			fmt.Println(display.Muted(fmt.Sprintf("   WOEID: %d", trendsWOEID)))
		} else {
			fmt.Println(display.Muted("   Worldwide"))
		}
		fmt.Println()

		for i, trend := range trends {
			line := display.Numbered(i+1, display.Bold(trend.Name))
			if trend.TweetVolume > 0 {
				volume := formatVolume(trend.TweetVolume)
				line = line + " " + display.Muted(volume)
			}
			if trend.IsPromoted {
				line = line + " " + display.Warning("· Promoted")
			}
			fmt.Println(line)
		}

		fmt.Println()
		fmt.Println(display.Muted(fmt.Sprintf("Showing top %d trends", len(trends))))
	},
}

func formatVolume(volume int) string {
	if volume >= 1000000 {
		return fmt.Sprintf("%.1fM posts", float64(volume)/1000000)
	} else if volume >= 1000 {
		return fmt.Sprintf("%.1fK posts", float64(volume)/1000)
	}
	return fmt.Sprintf("%d posts", volume)
}

func init() {
	rootCmd.AddCommand(trendsCmd)

	trendsCmd.Flags().StringVarP(&trendsLocation, "location", "l", "", "Location name (e.g., 'Paris', 'France')")
	trendsCmd.Flags().IntVarP(&trendsWOEID, "woeid", "w", 0, "Where On Earth ID (e.g., 1 for worldwide, 615702 for Paris)")
}
