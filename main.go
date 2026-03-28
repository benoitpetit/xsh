/*
xsh - Twitter/X from your terminal. No API keys.

A command-line interface for Twitter/X using cookie-based authentication.

Usage:
  xsh [command]

Available Commands:
  auth          Manage authentication (login, logout, import, accounts, switch, status, whoami)
  feed          View your timeline
  search        Search for tweets
  tweet         View a tweet and its thread
  user          View a user's profile
  post          Post a new tweet
  delete        Delete a tweet
  like          Like a tweet
  unlike        Unlike a tweet
  retweet       Retweet a tweet
  unretweet     Undo a retweet
  bookmark      Bookmark a tweet
  unbookmark    Remove a bookmark
  bookmarks     View your bookmarks
  tweets        List tweets from a user
  follow        Follow a user
  unfollow      Unfollow a user
  block         Block a user
  unblock       Unblock a user
  mute          Mute a user
  unmute        Unmute a user
  lists         Manage lists (list, create, delete, add, remove)
  dm            Direct messages (list, view, send)
  schedule      Schedule tweets
  scheduled     List scheduled tweets
  unschedule    Cancel a scheduled tweet
  jobs          Search job listings
  trends        View trending topics
  download      Download media from tweets
  endpoints     Manage API endpoints (list, check, refresh, status, update, reset)
  auto-update   Automatically update obsolete endpoints
  config        Show current configuration
  doctor        Diagnose common issues
  mcp           Start the MCP server
  version       Show version information

Flags:
  --json        Output as JSON
  --yaml        Output as YAML
  --compact, -c Compact output (essential fields only)
  --account     Account name to use
  -v, --verbose Verbose output (show HTTP requests)
  -h, --help    Help for xsh

Use "xsh [command] --help" for more information about a command.
*/
package main

import (
	"os"
	"strings"

	"github.com/benoitpetit/xsh/cmd"
	"github.com/benoitpetit/xsh/core"
)

// globalMonitor holds the endpoint monitor for cleanup
var globalMonitor *core.EndpointMonitor

// stopEndpointMonitoring stops the global monitor if running
func stopEndpointMonitoring() {
	if globalMonitor != nil {
		globalMonitor.Stop()
	}
}

func main() {
	// Ensure cleanup on exit
	defer stopEndpointMonitoring()
	
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
	
	globalMonitor = monitor
	monitor.Start()
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
