package cmd

import (
	"fmt"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/benoitpetit/xsh/utils"
	"github.com/spf13/cobra"
)

var tweetGetCmd = &cobra.Command{
	Use:   "get <tweet-id>",
	Short: "Fetch a single tweet without its conversation thread",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !utils.ValidateTweetID(args[0]) {
			return fmt.Errorf("invalid tweet ID: %s", args[0])
		}
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Get tweet: %v", err)))
			abortCommand(core.ExitAuthError)
			return nil
		}
		defer client.Close()

		tweet, err := core.GetTweetByID(client, args[0])
		if err != nil {
			return fmt.Errorf("failed to fetch tweet: %w", err)
		}
		if tweet == nil {
			return fmt.Errorf("tweet %s not found", args[0])
		}
		return output(tweet, func() {
			fmt.Fprintln(runtimeOutput(), display.FormatSingleTweet(tweet))
		})
	},
}
