package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/benoitpetit/xsh/models"
	"github.com/benoitpetit/xsh/utils"
)

var (
	tweetThread bool
	tweetCount  int
)

// tweetCmd represents the tweet command (parent for tweet operations)
var tweetCmd = &cobra.Command{
	Use:   "tweet",
	Short: "Tweet operations - view, post, like, retweet, etc.",
	Long: `View tweets and perform tweet operations.

Examples:
  xsh tweet view 1234567890                # View a specific tweet
  xsh tweet view 1234567890 --thread       # View tweet with replies as thread tree
  xsh tweet post "Hello world"             # Post a new tweet
  xsh tweet like 1234567890                # Like a tweet
  xsh tweet unlike 1234567890              # Unlike a tweet
  xsh tweet retweet 1234567890             # Retweet
  xsh tweet unretweet 1234567890           # Undo retweet
  xsh tweet bookmark 1234567890            # Bookmark a tweet
  xsh tweet unbookmark 1234567890          # Remove bookmark
  xsh tweet delete 1234567890              # Delete your tweet`,
}

// tweetViewCmd views a tweet and its thread
var tweetViewCmd = &cobra.Command{
	Use:   "view [id]",
	Short: "View a tweet and its thread",
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

		tweets, err := core.GetTweetDetail(client, args[0], tweetCount)
		if err != nil {
			fmt.Println(display.PrintError(fmt.Sprintf("Failed to fetch tweet: %v", err)))
			return
		}

		if len(tweets) == 0 {
			fmt.Println(display.PrintError(fmt.Sprintf("Tweet %s not found", args[0])))
			os.Exit(core.ExitError)
			return
		}

		if isJSONMode() {
			outputJSON(tweets)
		} else if tweetThread {
			// Show thread as tree structure
			fmt.Println(display.FormatThread(tweets, args[0]))
		} else {
			// Default: show only the focal tweet in simple style
			var focal *models.Tweet
			for _, t := range tweets {
				if t.ID == args[0] {
					focal = t
					break
				}
			}
			if focal == nil {
				focal = tweets[0]
			}
			// Show single tweet in simple format (like thread but single item)
			fmt.Println(display.FormatSingleTweet(focal))
		}
	},
}

// tweetPostCmd posts a new tweet
var tweetPostCmd = &cobra.Command{
	Use:   "post [text]",
	Short: "Post a new tweet",
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
					if verbose {
						fmt.Printf("  Note: %s\n", note)
					}
				} else {
					fmt.Println(display.PrintSuccess("Tweet posted!"))
				}
			}
		}
	},
}

// tweetDeleteCmd deletes a tweet
var tweetDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete your tweet",
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

// tweetLikeCmd likes a tweet
var tweetLikeCmd = &cobra.Command{
	Use:   "like [id]",
	Short: "Like a tweet",
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

// tweetUnlikeCmd unlikes a tweet
var tweetUnlikeCmd = &cobra.Command{
	Use:   "unlike [id]",
	Short: "Unlike a tweet",
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

// tweetRetweetCmd retweets a tweet
var tweetRetweetCmd = &cobra.Command{
	Use:   "retweet [id]",
	Short: "Retweet a tweet",
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

// tweetUnretweetCmd undoes a retweet
var tweetUnretweetCmd = &cobra.Command{
	Use:   "unretweet [id]",
	Short: "Undo a retweet",
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

// tweetBookmarkCmd bookmarks a tweet
var tweetBookmarkCmd = &cobra.Command{
	Use:   "bookmark [id]",
	Short: "Bookmark a tweet",
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

// tweetUnbookmarkCmd removes a bookmark
var tweetUnbookmarkCmd = &cobra.Command{
	Use:   "unbookmark [id]",
	Short: "Remove a bookmark",
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

// bookmarksCmd views your bookmarks
var bookmarksCmd = &cobra.Command{
	Use:   "bookmarks",
	Short: "View your bookmarks",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.PrintError(err.Error()))
			return
		}
		defer client.Close()

		count, _ := cmd.Flags().GetInt("count")
		response, err := core.GetBookmarks(client, count, "")
		if err != nil {
			fmt.Println(display.PrintError(fmt.Sprintf("Failed to fetch bookmarks: %v", err)))
			return
		}

		if isJSONMode() {
			outputJSON(response.Tweets)
		} else {
			fmt.Println(display.FormatTweetList(response.Tweets))
		}
	},
}

// extractTweetIDFromResult extracts the tweet ID from API response
func extractTweetIDFromResult(result map[string]interface{}) string {
	if result == nil {
		return ""
	}

	if data, ok := result["data"].(map[string]interface{}); ok {
		if createTweet, ok := data["create_tweet"].(map[string]interface{}); ok {
			if tweetResults, ok := createTweet["tweet_results"].(map[string]interface{}); ok {
				if result, ok := tweetResults["result"].(map[string]interface{}); ok {
					// Try rest_id first
					if restID, ok := result["rest_id"].(string); ok && restID != "" {
						return restID
					}
					// Fallback to legacy.id_str
					if legacy, ok := result["legacy"].(map[string]interface{}); ok {
						if idStr, ok := legacy["id_str"].(string); ok && idStr != "" {
							return idStr
						}
					}
				}
			}
		}
	}

	return ""
}

func init() {
	rootCmd.AddCommand(tweetCmd)
	rootCmd.AddCommand(bookmarksCmd)

	// Add subcommands to tweet
	tweetCmd.AddCommand(tweetViewCmd)
	tweetCmd.AddCommand(tweetPostCmd)
	tweetCmd.AddCommand(tweetDeleteCmd)
	tweetCmd.AddCommand(tweetLikeCmd)
	tweetCmd.AddCommand(tweetUnlikeCmd)
	tweetCmd.AddCommand(tweetRetweetCmd)
	tweetCmd.AddCommand(tweetUnretweetCmd)
	tweetCmd.AddCommand(tweetBookmarkCmd)
	tweetCmd.AddCommand(tweetUnbookmarkCmd)

	tweetViewCmd.Flags().BoolVar(&tweetThread, "thread", false, "Show tweet with all replies as a thread tree")
	tweetViewCmd.Flags().IntVarP(&tweetCount, "count", "n", 20, "Number of tweets/comments to fetch (max 100)")
	tweetPostCmd.Flags().String("reply-to", "", "Tweet ID to reply to")
	tweetPostCmd.Flags().String("quote", "", "Tweet URL to quote")
	tweetDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation")
	bookmarksCmd.Flags().IntP("count", "n", 20, "Number of tweets")
}
