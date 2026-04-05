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
	searchFilter    string
	searchTopN      int
	searchThreshold float64
)

var (
	searchType   string
	searchCount  int
	searchPages  int
	searchCursor string
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for tweets",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Sanitize search query (max 500 chars for Twitter search)
		query := utils.SanitizeInputWithMaxLength(args[0], 500)
		if query == "" {
			fmt.Println(display.Error("Search query cannot be empty"))
			os.Exit(core.ExitError)
			return
		}

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			os.Exit(core.ExitAuthError)
			return
		}
		defer client.Close()

		runWithWatch(func() error {
			var allTweets []*models.Tweet
			cursor := searchCursor

			for i := 0; i < searchPages; i++ {
				response, err := core.SearchTweets(client, query, searchType, searchCount, cursor)
				if err != nil {
					return err
				}
				allTweets = append(allTweets, response.Tweets...)
				cursor = response.CursorBottom
				if !response.HasMore {
					break
				}
			}

			// Apply filter if specified
			if searchFilter != "" {
				cfg, _ := core.LoadConfig()
				allTweets = utils.FilterTweets(allTweets, searchFilter, searchThreshold, searchTopN, &cfg.Filter)
			}

			output(allTweets, func() {
				fmt.Println(display.FormatTweetList(allTweets))
			})
			return nil
		})
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVarP(&searchType, "type", "t", "Top", "Search type: Top, Latest, Photos, Videos")
	searchCmd.Flags().IntVarP(&searchCount, "count", "n", 20, "Number of tweets")
	searchCmd.Flags().IntVarP(&searchPages, "pages", "p", 1, "Number of pages")
	searchCmd.Flags().StringVar(&searchCursor, "cursor", "", "Cursor for pagination (from previous response)")
	searchCmd.Flags().StringVar(&searchFilter, "filter", "", "Filter: all, top, score")
	searchCmd.Flags().IntVar(&searchTopN, "top", 10, "Top N for filter mode")
	searchCmd.Flags().Float64Var(&searchThreshold, "threshold", 0.0, "Score threshold")
}
