package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
)

var (
	composeFile   string
	composeDryRun bool
)

// composeCmd represents the compose command for creating threads
var composeCmd = &cobra.Command{
	Use:   "compose",
	Short: "Compose a tweet thread interactively",
	Long: `Compose a tweet thread interactively or from a file.

Splits long text into multiple tweets automatically or allows manual entry.`,
	Example: `  # Interactive mode
  xsh compose

  # From file
  xsh compose --file thread.txt

  # From stdin
  cat thread.txt | xsh compose

  # Preview without posting
  xsh compose --file thread.txt --dry-run`,
	Run: func(cmd *cobra.Command, args []string) {
		var tweets []string

		if composeFile != "" {
			// Read from file
			content, err := os.ReadFile(composeFile)
			if err != nil {
				fmt.Printf("Error reading file: %v\n", err)
				os.Exit(1)
				return
			}
			tweets = parseThreadFile(string(content))
		} else {
			// Interactive mode
			tweets = interactiveCompose()
		}

		if len(tweets) == 0 {
			fmt.Println("No tweets to post.")
			return
		}

		// Preview
		fmt.Println()
		previewThread(tweets)

		if composeDryRun {
			fmt.Println()
			fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#8899A6")).Render("Dry run mode - no tweets posted."))
			return
		}

		// Confirm
		fmt.Println()
		fmt.Print("Post this thread? [y/N] ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" {
			fmt.Println("Aborted.")
			return
		}

		// Post thread
		postThread(tweets)
	},
}

func interactiveCompose() []string {
	var tweets []string
	reader := bufio.NewReader(os.Stdin)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#1DA1F2"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#8899A6"))

	fmt.Println(titleStyle.Render("📝 Thread Composer"))
	fmt.Println(helpStyle.Render("Enter your tweets. Leave empty to finish. Use --- for new tweet."))
	fmt.Println()

	tweetNum := 1
	var currentTweet strings.Builder

	for {
		fmt.Printf("Tweet %d: ", tweetNum)
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimRight(line, "\n")

		// Check for separator
		if line == "---" {
			if currentTweet.Len() > 0 {
				tweets = append(tweets, strings.TrimSpace(currentTweet.String()))
				currentTweet.Reset()
				tweetNum++
			}
			continue
		}

		// Check for empty line (end)
		if line == "" && currentTweet.Len() == 0 {
			break
		}

		// Add line to current tweet
		if currentTweet.Len() > 0 {
			currentTweet.WriteString("\n")
		}
		currentTweet.WriteString(line)

		// Check length
		if currentTweet.Len() > 280 {
			warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAD1F"))
			fmt.Printf("  %s (%d/280)\n", warningStyle.Render("⚠️ Warning: Tweet too long"), currentTweet.Len())
		}
	}

	// Don't forget the last tweet
	if currentTweet.Len() > 0 {
		tweets = append(tweets, strings.TrimSpace(currentTweet.String()))
	}

	return tweets
}

func parseThreadFile(content string) []string {
	var tweets []string
	parts := strings.Split(content, "===")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Remove tweet number headers like "Tweet 1" or "1."
		lines := strings.Split(part, "\n")
		var cleanLines []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			// Skip empty lines and headers
			if line == "" {
				continue
			}
			// Skip lines that look like headers
			if matched, _ := regexp.MatchString(`^(Tweet \d+[.:]?|\d+[.:])`, line); matched {
				continue
			}
			cleanLines = append(cleanLines, line)
		}

		if len(cleanLines) > 0 {
			tweets = append(tweets, strings.Join(cleanLines, "\n"))
		}
	}

	// If no === separators found, try splitting by double newline
	if len(tweets) == 0 {
		parts = strings.Split(content, "\n\n")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				tweets = append(tweets, part)
			}
		}
	}

	return tweets
}

func previewThread(tweets []string) {
	titleStyle := lipgloss.NewStyle().Bold(true)
	tweetStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1).Width(60)
	countStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#8899A6"))

	fmt.Println(titleStyle.Render(fmt.Sprintf("Thread Preview (%d tweets):", len(tweets))))
	fmt.Println()

	for i, tweet := range tweets {
		charCount := len([]rune(tweet))
		var countStr string
		if charCount > 280 {
			countStr = lipgloss.NewStyle().Foreground(lipgloss.Color("#F4212E")).Render(fmt.Sprintf("%d/280", charCount))
		} else if charCount > 250 {
			countStr = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAD1F")).Render(fmt.Sprintf("%d/280", charCount))
		} else {
			countStr = lipgloss.NewStyle().Foreground(lipgloss.Color("#00BA7C")).Render(fmt.Sprintf("%d/280", charCount))
		}

		fmt.Printf("Tweet %d %s\n", i+1, countStyle.Render(countStr))
		fmt.Println(tweetStyle.Render(tweet))
		fmt.Println()
	}
}

func postThread(tweets []string) {
	client, err := getClient("")
	if err != nil {
		fmt.Println(display.PrintError(err.Error()))
		return
	}
	defer client.Close()

	var lastTweetID string

	for i, text := range tweets {
		// Truncate if too long (safety)
		if len([]rune(text)) > 280 {
			runes := []rune(text)
			text = string(runes[:277]) + "..."
		}

		result, err := core.CreateTweet(client, text, lastTweetID, "", nil)
		if err != nil {
			fmt.Println(display.PrintError(fmt.Sprintf("Failed to post tweet %d: %v", i+1, err)))
			return
		}

		// Extract tweet ID from result for threading
		if data, ok := result["data"].(map[string]interface{}); ok {
			if createTweet, ok := data["create_tweet"].(map[string]interface{}); ok {
				if tweetResult, ok := createTweet["tweet_results"].(map[string]interface{}); ok {
					if result, ok := tweetResult["result"].(map[string]interface{}); ok {
						if restID, ok := result["rest_id"].(string); ok {
							lastTweetID = restID
						}
					}
				}
			}
		}

		// Fallback: try to get ID from different response format
		if lastTweetID == "" {
			lastTweetID = extractTweetID(result)
		}

		fmt.Printf("✓ Posted tweet %d/%d\n", i+1, len(tweets))
	}

	fmt.Println()
	fmt.Println(display.PrintSuccess(fmt.Sprintf("Thread posted! (%d tweets)", len(tweets))))
}

func extractTweetID(result map[string]interface{}) string {
	// Try various paths to extract tweet ID
	if data, ok := result["data"].(map[string]interface{}); ok {
		for _, key := range []string{"create_tweet", "tweet_v2", "tweet"} {
			if obj, ok := data[key].(map[string]interface{}); ok {
				if id, ok := obj["rest_id"].(string); ok {
					return id
				}
				if id, ok := obj["id"].(string); ok {
					return id
				}
				if id, ok := obj["id_str"].(string); ok {
					return id
				}
			}
		}
	}
	return ""
}



func init() {
	rootCmd.AddCommand(composeCmd)

	composeCmd.Flags().StringVarP(&composeFile, "file", "f", "", "Read thread from file")
	composeCmd.Flags().BoolVar(&composeDryRun, "dry-run", false, "Preview without posting")
}
