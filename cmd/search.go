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
			fmt.Println(display.PrintError("Search query cannot be empty"))
			os.Exit(core.ExitError)
			return
		}

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.PrintError(err.Error()))
			os.Exit(core.ExitAuthError)
			return
		}
		defer client.Close()

		var allTweets []*models.Tweet
		cursor := ""

		for i := 0; i < searchPages; i++ {
			response, err := core.SearchTweets(client, query, searchType, searchCount, cursor)
			if err != nil {
				// Check if it's an obsolete endpoint error
				if apiErr, ok := err.(*core.APIError); ok && apiErr.StatusCode == 404 {
					fmt.Println(display.PrintError(fmt.Sprintf("Search failed: %v", err)))
					fmt.Println()
					fmt.Println(display.PrintWarning("The API endpoint may be outdated. Try:"))
					fmt.Println("  xsh endpoints check SearchTimeline")
					fmt.Println("  xsh endpoints list")
				} else {
					fmt.Println(display.PrintError(fmt.Sprintf("Failed to search: %v", err)))
				}
				os.Exit(core.ExitError)
				return
			}
			allTweets = append(allTweets, response.Tweets...)
			cursor = response.CursorBottom
			if !response.HasMore {
				break
			}
		}

		if isYAMLMode() {
			outputYAML(allTweets)
		} else if isJSONMode() {
			outputJSON(allTweets)
		} else {
			fmt.Println(display.FormatTweetList(allTweets))
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVarP(&searchType, "type", "t", "Top", "Search type: Top, Latest, Photos, Videos")
	searchCmd.Flags().IntVarP(&searchCount, "count", "n", 20, "Number of tweets")
	searchCmd.Flags().IntVarP(&searchPages, "pages", "p", 1, "Number of pages")
	searchCmd.Flags().StringVar(&searchCursor, "cursor", "", "Cursor for pagination (from previous response)")
}
