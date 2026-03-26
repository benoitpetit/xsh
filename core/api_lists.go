// Package core provides list management operations for Twitter/X.
package core

import (
	"strings"

	"github.com/benoitpetit/xsh/models"
)

// ListInfo represents information about a Twitter list
type ListInfo struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	MemberCount     int    `json:"member_count"`
	SubscriberCount int    `json:"subscriber_count"`
	Mode            string `json:"mode"` // "Private" or "Public"
	IsPinned        bool   `json:"is_pinned"`
}

// GetUserLists fetches the authenticated user's lists
func GetUserLists(client *XClient) ([]ListInfo, error) {
	lists := []ListInfo{}
	variables := map[string]interface{}{
		"count": 100,
	}

	data, err := client.GraphQLGet("ListsManagementPageTimeline", variables)
	if err != nil {
		return lists, err
	}

	parsed := parseUserLists(data)
	if parsed != nil {
		lists = parsed
	}
	return lists, nil
}

// parseUserLists parses the ListsManagementPageTimeline response
func parseUserLists(data map[string]interface{}) []ListInfo {
	var lists []ListInfo

	// Navigate to the timeline
	viewer, ok := data["data"].(map[string]interface{})
	if !ok {
		return lists
	}

	listMgmt, ok := viewer["viewer"].(map[string]interface{})
	if !ok {
		return lists
	}

	timeline, ok := listMgmt["list_management_timeline"].(map[string]interface{})
	if !ok {
		return lists
	}

	timeline2, ok := timeline["timeline"].(map[string]interface{})
	if !ok {
		return lists
	}

	instructions, ok := timeline2["instructions"].([]interface{})
	if !ok {
		return lists
	}

	for _, inst := range instructions {
		instruction, ok := inst.(map[string]interface{})
		if !ok {
			continue
		}

		entries, ok := instruction["entries"].([]interface{})
		if !ok {
			continue
		}

		for _, entry := range entries {
			entryMap, ok := entry.(map[string]interface{})
			if !ok {
				continue
			}

			entryID, _ := entryMap["entryId"].(string)
			// Only parse owned/subscribed lists, skip "Discover new Lists"
			if !containsSubstring(entryID, "owned-subscribed") {
				continue
			}

			content, ok := entryMap["content"].(map[string]interface{})
			if !ok {
				continue
			}

			items, ok := content["items"].([]interface{})
			if !ok {
				continue
			}

			for _, item := range items {
				itemMap, ok := item.(map[string]interface{})
				if !ok {
					continue
				}

				innerItem, ok := itemMap["item"].(map[string]interface{})
				if !ok {
					continue
				}

				itemContent, ok := innerItem["itemContent"].(map[string]interface{})
				if !ok {
					continue
				}

				listResult, ok := itemContent["list"].(map[string]interface{})
				if !ok {
					continue
				}

				listInfo := ListInfo{
					ID:              getString(listResult, "id_str"),
					Name:            getString(listResult, "name"),
					Description:     getString(listResult, "description"),
					Mode:            getString(listResult, "mode"),
					MemberCount:     int(getFloat64(listResult, "member_count")),
					SubscriberCount: int(getFloat64(listResult, "subscriber_count")),
				}

				if listInfo.ID != "" {
					lists = append(lists, listInfo)
				}
			}
		}
	}

	return lists
}

// CreateList creates a new list
func CreateList(client *XClient, name, description string, isPrivate bool) (map[string]interface{}, error) {
	variables := map[string]interface{}{
		"isPrivate":   isPrivate,
		"name":        name,
		"description": description,
	}
	return client.GraphQLPost("CreateList", variables)
}

// DeleteList deletes a list
func DeleteList(client *XClient, listID string) (map[string]interface{}, error) {
	variables := map[string]interface{}{
		"listId": listID,
	}
	return client.GraphQLPost("DeleteList", variables)
}

// GetListMembers fetches members of a list
func GetListMembers(client *XClient, listID string, count int, cursor string) ([]*models.User, string, error) {
	variables := map[string]interface{}{
		"listId": listID,
		"count":  count,
	}
	if cursor != "" {
		variables["cursor"] = cursor
	}

	data, err := client.GraphQLGet("ListMembers", variables)
	if err != nil {
		return nil, "", err
	}

	return extractUsersFromTimeline(data)
}

// AddListMember adds a user to a list
func AddListMember(client *XClient, listID, userID string) (map[string]interface{}, error) {
	variables := map[string]interface{}{
		"listId": listID,
		"userId": userID,
	}
	return client.GraphQLPost("ListAddMember", variables)
}

// RemoveListMember removes a user from a list
func RemoveListMember(client *XClient, listID, userID string) (map[string]interface{}, error) {
	variables := map[string]interface{}{
		"listId": listID,
		"userId": userID,
	}
	return client.GraphQLPost("ListRemoveMember", variables)
}

// PinList pins a list to the user's profile
func PinList(client *XClient, listID string) (map[string]interface{}, error) {
	variables := map[string]interface{}{
		"timeline_id": listID,
	}
	return client.GraphQLPost("PinTimeline", variables)
}

// UnpinList unpins a list from the user's profile
func UnpinList(client *XClient, listID string) (map[string]interface{}, error) {
	variables := map[string]interface{}{
		"timeline_id": listID,
	}
	return client.GraphQLPost("UnpinTimeline", variables)
}

// Helper functions

// getFloat64 gets a float64 from a map
func getFloat64(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}

// containsSubstring checks if s contains substr
func containsSubstring(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && strings.Contains(s, substr)
}
