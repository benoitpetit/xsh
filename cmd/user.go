// Package cmd provides user-related CLI commands.
package cmd

import (
	"fmt"

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
  user media <handle>     View user's media posts
  user likes <handle>     View user's liked tweets  
  user followers <handle> View user's followers
  user following <handle>  View who a user follows
  user followers-you-know <handle>      Discover likely connections
  user blue-verified-followers <handle> View blue-verified followers`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			abortCommand(core.ExitAuthError)
			return nil
		}
		defer client.Close()

		handle, valid := utils.ValidateTwitterHandle(args[0])
		if !valid {
			fmt.Println(display.Error(fmt.Sprintf("Invalid Twitter handle: %s", args[0])))
			abortCommand(core.ExitError)
			return nil
		}
		user, err := core.GetUserByHandle(client, handle)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to fetch user: %v", err)))
			abortCommand(core.ExitError)
			return nil
		}

		if user == nil {
			fmt.Println(display.Error(fmt.Sprintf("User @%s not found", handle)))
			abortCommand(core.ExitError)
			return nil
		}

		output(user, func() {
			fmt.Println(display.FormatUser(user))
		})

		return nil
	},
}

// userTweetsCmd represents the user tweets subcommand
var userTweetsCmd = &cobra.Command{
	Use:   "tweets [handle]",
	Short: "View a user's tweets",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			abortCommand(core.ExitAuthError)
			return nil
		}
		defer client.Close()

		handle, valid := utils.ValidateTwitterHandle(args[0])
		if !valid {
			fmt.Println(display.Error(fmt.Sprintf("Invalid Twitter handle: %s", args[0])))
			abortCommand(core.ExitError)
			return nil
		}
		user, err := core.GetUserByHandle(client, handle)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to fetch user: %v", err)))
			abortCommand(core.ExitError)
			return nil
		}

		if user == nil {
			fmt.Println(display.Error(fmt.Sprintf("User @%s not found", handle)))
			abortCommand(core.ExitError)
			return nil
		}

		runWithWatch(func() error {
			cursor, _ := cmd.Flags().GetString("cursor")
			response, err := core.GetUserTweets(client, user.ID, userCount, cursor, userReplies)
			if err != nil {
				return fmt.Errorf("failed to fetch tweets: %w", err)
			}

			outputPage(response.Tweets, response.CursorBottom, response.HasMore, func() {
				fmt.Println(display.FormatTweetList(response.Tweets))
			})
			return nil
		})

		return nil
	},
}

// userLikesCmd represents the user likes subcommand
var userLikesCmd = &cobra.Command{
	Use:   "likes [handle]",
	Short: "View a user's liked tweets",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			abortCommand(core.ExitAuthError)
			return nil
		}
		defer client.Close()

		handle, valid := utils.ValidateTwitterHandle(args[0])
		if !valid {
			fmt.Println(display.Error(fmt.Sprintf("Invalid Twitter handle: %s", args[0])))
			abortCommand(core.ExitError)
			return nil
		}
		user, err := core.GetUserByHandle(client, handle)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to fetch user: %v", err)))
			abortCommand(core.ExitError)
			return nil
		}

		if user == nil {
			fmt.Println(display.Error(fmt.Sprintf("User @%s not found", handle)))
			abortCommand(core.ExitError)
			return nil
		}

		runWithWatch(func() error {
			cursor, _ := cmd.Flags().GetString("cursor")
			response, err := core.GetUserLikes(client, user.ID, userCount, cursor)
			if err != nil {
				return fmt.Errorf("failed to fetch likes: %w", err)
			}

			outputPage(response.Tweets, response.CursorBottom, response.HasMore, func() {
				fmt.Println(display.FormatTweetList(response.Tweets))
			})
			return nil
		})

		return nil
	},
}

// userMediaCmd represents the user media subcommand
var userMediaCmd = &cobra.Command{
	Use:   "media [handle]",
	Short: "View a user's media posts",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			abortCommand(core.ExitAuthError)
			return nil
		}
		defer client.Close()

		handle, valid := utils.ValidateTwitterHandle(args[0])
		if !valid {
			fmt.Println(display.Error(fmt.Sprintf("Invalid Twitter handle: %s", args[0])))
			abortCommand(core.ExitError)
			return nil
		}
		user, err := core.GetUserByHandle(client, handle)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to fetch user: %v", err)))
			abortCommand(core.ExitError)
			return nil
		}
		if user == nil {
			fmt.Println(display.Error(fmt.Sprintf("User @%s not found", handle)))
			abortCommand(core.ExitError)
			return nil
		}

		runWithWatch(func() error {
			cursor, _ := cmd.Flags().GetString("cursor")
			response, err := core.GetUserMedia(client, user.ID, userCount, cursor)
			if err != nil {
				return fmt.Errorf("failed to fetch media: %w", err)
			}
			outputPage(response.Tweets, response.CursorBottom, response.HasMore, func() {
				fmt.Println(display.FormatTweetList(response.Tweets))
			})
			return nil
		})

		return nil
	},
}

func showUserFollowersError(message string) {
	if isJSONMode() || isYAMLMode() {
		_ = output(map[string]string{"error": message}, func() {})
		return
	}
	fmt.Fprintln(runtimeOutput(), display.Error(message))
}

// userFollowersCmd represents the user followers subcommand
var userFollowersCmd = &cobra.Command{
	Use:   "followers [handle]",
	Short: "View a user's followers",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient("")
		if err != nil {
			showUserFollowersError(err.Error())
			abortCommand(core.ExitAuthError)
			return nil
		}
		defer client.Close()

		handle, valid := utils.ValidateTwitterHandle(args[0])
		if !valid {
			showUserFollowersError(fmt.Sprintf("Invalid Twitter handle: %s", args[0]))
			abortCommand(core.ExitError)
			return nil
		}
		user, err := core.GetUserByHandle(client, handle)
		if err != nil {
			showUserFollowersError(fmt.Sprintf("Failed to fetch user: %v", err))
			abortCommand(core.ExitError)
			return nil
		}

		if user == nil {
			showUserFollowersError(fmt.Sprintf("User @%s not found", handle))
			abortCommand(core.ExitError)
			return nil
		}

		cursor, _ := cmd.Flags().GetString("cursor")
		users, nextCursor, err := core.GetFollowers(client, user.ID, userCount, cursor)
		if err != nil {
			showUserFollowersError(fmt.Sprintf("Failed to fetch followers: %v", err))
			abortCommand(core.ExitError)
			return nil
		}

		outputPage(users, nextCursor, nextCursor != "", func() {
			fmt.Println(display.FormatUserList(users))
		})

		return nil
	},
}

// userFollowingCmd represents the user following subcommand
var userFollowingCmd = &cobra.Command{
	Use:   "following [handle]",
	Short: "View who a user follows",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			abortCommand(core.ExitAuthError)
			return nil
		}
		defer client.Close()

		handle, valid := utils.ValidateTwitterHandle(args[0])
		if !valid {
			fmt.Println(display.Error(fmt.Sprintf("Invalid Twitter handle: %s", args[0])))
			abortCommand(core.ExitError)
			return nil
		}
		user, err := core.GetUserByHandle(client, handle)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to fetch user: %v", err)))
			abortCommand(core.ExitError)
			return nil
		}

		if user == nil {
			fmt.Println(display.Error(fmt.Sprintf("User @%s not found", handle)))
			abortCommand(core.ExitError)
			return nil
		}

		cursor, _ := cmd.Flags().GetString("cursor")
		users, nextCursor, err := core.GetFollowing(client, user.ID, userCount, cursor)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to fetch following: %v", err)))
			abortCommand(core.ExitError)
			return nil
		}

		outputPage(users, nextCursor, nextCursor != "", func() {
			fmt.Println(display.FormatUserList(users))
		})

		return nil
	},
}

// userFollowersYouKnowCmd shows accounts followed by people a user follows.
var userFollowersYouKnowCmd = &cobra.Command{
	Use:   "followers-you-know [handle]",
	Short: "View followers you may know",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			abortCommand(core.ExitAuthError)
		}
		defer client.Close()

		handle, valid := utils.ValidateTwitterHandle(args[0])
		if !valid {
			fmt.Println(display.Error(fmt.Sprintf("Invalid Twitter handle: %s", args[0])))
			abortCommand(core.ExitError)
		}
		user, err := core.GetUserByHandle(client, handle)
		if err != nil || user == nil {
			if err != nil {
				fmt.Println(display.Error(fmt.Sprintf("Failed to fetch user: %v", err)))
			} else {
				fmt.Println(display.Error(fmt.Sprintf("User @%s not found", handle)))
			}
			abortCommand(core.ExitError)
		}

		count, _ := cmd.Flags().GetInt("count")
		cursor, _ := cmd.Flags().GetString("cursor")
		users, nextCursor, err := core.GetFollowersYouKnow(client, user.ID, count, cursor)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to fetch followers you may know: %v", err)))
			abortCommand(core.ExitError)
		}
		outputPage(users, nextCursor, nextCursor != "", func() { fmt.Println(display.FormatUserList(users)) })

		return nil
	},
}

// userBlueVerifiedFollowersCmd shows verified followers of a user.
var userBlueVerifiedFollowersCmd = &cobra.Command{
	Use:   "blue-verified-followers [handle]",
	Short: "View a user's blue-verified followers",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			abortCommand(core.ExitAuthError)
		}
		defer client.Close()

		handle, valid := utils.ValidateTwitterHandle(args[0])
		if !valid {
			fmt.Println(display.Error(fmt.Sprintf("Invalid Twitter handle: %s", args[0])))
			abortCommand(core.ExitError)
		}
		user, err := core.GetUserByHandle(client, handle)
		if err != nil || user == nil {
			if err != nil {
				fmt.Println(display.Error(fmt.Sprintf("Failed to fetch user: %v", err)))
			} else {
				fmt.Println(display.Error(fmt.Sprintf("User @%s not found", handle)))
			}
			abortCommand(core.ExitError)
		}

		count, _ := cmd.Flags().GetInt("count")
		cursor, _ := cmd.Flags().GetString("cursor")
		users, nextCursor, err := core.GetBlueVerifiedFollowers(client, user.ID, count, cursor)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to fetch blue-verified followers: %v", err)))
			abortCommand(core.ExitError)
		}
		outputPage(users, nextCursor, nextCursor != "", func() { fmt.Println(display.FormatUserList(users)) })

		return nil
	},
}

func init() {
	rootCmd.AddCommand(userCmd)

	// Add subcommands to user command only (not to root)
	userCmd.AddCommand(userTweetsCmd)
	userCmd.AddCommand(userMediaCmd)
	userCmd.AddCommand(userLikesCmd)
	userCmd.AddCommand(userFollowersCmd)
	userCmd.AddCommand(userFollowingCmd)
	userCmd.AddCommand(userFollowersYouKnowCmd)
	userCmd.AddCommand(userBlueVerifiedFollowersCmd)

	// Flags
	userTweetsCmd.Flags().IntVarP(&userCount, "count", "n", 20, "Number of tweets")
	userTweetsCmd.Flags().BoolVar(&userReplies, "replies", false, "Include replies")
	userTweetsCmd.Flags().String("cursor", "", "Pagination cursor from a previous response")
	userMediaCmd.Flags().IntVarP(&userCount, "count", "n", 20, "Number of tweets")
	userMediaCmd.Flags().String("cursor", "", "Pagination cursor from a previous response")
	userLikesCmd.Flags().IntVarP(&userCount, "count", "n", 20, "Number of tweets")
	userLikesCmd.Flags().String("cursor", "", "Pagination cursor from a previous response")
	userFollowersCmd.Flags().IntVarP(&userCount, "count", "n", 20, "Number of users")
	userFollowersCmd.Flags().String("cursor", "", "Pagination cursor from a previous response")
	userFollowingCmd.Flags().IntVarP(&userCount, "count", "n", 20, "Number of users")
	userFollowingCmd.Flags().String("cursor", "", "Pagination cursor from a previous response")
	userFollowersYouKnowCmd.Flags().IntP("count", "n", 20, "Number of users")
	userFollowersYouKnowCmd.Flags().String("cursor", "", "Pagination cursor from a previous response")
	userBlueVerifiedFollowersCmd.Flags().IntP("count", "n", 20, "Number of users")
	userBlueVerifiedFollowersCmd.Flags().String("cursor", "", "Pagination cursor from a previous response")
}
