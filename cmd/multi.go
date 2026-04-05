// Package cmd provides multi-account batch operations for xsh.
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/spf13/cobra"
)

// multiCmd runs the same operation across multiple stored accounts
var multiCmd = &cobra.Command{
	Use:   "multi <action> [args...]",
	Short: "Run an action across multiple accounts",
	Long: `Execute the same action using all stored accounts (or a subset).
Useful for monitoring feeds, checking notifications, or posting from
multiple accounts in a single command.

Supported actions:
  whoami          Verify all accounts are valid
  feed            Fetch latest from home timeline per account
  notifications   Check notifications per account
  tweet <text>    Post a tweet from each account
  like <id>       Like a tweet from each account

Examples:
  xsh multi whoami                             # Verify all accounts
  xsh multi feed                               # Latest feed per account
  xsh multi notifications                      # Notifications per account
  xsh multi tweet "Hello from all accounts"    # Post from all
  xsh multi like 1234567890                    # Like from all
  xsh multi feed --accounts work,personal      # Only specific accounts`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		action := args[0]
		actionArgs := args[1:]

		accountFilter, _ := cmd.Flags().GetString("accounts")
		accounts, err := resolveMultiAccounts(accountFilter)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to list accounts: %v", err)))
			os.Exit(core.ExitError)
			return
		}

		if len(accounts) == 0 {
			fmt.Println(display.Error("No accounts configured. Run 'xsh auth login' first."))
			os.Exit(core.ExitAuthError)
			return
		}

		if len(accounts) == 1 {
			fmt.Println(display.Warning("Only 1 account found. Use 'xsh auth login --account <name>' to add more."))
		}

		var results []multiResult

		for i, acc := range accounts {
			fmt.Printf("\r%s [%d/%d] %s: %s...",
				display.Muted("..."), i+1, len(accounts), acc, action)

			result := runMultiAction(acc, action, actionArgs)
			results = append(results, result)
		}
		fmt.Println()

		output(results, func() {
			displayMultiResults(results, action)
		})
	},
}

type multiResult struct {
	Account string      `json:"account"`
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func resolveMultiAccounts(filter string) ([]string, error) {
	all, err := core.ListAccounts()
	if err != nil {
		return nil, err
	}

	if filter == "" {
		return all, nil
	}

	wanted := make(map[string]bool)
	for _, a := range strings.Split(filter, ",") {
		wanted[strings.TrimSpace(a)] = true
	}

	var result []string
	for _, a := range all {
		if wanted[a] {
			result = append(result, a)
		}
	}
	return result, nil
}

func runMultiAction(acc, action string, args []string) multiResult {
	client, err := getClient(acc)
	if err != nil {
		return multiResult{Account: acc, Success: false, Error: err.Error()}
	}
	defer client.Close()

	switch action {
	case "whoami":
		return multiWhoami(client, acc)
	case "feed":
		return multiFeed(client, acc)
	case "notifications":
		return multiNotifications(client, acc)
	case "tweet":
		if len(args) == 0 {
			return multiResult{Account: acc, Success: false, Error: "tweet text required"}
		}
		return multiTweet(client, acc, strings.Join(args, " "))
	case "like":
		if len(args) == 0 {
			return multiResult{Account: acc, Success: false, Error: "tweet ID required"}
		}
		return multiLike(client, acc, args[0])
	default:
		return multiResult{Account: acc, Success: false, Error: fmt.Sprintf("unknown action: %s", action)}
	}
}

func multiWhoami(client *core.XClient, acc string) multiResult {
	creds, err := core.GetCredentials(acc)
	if err != nil || creds == nil {
		errMsg := "not authenticated"
		if err != nil {
			errMsg = err.Error()
		}
		return multiResult{Account: acc, Success: false, Error: errMsg}
	}

	// Verify by fetching timeline
	_, err = core.GetHomeTimeline(client, "ForYou", 1, "")
	if err != nil {
		return multiResult{Account: acc, Success: false, Error: fmt.Sprintf("auth invalid: %v", err)}
	}

	return multiResult{
		Account: acc,
		Success: true,
		Data: map[string]interface{}{
			"name":      creds.AccountName,
			"api_valid": true,
		},
	}
}

func multiFeed(client *core.XClient, acc string) multiResult {
	response, err := core.GetHomeTimeline(client, "ForYou", 5, "")
	if err != nil {
		return multiResult{Account: acc, Success: false, Error: err.Error()}
	}

	summary := map[string]interface{}{
		"tweet_count": len(response.Tweets),
	}
	if len(response.Tweets) > 0 {
		first := response.Tweets[0]
		text := first.Text
		if len(text) > 80 {
			text = text[:77] + "..."
		}
		summary["latest"] = fmt.Sprintf("@%s: %s", first.AuthorHandle, text)
	}
	return multiResult{Account: acc, Success: true, Data: summary}
}

func multiNotifications(client *core.XClient, acc string) multiResult {
	resp, err := core.GetNotifications(client, 10, "")
	if err != nil {
		return multiResult{Account: acc, Success: false, Error: err.Error()}
	}
	return multiResult{
		Account: acc,
		Success: true,
		Data: map[string]interface{}{
			"count": len(resp.Notifications),
		},
	}
}

func multiTweet(client *core.XClient, acc, text string) multiResult {
	result, err := core.CreateTweet(client, text, "", "", nil, "")
	if err != nil {
		return multiResult{Account: acc, Success: false, Error: err.Error()}
	}
	tweetID := ""
	if id, ok := result["id"].(string); ok {
		tweetID = id
	}
	return multiResult{
		Account: acc,
		Success: true,
		Data: map[string]interface{}{
			"tweet_id": tweetID,
		},
	}
}

func multiLike(client *core.XClient, acc, tweetID string) multiResult {
	_, err := core.LikeTweet(client, tweetID)
	if err != nil {
		return multiResult{Account: acc, Success: false, Error: err.Error()}
	}
	return multiResult{
		Account: acc,
		Success: true,
		Data: map[string]interface{}{
			"liked": tweetID,
		},
	}
}

func displayMultiResults(results []multiResult, action string) {
	fmt.Println(display.Subtitle(fmt.Sprintf("Multi-Account · %s · %d accounts", action, len(results))))
	fmt.Println()

	succeeded, failed := 0, 0
	for _, r := range results {
		if r.Success {
			succeeded++
		} else {
			failed++
		}
	}

	for _, r := range results {
		status := display.Success("OK")
		if !r.Success {
			status = display.Error("FAIL")
		}

		fmt.Printf("  %s  %-20s", status, r.Account)

		if r.Error != "" {
			fmt.Printf("  %s", display.Muted(r.Error))
		} else if r.Data != nil {
			// Print summary data inline
			switch d := r.Data.(type) {
			case map[string]interface{}:
				parts := make([]string, 0)
				for k, v := range d {
					parts = append(parts, fmt.Sprintf("%s=%v", k, v))
				}
				fmt.Printf("  %s", display.Muted(strings.Join(parts, " ")))
			}
		}
		fmt.Println()
	}

	fmt.Println()
	summary := fmt.Sprintf("%d succeeded", succeeded)
	if failed > 0 {
		summary += fmt.Sprintf(", %d failed", failed)
	}
	fmt.Println(display.Muted(summary))
}

func init() {
	rootCmd.AddCommand(multiCmd)
	multiCmd.Flags().String("accounts", "", "Comma-separated list of account names (default: all)")
}
