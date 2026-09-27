// Package cmd provides NDJSON streaming output for xsh.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/benoitpetit/xsh/models"
	"github.com/spf13/cobra"
)

var (
	streamInterval int
	streamSource   string
)

// streamCmd emits tweets as NDJSON lines (one JSON object per line)
var streamCmd = &cobra.Command{
	Use:   "stream [source]",
	Short: "Stream tweets as NDJSON for piping",
	Long: `Poll a tweet source at regular intervals and emit new tweets as
newline-delimited JSON (NDJSON). Each line is a complete JSON object.

Perfect for piping to jq, webhook endpoints, log aggregators, or other tools.

Sources:
  feed              Home timeline (default)
  search <query>    Search results
  user <handle>     User's tweets
  notifications     Notification events

Examples:
  xsh stream feed                         # Stream home feed as NDJSON
  xsh stream feed | jq '.text'            # Pipe to jq
  xsh stream search golang --interval 30  # Search every 30s
  xsh stream user @elonmusk               # Stream user tweets
  xsh stream feed | while read line; do   # Webhook
    curl -X POST https://hook.example.com -d "$line"
  done`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		source := args[0]
		sourceArgs := args[1:]

		client, err := getClient("")
		if err != nil {
			fmt.Fprintln(runtimeError(), display.Error(err.Error()))
			abortCommand(core.ExitAuthError)
			return nil
		}
		defer client.Close()

		if streamInterval < 10 {
			streamInterval = 10
		}

		fmt.Fprintln(runtimeError(), display.Muted(fmt.Sprintf("Streaming %s every %ds (Ctrl+C to stop)...", source, streamInterval)))

		// Track seen tweet IDs to only emit new ones
		seen := make(map[string]bool)
		encoder := json.NewEncoder(runtimeOutput())

		// Signal handling
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigCh)

		ticker := time.NewTicker(time.Duration(streamInterval) * time.Second)
		defer ticker.Stop()

		// Run immediately, then on each tick
		if err := streamOnce(client, source, sourceArgs, seen, encoder); err != nil {
			return err
		}

		for {
			select {
			case <-sigCh:
				fmt.Fprintln(runtimeError(), display.Muted("\nStream stopped."))
				return nil
			case <-runtimeContext().Done():
				return runtimeContext().Err()
			case <-ticker.C:
				if err := streamOnce(client, source, sourceArgs, seen, encoder); err != nil {
					return err
				}
			}
		}
	},
}

func streamOnce(client *core.XClient, source string, args []string, seen map[string]bool, encoder *json.Encoder) error {
	switch source {
	case "feed":
		return streamFeed(client, seen, encoder)
	case "search":
		if len(args) == 0 {
			return fmt.Errorf("search requires a query argument")
		}
		return streamSearch(client, strings.Join(args, " "), seen, encoder)
	case "user":
		if len(args) == 0 {
			return fmt.Errorf("user requires a handle argument")
		}
		return streamUser(client, args[0], seen, encoder)
	case "notifications":
		return streamNotifications(client, seen, encoder)
	default:
		return fmt.Errorf("unknown source: %s", source)
	}
}

func streamFeed(client *core.XClient, seen map[string]bool, encoder *json.Encoder) error {
	response, err := core.GetHomeTimeline(client, "ForYou", 20, "")
	if err != nil {
		fmt.Fprintln(runtimeError(), display.Muted(fmt.Sprintf("Feed error: %v", err)))
		return err
	}
	return emitNewTweets(response.Tweets, seen, encoder)
}

func streamSearch(client *core.XClient, query string, seen map[string]bool, encoder *json.Encoder) error {
	response, err := core.SearchTweets(client, query, "Latest", 20, "")
	if err != nil {
		fmt.Fprintln(runtimeError(), display.Muted(fmt.Sprintf("Search error: %v", err)))
		return err
	}
	return emitNewTweets(response.Tweets, seen, encoder)
}

func streamUser(client *core.XClient, handle string, seen map[string]bool, encoder *json.Encoder) error {
	handle = strings.TrimPrefix(handle, "@")
	user, err := core.GetUserByHandle(client, handle)
	if err != nil || user == nil {
		return fmt.Errorf("user lookup error: %w", err)
	}

	response, err := core.GetUserTweets(client, user.ID, 20, "", false)
	if err != nil {
		return fmt.Errorf("user tweets error: %w", err)
	}
	return emitNewTweets(response.Tweets, seen, encoder)
}

func streamNotifications(client *core.XClient, seen map[string]bool, encoder *json.Encoder) error {
	resp, err := core.GetNotifications(client, 20, "")
	if err != nil {
		return fmt.Errorf("notifications error: %w", err)
	}

	for _, n := range resp.Notifications {
		if seen[n.ID] {
			continue
		}
		seen[n.ID] = true

		event := map[string]interface{}{
			"type":       "notification",
			"id":         n.ID,
			"notif_type": n.Type,
			"message":    n.Message,
			"user":       n.UserHandle,
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		}
		if n.TweetText != "" {
			event["tweet_text"] = n.TweetText
		}
		if err := encoder.Encode(event); err != nil {
			return err
		}
	}
	return nil
}

func emitNewTweets(tweets []*models.Tweet, seen map[string]bool, encoder *json.Encoder) error {
	// Emit in reverse chronological order (newest last for streaming)
	newTweets := make([]*models.Tweet, 0)
	for _, t := range tweets {
		if !seen[t.ID] {
			seen[t.ID] = true
			newTweets = append(newTweets, t)
		}
	}

	// Reverse to emit oldest first (natural streaming order)
	for i, j := 0, len(newTweets)-1; i < j; i, j = i+1, j-1 {
		newTweets[i], newTweets[j] = newTweets[j], newTweets[i]
	}

	for _, t := range newTweets {
		event := map[string]interface{}{
			"type":       "tweet",
			"id":         t.ID,
			"text":       t.Text,
			"author":     t.AuthorHandle,
			"author_id":  t.AuthorID,
			"created_at": t.CreatedAt,
			"likes":      t.Engagement.Likes,
			"retweets":   t.Engagement.Retweets,
			"replies":    t.Engagement.Replies,
			"views":      t.Engagement.Views,
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		}

		if len(t.Media) > 0 {
			mediaURLs := make([]string, 0, len(t.Media))
			for _, m := range t.Media {
				if m.URL != "" {
					mediaURLs = append(mediaURLs, m.URL)
				}
			}
			if len(mediaURLs) > 0 {
				event["media"] = mediaURLs
			}
		}

		if t.QuotedTweet != nil {
			event["quoted_tweet_id"] = t.QuotedTweet.ID
		}

		if err := encoder.Encode(event); err != nil {
			return err
		}
	}

	if len(newTweets) > 0 {
		fmt.Fprintln(runtimeError(), display.Muted(fmt.Sprintf("  +%d new items", len(newTweets))))
	}
	return nil
}

func init() {
	rootCmd.AddCommand(streamCmd)
	streamCmd.Flags().IntVar(&streamInterval, "interval", 30, "Polling interval in seconds (minimum 10)")
}
