// Package cmd provides extended MCP tool handlers for xsh.
package cmd

import (
	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/models"
	"github.com/mark3labs/mcp-go/mcp"
)

// ===================== READ TOOLS =====================

func handleGetBookmarkFolders(args map[string]interface{}) (*mcp.CallToolResult, error) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	folders, err := core.GetBookmarkFolders(client)
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(folders), nil
}

func handleGetBookmarkFolderTimeline(args map[string]interface{}) (*mcp.CallToolResult, error) {
	folderID, _ := args["folder_id"].(string)
	count := 20.0
	if c, ok := args["count"].(float64); ok {
		count = c
	}

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	response, err := core.GetBookmarkFolderTimeline(client, folderID, int(count), "")
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(response), nil
}

func handleGetLists(args map[string]interface{}) (*mcp.CallToolResult, error) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	lists, err := core.GetUserLists(client)
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(lists), nil
}

func handleGetListTimeline(args map[string]interface{}) (*mcp.CallToolResult, error) {
	listID, _ := args["list_id"].(string)
	count := 20.0
	if c, ok := args["count"].(float64); ok {
		count = c
	}

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	response, err := core.GetListTweets(client, listID, int(count), "")
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(response), nil
}

func handleGetListMembers(args map[string]interface{}) (*mcp.CallToolResult, error) {
	listID, _ := args["list_id"].(string)
	count := 20.0
	if c, ok := args["count"].(float64); ok {
		count = c
	}

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	users, nextCursor, err := core.GetListMembers(client, listID, int(count), "")
	if err != nil {
		return errorResult(err), nil
	}

	result := map[string]interface{}{
		"users":       users,
		"next_cursor": nextCursor,
	}
	return serializeResult(result), nil
}

func handleGetTweetsBatch(args map[string]interface{}) (*mcp.CallToolResult, error) {
	tweetIDsRaw, ok := args["tweet_ids"].([]interface{})
	if !ok {
		return errorResultString("tweet_ids must be an array"), nil
	}

	tweetIDs := make([]string, 0, len(tweetIDsRaw))
	for _, id := range tweetIDsRaw {
		if s, ok := id.(string); ok {
			tweetIDs = append(tweetIDs, s)
		}
	}

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	tweets, err := core.GetTweetsByIDs(client, tweetIDs)
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(tweets), nil
}

func handleGetUsersBatch(args map[string]interface{}) (*mcp.CallToolResult, error) {
	handlesRaw, ok := args["handles"].([]interface{})
	if !ok {
		return errorResultString("handles must be an array"), nil
	}

	handles := make([]string, 0, len(handlesRaw))
	for _, h := range handlesRaw {
		if s, ok := h.(string); ok {
			handles = append(handles, s)
		}
	}

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	users := make([]*models.User, 0, len(handles))
	for _, handle := range handles {
		user, err := core.GetUserByHandle(client, handle)
		if err == nil && user != nil {
			users = append(users, user)
		}
	}

	return serializeResult(users), nil
}

func handleGetUserTweets(args map[string]interface{}) (*mcp.CallToolResult, error) {
	handle, _ := args["handle"].(string)
	count := 20.0
	if c, ok := args["count"].(float64); ok {
		count = c
	}

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	user, err := core.GetUserByHandle(client, handle)
	if err != nil || user == nil {
		return errorResultString("User not found"), nil
	}

	response, err := core.GetUserTweets(client, user.ID, int(count), "", false)
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(response), nil
}

func handleGetUserLikes(args map[string]interface{}) (*mcp.CallToolResult, error) {
	handle, _ := args["handle"].(string)
	count := 20.0
	if c, ok := args["count"].(float64); ok {
		count = c
	}

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	user, err := core.GetUserByHandle(client, handle)
	if err != nil || user == nil {
		return errorResultString("User not found"), nil
	}

	response, err := core.GetUserLikes(client, user.ID, int(count), "")
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(response), nil
}

func handleGetFollowers(args map[string]interface{}) (*mcp.CallToolResult, error) {
	handle, _ := args["handle"].(string)
	count := 20.0
	if c, ok := args["count"].(float64); ok {
		count = c
	}

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	user, err := core.GetUserByHandle(client, handle)
	if err != nil || user == nil {
		return errorResultString("User not found"), nil
	}

	users, nextCursor, err := core.GetFollowers(client, user.ID, int(count), "")
	if err != nil {
		return errorResult(err), nil
	}

	result := map[string]interface{}{
		"users":       users,
		"next_cursor": nextCursor,
	}
	return serializeResult(result), nil
}

func handleGetFollowing(args map[string]interface{}) (*mcp.CallToolResult, error) {
	handle, _ := args["handle"].(string)
	count := 20.0
	if c, ok := args["count"].(float64); ok {
		count = c
	}

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	user, err := core.GetUserByHandle(client, handle)
	if err != nil || user == nil {
		return errorResultString("User not found"), nil
	}

	users, nextCursor, err := core.GetFollowing(client, user.ID, int(count), "")
	if err != nil {
		return errorResult(err), nil
	}

	result := map[string]interface{}{
		"users":       users,
		"next_cursor": nextCursor,
	}
	return serializeResult(result), nil
}

func handleDMInbox(args map[string]interface{}) (*mcp.CallToolResult, error) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	conversations, err := core.GetDMInbox(client)
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(conversations), nil
}

// ===================== WRITE TOOLS =====================

func handleCreateList(args map[string]interface{}) (*mcp.CallToolResult, error) {
	name, _ := args["name"].(string)
	description, _ := args["description"].(string)
	isPrivate := false
	if p, ok := args["is_private"].(bool); ok {
		isPrivate = p
	}

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	result, err := core.CreateList(client, name, description, isPrivate)
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(result), nil
}

func handleDeleteList(args map[string]interface{}) (*mcp.CallToolResult, error) {
	listID, _ := args["list_id"].(string)

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	result, err := core.DeleteList(client, listID)
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(result), nil
}

func handleAddListMember(args map[string]interface{}) (*mcp.CallToolResult, error) {
	listID, _ := args["list_id"].(string)
	handle, _ := args["handle"].(string)

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	user, err := core.GetUserByHandle(client, handle)
	if err != nil || user == nil {
		return errorResultString("User not found"), nil
	}

	_, err = core.AddListMember(client, listID, user.ID)
	if err != nil {
		return errorResult(err), nil
	}

	return successResult("add_member", handle), nil
}

func handleRemoveListMember(args map[string]interface{}) (*mcp.CallToolResult, error) {
	listID, _ := args["list_id"].(string)
	handle, _ := args["handle"].(string)

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	user, err := core.GetUserByHandle(client, handle)
	if err != nil || user == nil {
		return errorResultString("User not found"), nil
	}

	_, err = core.RemoveListMember(client, listID, user.ID)
	if err != nil {
		return errorResult(err), nil
	}

	return successResult("remove_member", handle), nil
}

func handlePinList(args map[string]interface{}) (*mcp.CallToolResult, error) {
	listID, _ := args["list_id"].(string)

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	_, err = core.PinList(client, listID)
	if err != nil {
		return errorResult(err), nil
	}

	return successResult("pin", listID), nil
}

func handleUnpinList(args map[string]interface{}) (*mcp.CallToolResult, error) {
	listID, _ := args["list_id"].(string)

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	_, err = core.UnpinList(client, listID)
	if err != nil {
		return errorResult(err), nil
	}

	return successResult("unpin", listID), nil
}

func handleDMSend(args map[string]interface{}) (*mcp.CallToolResult, error) {
	handle, _ := args["handle"].(string)
	text, _ := args["text"].(string)

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	user, err := core.GetUserByHandle(client, handle)
	if err != nil || user == nil {
		return errorResultString("User not found"), nil
	}

	result, err := core.SendDM(client, user.ID, text)
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(result), nil
}

func handleDMDelete(args map[string]interface{}) (*mcp.CallToolResult, error) {
	messageID, _ := args["message_id"].(string)

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	result, err := core.DeleteDM(client, messageID)
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(result), nil
}

func handleScheduleTweet(args map[string]interface{}) (*mcp.CallToolResult, error) {
	text, _ := args["text"].(string)
	executeAt := int64(0)
	if e, ok := args["execute_at"].(float64); ok {
		executeAt = int64(e)
	}

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	result, err := core.CreateScheduledTweet(client, text, executeAt, nil)
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(result), nil
}

func handleListScheduledTweets(args map[string]interface{}) (*mcp.CallToolResult, error) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	tweets, err := core.GetScheduledTweets(client)
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(tweets), nil
}

func handleCancelScheduledTweet(args map[string]interface{}) (*mcp.CallToolResult, error) {
	id, _ := args["id"].(string)

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	result, err := core.DeleteScheduledTweet(client, id)
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(result), nil
}

func handleDownloadMedia(args map[string]interface{}) (*mcp.CallToolResult, error) {
	tweetID, _ := args["tweet_id"].(string)
	outputDir := "."
	if o, ok := args["output_dir"].(string); ok {
		outputDir = o
	}

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	files, err := core.DownloadTweetMedia(client, tweetID, outputDir)
	if err != nil {
		return errorResult(err), nil
	}

	result := map[string]interface{}{
		"tweet_id": tweetID,
		"files":    files,
		"count":    len(files),
	}
	return serializeResult(result), nil
}

func handleFollow(args map[string]interface{}) (*mcp.CallToolResult, error) {
	handle, _ := args["handle"].(string)
	client, _ := core.NewXClient(nil, "", "")
	defer client.Close()

	user, err := core.GetUserByHandle(client, handle)
	if err != nil || user == nil {
		return errorResultString("User not found"), nil
	}

	_, err = core.FollowUser(client, user.ID)
	if err != nil {
		return errorResult(err), nil
	}
	return successResult("follow", handle), nil
}

func handleUnfollow(args map[string]interface{}) (*mcp.CallToolResult, error) {
	handle, _ := args["handle"].(string)
	client, _ := core.NewXClient(nil, "", "")
	defer client.Close()

	user, err := core.GetUserByHandle(client, handle)
	if err != nil || user == nil {
		return errorResultString("User not found"), nil
	}

	_, err = core.UnfollowUser(client, user.ID)
	if err != nil {
		return errorResult(err), nil
	}
	return successResult("unfollow", handle), nil
}

func handleBlock(args map[string]interface{}) (*mcp.CallToolResult, error) {
	handle, _ := args["handle"].(string)
	client, _ := core.NewXClient(nil, "", "")
	defer client.Close()

	user, err := core.GetUserByHandle(client, handle)
	if err != nil || user == nil {
		return errorResultString("User not found"), nil
	}

	_, err = core.BlockUser(client, user.ID)
	if err != nil {
		return errorResult(err), nil
	}
	return successResult("block", handle), nil
}

func handleUnblock(args map[string]interface{}) (*mcp.CallToolResult, error) {
	handle, _ := args["handle"].(string)
	client, _ := core.NewXClient(nil, "", "")
	defer client.Close()

	user, err := core.GetUserByHandle(client, handle)
	if err != nil || user == nil {
		return errorResultString("User not found"), nil
	}

	_, err = core.UnblockUser(client, user.ID)
	if err != nil {
		return errorResult(err), nil
	}
	return successResult("unblock", handle), nil
}

func handleMute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	handle, _ := args["handle"].(string)
	client, _ := core.NewXClient(nil, "", "")
	defer client.Close()

	user, err := core.GetUserByHandle(client, handle)
	if err != nil || user == nil {
		return errorResultString("User not found"), nil
	}

	_, err = core.MuteUser(client, user.ID)
	if err != nil {
		return errorResult(err), nil
	}
	return successResult("mute", handle), nil
}

func handleUnmute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	handle, _ := args["handle"].(string)
	client, _ := core.NewXClient(nil, "", "")
	defer client.Close()

	user, err := core.GetUserByHandle(client, handle)
	if err != nil || user == nil {
		return errorResultString("User not found"), nil
	}

	_, err = core.UnmuteUser(client, user.ID)
	if err != nil {
		return errorResult(err), nil
	}
	return successResult("unmute", handle), nil
}

// ===================== JOBS TOOLS =====================

func handleSearchJobs(args map[string]interface{}) (*mcp.CallToolResult, error) {
	keyword, _ := args["keyword"].(string)
	location, _ := args["location"].(string)
	company, _ := args["company"].(string)
	industry, _ := args["industry"].(string)
	count := 25.0
	if c, ok := args["count"].(float64); ok {
		count = c
	}

	// Handle array parameters
	var locationTypes, employmentTypes, seniority []string
	if lt, ok := args["location_type"].([]interface{}); ok {
		for _, v := range lt {
			if s, ok := v.(string); ok {
				locationTypes = append(locationTypes, s)
			}
		}
	}
	if et, ok := args["employment_type"].([]interface{}); ok {
		for _, v := range et {
			if s, ok := v.(string); ok {
				employmentTypes = append(employmentTypes, s)
			}
		}
	}
	if s, ok := args["seniority"].([]interface{}); ok {
		for _, v := range s {
			if str, ok := v.(string); ok {
				seniority = append(seniority, str)
			}
		}
	}

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	response, err := core.SearchJobs(client, keyword, location, locationTypes, employmentTypes, seniority, company, industry, int(count), "")
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(response), nil
}

func handleGetJob(args map[string]interface{}) (*mcp.CallToolResult, error) {
	jobID, _ := args["job_id"].(string)

	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	job, err := core.GetJobDetail(client, jobID)
	if err != nil {
		return errorResult(err), nil
	}
	if job == nil {
		return errorResultString("Job not found"), nil
	}

	return serializeResult(job), nil
}

// ===================== TRENDING TOOLS =====================

func handleGetTrending(args map[string]interface{}) (*mcp.CallToolResult, error) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		return errorResult(err), nil
	}
	defer client.Close()

	// Use WOEID 1 for worldwide trends
	trends, err := core.GetTrends(client, 1)
	if err != nil {
		return errorResult(err), nil
	}

	return serializeResult(trends), nil
}
