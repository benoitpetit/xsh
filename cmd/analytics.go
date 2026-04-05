// Package cmd provides analytics commands for xsh.
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

var analyticsCount int

// analyticsCmd shows engagement analytics for a user's tweets
var analyticsCmd = &cobra.Command{
	Use:   "analytics [handle]",
	Short: "View tweet engagement analytics for a user",
	Long: `Fetch a user's recent tweets and display engagement analytics:
total impressions, average engagement rate, top performing tweets, etc.

Examples:
  xsh analytics elonmusk              # Analyze @elonmusk's recent tweets
  xsh analytics @jack -n 50           # Analyze last 50 tweets
  xsh analytics myhandle --json       # Get analytics as JSON`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			os.Exit(core.ExitAuthError)
			return
		}
		defer client.Close()

		handle, valid := utils.ValidateTwitterHandle(args[0])
		if !valid {
			fmt.Println(display.Error(fmt.Sprintf("Invalid Twitter handle: %s", args[0])))
			os.Exit(core.ExitError)
			return
		}

		user, err := core.GetUserByHandle(client, handle)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to fetch user: %v", err)))
			os.Exit(core.ExitError)
			return
		}
		if user == nil {
			fmt.Println(display.Error(fmt.Sprintf("User @%s not found", handle)))
			os.Exit(core.ExitError)
			return
		}

		response, err := core.GetUserTweets(client, user.ID, analyticsCount, "", false)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to fetch tweets: %v", err)))
			os.Exit(core.ExitError)
			return
		}

		tweets := response.Tweets
		if len(tweets) == 0 {
			fmt.Println(display.Muted("No tweets found for analytics"))
			return
		}

		stats := models.ComputeAnalytics(tweets)

		output(stats, func() {
			fmt.Println(display.FormatAnalytics(stats, handle))
		})
	},
}

func init() {
	rootCmd.AddCommand(analyticsCmd)
	analyticsCmd.Flags().IntVarP(&analyticsCount, "count", "n", 20, "Number of recent tweets to analyze")
}
