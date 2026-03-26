// Package cmd provides batch operation commands for xsh.
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/spf13/cobra"
)

// tweetsBatchCmd fetches multiple tweets by ID
var tweetsBatchCmd = &cobra.Command{
	Use:   "tweets <tweet-id>...",
	Short: "Fetch multiple tweets by ID",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := getClient("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		tweets, err := core.GetTweetsByIDs(client, args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		output(tweets, func() {
			fmt.Println(display.FormatTweets(tweets))
		})
	},
}

// usersBatchCmd fetches multiple users by handle
var usersBatchCmd = &cobra.Command{
	Use:   "users <handle>...",
	Short: "Fetch multiple user profiles",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Clean handles
		for i, h := range args {
			args[i] = strings.TrimPrefix(h, "@")
		}

		client, err := getClient("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		users, err := core.GetUsersByHandles(client, args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		output(users, func() {
			fmt.Println(display.FormatUsers(users))
		})
	},
}

func init() {
	rootCmd.AddCommand(tweetsBatchCmd)
	rootCmd.AddCommand(usersBatchCmd)
}
