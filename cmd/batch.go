// Package cmd provides batch operation commands for xsh.
package cmd

import (
	"fmt"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			abortCommand(core.ExitAuthError)
		}
		defer client.Close()

		tweets, err := core.GetTweetsByIDs(client, args)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			abortCommand(core.ExitError)
		}

		output(tweets, func() {
			fmt.Println(display.FormatTweets(tweets))
		})

		return nil
	},
}

// usersBatchCmd fetches multiple users by handle
var usersBatchCmd = &cobra.Command{
	Use:   "users <handle>...",
	Short: "Fetch multiple user profiles",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Clean handles
		for i, h := range args {
			args[i] = strings.TrimPrefix(h, "@")
		}

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			abortCommand(core.ExitAuthError)
		}
		defer client.Close()

		users, err := core.GetUsersByHandles(client, args)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			abortCommand(core.ExitError)
		}

		output(users, func() {
			fmt.Println(display.FormatUsers(users))
		})

		return nil
	},
}

func init() {
	rootCmd.AddCommand(tweetsBatchCmd)
	rootCmd.AddCommand(usersBatchCmd)
}
