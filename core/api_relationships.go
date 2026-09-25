package core

import (
	"strings"

	"github.com/benoitpetit/xsh/models"
)

// GetViewer fetches the authenticated user's profile.
func GetViewer(client *XClient) (*models.User, error) {
	data, err := client.GraphQLGet("Viewer", map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	dataMap, _ := data["data"].(map[string]interface{})
	viewer, _ := dataMap["viewer"].(map[string]interface{})
	if userResults, ok := viewer["user_results"].(map[string]interface{}); ok {
		if result, ok := userResults["result"].(map[string]interface{}); ok {
			return models.UserFromAPIResult(result), nil
		}
	}
	if result, ok := viewer["result"].(map[string]interface{}); ok {
		return models.UserFromAPIResult(result), nil
	}
	return nil, nil
}

// GetFollowersYouKnow fetches accounts followed by people the target user follows.
func GetFollowersYouKnow(client *XClient, userID string, count int, cursor string) ([]*models.User, string, error) {
	return getRelationshipUsers(client, "FollowersYouKnow", userID, count, cursor)
}

// GetBlueVerifiedFollowers fetches verified followers of a user.
func GetBlueVerifiedFollowers(client *XClient, userID string, count int, cursor string) ([]*models.User, string, error) {
	return getRelationshipUsers(client, "BlueVerifiedFollowers", userID, count, cursor)
}

func getRelationshipUsers(client *XClient, operation, userID string, count int, cursor string) ([]*models.User, string, error) {
	variables := map[string]interface{}{
		"userId":                 userID,
		"count":                  count,
		"includePromotedContent": false,
	}
	if cursor != "" {
		variables["cursor"] = cursor
	}

	data, err := client.GraphQLGet(operation, variables)
	if err != nil {
		return nil, "", err
	}
	return parseRelationshipUsers(data)
}

// GetBlockedAccounts fetches the authenticated account's blocked users.
func GetBlockedAccounts(client *XClient, count int, cursor string) ([]*models.User, string, error) {
	return getViewerRelationshipUsers(client, "BlockedAccountsAll", count, cursor)
}

// GetMutedAccounts fetches the authenticated account's muted users.
func GetMutedAccounts(client *XClient, count int, cursor string) ([]*models.User, string, error) {
	return getViewerRelationshipUsers(client, "MutedAccounts", count, cursor)
}

func getViewerRelationshipUsers(client *XClient, operation string, count int, cursor string) ([]*models.User, string, error) {
	variables := map[string]interface{}{"count": count}
	if cursor != "" {
		variables["cursor"] = cursor
	}

	data, err := client.GraphQLGet(operation, variables)
	if err != nil {
		return nil, "", err
	}
	return parseRelationshipUsers(data)
}

func parseRelationshipUsers(data map[string]interface{}) ([]*models.User, string, error) {
	users, cursor, err := extractUsersFromTimeline(data)
	if users == nil {
		users = []*models.User{}
	}
	return users, cursor, err
}

// GetListInfo fetches metadata for one list.
func GetListInfo(client *XClient, listID string) (*ListInfo, error) {
	data, err := client.GraphQLGet("ListByRestId", map[string]interface{}{"listId": listID})
	if err != nil {
		return nil, err
	}
	return parseListInfo(data), nil
}

func parseListInfo(data map[string]interface{}) *ListInfo {
	dataMap, _ := data["data"].(map[string]interface{})
	listMap, _ := dataMap["list"].(map[string]interface{})
	result, _ := listMap["result"].(map[string]interface{})
	if len(result) == 0 {
		result = listMap
	}
	if len(result) == 0 {
		return nil
	}

	list := listInfoFromMap(result)
	if list.ID == "" {
		return nil
	}
	return &list
}

// GetListMemberships fetches lists the authenticated user belongs to.
func GetListMemberships(client *XClient, userID string, count int, cursor string) ([]ListInfo, string, error) {
	variables := map[string]interface{}{"count": count}
	if userID != "" {
		variables["userId"] = userID
	}
	if cursor != "" {
		variables["cursor"] = cursor
	}

	data, err := client.GraphQLGet("ListMemberships", variables)
	if err != nil {
		return nil, "", err
	}
	return parseListsFromTimeline(data)
}

func parseListsFromTimeline(data map[string]interface{}) ([]ListInfo, string, error) {
	var lists []ListInfo
	var nextCursor string

	for _, instruction := range findInstructions(data) {
		entries, _ := instruction["entries"].([]interface{})
		for _, rawEntry := range entries {
			entry, _ := rawEntry.(map[string]interface{})
			entryID := getString(entry, "entryId")
			content, _ := entry["content"].(map[string]interface{})
			if strings.HasPrefix(entryID, "cursor-bottom") {
				nextCursor = extractCursor(content)
				continue
			}

			if itemContent, ok := content["itemContent"].(map[string]interface{}); ok {
				if listMap, ok := itemContent["list"].(map[string]interface{}); ok {
					lists = append(lists, listInfoFromMap(listMap))
				}
			}
			items, _ := content["items"].([]interface{})
			for _, rawItem := range items {
				item, _ := rawItem.(map[string]interface{})
				inner, _ := item["item"].(map[string]interface{})
				innerContent, _ := inner["itemContent"].(map[string]interface{})
				if listMap, ok := innerContent["list"].(map[string]interface{}); ok {
					lists = append(lists, listInfoFromMap(listMap))
				}
			}
		}
	}

	filtered := lists[:0]
	for _, list := range lists {
		if list.ID != "" {
			filtered = append(filtered, list)
		}
	}
	if filtered == nil {
		filtered = []ListInfo{}
	}
	return filtered, nextCursor, nil
}

func listInfoFromMap(result map[string]interface{}) ListInfo {
	id := getString(result, "rest_id")
	if id == "" {
		id = getString(result, "id_str")
	}
	if id == "" {
		id = getString(result, "id")
	}

	isPinned, _ := result["is_pinned"].(bool)
	return ListInfo{
		ID:              id,
		Name:            getString(result, "name"),
		Description:     getString(result, "description"),
		MemberCount:     int(getFloat64(result, "member_count")),
		SubscriberCount: int(getFloat64(result, "subscriber_count")),
		Mode:            getString(result, "mode"),
		IsPinned:        isPinned,
	}
}
