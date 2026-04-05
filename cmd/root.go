// Package cmd provides the CLI commands for xsh.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/benoitpetit/xsh/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	// Global flags
	jsonOutput    bool
	yamlOutput    bool
	compactMode   bool
	account       string
	verbose       bool
	watchInterval int // seconds, 0 = disabled
)

const logo = `
 ▀▄▀ ▄▀▀ █▄█
 █ █ ▄██ █ █
`

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "xsh",
	Short: "Twitter/X from your terminal. No API keys.",
	Long: logo + `
xsh is a command-line interface for Twitter/X using cookie-based authentication.

No API keys required. Just log in with your browser, and you're in.
Works for humans (rich terminal output) and AI agents (structured JSON).

Get started:
  xsh auth login               # Authenticate with your browser cookies
  xsh feed                     # View your timeline
  xsh tweet view <id>          # View a specific tweet
  xsh search "golang"          # Search for tweets
  xsh user <handle>            # View user profile`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Propagate verbose flag to core package
		if verbose {
			core.Verbose = true
		}
	},
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(core.ExitError)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	rootCmd.PersistentFlags().BoolVar(&yamlOutput, "yaml", false, "Output as YAML")
	rootCmd.PersistentFlags().BoolVarP(&compactMode, "compact", "c", false, "Compact output for AI agents (essential fields only)")
	rootCmd.PersistentFlags().StringVar(&account, "account", "", "Account name to use")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output (show HTTP requests)")
	rootCmd.PersistentFlags().IntVarP(&watchInterval, "watch", "w", 0, "Watch mode: auto-refresh every N seconds (0 = disabled)")
}

// isJSONMode determines if output should be JSON (explicit flag or non-TTY)
func isJSONMode() bool {
	if jsonOutput {
		return true
	}
	// Auto-detect pipe/redirect like Python (but only if yaml/compact are not explicitly set)
	if !yamlOutput && !compactMode {
		stat, _ := os.Stdout.Stat()
		return (stat.Mode() & os.ModeCharDevice) == 0
	}
	return false
}

// isYAMLMode determines if output should be YAML
func isYAMLMode() bool {
	return yamlOutput
}

// isCompactMode determines if output should be compact (for AI agents)
func isCompactMode() bool {
	return compactMode
}

// outputJSON prints data as JSON
func outputJSON(data interface{}) {
	var output interface{}

	switch v := data.(type) {
	case *core.AuthCredentials:
		output = map[string]interface{}{
			"auth_token": v.AuthToken[:8] + "...",
			"ct0":        v.Ct0[:8] + "...",
			"account":    v.AccountName,
		}
	default:
		output = data
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(output)
}

// outputYAML prints data as YAML
func outputYAML(data interface{}) {
	var output interface{}

	switch v := data.(type) {
	case *core.AuthCredentials:
		output = map[string]interface{}{
			"auth_token": v.AuthToken[:8] + "...",
			"ct0":        v.Ct0[:8] + "...",
			"account":    v.AccountName,
		}
	default:
		output = data
	}

	yamlData, err := yaml.Marshal(output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling YAML: %v\n", err)
		return
	}
	os.Stdout.Write(yamlData)
}

// getClient creates an XClient with error handling
func getClient(acc string) (*core.XClient, error) {
	if acc == "" {
		acc = account
	}

	// Load config to get proxy settings
	cfg, err := core.LoadConfig()
	if err != nil {
		cfg = core.DefaultConfig()
	}

	return core.NewXClient(nil, acc, cfg.Network.Proxy)
}

// output handles output in the appropriate format (YAML, JSON, Compact, or human-readable)
// humanOutput should be a function that prints human-readable output
func output(data interface{}, humanOutput func()) {
	if isCompactMode() {
		outputCompact(data)
	} else if isYAMLMode() {
		outputYAML(data)
	} else if isJSONMode() {
		outputJSON(data)
	} else {
		humanOutput()
	}
}

// outputCompact prints minimal data for AI agents (compact JSON with essential fields)
func outputCompact(data interface{}) {
	compact := toCompact(data)
	outputJSON(compact)
}

// isWatchMode returns true if the watch flag is set
func isWatchMode() bool {
	return watchInterval > 0
}

// runWithWatch runs a fetch-and-display function once, or in a polling loop if --watch is set.
// fetchAndDisplay should fetch data and call output() or print directly. It returns an error
// if the fetch fails. The loop catches SIGINT/SIGTERM for graceful exit.
func runWithWatch(fetchAndDisplay func() error) {
	// Run once immediately
	if err := fetchAndDisplay(); err != nil {
		fmt.Println(display.Error(err.Error()))
		os.Exit(core.ExitError)
		return
	}

	if !isWatchMode() {
		return
	}

	// Watch mode: poll at the given interval
	interval := time.Duration(watchInterval) * time.Second
	if interval < 5*time.Second {
		interval = 5 * time.Second // minimum 5s to avoid hammering
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			fmt.Println("\n" + display.Muted("Watch mode stopped."))
			return
		case <-ticker.C:
			// Clear terminal
			fmt.Print("\033[2J\033[H")
			fmt.Println(display.Muted(fmt.Sprintf("Auto-refresh every %ds · %s · Ctrl+C to stop",
				watchInterval, time.Now().Format("15:04:05"))))
			fmt.Println()

			if err := fetchAndDisplay(); err != nil {
				fmt.Println(display.Error(fmt.Sprintf("Refresh failed: %v", err)))
				// Don't exit — keep trying on next tick
			}
		}
	}
}

// toCompact converts data to compact format with only essential fields
func toCompact(data interface{}) interface{} {
	switch v := data.(type) {
	case []*models.Tweet:
		var result []map[string]interface{}
		for _, t := range v {
			result = append(result, map[string]interface{}{
				"id":       t.ID,
				"text":     t.Text,
				"author":   t.AuthorHandle,
				"created":  t.CreatedAt,
				"likes":    t.Engagement.Likes,
				"retweets": t.Engagement.Retweets,
				"replies":  t.Engagement.Replies,
			})
		}
		return result
	case *models.Tweet:
		return map[string]interface{}{
			"id":       v.ID,
			"text":     v.Text,
			"author":   v.AuthorHandle,
			"created":  v.CreatedAt,
			"likes":    v.Engagement.Likes,
			"retweets": v.Engagement.Retweets,
			"replies":  v.Engagement.Replies,
		}
	case []*models.User:
		var result []map[string]interface{}
		for _, u := range v {
			result = append(result, map[string]interface{}{
				"id":        u.ID,
				"handle":    u.Handle,
				"name":      u.Name,
				"followers": u.FollowersCount,
				"following": u.FollowingCount,
			})
		}
		return result
	case *models.User:
		return map[string]interface{}{
			"id":        v.ID,
			"handle":    v.Handle,
			"name":      v.Name,
			"bio":       v.Bio,
			"followers": v.FollowersCount,
			"following": v.FollowingCount,
			"verified":  v.Verified,
		}
	case []*models.DMConversation:
		var result []map[string]interface{}
		for _, c := range v {
			result = append(result, map[string]interface{}{
				"id":           c.ID,
				"type":         c.Type,
				"participants": len(c.Participants),
				"last_message": c.LastMessage,
				"unread":       c.Unread,
			})
		}
		return result
	case *models.DMConversation:
		return map[string]interface{}{
			"id":           v.ID,
			"type":         v.Type,
			"participants": len(v.Participants),
			"last_message": v.LastMessage,
			"unread":       v.Unread,
		}
	case []*models.Job:
		var result []map[string]interface{}
		for _, j := range v {
			result = append(result, map[string]interface{}{
				"id":       j.ID,
				"title":    j.Title,
				"company":  j.Company.Name,
				"location": j.Location,
				"salary":   j.FormattedSalary,
			})
		}
		return result
	case *models.Job:
		return map[string]interface{}{
			"id":          v.ID,
			"title":       v.Title,
			"company":     v.Company.Name,
			"location":    v.Location,
			"salary":      v.FormattedSalary,
			"employment":  v.EmploymentType,
			"description": v.Description,
		}
	case []*models.Trend:
		var result []map[string]interface{}
		for _, t := range v {
			result = append(result, map[string]interface{}{
				"name":   t.Name,
				"volume": t.TweetVolume,
				"rank":   t.Rank,
			})
		}
		return result
	case *models.Trend:
		return map[string]interface{}{
			"name":   v.Name,
			"query":  v.Query,
			"volume": v.TweetVolume,
			"rank":   v.Rank,
		}
	case *core.Community:
		return map[string]interface{}{
			"id":          v.ID,
			"name":        v.Name,
			"description": v.Description,
			"members":     v.MemberCount,
			"role":        v.Role,
		}
	case *core.Space:
		return map[string]interface{}{
			"id":           v.ID,
			"title":        v.Title,
			"state":        v.State,
			"participants": v.ParticipantCount,
			"started_at":   v.StartedAt,
		}
	case []core.Space:
		var result []map[string]interface{}
		for _, s := range v {
			result = append(result, map[string]interface{}{
				"id":           s.ID,
				"title":        s.Title,
				"state":        s.State,
				"participants": s.ParticipantCount,
			})
		}
		return result
	case []core.Notification:
		var result []map[string]interface{}
		for _, n := range v {
			result = append(result, map[string]interface{}{
				"id":      n.ID,
				"type":    n.Type,
				"message": n.Message,
				"user":    n.UserHandle,
			})
		}
		return result
	case *models.TweetAnalytics:
		return map[string]interface{}{
			"total_tweets":    v.TotalTweets,
			"total_views":     v.TotalViews,
			"total_likes":     v.TotalLikes,
			"total_retweets":  v.TotalRetweets,
			"total_replies":   v.TotalReplies,
			"total_bookmarks": v.TotalBookmarks,
			"avg_views":       v.AvgViews,
			"avg_likes":       v.AvgLikes,
			"engagement_rate": v.EngagementRate,
			"media_breakdown": v.MediaBreakdown,
		}
	case map[string]interface{}:
		// For simple API responses (post/delete/etc), return as-is but filter to essential fields
		compact := make(map[string]interface{})
		if id, ok := v["id"]; ok {
			compact["id"] = id
		}
		if success, ok := v["success"]; ok {
			compact["success"] = success
		}
		if message, ok := v["message"]; ok {
			compact["message"] = message
		}
		if len(compact) == 0 {
			return v // Return original if no recognized fields
		}
		return compact
	default:
		return data
	}
}
