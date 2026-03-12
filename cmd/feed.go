// Package cmd provides feed/timeline CLI commands.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/benoitpetit/xsh/models"
	"github.com/benoitpetit/xsh/utils"
)

var (
	feedType      string
	feedCount     int
	feedPages     int
	feedFilter    string
	feedTopN      int
	feedThreshold float64
)

// feedCmd represents the feed command
var feedCmd = &cobra.Command{
	Use:   "feed",
	Short: "View your timeline",
	Long:  "Fetch your home timeline (for-you or following).",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.PrintError(err.Error()))
			os.Exit(core.ExitAuthError)
			return
		}
		defer client.Close()

		allTweets := []*models.Tweet{}
		cursor := ""

		// Like Python: fetch count tweets per page, for pages iterations
		for i := 0; i < feedPages; i++ {
			response, err := core.GetHomeTimeline(client, feedType, feedCount, cursor)
			if err != nil {
				fmt.Println(display.PrintError(fmt.Sprintf("Failed to fetch timeline: %v", err)))
				os.Exit(core.ExitError)
				return
			}
			allTweets = append(allTweets, response.Tweets...)
			cursor = response.CursorBottom
			if !response.HasMore {
				break
			}
		}

		// Truncate to exact count requested (API may return more than requested)
		totalRequested := feedCount * feedPages
		if len(allTweets) > totalRequested {
			allTweets = allTweets[:totalRequested]
		}

		// Apply filter if specified
		if feedFilter != "" {
			cfg, _ := core.LoadConfig()
			allTweets = utils.FilterTweets(allTweets, feedFilter, feedThreshold, feedTopN, &cfg.Filter)
		}

		if isJSONMode() {
			outputJSON(allTweets)
		} else {
			fmt.Println(display.FormatTweetList(allTweets))
		}
	},
}

func init() {
	rootCmd.AddCommand(feedCmd)

	feedCmd.Flags().StringVarP(&feedType, "type", "t", "for-you", "Timeline type: for-you, following")
	feedCmd.Flags().IntVarP(&feedCount, "count", "n", 20, "Number of tweets per page")
	feedCmd.Flags().IntVarP(&feedPages, "pages", "p", 1, "Number of pages to fetch")
	feedCmd.Flags().StringVar(&feedFilter, "filter", "", "Filter: all, top, score")
	feedCmd.Flags().IntVar(&feedTopN, "top", 10, "Top N for filter mode")
	feedCmd.Flags().Float64Var(&feedThreshold, "threshold", 0.0, "Score threshold")
}
