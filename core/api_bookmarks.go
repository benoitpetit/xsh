// Package core provides bookmark folder operations for Twitter/X.
package core

import (
	"github.com/benoitpetit/xsh/models"
)

// Hardcoded Query IDs for bookmark folders
const (
	QueryBookmarkFolders      = "i78YDd0Tza-dV4SYs58kRg"
	QueryBookmarkFolderTL     = "hNY7X2xE2N7HVF6Qb_mu6w"
)

// BookmarkFolder represents a bookmark folder
type BookmarkFolder struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GetBookmarkFolders fetches the user's bookmark folders
func GetBookmarkFolders(client *XClient) ([]BookmarkFolder, error) {
	folders := []BookmarkFolder{}
	var cursor string

	// Paginate through results
	for i := 0; i < 10; i++ { // Max 10 pages
		variables := map[string]interface{}{}
		if cursor != "" {
			variables["cursor"] = cursor
		}

		data, err := client.GraphQLGetRaw(QueryBookmarkFolders, "BookmarkFoldersSlice", variables)
		if err != nil {
			return folders, err
		}

		// Parse folders
		pageFolders, nextCursor := parseBookmarkFoldersPage(data)
		folders = append(folders, pageFolders...)

		if nextCursor == "" {
			break
		}
		cursor = nextCursor
	}

	return folders, nil
}

// parseBookmarkFoldersPage parses a single page of bookmark folders
func parseBookmarkFoldersPage(data map[string]interface{}) ([]BookmarkFolder, string) {
	var folders []BookmarkFolder
	var nextCursor string

	// Navigate to items
	viewer, ok := data["data"].(map[string]interface{})
	if !ok {
		return folders, nextCursor
	}

	viewer2, ok := viewer["viewer"].(map[string]interface{})
	if !ok {
		return folders, nextCursor
	}

	userResults, ok := viewer2["user_results"].(map[string]interface{})
	if !ok {
		return folders, nextCursor
	}

	result, ok := userResults["result"].(map[string]interface{})
	if !ok {
		return folders, nextCursor
	}

	bookmarkSlice, ok := result["bookmark_collections_slice"].(map[string]interface{})
	if !ok {
		return folders, nextCursor
	}

	items, ok := bookmarkSlice["items"].([]interface{})
	if ok {
		for _, item := range items {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			folder := BookmarkFolder{
				ID:   getString(itemMap, "id"),
				Name: getString(itemMap, "name"),
			}

			if folder.ID != "" {
				folders = append(folders, folder)
			}
		}
	}

	// Extract next cursor
	if sliceInfo, ok := bookmarkSlice["slice_info"].(map[string]interface{}); ok {
		nextCursor = getString(sliceInfo, "next_cursor")
	}

	return folders, nextCursor
}

// GetBookmarkFolderTimeline fetches tweets from a bookmark folder
func GetBookmarkFolderTimeline(client *XClient, folderID string, count int, cursor string) (*models.TimelineResponse, error) {
	variables := map[string]interface{}{
		"bookmark_collection_id": folderID,
		"count":                  count,
		"includePromotedContent": false,
	}

	if cursor != "" {
		variables["cursor"] = cursor
	}

	data, err := client.GraphQLGetRaw(QueryBookmarkFolderTL, "BookmarkFolderTimeline", variables)
	if err != nil {
		return nil, err
	}

	return extractTweetsFromTimeline(data), nil
}
