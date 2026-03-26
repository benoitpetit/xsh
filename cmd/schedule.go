// Package cmd provides scheduled tweet commands for xsh.
package cmd

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/spf13/cobra"
)

// scheduleCmd schedules a tweet
var scheduleCmd = &cobra.Command{
	Use:   "schedule <text>",
	Short: "Schedule a tweet for future posting",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		text := args[0]
		atStr, _ := cmd.Flags().GetString("at")

		// Parse the schedule time
		scheduleTime, err := parseScheduleTime(atStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing time: %v\n", err)
			os.Exit(1)
		}

		if scheduleTime.Before(time.Now()) {
			fmt.Fprintf(os.Stderr, "Error: Schedule time must be in the future\n")
			os.Exit(1)
		}

		executeAt := scheduleTime.Unix()

		client, err := getClient("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		result, err := core.CreateScheduledTweet(client, text, executeAt, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		output(result, func() {
			fmt.Printf("✓ Tweet scheduled for %s\n", scheduleTime.Format("2006-01-02 15:04"))
		})
	},
}

// scheduledCmd lists scheduled tweets
var scheduledCmd = &cobra.Command{
	Use:   "scheduled",
	Short: "List scheduled tweets",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := getClient("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		tweets, err := core.GetScheduledTweets(client)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		output(tweets, func() {
			fmt.Println(display.FormatScheduledTweets(tweets))
		})
	},
}

// unscheduleCmd cancels a scheduled tweet
var unscheduleCmd = &cobra.Command{
	Use:   "unschedule <scheduled-tweet-id>",
	Short: "Cancel a scheduled tweet",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scheduledTweetID := args[0]

		client, err := getClient("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		_, err = core.DeleteScheduledTweet(client, scheduledTweetID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		output(map[string]string{
			"action":             "unschedule",
			"scheduled_tweet_id": scheduledTweetID,
			"status":             "success",
		}, func() {
			fmt.Printf("✓ Cancelled scheduled tweet %s\n", scheduledTweetID)
		})
	},
}

// parseScheduleTime parses a time string in various formats
func parseScheduleTime(timeStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	// Try Unix timestamp
	if unixTs, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
		return time.Unix(unixTs, 0), nil
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}

func init() {
	rootCmd.AddCommand(scheduleCmd)
	rootCmd.AddCommand(scheduledCmd)
	rootCmd.AddCommand(unscheduleCmd)

	scheduleCmd.Flags().String("at", "", "Schedule time (ISO format or 'YYYY-MM-DD HH:MM')")
	scheduleCmd.MarkFlagRequired("at")
}
