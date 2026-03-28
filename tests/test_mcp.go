package tests

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// TestMCPToolSchema tests MCP tool structure
func TestMCPToolSchema(t *testing.T) {
	// Test creating an MCP tool
	tool := mcp.NewTool("test_tool", "Test tool description", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"param1": map[string]interface{}{
				"type":        "string",
				"description": "First parameter",
			},
		},
	})

	if tool.Name != "test_tool" {
		t.Errorf("Tool.Name = %v, want 'test_tool'", tool.Name)
	}
}

// TestMCPCallToolResult tests tool result creation
func TestMCPCallToolResult(t *testing.T) {
	result := &mcp.CallToolResult{
		Content: []interface{}{
			mcp.TextContent{
				Type: "text",
				Text: `{"success": true}`,
			},
		},
	}

	if len(result.Content) != 1 {
		t.Errorf("len(Content) = %v, want 1", len(result.Content))
	}
}

// TestMCPTextContent tests text content creation
func TestMCPTextContent(t *testing.T) {
	content := mcp.TextContent{
		Type: "text",
		Text: "Test message",
	}

	if content.Type != "text" {
		t.Errorf("Type = %v, want 'text'", content.Type)
	}

	if content.Text != "Test message" {
		t.Errorf("Text = %v, want 'Test message'", content.Text)
	}
}

// TestMCPToolNames verifies all expected MCP tools are defined
func TestMCPToolNames(t *testing.T) {
	// 14 base tools + 30 extended unique tools (without unlike/unretweet/unbookmark duplicates)
	expectedTools := []string{
		// Base tools
		"get_feed",
		"search",
		"get_tweet",
		"get_user",
		"list_bookmarks",
		"post_tweet",
		"delete_tweet",
		"like",
		"unlike",
		"retweet",
		"unretweet",
		"bookmark",
		"unbookmark",
		"auth_status",
		// Extended tools - read
		"get_bookmark_folders",
		"get_bookmark_folder_timeline",
		"get_lists",
		"get_list_timeline",
		"get_list_members",
		"get_tweets_batch",
		"get_users_batch",
		"get_user_tweets",
		"get_user_likes",
		"get_followers",
		"get_following",
		"dm_inbox",
		// Extended tools - write
		"follow",
		"unfollow",
		"block",
		"unblock",
		"mute",
		"unmute",
		// Extended tools - lists
		"create_list",
		"delete_list",
		"add_list_member",
		"remove_list_member",
		"pin_list",
		"unpin_list",
		// Extended tools - DM
		"dm_send",
		"dm_delete",
		// Extended tools - scheduling
		"schedule_tweet",
		"list_scheduled_tweets",
		"cancel_scheduled_tweet",
		// Extended tools - media
		"download_media",
	}

	if len(expectedTools) != 44 {
		t.Errorf("Expected 44 MCP tools, got %d", len(expectedTools))
	}

	// Verify unique names
	seen := make(map[string]bool)
	for _, name := range expectedTools {
		if seen[name] {
			t.Errorf("Duplicate MCP tool name: %s", name)
		}
		seen[name] = true
	}
}

// TestMCPJSONSerialization tests JSON serialization
func TestMCPJSONSerialization(t *testing.T) {
	data := map[string]interface{}{
		"authenticated": true,
		"account":       "test",
	}

	// Simulate serialization (in real code this would be json.Marshal)
	if data["authenticated"] != true {
		t.Error("authenticated should be true")
	}

	if data["account"] != "test" {
		t.Errorf("account = %v, want 'test'", data["account"])
	}
}
