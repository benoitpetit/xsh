// Package cmd provides quote tweet and thread commands for xsh.
package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/benoitpetit/xsh/models"
	"github.com/benoitpetit/xsh/utils"
	"github.com/spf13/cobra"
)

// quotesCmd fetches quote tweets for a given tweet ID
var quotesCmd = &cobra.Command{
	Use:   "quotes [tweet-id]",
	Short: "View quote tweets of a tweet",
	Long: `View tweets that quote a specific tweet.

Examples:
  xsh quotes 1234567890             # View quote tweets
  xsh quotes 1234567890 -n 50       # View up to 50 quote tweets`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		count, _ := strconv.Atoi(cmd.Flag("count").Value.String())

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		response, err := core.GetQuoteTweets(client, args[0], count, "")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to fetch quote tweets: %v", err)))
			os.Exit(core.ExitError)
		}

		output(response.Tweets, func() {
			fmt.Println(display.FormatTweetList(response.Tweets))
		})
	},
}

// threadCmd views a full conversation thread
var threadCmd = &cobra.Command{
	Use:   "thread [tweet-id]",
	Short: "View a full conversation thread",
	Long: `View the full reply thread for a tweet, showing the conversation tree.

Examples:
  xsh thread 1234567890             # View full conversation thread
  xsh thread 1234567890 -n 50       # Fetch up to 50 replies`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		count, _ := strconv.Atoi(cmd.Flag("count").Value.String())

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		tweets, err := core.GetTweetDetail(client, args[0], count)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to fetch thread: %v", err)))
			os.Exit(core.ExitError)
		}

		output(tweets, func() {
			fmt.Println(display.FormatThread(tweets, args[0]))
		})
	},
}

// pinnedCmd views a user's pinned tweet
var pinnedCmd = &cobra.Command{
	Use:   "pinned [handle]",
	Short: "View a user's pinned tweet",
	Long: `View the pinned tweet of a Twitter/X user.

Examples:
  xsh pinned elonmusk               # View @elonmusk's pinned tweet
  xsh pinned @jack                   # View @jack's pinned tweet`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		handle, valid := utils.ValidateTwitterHandle(args[0])
		if !valid {
			fmt.Println(display.Error(fmt.Sprintf("Invalid Twitter handle: %s", args[0])))
			os.Exit(core.ExitError)
		}

		user, err := core.GetUserByHandle(client, handle)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to fetch user: %v", err)))
			os.Exit(core.ExitError)
		}
		if user == nil {
			fmt.Println(display.Error(fmt.Sprintf("User @%s not found", handle)))
			os.Exit(core.ExitError)
		}

		if user.PinnedTweetID == "" {
			fmt.Println(display.Muted(fmt.Sprintf("@%s has no pinned tweet", handle)))
			return
		}

		tweets, err := core.GetTweetDetail(client, user.PinnedTweetID, 1)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to fetch pinned tweet: %v", err)))
			os.Exit(core.ExitError)
		}

		if len(tweets) == 0 {
			fmt.Println(display.Muted("Pinned tweet not found or unavailable"))
			return
		}

		// Show the focal (pinned) tweet
		output(tweets[0], func() {
			fmt.Println(display.Subtitle(fmt.Sprintf("Pinned tweet by @%s", handle)))
			fmt.Println(display.FormatTweet(tweets[0], "", true, false))
		})
	},
}

// unrollCmd unrolls a thread into a clean document
var unrollCmd = &cobra.Command{
	Use:   "unroll [tweet-id]",
	Short: "Unroll a thread into a clean readable document",
	Long: `Unroll a thread (self-reply chain by the same author) into a clean,
concatenated document. Only tweets by the original author are included.

Examples:
  xsh unroll 1234567890             # Unroll the thread
  xsh unroll 1234567890 -n 100      # Fetch more tweets in the thread`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		count, _ := strconv.Atoi(cmd.Flag("count").Value.String())

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		tweets, err := core.GetTweetDetail(client, args[0], count)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to fetch thread: %v", err)))
			os.Exit(core.ExitError)
		}

		if len(tweets) == 0 {
			fmt.Println(display.Muted("Thread not found or empty"))
			return
		}

		// Find the focal tweet and determine the thread author
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
		author := focal.AuthorHandle

		// Build the self-reply chain: tweets by the same author that form
		// a reply chain starting from the conversation root.
		// First, build a map of tweet ID -> tweet
		tweetMap := make(map[string]*models.Tweet)
		for _, t := range tweets {
			tweetMap[t.ID] = t
		}

		// Collect tweets by the author, ordered by their position in the chain
		var chain []*models.Tweet
		// Start from the conversation root and follow self-replies
		// Find all tweets by the author
		authorTweets := make(map[string]*models.Tweet)
		for _, t := range tweets {
			if t.AuthorHandle == author {
				authorTweets[t.ID] = t
			}
		}

		// Find the root of the thread (author tweet with no parent by the same author)
		roots := []*models.Tweet{}
		for _, t := range authorTweets {
			if _, parentIsAuthor := authorTweets[t.ReplyToID]; !parentIsAuthor {
				roots = append(roots, t)
			}
		}

		// Build chain by following reply-to links
		// Build children map
		children := make(map[string]*models.Tweet)
		for _, t := range authorTweets {
			if _, ok := authorTweets[t.ReplyToID]; ok {
				children[t.ReplyToID] = t
			}
		}

		// Walk from each root
		if len(roots) > 0 {
			current := roots[0]
			for current != nil {
				chain = append(chain, current)
				current = children[current.ID]
			}
		}

		// Fallback: if chain building failed, just use all author tweets in order
		if len(chain) == 0 {
			for _, t := range tweets {
				if t.AuthorHandle == author {
					chain = append(chain, t)
				}
			}
		}

		output(chain, func() {
			fmt.Println(display.FormatUnrolledThread(chain, author))
		})
	},
}

func init() {
	rootCmd.AddCommand(quotesCmd)
	rootCmd.AddCommand(threadCmd)
	rootCmd.AddCommand(pinnedCmd)
	rootCmd.AddCommand(unrollCmd)

	quotesCmd.Flags().IntP("count", "n", 20, "Number of quote tweets to fetch")
	threadCmd.Flags().IntP("count", "n", 40, "Number of replies to fetch")
	unrollCmd.Flags().IntP("count", "n", 100, "Max tweets to fetch for unrolling")
}
