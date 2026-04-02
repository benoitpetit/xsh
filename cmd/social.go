// Package cmd provides social action commands for xsh.
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/spf13/cobra"
)

// socialCmd represents the social command group
var socialCmd = &cobra.Command{
	Use:   "social",
	Short: "Social actions (follow, block, mute)",
}

// followCmd follows a user
var followCmd = &cobra.Command{
	Use:   "follow <handle>",
	Short: "Follow a user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handle := strings.TrimPrefix(args[0], "@")

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		user, err := core.GetUserByHandle(client, handle)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error fetching user: %v", err)))
			os.Exit(core.ExitError)
		}
		if user == nil {
			fmt.Println(display.Error(fmt.Sprintf("User @%s not found", handle)))
			os.Exit(core.ExitError)
		}

		_, err = core.FollowUser(client, user.ID)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitError)
		}

		output(map[string]string{
			"action": "follow",
			"handle": handle,
			"status": "success",
		}, func() {
			fmt.Println(display.Success(fmt.Sprintf("Followed @%s", handle)))
		})
	},
}

// unfollowCmd unfollows a user
var unfollowCmd = &cobra.Command{
	Use:   "unfollow <handle>",
	Short: "Unfollow a user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handle := strings.TrimPrefix(args[0], "@")

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		user, err := core.GetUserByHandle(client, handle)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error fetching user: %v", err)))
			os.Exit(core.ExitError)
		}
		if user == nil {
			fmt.Println(display.Error(fmt.Sprintf("User @%s not found", handle)))
			os.Exit(core.ExitError)
		}

		_, err = core.UnfollowUser(client, user.ID)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitError)
		}

		output(map[string]string{
			"action": "unfollow",
			"handle": handle,
			"status": "success",
		}, func() {
			fmt.Println(display.Success(fmt.Sprintf("Unfollowed @%s", handle)))
		})
	},
}

// blockCmd blocks a user
var blockCmd = &cobra.Command{
	Use:   "block <handle>",
	Short: "Block a user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handle := strings.TrimPrefix(args[0], "@")

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		user, err := core.GetUserByHandle(client, handle)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error fetching user: %v", err)))
			os.Exit(core.ExitError)
		}
		if user == nil {
			fmt.Println(display.Error(fmt.Sprintf("User @%s not found", handle)))
			os.Exit(core.ExitError)
		}

		_, err = core.BlockUser(client, user.ID)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitError)
		}

		output(map[string]string{
			"action": "block",
			"handle": handle,
			"status": "success",
		}, func() {
			fmt.Println(display.Success(fmt.Sprintf("Blocked @%s", handle)))
		})
	},
}

// unblockCmd unblocks a user
var unblockCmd = &cobra.Command{
	Use:   "unblock <handle>",
	Short: "Unblock a user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handle := strings.TrimPrefix(args[0], "@")

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		user, err := core.GetUserByHandle(client, handle)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error fetching user: %v", err)))
			os.Exit(core.ExitError)
		}
		if user == nil {
			fmt.Println(display.Error(fmt.Sprintf("User @%s not found", handle)))
			os.Exit(core.ExitError)
		}

		_, err = core.UnblockUser(client, user.ID)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitError)
		}

		output(map[string]string{
			"action": "unblock",
			"handle": handle,
			"status": "success",
		}, func() {
			fmt.Println(display.Success(fmt.Sprintf("Unblocked @%s", handle)))
		})
	},
}

// muteCmd mutes a user
var muteCmd = &cobra.Command{
	Use:   "mute <handle>",
	Short: "Mute a user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handle := strings.TrimPrefix(args[0], "@")

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		user, err := core.GetUserByHandle(client, handle)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error fetching user: %v", err)))
			os.Exit(core.ExitError)
		}
		if user == nil {
			fmt.Println(display.Error(fmt.Sprintf("User @%s not found", handle)))
			os.Exit(core.ExitError)
		}

		_, err = core.MuteUser(client, user.ID)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitError)
		}

		output(map[string]string{
			"action": "mute",
			"handle": handle,
			"status": "success",
		}, func() {
			fmt.Println(display.Success(fmt.Sprintf("Muted @%s", handle)))
		})
	},
}

// unmuteCmd unmutes a user
var unmuteCmd = &cobra.Command{
	Use:   "unmute <handle>",
	Short: "Unmute a user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handle := strings.TrimPrefix(args[0], "@")

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		user, err := core.GetUserByHandle(client, handle)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error fetching user: %v", err)))
			os.Exit(core.ExitError)
		}
		if user == nil {
			fmt.Println(display.Error(fmt.Sprintf("User @%s not found", handle)))
			os.Exit(core.ExitError)
		}

		_, err = core.UnmuteUser(client, user.ID)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitError)
		}

		output(map[string]string{
			"action": "unmute",
			"handle": handle,
			"status": "success",
		}, func() {
			fmt.Println(display.Success(fmt.Sprintf("Unmuted @%s", handle)))
		})
	},
}

func init() {
	rootCmd.AddCommand(socialCmd)

	socialCmd.AddCommand(followCmd)
	socialCmd.AddCommand(unfollowCmd)
	socialCmd.AddCommand(blockCmd)
	socialCmd.AddCommand(unblockCmd)
	socialCmd.AddCommand(muteCmd)
	socialCmd.AddCommand(unmuteCmd)

	rootCmd.AddCommand(followCmd)
	rootCmd.AddCommand(unfollowCmd)
	rootCmd.AddCommand(blockCmd)
	rootCmd.AddCommand(unblockCmd)
	rootCmd.AddCommand(muteCmd)
	rootCmd.AddCommand(unmuteCmd)
}
