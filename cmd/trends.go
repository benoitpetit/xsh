package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
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
			fmt.Println(display.PrintError(err.Error()))
			return
		}
		defer client.Close()

		// Get trends
		trends, err := core.GetTrends(client, trendsWOEID)
		if err != nil {
			fmt.Println(display.PrintError(fmt.Sprintf("Failed to fetch trends: %v", err)))
			return
		}

		if isJSONMode() {
			outputJSON(trends)
			return
		}

		// Display trends
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#1DA1F2"))
		locationStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#8899A6"))
		rankStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#8899A6")).Width(4).Align(lipgloss.Right)
		trendStyle := lipgloss.NewStyle().Bold(true)
		volumeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#8899A6")).Italic(true)

		fmt.Println(titleStyle.Render("🔥 Trending Topics"))
		if trendsLocation != "" {
			fmt.Println(locationStyle.Render(fmt.Sprintf("   Location: %s", trendsLocation)))
		} else if trendsWOEID != 0 {
			fmt.Println(locationStyle.Render(fmt.Sprintf("   WOEID: %d", trendsWOEID)))
		} else {
			fmt.Println(locationStyle.Render("   Worldwide"))
		}
		fmt.Println()

		for i, trend := range trends {
			rank := fmt.Sprintf("%d.", i+1)
			fmt.Printf("%s %s", rankStyle.Render(rank), trendStyle.Render(trend.Name))
			
			if trend.TweetVolume > 0 {
				volume := formatVolume(trend.TweetVolume)
				fmt.Printf(" %s", volumeStyle.Render(volume))
			}
			
			if trend.IsPromoted {
				fmt.Print(" " + lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAD1F")).Render("· Promoted"))
			}
			
			fmt.Println()
		}

		fmt.Println()
		fmt.Println(locationStyle.Render(fmt.Sprintf("Showing top %d trends", len(trends))))
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
