// Package tests provides unit tests for bookmark folder operations.
package tests

import (
	"testing"

	"github.com/benoitpetit/xsh/core"
)

func TestGetBookmarkFolders(t *testing.T) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		t.Skip("No credentials available:", err)
	}
	defer client.Close()

	folders, err := core.GetBookmarkFolders(client)
	if err != nil {
		t.Logf("GetBookmarkFolders() error = %v", err)
		return
	}

	if folders == nil {
		t.Error("GetBookmarkFolders() returned nil, expected slice")
	}

	t.Logf("Found %d bookmark folders", len(folders))

	// Validate folder structure
	for _, folder := range folders {
		if folder.ID == "" {
			t.Error("Bookmark folder has empty ID")
		}
		if folder.Name == "" {
			t.Error("Bookmark folder has empty name")
		}
	}
}

func TestGetBookmarkFolderTimeline(t *testing.T) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		t.Skip("No credentials available:", err)
	}
	defer client.Close()

	// First get folders to find a valid ID
	folders, err := core.GetBookmarkFolders(client)
	if err != nil {
		t.Skip("Cannot get folders:", err)
	}

	if len(folders) == 0 {
		t.Skip("No bookmark folders to test with")
	}

	folderID := folders[0].ID

	response, err := core.GetBookmarkFolderTimeline(client, folderID, 10, "")
	if err != nil {
		t.Logf("GetBookmarkFolderTimeline() error = %v", err)
		return
	}

	if response == nil {
		t.Error("GetBookmarkFolderTimeline() returned nil response")
		return
	}

	t.Logf("Found %d tweets in folder %s", len(response.Tweets), folderID)
}
