/*
xsh - Twitter/X from your terminal. No API keys.

A command-line interface for Twitter/X using cookie-based authentication.

Usage:
  xsh [command]

Available Commands:
  auth         Manage authentication
  feed         View your timeline
  search       Search for tweets
  tweet        View a tweet and its thread
  user         View a user's profile
  post         Post a new tweet
  delete       Delete a tweet
  like         Like a tweet
  unlike       Unlike a tweet
  retweet      Retweet a tweet
  unretweet    Undo a retweet
  bookmark     Bookmark a tweet
  unbookmark   Remove a bookmark
  bookmarks    View your bookmarks
  config       Show current configuration
  mcp          Start the MCP server

Flags:
  --json       Output as JSON
  --account    Account name to use
  -h, --help   Help for xsh

Use "xsh [command] --help" for more information about a command.
*/
package main

import (
	"os"
	"strings"

	"github.com/benoitpetit/xsh/cmd"
	"github.com/benoitpetit/xsh/core"
)

func main() {
	// Run startup check for endpoint obsolescence
	// Only for specific commands that need API access, not for help/config
	if shouldRunStartupCheck() {
		// Run synchronously for critical API commands
		core.RunStartupCheck()
		
		// Start background endpoint monitoring for long-running commands
		if isLongRunningCommand() {
			startEndpointMonitoring()
		}
	}
	
	cmd.Execute()
}

// isLongRunningCommand returns true for commands that might run for a while
func isLongRunningCommand() bool {
	longRunning := []string{"mcp", "feed", "search"}
	for _, arg := range os.Args[1:] {
		for _, cmd := range longRunning {
			if arg == cmd {
				return true
			}
		}
	}
	return false
}

// startEndpointMonitoring starts the background endpoint monitor
func startEndpointMonitoring() {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return
	}
	
	monitor, err := core.NewEndpointMonitor(client, core.Verbose)
	if err != nil {
		client.Close()
		return
	}
	
	monitor.Start()
	// Note: monitor.Stop() should be called when the application exits
	// This is handled via client.Close() chain in normal flow
}

// shouldRunStartupCheck returns true if we should run the endpoint check
func shouldRunStartupCheck() bool {
	// Don't check for help commands
	for _, arg := range os.Args[1:] {
		if arg == "-h" || arg == "--help" || arg == "help" {
			return false
		}
		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") {
			// Short flags like -v, don't trigger check
			continue
		}
	}
	
	// Check if it's a command that needs API
	apiCommands := []string{"feed", "search", "user", "tweet",
		"unlike", "unretweet", "unbookmark",
		"bookmarks", "tweets", "follow", "unfollow", "block", "unblock", "mute", "unmute",
		"lists", "dm", "schedule", "scheduled", "unschedule", "jobs", "trends", "download"}
	
	for _, arg := range os.Args[1:] {
		for _, cmd := range apiCommands {
			if arg == cmd {
				return true
			}
		}
	}
	
	return false
}
