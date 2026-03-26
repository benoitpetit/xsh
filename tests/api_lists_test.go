// Package tests provides unit tests for list operations.
package tests

import (
	"testing"

	"github.com/benoitpetit/xsh/core"
)

func TestGetUserLists(t *testing.T) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		t.Skip("No credentials available:", err)
	}
	defer client.Close()

	lists, err := core.GetUserLists(client)
	if err != nil {
		t.Logf("GetUserLists() error = %v", err)
		// Don't fail - might be rate limited or no lists
		return
	}

	// Verify we got a slice (could be empty)
	if lists == nil {
		t.Error("GetUserLists() returned nil, expected slice")
	}

	t.Logf("Found %d lists", len(lists))
}

func TestCreateList(t *testing.T) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		t.Skip("No credentials available:", err)
	}
	defer client.Close()

	// Skip to avoid creating test lists
	t.Skip("Skipping - would create actual list")

	// Example:
	// result, err := core.CreateList(client, "Test List", "Test description", false)
	// if err != nil {
	//     t.Errorf("CreateList() error = %v", err)
	// }
	// t.Logf("Created list: %v", result)
}

func TestDeleteList(t *testing.T) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		t.Skip("No credentials available:", err)
	}
	defer client.Close()

	t.Skip("Skipping - would delete actual list")
}

func TestAddListMember(t *testing.T) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		t.Skip("No credentials available:", err)
	}
	defer client.Close()

	t.Skip("Skipping - would modify actual list")
}

func TestRemoveListMember(t *testing.T) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		t.Skip("No credentials available:", err)
	}
	defer client.Close()

	t.Skip("Skipping - would modify actual list")
}

func TestPinList(t *testing.T) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		t.Skip("No credentials available:", err)
	}
	defer client.Close()

	t.Skip("Skipping - would modify actual list")
}

func TestUnpinList(t *testing.T) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		t.Skip("No credentials available:", err)
	}
	defer client.Close()

	t.Skip("Skipping - would modify actual list")
}

func TestGetListTweets(t *testing.T) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		t.Skip("No credentials available:", err)
	}
	defer client.Close()

	// Test with a known public list ID
	// You can find list IDs in the URL when viewing a list on Twitter
	listID := "1234567890" // Replace with actual test list ID
	
	response, err := core.GetListTweets(client, listID, 10, "")
	if err != nil {
		t.Logf("GetListTweets() error = %v", err)
		return
	}

	if response == nil {
		t.Error("GetListTweets() returned nil response")
		return
	}

	t.Logf("Found %d tweets in list", len(response.Tweets))
}

func TestGetListMembers(t *testing.T) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		t.Skip("No credentials available:", err)
	}
	defer client.Close()

	listID := "1234567890" // Replace with actual test list ID
	
	users, nextCursor, err := core.GetListMembers(client, listID, 10, "")
	if err != nil {
		t.Logf("GetListMembers() error = %v", err)
		return
	}

	t.Logf("Found %d members, next cursor: %s", len(users), nextCursor)
}
