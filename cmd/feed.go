// Package cmd provides feed/timeline CLI commands.
package cmd

import (
	"fmt"
	"os"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/benoitpetit/xsh/models"
	"github.com/benoitpetit/xsh/utils"
	"github.com/spf13/cobra"
)

var (
	feedType      string
	feedCount     int
	feedPages     int
	feedCursor    string
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
			fmt.Println(display.Error(err.Error()))
			os.Exit(core.ExitAuthError)
			return
		}
		defer client.Close()

		runWithWatch(func() error {
			allTweets := []*models.Tweet{}
			cursor := feedCursor

			// Like Python: fetch count tweets per page, for pages iterations
			for i := 0; i < feedPages; i++ {
				response, err := core.GetHomeTimeline(client, feedType, feedCount, cursor)
				if err != nil {
					return fmt.Errorf("failed to fetch timeline: %w", err)
				}
				allTweets = append(allTweets, response.Tweets...)
				cursor = response.CursorBottom
				if !response.HasMore {
					break
				}
			}

			// Print cursor for next page if available (only on single run)
			if cursor != "" && !isJSONMode() && !isWatchMode() {
				fmt.Fprintln(os.Stderr, display.Info(fmt.Sprintf("Next cursor: %s", cursor)))
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

			output(allTweets, func() {
				fmt.Println(display.FormatTweetList(allTweets))
			})
			return nil
		})
	},
}

func init() {
	rootCmd.AddCommand(feedCmd)

	feedCmd.Flags().StringVarP(&feedType, "type", "t", "for-you", "Timeline type: for-you, following")
	feedCmd.Flags().IntVarP(&feedCount, "count", "n", 20, "Number of tweets per page")
	feedCmd.Flags().IntVarP(&feedPages, "pages", "p", 1, "Number of pages to fetch")
	feedCmd.Flags().StringVar(&feedCursor, "cursor", "", "Pagination cursor from previous response")
	feedCmd.Flags().StringVar(&feedFilter, "filter", "", "Filter: all, top, score")
	feedCmd.Flags().IntVar(&feedTopN, "top", 10, "Top N for filter mode")
	feedCmd.Flags().Float64Var(&feedThreshold, "threshold", 0.0, "Score threshold")
}
