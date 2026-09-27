// Package cmd provides community commands for xsh.
package cmd

import (
	"fmt"
	"strconv"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/spf13/cobra"
)

// communityCmd represents the community parent command
var communityCmd = &cobra.Command{
	Use:   "community",
	Short: "Community operations - view, tweets, join, leave",
	Long: `View and interact with Twitter/X Communities.

Examples:
  xsh community view 123456          # View community details
  xsh community tweets 123456        # View community tweets
  xsh community join 123456          # Join a community
  xsh community leave 123456         # Leave a community`,
}

// communityViewCmd views a community's details
var communityViewCmd = &cobra.Command{
	Use:   "view [community-id]",
	Short: "View a community's details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			abortCommand(core.ExitAuthError)
		}
		defer client.Close()

		community, err := core.GetCommunity(client, args[0])
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to fetch community: %v", err)))
			abortCommand(core.ExitError)
		}

		output(community, func() {
			fmt.Println(display.FormatCommunity(community))
		})

		return nil
	},
}

// communityTweetsCmd views tweets from a community
var communityTweetsCmd = &cobra.Command{
	Use:   "tweets [community-id]",
	Short: "View tweets from a community",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		count, _ := strconv.Atoi(cmd.Flag("count").Value.String())

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			abortCommand(core.ExitAuthError)
		}
		defer client.Close()

		response, err := core.GetCommunityTimeline(client, args[0], count, "")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to fetch community tweets: %v", err)))
			abortCommand(core.ExitError)
		}

		output(response.Tweets, func() {
			fmt.Println(display.FormatTweetList(response.Tweets))
		})

		return nil
	},
}

// communityJoinCmd joins a community
var communityJoinCmd = &cobra.Command{
	Use:   "join [community-id]",
	Short: "Join a community",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			abortCommand(core.ExitAuthError)
		}
		defer client.Close()

		_, err = core.JoinCommunity(client, args[0])
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to join community: %v", err)))
			abortCommand(core.ExitError)
		}

		fmt.Println(display.Success(fmt.Sprintf("Joined community %s", args[0])))

		return nil
	},
}

// communityLeaveCmd leaves a community
var communityLeaveCmd = &cobra.Command{
	Use:   "leave [community-id]",
	Short: "Leave a community",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			abortCommand(core.ExitAuthError)
		}
		defer client.Close()

		_, err = core.LeaveCommunity(client, args[0])
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to leave community: %v", err)))
			abortCommand(core.ExitError)
		}

		fmt.Println(display.Success(fmt.Sprintf("Left community %s", args[0])))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(communityCmd)
	communityCmd.AddCommand(communityViewCmd)
	communityCmd.AddCommand(communityTweetsCmd)
	communityCmd.AddCommand(communityJoinCmd)
	communityCmd.AddCommand(communityLeaveCmd)

	communityTweetsCmd.Flags().IntP("count", "n", 20, "Number of tweets to fetch")
}
