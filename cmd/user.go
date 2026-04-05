// Package cmd provides user-related CLI commands.
package cmd

import (
	"fmt"
	"os"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/benoitpetit/xsh/utils"
	"github.com/spf13/cobra"
)

var (
	userCount   int
	userReplies bool
)

// userCmd represents the user command (parent only, no subcommands at root level)
var userCmd = &cobra.Command{
	Use:   "user [handle]",
	Short: "View a user's profile",
	Long: `View a user's profile and manage user-related operations.
		
Use subcommands for specific actions:
  user tweets <handle>    View user's tweets
  user likes <handle>     View user's liked tweets  
  user followers <handle> View user's followers
  user following <handle>  View who a user follows`,
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

		output(user, func() {
			fmt.Println(display.FormatUser(user))
		})
	},
}

// userTweetsCmd represents the user tweets subcommand
var userTweetsCmd = &cobra.Command{
	Use:   "tweets [handle]",
	Short: "View a user's tweets",
	Args:  cobra.ExactArgs(1),
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

		runWithWatch(func() error {
			response, err := core.GetUserTweets(client, user.ID, userCount, "", userReplies)
			if err != nil {
				return fmt.Errorf("failed to fetch tweets: %w", err)
			}

			output(response.Tweets, func() {
				fmt.Println(display.FormatTweetList(response.Tweets))
			})
			return nil
		})
	},
}

// userLikesCmd represents the user likes subcommand
var userLikesCmd = &cobra.Command{
	Use:   "likes [handle]",
	Short: "View a user's liked tweets",
	Args:  cobra.ExactArgs(1),
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

		runWithWatch(func() error {
			response, err := core.GetUserLikes(client, user.ID, userCount, "")
			if err != nil {
				return fmt.Errorf("failed to fetch likes: %w", err)
			}

			output(response.Tweets, func() {
				fmt.Println(display.FormatTweetList(response.Tweets))
			})
			return nil
		})
	},
}

// userFollowersCmd represents the user followers subcommand
var userFollowersCmd = &cobra.Command{
	Use:   "followers [handle]",
	Short: "View a user's followers",
	Args:  cobra.ExactArgs(1),
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

		users, _, err := core.GetFollowers(client, user.ID, userCount, "")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to fetch followers: %v", err)))
			os.Exit(core.ExitError)
			return
		}

		output(users, func() {
			fmt.Println(display.FormatUserList(users))
		})
	},
}

// userFollowingCmd represents the user following subcommand
var userFollowingCmd = &cobra.Command{
	Use:   "following [handle]",
	Short: "View who a user follows",
	Args:  cobra.ExactArgs(1),
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

		users, _, err := core.GetFollowing(client, user.ID, userCount, "")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to fetch following: %v", err)))
			os.Exit(core.ExitError)
			return
		}

		output(users, func() {
			fmt.Println(display.FormatUserList(users))
		})
	},
}

func init() {
	rootCmd.AddCommand(userCmd)

	// Add subcommands to user command only (not to root)
	userCmd.AddCommand(userTweetsCmd)
	userCmd.AddCommand(userLikesCmd)
	userCmd.AddCommand(userFollowersCmd)
	userCmd.AddCommand(userFollowingCmd)

	// Flags
	userTweetsCmd.Flags().IntVarP(&userCount, "count", "n", 20, "Number of tweets")
	userTweetsCmd.Flags().BoolVar(&userReplies, "replies", false, "Include replies")
	userLikesCmd.Flags().IntVarP(&userCount, "count", "n", 20, "Number of tweets")
	userFollowersCmd.Flags().IntVarP(&userCount, "count", "n", 20, "Number of users")
	userFollowingCmd.Flags().IntVarP(&userCount, "count", "n", 20, "Number of users")
}
