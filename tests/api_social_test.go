// Package tests provides unit tests for social actions.
package tests

import (
	"testing"

	"github.com/benoitpetit/xsh/core"
)

func TestFollowUser(t *testing.T) {
	// This test requires valid credentials
	// Skip if no auth available
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		t.Skip("No credentials available:", err)
	}
	defer client.Close()

	// Test with a known user ID (Elon Musk)
	// Note: This will actually follow the user!
	// In real tests, use a test account
	_ = client
	t.Skip("Skipping actual API call - would follow user")
}

func TestUnfollowUser(t *testing.T) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		t.Skip("No credentials available:", err)
	}
	defer client.Close()

	t.Skip("Skipping actual API call")
}

func TestBlockUser(t *testing.T) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		t.Skip("No credentials available:", err)
	}
	defer client.Close()

	t.Skip("Skipping actual API call")
}

func TestUnblockUser(t *testing.T) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		t.Skip("No credentials available:", err)
	}
	defer client.Close()

	t.Skip("Skipping actual API call")
}

func TestMuteUser(t *testing.T) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		t.Skip("No credentials available:", err)
	}
	defer client.Close()

	t.Skip("Skipping actual API call")
}

func TestUnmuteUser(t *testing.T) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		t.Skip("No credentials available:", err)
	}
	defer client.Close()

	t.Skip("Skipping actual API call")
}
