package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/benoitpetit/xsh/models"
	"github.com/benoitpetit/xsh/utils"
)

var (
	exportFormat string
	exportOutput string
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export tweets to various formats",
	Long: `Export tweets to CSV, JSONL, TSV, or Markdown formats.

Supports exporting from feed, search, bookmarks, or user tweets.`,
}

// exportFeedCmd exports timeline
var exportFeedCmd = &cobra.Command{
	Use:   "feed",
	Short: "Export timeline to file",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			return
		}
		defer client.Close()

		feedType, _ := cmd.Flags().GetString("type")
		count, _ := cmd.Flags().GetInt("count")
		pages, _ := cmd.Flags().GetInt("pages")
		filter, _ := cmd.Flags().GetString("filter")

		var allTweets []*models.Tweet
		cursor := ""

		for i := 0; i < pages; i++ {
			response, err := core.GetHomeTimeline(client, feedType, count, cursor)
			if err != nil {
				fmt.Println(display.Error(fmt.Sprintf("Failed to fetch timeline: %v", err)))
				return
			}
			allTweets = append(allTweets, response.Tweets...)
			cursor = response.CursorBottom
			if !response.HasMore {
				break
			}
		}

		tweets := allTweets

		// Apply filter if specified
		if filter != "" {
			cfg, _ := core.LoadConfig()
			tweets = utils.FilterTweets(tweets, filter, 0, count, &cfg.Filter)
		}

		if err := exportTweets(tweets, exportFormat, exportOutput); err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Export failed: %v", err)))
			return
		}

		fmt.Println(display.Success(fmt.Sprintf("Exported %d tweets to %s", len(tweets), exportOutput)))
	},
}

// exportSearchCmd exports search results
var exportSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Export search results to file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			return
		}
		defer client.Close()

		searchType, _ := cmd.Flags().GetString("type")
		count, _ := cmd.Flags().GetInt("count")
		pages, _ := cmd.Flags().GetInt("pages")

		var allTweets []*models.Tweet
		cursor := ""

		for i := 0; i < pages; i++ {
			response, err := core.SearchTweets(client, args[0], searchType, count, cursor)
			if err != nil {
				fmt.Println(display.Error(fmt.Sprintf("Failed to search: %v", err)))
				return
			}
			allTweets = append(allTweets, response.Tweets...)
			cursor = response.CursorBottom
			if !response.HasMore {
				break
			}
		}

		if err := exportTweets(allTweets, exportFormat, exportOutput); err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Export failed: %v", err)))
			return
		}

		fmt.Println(display.Success(fmt.Sprintf("Exported %d tweets to %s", len(allTweets), exportOutput)))
	},
}

// exportBookmarksCmd exports bookmarks
var exportBookmarksCmd = &cobra.Command{
	Use:   "bookmarks",
	Short: "Export bookmarks to file",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			return
		}
		defer client.Close()

		count, _ := cmd.Flags().GetInt("count")

		response, err := core.GetBookmarks(client, count, "")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to fetch bookmarks: %v", err)))
			return
		}

		if err := exportTweets(response.Tweets, exportFormat, exportOutput); err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Export failed: %v", err)))
			return
		}

		fmt.Println(display.Success(fmt.Sprintf("Exported %d bookmarks to %s", len(response.Tweets), exportOutput)))
	},
}

// exportTweets exports tweets to specified format
func exportTweets(tweets []*models.Tweet, format, output string) error {
	var outputWriter *os.File
	var err error

	if output == "-" {
		outputWriter = os.Stdout
	} else {
		outputWriter, err = os.Create(output)
		if err != nil {
			return err
		}
		defer outputWriter.Close()
	}

	switch format {
	case "json":
		return exportJSON(tweets, outputWriter)
	case "jsonl":
		return exportJSONL(tweets, outputWriter)
	case "csv":
		return exportCSV(tweets, outputWriter)
	case "tsv":
		return exportTSV(tweets, outputWriter)
	case "md", "markdown":
		return exportMarkdown(tweets, outputWriter)
	default:
		return fmt.Errorf("unsupported format: %s (use: json, jsonl, csv, tsv, md)", format)
	}
}

func exportJSON(tweets []*models.Tweet, w *os.File) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(tweets)
}

func exportJSONL(tweets []*models.Tweet, w *os.File) error {
	encoder := json.NewEncoder(w)
	for _, tweet := range tweets {
		if err := encoder.Encode(tweet); err != nil {
			return err
		}
	}
	return nil
}

func exportCSV(tweets []*models.Tweet, w *os.File) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Header
	header := []string{"id", "created_at", "author_handle", "author_name", "text", "likes", "retweets", "replies", "views", "tweet_url"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Data
	for _, t := range tweets {
		row := []string{
			t.ID,
			t.CreatedAt.Format("2006-01-02 15:04:05"),
			t.AuthorHandle,
			t.AuthorName,
			cleanTextForCSV(t.Text),
			fmt.Sprintf("%d", t.Engagement.Likes),
			fmt.Sprintf("%d", t.Engagement.Retweets),
			fmt.Sprintf("%d", t.Engagement.Replies),
			fmt.Sprintf("%d", t.Engagement.Views),
			t.TweetURL(),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	return nil
}

func exportTSV(tweets []*models.Tweet, w *os.File) error {
	// Header
	fmt.Fprintln(w, "id\tcreated_at\tauthor_handle\tauthor_name\ttext\tlikes\tretweets\treplies\tviews\ttweet_url")

	// Data
	for _, t := range tweets {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\t%d\t%d\t%d\t%s\n",
			t.ID,
			t.CreatedAt.Format("2006-01-02 15:04:05"),
			t.AuthorHandle,
			t.AuthorName,
			strings.ReplaceAll(t.Text, "\t", " "),
			t.Engagement.Likes,
			t.Engagement.Retweets,
			t.Engagement.Replies,
			t.Engagement.Views,
			t.TweetURL(),
		)
	}
	return nil
}

func exportMarkdown(tweets []*models.Tweet, w *os.File) error {
	fmt.Fprintln(w, "# Tweet Export")
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "*Generated on %s*\n\n", time.Now().Format("2006-01-02"))

	for i, t := range tweets {
		fmt.Fprintf(w, "## %d. @%s\n\n", i+1, t.AuthorHandle)
		fmt.Fprintf(w, "**%s** (@%s) · %s\n\n", t.AuthorName, t.AuthorHandle, t.CreatedAt.Format("Jan 2, 2006"))
		fmt.Fprintf(w, "%s\n\n", t.Text)
		fmt.Fprintf(w, "🔗 [View on X](%s)\n\n", t.TweetURL())
		fmt.Fprintf(w, "❤️ %d · 🔁 %d · 💬 %d\n\n", t.Engagement.Likes, t.Engagement.Retweets, t.Engagement.Replies)
		fmt.Fprintln(w, "---")
		fmt.Fprintln(w, "")
	}
	return nil
}

func cleanTextForCSV(text string) string {
	// Remove newlines and quotes for CSV compatibility
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\"", "'")
	return text
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.AddCommand(exportFeedCmd)
	exportCmd.AddCommand(exportSearchCmd)
	exportCmd.AddCommand(exportBookmarksCmd)

	// Global export flags
	exportCmd.PersistentFlags().StringVarP(&exportFormat, "format", "f", "jsonl", "Export format: json, jsonl, csv, tsv, md")
	exportCmd.PersistentFlags().StringVarP(&exportOutput, "output", "o", "tweets.jsonl", "Output file (use - for stdout)")

	// Feed export flags
	exportFeedCmd.Flags().StringP("type", "t", "for-you", "Timeline type: for-you, following")
	exportFeedCmd.Flags().IntP("count", "n", 100, "Number of tweets per page")
	exportFeedCmd.Flags().IntP("pages", "p", 1, "Number of pages to fetch")
	exportFeedCmd.Flags().String("filter", "", "Filter: all, top, score")

	// Search export flags
	exportSearchCmd.Flags().StringP("type", "t", "Top", "Search type: Top, Latest, Photos, Videos")
	exportSearchCmd.Flags().IntP("count", "n", 100, "Number of tweets per page")
	exportSearchCmd.Flags().IntP("pages", "p", 1, "Number of pages")

	// Bookmarks export flags
	exportBookmarksCmd.Flags().IntP("count", "n", 100, "Number of bookmarks")
}
