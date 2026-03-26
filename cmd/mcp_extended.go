// Package cmd provides extended MCP tools for xsh.
package cmd

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerExtendedTools registers all the extended MCP tools
func registerExtendedTools(s *server.MCPServer) {
	registerReadToolsExtended(s)
	registerWriteToolsExtended(s)
	registerListTools(s)
	registerDMTools(s)
	registerScheduledTools(s)
	registerMediaTools(s)
	registerJobsTools(s)
	registerTrendingTools(s)
}

func registerReadToolsExtended(s *server.MCPServer) {
	// get_bookmark_folders
	s.AddTool(createTool("get_bookmark_folders", "Fetch the user's bookmark folders", map[string]interface{}{}), handleGetBookmarkFolders)

	// get_bookmark_folder_timeline
	s.AddTool(createTool("get_bookmark_folder_timeline", "Fetch tweets from a bookmark folder", map[string]interface{}{
		"folder_id": map[string]interface{}{"type": "string", "description": "The bookmark folder ID", "required": true},
		"count":     map[string]interface{}{"type": "number", "description": "Number of tweets to fetch", "default": 20},
	}), handleGetBookmarkFolderTimeline)

	// get_lists
	s.AddTool(createTool("get_lists", "Fetch the user's lists", map[string]interface{}{}), handleGetLists)

	// get_list_timeline
	s.AddTool(createTool("get_list_timeline", "Fetch tweets from a list", map[string]interface{}{
		"list_id": map[string]interface{}{"type": "string", "description": "The list ID", "required": true},
		"count":   map[string]interface{}{"type": "number", "description": "Number of tweets to fetch", "default": 20},
	}), handleGetListTimeline)

	// get_list_members
	s.AddTool(createTool("get_list_members", "Fetch members of a list", map[string]interface{}{
		"list_id": map[string]interface{}{"type": "string", "description": "The list ID", "required": true},
		"count":   map[string]interface{}{"type": "number", "description": "Number of members to fetch", "default": 20},
	}), handleGetListMembers)

	// get_tweets_batch
	s.AddTool(createTool("get_tweets_batch", "Fetch multiple tweets by their IDs", map[string]interface{}{
		"tweet_ids": map[string]interface{}{"type": "array", "description": "List of tweet IDs to fetch", "required": true},
	}), handleGetTweetsBatch)

	// get_users_batch
	s.AddTool(createTool("get_users_batch", "Fetch multiple users by their handles", map[string]interface{}{
		"handles": map[string]interface{}{"type": "array", "description": "List of user handles (without @)", "required": true},
	}), handleGetUsersBatch)

	// get_user_tweets
	s.AddTool(createTool("get_user_tweets", "Fetch tweets posted by a user", map[string]interface{}{
		"handle": map[string]interface{}{"type": "string", "description": "The user's handle (without @)", "required": true},
		"count":  map[string]interface{}{"type": "number", "description": "Number of tweets to fetch", "default": 20},
	}), handleGetUserTweets)

	// get_user_likes
	s.AddTool(createTool("get_user_likes", "Fetch tweets liked by a user", map[string]interface{}{
		"handle": map[string]interface{}{"type": "string", "description": "The user's handle (without @)", "required": true},
		"count":  map[string]interface{}{"type": "number", "description": "Number of tweets to fetch", "default": 20},
	}), handleGetUserLikes)

	// get_followers
	s.AddTool(createTool("get_followers", "Fetch followers of a user", map[string]interface{}{
		"handle": map[string]interface{}{"type": "string", "description": "The user's handle (without @)", "required": true},
		"count":  map[string]interface{}{"type": "number", "description": "Number of followers to fetch", "default": 20},
	}), handleGetFollowers)

	// get_following
	s.AddTool(createTool("get_following", "Fetch users followed by a user", map[string]interface{}{
		"handle": map[string]interface{}{"type": "string", "description": "The user's handle (without @)", "required": true},
		"count":  map[string]interface{}{"type": "number", "description": "Number of users to fetch", "default": 20},
	}), handleGetFollowing)

	// dm_inbox
	s.AddTool(createTool("dm_inbox", "Fetch DM inbox conversations", map[string]interface{}{}), handleDMInbox)
}

func registerWriteToolsExtended(s *server.MCPServer) {
	// unlike
	s.AddTool(createTool("unlike", "Unlike a tweet", map[string]interface{}{
		"id": map[string]interface{}{"type": "string", "description": "The tweet ID to unlike", "required": true},
	}), handleUnlike)

	// unretweet
	s.AddTool(createTool("unretweet", "Undo a retweet", map[string]interface{}{
		"id": map[string]interface{}{"type": "string", "description": "The tweet ID to unretweet", "required": true},
	}), handleUnretweet)

	// unbookmark
	s.AddTool(createTool("unbookmark", "Remove a bookmark", map[string]interface{}{
		"id": map[string]interface{}{"type": "string", "description": "The tweet ID to unbookmark", "required": true},
	}), handleUnbookmark)

	// follow
	s.AddTool(createTool("follow", "Follow a user", map[string]interface{}{
		"handle": map[string]interface{}{"type": "string", "description": "The user's handle (without @)", "required": true},
	}), handleFollow)

	// unfollow
	s.AddTool(createTool("unfollow", "Unfollow a user", map[string]interface{}{
		"handle": map[string]interface{}{"type": "string", "description": "The user's handle (without @)", "required": true},
	}), handleUnfollow)

	// block
	s.AddTool(createTool("block", "Block a user", map[string]interface{}{
		"handle": map[string]interface{}{"type": "string", "description": "The user's handle (without @)", "required": true},
	}), handleBlock)

	// unblock
	s.AddTool(createTool("unblock", "Unblock a user", map[string]interface{}{
		"handle": map[string]interface{}{"type": "string", "description": "The user's handle (without @)", "required": true},
	}), handleUnblock)

	// mute
	s.AddTool(createTool("mute", "Mute a user", map[string]interface{}{
		"handle": map[string]interface{}{"type": "string", "description": "The user's handle (without @)", "required": true},
	}), handleMute)

	// unmute
	s.AddTool(createTool("unmute", "Unmute a user", map[string]interface{}{
		"handle": map[string]interface{}{"type": "string", "description": "The user's handle (without @)", "required": true},
	}), handleUnmute)
}

func registerListTools(s *server.MCPServer) {
	// create_list
	s.AddTool(createTool("create_list", "Create a new list", map[string]interface{}{
		"name":        map[string]interface{}{"type": "string", "description": "Name for the new list", "required": true},
		"description": map[string]interface{}{"type": "string", "description": "List description"},
		"is_private":  map[string]interface{}{"type": "boolean", "description": "Whether the list should be private", "default": false},
	}), handleCreateList)

	// delete_list
	s.AddTool(createTool("delete_list", "Delete a list", map[string]interface{}{
		"list_id": map[string]interface{}{"type": "string", "description": "The list ID to delete", "required": true},
	}), handleDeleteList)

	// add_list_member
	s.AddTool(createTool("add_list_member", "Add a member to a list", map[string]interface{}{
		"list_id": map[string]interface{}{"type": "string", "description": "The list ID", "required": true},
		"handle":  map[string]interface{}{"type": "string", "description": "The user's handle to add", "required": true},
	}), handleAddListMember)

	// remove_list_member
	s.AddTool(createTool("remove_list_member", "Remove a member from a list", map[string]interface{}{
		"list_id": map[string]interface{}{"type": "string", "description": "The list ID", "required": true},
		"handle":  map[string]interface{}{"type": "string", "description": "The user's handle to remove", "required": true},
	}), handleRemoveListMember)

	// pin_list
	s.AddTool(createTool("pin_list", "Pin a list", map[string]interface{}{
		"list_id": map[string]interface{}{"type": "string", "description": "The list ID to pin", "required": true},
	}), handlePinList)

	// unpin_list
	s.AddTool(createTool("unpin_list", "Unpin a list", map[string]interface{}{
		"list_id": map[string]interface{}{"type": "string", "description": "The list ID to unpin", "required": true},
	}), handleUnpinList)
}

func registerDMTools(s *server.MCPServer) {
	// dm_send
	s.AddTool(createTool("dm_send", "Send a direct message", map[string]interface{}{
		"handle": map[string]interface{}{"type": "string", "description": "The user's handle (without @)", "required": true},
		"text":   map[string]interface{}{"type": "string", "description": "The message text", "required": true},
	}), handleDMSend)

	// dm_delete
	s.AddTool(createTool("dm_delete", "Delete a DM message", map[string]interface{}{
		"message_id": map[string]interface{}{"type": "string", "description": "The message ID to delete", "required": true},
	}), handleDMDelete)
}

func registerScheduledTools(s *server.MCPServer) {
	// schedule_tweet
	s.AddTool(createTool("schedule_tweet", "Schedule a tweet for future posting", map[string]interface{}{
		"text":       map[string]interface{}{"type": "string", "description": "The tweet text content", "required": true},
		"execute_at": map[string]interface{}{"type": "number", "description": "Unix timestamp for when to post", "required": true},
	}), handleScheduleTweet)

	// list_scheduled_tweets
	s.AddTool(createTool("list_scheduled_tweets", "List all scheduled tweets", map[string]interface{}{}), handleListScheduledTweets)

	// cancel_scheduled_tweet
	s.AddTool(createTool("cancel_scheduled_tweet", "Cancel a scheduled tweet", map[string]interface{}{
		"id": map[string]interface{}{"type": "string", "description": "The scheduled tweet ID to cancel", "required": true},
	}), handleCancelScheduledTweet)
}

func registerMediaTools(s *server.MCPServer) {
	// download_media
	s.AddTool(createTool("download_media", "Download media from a tweet", map[string]interface{}{
		"tweet_id":   map[string]interface{}{"type": "string", "description": "The tweet ID to download media from", "required": true},
		"output_dir": map[string]interface{}{"type": "string", "description": "Directory to save files to", "default": "."},
	}), handleDownloadMedia)
}

func registerJobsTools(s *server.MCPServer) {
	// search_jobs
	s.AddTool(createTool("search_jobs", "Search for job listings on X/Twitter", map[string]interface{}{
		"keyword":         map[string]interface{}{"type": "string", "description": "Search keyword (e.g. 'data engineer', 'product manager')"},
		"location":        map[string]interface{}{"type": "string", "description": "Location filter (e.g. 'Paris', 'New York')"},
		"location_type":   map[string]interface{}{"type": "array", "description": "Location types: remote, onsite, hybrid"},
		"employment_type": map[string]interface{}{"type": "array", "description": "Employment types: full_time, part_time, contract, internship"},
		"seniority":       map[string]interface{}{"type": "array", "description": "Seniority levels: entry_level, mid_level, senior"},
		"company":         map[string]interface{}{"type": "string", "description": "Company name filter"},
		"industry":        map[string]interface{}{"type": "string", "description": "Industry filter"},
		"count":           map[string]interface{}{"type": "number", "description": "Number of results", "default": 25},
	}), handleSearchJobs)

	// get_job
	s.AddTool(createTool("get_job", "Get detailed information about a specific job listing", map[string]interface{}{
		"job_id": map[string]interface{}{"type": "string", "description": "The job listing ID", "required": true},
	}), handleGetJob)
}

func registerTrendingTools(s *server.MCPServer) {
	// get_trending
	s.AddTool(createTool("get_trending", "Get currently trending topics on Twitter/X", map[string]interface{}{}), handleGetTrending)
}

// successResult creates a success result for MCP
func successResult(action, target string) *mcp.CallToolResult {
	result := map[string]string{
		"action": action,
		"target": target,
		"status": "success",
	}
	return serializeResult(result)
}
