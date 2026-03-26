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
	tweetThread    bool
	tweetCount     int
	exportArticle  string
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
	Long: `View a tweet and its thread. 
	
For tweets containing articles (long-form content), use --export to save as Markdown:
  xsh tweet view <id> --export article.md`,
	Args: cobra.ExactArgs(1),
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

		// Get focal tweet
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

		// Check for article and handle export
		if exportArticle != "" {
			articleData, err := core.GetArticle(client, args[0])
			if err != nil || articleData == nil {
				fmt.Println(display.PrintWarning("No article found in this tweet"))
				os.Exit(core.ExitError)
				return
			}

			if err := core.ExportArticleToFile(articleData, focal, exportArticle); err != nil {
				fmt.Println(display.PrintError(fmt.Sprintf("Failed to export article: %v", err)))
				os.Exit(core.ExitError)
				return
			}

			fmt.Println(display.PrintSuccess(fmt.Sprintf("Article exported to %s", exportArticle)))
			return
		}

		// Check if this is an article tweet (for display purposes)
		articleData, _ := core.GetArticle(client, args[0])

		if isJSONMode() {
			outputJSON(tweets)
		} else if isYAMLMode() {
			outputYAML(tweets)
		} else if tweetThread {
			// Show thread as tree structure
			fmt.Println(display.FormatThread(tweets, args[0]))
		} else if articleData != nil {
			// Display article
			metadata := utils.ExtractArticleMetadata(articleData)
			contentMD := utils.ArticleToMarkdown(articleData)
			fmt.Println(display.FormatArticle(
				models.GetString(metadata, "title"),
				focal.AuthorHandle,
				contentMD,
				focal.Engagement,
			))
		} else {
			// Default: show only the focal tweet in simple style
			fmt.Println(display.FormatSingleTweet(focal))
		}
	},
}

// tweetPostCmd posts a new tweet
var tweetPostCmd = &cobra.Command{
	Use:   "post [text]",
	Short: "Post a new tweet",
	Args:  cobra.ExactArgs(1),
	Long: `Post a new tweet, optionally with images (up to 4).

Examples:
  xsh tweet post "Hello world!"
  xsh tweet post "Check this out" --image photo.jpg
  xsh tweet post "My photos" -i img1.jpg -i img2.jpg -i img3.jpg
  xsh tweet post "Replying" --reply-to 1234567890
  xsh tweet post "Quoting" --quote https://x.com/user/status/1234567890`,
	Run: func(cmd *cobra.Command, args []string) {
		text, valid := utils.ValidateTweetTextWithLimit(args[0], 280)
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
		images, _ := cmd.Flags().GetStringArray("image")

		if replyTo != "" && !utils.ValidateTweetID(replyTo) {
			fmt.Println(display.PrintError(fmt.Sprintf("Invalid reply-to tweet ID: %s", replyTo)))
			os.Exit(core.ExitError)
			return
		}

		// Validate max images
		if len(images) > 4 {
			fmt.Println(display.PrintError(fmt.Sprintf("Too many images: %d (max 4)", len(images))))
			os.Exit(core.ExitError)
			return
		}

		// Upload media if provided
		var mediaIDs []string
		if len(images) > 0 {
			if verbose {
				fmt.Printf("Uploading %d image(s)...\n", len(images))
			}
			for _, imgPath := range images {
				mediaID, err := core.UploadMediaFile(client, imgPath)
				if err != nil {
					fmt.Println(display.PrintError(fmt.Sprintf("Failed to upload image %s: %v", imgPath, err)))
					os.Exit(core.ExitError)
					return
				}
				mediaIDs = append(mediaIDs, mediaID)
				if verbose {
					fmt.Printf("  Uploaded: %s -> %s\n", imgPath, mediaID[:8]+"...")
				}
			}
		}

		result, err := core.CreateTweet(client, text, replyTo, quote, mediaIDs)
		if err != nil {
			fmt.Println(display.PrintError(fmt.Sprintf("Failed to post tweet: %v", err)))
			return
		}

		outputData := result
		if len(mediaIDs) > 0 {
			outputData["media_ids"] = mediaIDs
		}
		output(outputData, func() {
			tweetID := extractTweetIDFromResult(result)
			if tweetID != "" {
				suffix := ""
				if len(mediaIDs) > 0 {
					suffix = fmt.Sprintf(" with %d image(s)", len(mediaIDs))
				}
				fmt.Println(display.PrintSuccess(fmt.Sprintf("Tweet posted%s! ID: %s", suffix, tweetID)))
				fmt.Printf("  URL: https://x.com/i/web/status/%s\n", tweetID)
			} else {
				// Check if it's a "likely success" response (empty ID but no error)
				if note, ok := result["_note"].(string); ok && note != "" {
					suffix := ""
					if len(mediaIDs) > 0 {
						suffix = fmt.Sprintf(" with %d image(s)", len(mediaIDs))
					}
					fmt.Println(display.PrintSuccess(fmt.Sprintf("Tweet posted%s successfully!", suffix)))
					if verbose {
						fmt.Printf("  Note: %s\n", note)
					}
				} else {
					suffix := ""
					if len(mediaIDs) > 0 {
						suffix = fmt.Sprintf(" with %d image(s)", len(mediaIDs))
					}
					fmt.Println(display.PrintSuccess(fmt.Sprintf("Tweet posted%s!", suffix)))
				}
			}
		})
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

		output(result, func() {
			fmt.Println(display.PrintSuccess(fmt.Sprintf("Deleted tweet %s", args[0])))
		})
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

		output(result, func() {
			fmt.Println(display.PrintSuccess(fmt.Sprintf("Liked tweet %s", args[0])))
		})
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

		output(result, func() {
			fmt.Println(display.PrintSuccess(fmt.Sprintf("Unliked tweet %s", args[0])))
		})
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

		output(result, func() {
			fmt.Println(display.PrintSuccess(fmt.Sprintf("Retweeted %s", args[0])))
		})
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

		output(result, func() {
			fmt.Println(display.PrintSuccess(fmt.Sprintf("Unretweeted %s", args[0])))
		})
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

		output(result, func() {
			fmt.Println(display.PrintSuccess(fmt.Sprintf("Bookmarked tweet %s", args[0])))
		})
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

		output(result, func() {
			fmt.Println(display.PrintSuccess(fmt.Sprintf("Unbookmarked tweet %s", args[0])))
		})
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

		output(response.Tweets, func() {
			fmt.Println(display.FormatTweetList(response.Tweets))
		})
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
	tweetViewCmd.Flags().StringVarP(&exportArticle, "export", "o", "", "Export article to Markdown file (for tweets containing long-form articles)")
	tweetPostCmd.Flags().String("reply-to", "", "Tweet ID to reply to")
	tweetPostCmd.Flags().String("quote", "", "Tweet URL to quote")
	tweetPostCmd.Flags().StringArrayP("image", "i", nil, "Image to attach (can be used multiple times, max 4)")
	tweetDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation")
	bookmarksCmd.Flags().IntP("count", "n", 20, "Number of tweets")
}
