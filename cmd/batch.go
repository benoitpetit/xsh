// Package cmd provides batch operation commands for xsh.
package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/benoitpetit/xsh/models"
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

var usersByID []string

func selectUserBatchInput(handles, ids []string) (bool, []string, error) {
	if len(handles) > 0 && len(ids) > 0 {
		return false, nil, fmt.Errorf("pass user handles or --id values, not both")
	}
	if len(ids) > 0 {
		for _, id := range ids {
			if _, err := strconv.ParseUint(id, 10, 64); err != nil {
				return false, nil, fmt.Errorf("invalid numeric user ID %q", id)
			}
		}
		return true, ids, nil
	}
	if len(handles) == 0 {
		return false, nil, fmt.Errorf("provide one or more user handles or repeat --id <user-id>")
	}
	return false, handles, nil
}

// usersBatchCmd fetches multiple users by handle
var usersBatchCmd = &cobra.Command{
	Use:   "users [handle]...",
	Short: "Fetch multiple user profiles by handle or ID",
	Long: `Fetch multiple user profiles by handle, or by numeric user IDs with --id.

Examples:
  xsh users alice bob
  xsh users --id 44196397 --id 783214`,
	Args: func(cmd *cobra.Command, args []string) error {
		_, _, err := selectUserBatchInput(args, usersByID)
		return err
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		byID, values, err := selectUserBatchInput(args, usersByID)
		if err != nil {
			return err
		}

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			abortCommand(core.ExitAuthError)
		}
		defer client.Close()

		var users []*models.User
		if byID {
			users, err = core.GetUsersByIDs(client, values)
		} else {
			for i, handle := range values {
				values[i] = strings.TrimPrefix(handle, "@")
			}
			users, err = core.GetUsersByHandles(client, values)
		}
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
	usersBatchCmd.Flags().StringArrayVar(&usersByID, "id", nil, "Look up a numeric user ID (repeat for each ID; mutually exclusive with handles)")
}
