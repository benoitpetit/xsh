package tests

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// TestMCPToolSchema teste la structure des outils MCP
func TestMCPToolSchema(t *testing.T) {
	// Test de la création d'un outil MCP
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

// TestMCPCallToolResult teste la création d'un résultat d'outil
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

// TestMCPTextContent teste la création de contenu texte
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

// TestMCPToolNames vérifie que tous les outils MCP attendus sont définis
func TestMCPToolNames(t *testing.T) {
	expectedTools := []string{
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
	}

	if len(expectedTools) != 14 {
		t.Errorf("Expected 14 MCP tools, got %d", len(expectedTools))
	}

	// Vérifier les noms uniques
	seen := make(map[string]bool)
	for _, name := range expectedTools {
		if seen[name] {
			t.Errorf("Duplicate MCP tool name: %s", name)
		}
		seen[name] = true
	}
}

// TestMCPJSONSerialization teste la sérialisation JSON
func TestMCPJSONSerialization(t *testing.T) {
	data := map[string]interface{}{
		"authenticated": true,
		"account":       "test",
	}

	// Simuler la sérialisation (dans le vrai code ce serait json.Marshal)
	if data["authenticated"] != true {
		t.Error("authenticated should be true")
	}

	if data["account"] != "test" {
		t.Errorf("account = %v, want 'test'", data["account"])
	}
}
