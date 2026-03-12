package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	"github.com/benoitpetit/xsh/core"
)

// mcpCmd represents the mcp command
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the MCP server (stdio transport)",
	Long: `Launch the xsh MCP server for use with any MCP-compatible client.

This starts the server using stdio transport for communication with MCP clients
like Claude Desktop, Claude Code, or other MCP-compatible tools.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runMCPServer(); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}

func runMCPServer() error {
	// Create MCP server
	s := server.NewMCPServer(
		"xsh",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Register tools
	registerReadTools(s)
	registerWriteTools(s)
	registerInfoTools(s)

	// Start server with stdio transport
	return server.ServeStdio(s)
}

// createTool creates a tool with the given name, description and parameters
func createTool(name, description string, params map[string]interface{}) mcp.Tool {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": params,
	}
	
	// Extract required params
	var required []string
	for k, v := range params {
		if param, ok := v.(map[string]interface{}); ok {
			if req, ok := param["required"].(bool); ok && req {
				required = append(required, k)
				delete(param, "required")
			}
		}
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	
	return mcp.NewTool(name, description, schema)
}

func registerReadTools(s *server.MCPServer) {
	// get_feed
	getFeedTool := createTool("get_feed", "Fetch the home timeline", map[string]interface{}{
		"type": map[string]interface{}{
			"type":        "string",
			"description": "Timeline type — 'for-you' or 'following'",
			"default":     "for-you",
		},
		"count": map[string]interface{}{
			"type":        "number",
			"description": "Number of tweets to fetch (max 100)",
			"default":     20,
		},
	})
	s.AddTool(getFeedTool, handleGetFeed)

	// search
	searchTool := createTool("search", "Search for tweets", map[string]interface{}{
		"query": map[string]interface{}{
			"type":        "string",
			"description": "Search query string",
			"required":    true,
		},
		"type": map[string]interface{}{
			"type":        "string",
			"description": "Search type — 'Top', 'Latest', 'Photos', or 'Videos'",
			"default":     "Top",
		},
		"count": map[string]interface{}{
			"type":        "number",
			"description": "Number of results to return (max 100)",
			"default":     20,
		},
	})
	s.AddTool(searchTool, handleSearch)

	// get_tweet
	getTweetTool := createTool("get_tweet", "Fetch a tweet by ID, optionally with its conversation thread", map[string]interface{}{
		"id": map[string]interface{}{
			"type":        "string",
			"description": "The tweet ID",
			"required":    true,
		},
		"thread": map[string]interface{}{
			"type":        "boolean",
			"description": "If true, return the full conversation thread",
			"default":     false,
		},
	})
	s.AddTool(getTweetTool, handleGetTweet)

	// get_user
	getUserTool := createTool("get_user", "Fetch a user profile by handle", map[string]interface{}{
		"handle": map[string]interface{}{
			"type":        "string",
			"description": "The user's screen name (without @)",
			"required":    true,
		},
	})
	s.AddTool(getUserTool, handleGetUser)

	// list_bookmarks
	listBookmarksTool := createTool("list_bookmarks", "Fetch bookmarked tweets", map[string]interface{}{
		"count": map[string]interface{}{
			"type":        "number",
			"description": "Number of bookmarks to fetch (max 100)",
			"default":     20,
		},
	})
	s.AddTool(listBookmarksTool, handleListBookmarks)
}

func registerWriteTools(s *server.MCPServer) {
	// post_tweet
	postTweetTool := createTool("post_tweet", "Post a new tweet", map[string]interface{}{
		"text": map[string]interface{}{
			"type":        "string",
			"description": "The tweet text content",
			"required":    true,
		},
		"reply_to": map[string]interface{}{
			"type":        "string",
			"description": "Tweet ID to reply to (optional)",
		},
		"quote": map[string]interface{}{
			"type":        "string",
			"description": "URL of tweet to quote (optional)",
		},
	})
	s.AddTool(postTweetTool, handlePostTweet)

	// delete_tweet
	deleteTweetTool := createTool("delete_tweet", "Delete a tweet by ID", map[string]interface{}{
		"id": map[string]interface{}{
			"type":        "string",
			"description": "The tweet ID to delete",
			"required":    true,
		},
	})
	s.AddTool(deleteTweetTool, handleDeleteTweet)

	// like
	likeTool := createTool("like", "Like a tweet", map[string]interface{}{
		"id": map[string]interface{}{
			"type":        "string",
			"description": "The tweet ID to like",
			"required":    true,
		},
	})
	s.AddTool(likeTool, handleLike)

	// unlike
	unlikeTool := createTool("unlike", "Unlike a tweet", map[string]interface{}{
		"id": map[string]interface{}{
			"type":        "string",
			"description": "The tweet ID to unlike",
			"required":    true,
		},
	})
	s.AddTool(unlikeTool, handleUnlike)

	// retweet
	retweetTool := createTool("retweet", "Retweet a tweet", map[string]interface{}{
		"id": map[string]interface{}{
			"type":        "string",
			"description": "The tweet ID to retweet",
			"required":    true,
		},
	})
	s.AddTool(retweetTool, handleRetweet)

	// unretweet
	unretweetTool := createTool("unretweet", "Undo a retweet", map[string]interface{}{
		"id": map[string]interface{}{
			"type":        "string",
			"description": "The tweet ID to unretweet",
			"required":    true,
		},
	})
	s.AddTool(unretweetTool, handleUnretweet)

	// bookmark
	bookmarkTool := createTool("bookmark", "Bookmark a tweet", map[string]interface{}{
		"id": map[string]interface{}{
			"type":        "string",
			"description": "The tweet ID to bookmark",
			"required":    true,
		},
	})
	s.AddTool(bookmarkTool, handleBookmark)

	// unbookmark
	unbookmarkTool := createTool("unbookmark", "Remove a bookmark from a tweet", map[string]interface{}{
		"id": map[string]interface{}{
			"type":        "string",
			"description": "The tweet ID to unbookmark",
			"required":    true,
		},
	})
	s.AddTool(unbookmarkTool, handleUnbookmark)
}

func registerInfoTools(s *server.MCPServer) {
	// auth_status
	authStatusTool := createTool("auth_status", "Check authentication status and return credential info", map[string]interface{}{})
	s.AddTool(authStatusTool, handleAuthStatus)
}

// Tool handlers - signature: func(arguments map[string]interface{}) (*mcp.CallToolResult, error)

func handleGetFeed(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	timelineType := "for-you"
	if t, ok := arguments["type"].(string); ok {
		timelineType = t
	}

	count := 20.0
	if c, ok := arguments["count"].(float64); ok {
		count = c
	}

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	response, err := core.GetHomeTimeline(client, timelineType, int(count), "")
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(response), nil
}

func handleSearch(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	query, _ := arguments["query"].(string)

	searchType := "Top"
	if t, ok := arguments["type"].(string); ok {
		searchType = t
	}

	count := 20.0
	if c, ok := arguments["count"].(float64); ok {
		count = c
	}

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	response, err := core.SearchTweets(client, query, searchType, int(count), "")
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(response), nil
}

func handleGetTweet(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	tweetID, _ := arguments["id"].(string)

	showThread := false
	if t, ok := arguments["thread"].(bool); ok {
		showThread = t
	}

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	tweets, err := core.GetTweetDetail(client, tweetID, 20)
	if err != nil {
		return errorResult(err), nil
	}

	if len(tweets) == 0 {
		return errorResultString("Tweet not found"), nil
	}

	if showThread {
		return serializeResult(tweets), nil
	}

	// Return just the focal tweet
	for _, t := range tweets {
		if t.ID == tweetID {
			return serializeResult(t), nil
		}
	}
	return serializeResult(tweets[0]), nil
}

func handleGetUser(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	handle, _ := arguments["handle"].(string)

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	user, err := core.GetUserByHandle(client, handle)
	if err != nil {
		return errorResult(err), nil
	}

	if user == nil {
		return errorResultString("User not found"), nil
	}

	return serializeResult(user), nil
}

func handleListBookmarks(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	count := 20.0
	if c, ok := arguments["count"].(float64); ok {
		count = c
	}

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	response, err := core.GetBookmarks(client, int(count), "")
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(response), nil
}

func handlePostTweet(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	text, _ := arguments["text"].(string)

	replyTo := ""
	if r, ok := arguments["reply_to"].(string); ok {
		replyTo = r
	}

	quote := ""
	if q, ok := arguments["quote"].(string); ok {
		quote = q
	}

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	result, err := core.CreateTweet(client, text, replyTo, quote, nil)
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(result), nil
}

func handleDeleteTweet(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	tweetID, _ := arguments["id"].(string)

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	result, err := core.DeleteTweet(client, tweetID)
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(result), nil
}

func handleLike(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	tweetID, _ := arguments["id"].(string)

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	result, err := core.LikeTweet(client, tweetID)
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(result), nil
}

func handleUnlike(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	tweetID, _ := arguments["id"].(string)

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	result, err := core.UnlikeTweet(client, tweetID)
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(result), nil
}

func handleRetweet(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	tweetID, _ := arguments["id"].(string)

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	result, err := core.Retweet(client, tweetID)
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(result), nil
}

func handleUnretweet(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	tweetID, _ := arguments["id"].(string)

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	result, err := core.Unretweet(client, tweetID)
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(result), nil
}

func handleBookmark(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	tweetID, _ := arguments["id"].(string)

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	result, err := core.BookmarkTweet(client, tweetID)
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(result), nil
}

func handleUnbookmark(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	tweetID, _ := arguments["id"].(string)

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	result, err := core.UnbookmarkTweet(client, tweetID)
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(result), nil
}

func handleAuthStatus(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	creds, err := core.GetCredentials("")
	if err != nil {
		result := map[string]interface{}{
			"authenticated": false,
			"valid":         false,
			"error":         err.Error(),
		}
		return serializeResult(result), nil
	}

	result := map[string]interface{}{
		"authenticated": true,
		"valid":         creds.IsValid(),
		"account":       creds.AccountName,
		"has_cookies":   len(creds.Cookies) > 0,
	}
	return serializeResult(result), nil
}

// Helper functions

func serializeResult(data interface{}) *mcp.CallToolResult {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return errorResult(err)
	}
	return &mcp.CallToolResult{
		Content: []interface{}{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonData),
			},
		},
	}
}

func errorResult(err error) *mcp.CallToolResult {
	result := map[string]interface{}{
		"error": err.Error(),
		"type":  fmt.Sprintf("%T", err),
	}
	return serializeResult(result)
}

func errorResultString(msg string) *mcp.CallToolResult {
	result := map[string]interface{}{
		"error": msg,
		"type":  "Error",
	}
	return serializeResult(result)
}
