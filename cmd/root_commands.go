package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/benoitpetit/xsh/utils"
)

// postCmd - alias for tweet post
var postCmd = &cobra.Command{
	Use:   "post [text]",
	Short: "Post a new tweet (alias: tweet post)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		text, valid := utils.ValidateTweetText(args[0], 280)
		if !valid {
			fmt.Println(display.PrintError("Tweet text is empty or exceeds 280 characters"))
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

		replyTo, _ := cmd.Flags().GetString("reply-to")
		quote, _ := cmd.Flags().GetString("quote")

		if replyTo != "" && !utils.ValidateTweetID(replyTo) {
			fmt.Println(display.PrintError(fmt.Sprintf("Invalid reply-to tweet ID: %s", replyTo)))
			os.Exit(core.ExitError)
			return
		}

		result, err := core.CreateTweet(client, text, replyTo, quote, nil)
		if err != nil {
			fmt.Println(display.PrintError(fmt.Sprintf("Failed to post tweet: %v", err)))
			return
		}

		if isJSONMode() {
			outputJSON(result)
		} else {
			tweetID := extractTweetIDFromResult(result)
			if tweetID != "" {
				fmt.Println(display.PrintSuccess(fmt.Sprintf("Tweet posted! ID: %s", tweetID)))
				fmt.Printf("  URL: https://x.com/i/web/status/%s\n", tweetID)
			} else {
				// Check if it's a "likely success" response (empty ID but no error)
				if note, ok := result["_note"].(string); ok && note != "" {
					fmt.Println(display.PrintSuccess("Tweet posted successfully!"))
					if core.Verbose {
						fmt.Printf("  Note: %s\n", note)
					}
				} else {
					fmt.Println(display.PrintSuccess("Tweet posted!"))
				}
			}
		}
	},
}

// deleteCmd - alias for tweet delete
var deleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete your tweet (alias: tweet delete)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !utils.ValidateTweetID(args[0]) {
			fmt.Println(display.PrintError(fmt.Sprintf("Invalid tweet ID: %s", args[0])))
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

		force, _ := cmd.Flags().GetBool("force")
		if !force && !isJSONMode() {
			fmt.Printf("Delete tweet %s? [y/N] ", args[0])
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Aborted.")
				return
			}
		}

		result, err := core.DeleteTweet(client, args[0])
		if err != nil {
			fmt.Println(display.PrintError(fmt.Sprintf("Failed to delete tweet: %v", err)))
			return
		}

		if isJSONMode() {
			outputJSON(result)
		} else {
			fmt.Println(display.PrintSuccess(fmt.Sprintf("Deleted tweet %s", args[0])))
		}
	},
}

// likeCmd - alias for tweet like
var likeCmd = &cobra.Command{
	Use:   "like [id]",
	Short: "Like a tweet (alias: tweet like)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !utils.ValidateTweetID(args[0]) {
			fmt.Println(display.PrintError(fmt.Sprintf("Invalid tweet ID: %s", args[0])))
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

		result, err := core.LikeTweet(client, args[0])
		if err != nil {
			fmt.Println(display.PrintError(fmt.Sprintf("Failed to like tweet: %v", err)))
			return
		}

		if isJSONMode() {
			outputJSON(result)
		} else {
			fmt.Println(display.PrintSuccess(fmt.Sprintf("Liked tweet %s", args[0])))
		}
	},
}

// unlikeCmd - alias for tweet unlike
var unlikeCmd = &cobra.Command{
	Use:   "unlike [id]",
	Short: "Unlike a tweet (alias: tweet unlike)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !utils.ValidateTweetID(args[0]) {
			fmt.Println(display.PrintError(fmt.Sprintf("Invalid tweet ID: %s", args[0])))
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

		result, err := core.UnlikeTweet(client, args[0])
		if err != nil {
			fmt.Println(display.PrintError(fmt.Sprintf("Failed to unlike tweet: %v", err)))
			return
		}

		if isJSONMode() {
			outputJSON(result)
		} else {
			fmt.Println(display.PrintSuccess(fmt.Sprintf("Unliked tweet %s", args[0])))
		}
	},
}

// retweetCmd - alias for tweet retweet
var retweetCmd = &cobra.Command{
	Use:   "retweet [id]",
	Short: "Retweet a tweet (alias: tweet retweet)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !utils.ValidateTweetID(args[0]) {
			fmt.Println(display.PrintError(fmt.Sprintf("Invalid tweet ID: %s", args[0])))
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

		result, err := core.Retweet(client, args[0])
		if err != nil {
			fmt.Println(display.PrintError(fmt.Sprintf("Failed to retweet: %v", err)))
			return
		}

		if isJSONMode() {
			outputJSON(result)
		} else {
			fmt.Println(display.PrintSuccess(fmt.Sprintf("Retweeted %s", args[0])))
		}
	},
}

// unretweetCmd - alias for tweet unretweet
var unretweetCmd = &cobra.Command{
	Use:   "unretweet [id]",
	Short: "Undo a retweet (alias: tweet unretweet)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !utils.ValidateTweetID(args[0]) {
			fmt.Println(display.PrintError(fmt.Sprintf("Invalid tweet ID: %s", args[0])))
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

		result, err := core.Unretweet(client, args[0])
		if err != nil {
			fmt.Println(display.PrintError(fmt.Sprintf("Failed to unretweet: %v", err)))
			return
		}

		if isJSONMode() {
			outputJSON(result)
		} else {
			fmt.Println(display.PrintSuccess(fmt.Sprintf("Unretweeted %s", args[0])))
		}
	},
}

// bookmarkCmd - alias for tweet bookmark
var bookmarkCmd = &cobra.Command{
	Use:   "bookmark [id]",
	Short: "Bookmark a tweet (alias: tweet bookmark)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.PrintError(err.Error()))
			return
		}
		defer client.Close()

		result, err := core.BookmarkTweet(client, args[0])
		if err != nil {
			fmt.Println(display.PrintError(fmt.Sprintf("Failed to bookmark: %v", err)))
			return
		}

		if isJSONMode() {
			outputJSON(result)
		} else {
			fmt.Println(display.PrintSuccess(fmt.Sprintf("Bookmarked tweet %s", args[0])))
		}
	},
}

// unbookmarkCmd - alias for tweet unbookmark
var unbookmarkCmd = &cobra.Command{
	Use:   "unbookmark [id]",
	Short: "Remove a bookmark (alias: tweet unbookmark)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.PrintError(err.Error()))
			return
		}
		defer client.Close()

		result, err := core.UnbookmarkTweet(client, args[0])
		if err != nil {
			fmt.Println(display.PrintError(fmt.Sprintf("Failed to unbookmark: %v", err)))
			return
		}

		if isJSONMode() {
			outputJSON(result)
		} else {
			fmt.Println(display.PrintSuccess(fmt.Sprintf("Unbookmarked tweet %s", args[0])))
		}
	},
}

func init() {
	// Add quick action commands to root
	rootCmd.AddCommand(postCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(likeCmd)
	rootCmd.AddCommand(unlikeCmd)
	rootCmd.AddCommand(retweetCmd)
	rootCmd.AddCommand(unretweetCmd)
	rootCmd.AddCommand(bookmarkCmd)
	rootCmd.AddCommand(unbookmarkCmd)

	// Flags for post command
	postCmd.Flags().String("reply-to", "", "Tweet ID to reply to")
	postCmd.Flags().String("quote", "", "Tweet URL to quote")

	// Flags for delete command
	deleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation")
}
