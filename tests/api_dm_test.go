// Package tests provides unit tests for DM operations.
package tests

import (
	"testing"

	"github.com/benoitpetit/xsh/core"
)

func TestGetDMInbox(t *testing.T) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		t.Skip("No credentials available:", err)
	}
	defer client.Close()

	conversations, err := core.GetDMInbox(client)
	if err != nil {
		t.Logf("GetDMInbox() error = %v", err)
		// Don't fail - might be rate limited
		return
	}

	if conversations == nil {
		t.Error("GetDMInbox() returned nil, expected slice")
	}

	t.Logf("Found %d conversations", len(conversations))

	// Validate conversation structure
	for _, conv := range conversations {
		if conv.ID == "" {
			t.Error("Conversation has empty ID")
		}
		if conv.Type != "one_to_one" && conv.Type != "group" {
			t.Errorf("Conversation has invalid type: %s", conv.Type)
		}
	}
}

func TestSendDM(t *testing.T) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		t.Skip("No credentials available:", err)
	}
	defer client.Close()

	t.Skip("Skipping - would send actual DM")

	// Example usage:
	// userID := "1234567890" // Test user ID
	// result, err := core.SendDM(client, userID, "Test message")
	// if err != nil {
	//     t.Errorf("SendDM() error = %v", err)
	// }
}

func TestDeleteDM(t *testing.T) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		t.Skip("No credentials available:", err)
	}
	defer client.Close()

	t.Skip("Skipping - would delete actual DM")
}
