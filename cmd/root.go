// Package cmd provides the CLI commands for xsh.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/benoitpetit/xsh/core"
)

var (
	// Global flags
	jsonOutput bool
	account    string
	verbose    bool
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
  xsh login                    # Authenticate with your browser cookies
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
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	rootCmd.PersistentFlags().StringVar(&account, "account", "", "Account name to use")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output (show HTTP requests)")
}

// isJSONMode determines if output should be JSON (explicit flag or non-TTY)
func isJSONMode() bool {
	if jsonOutput {
		return true
	}
	// Auto-detect pipe/redirect like Python
	stat, _ := os.Stdout.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
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

// getClient creates an XClient with error handling
func getClient(acc string) (*core.XClient, error) {
	if acc == "" {
		acc = account
	}
	return core.NewXClient(nil, acc, "")
}
