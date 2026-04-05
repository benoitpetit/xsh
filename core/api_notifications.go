// Package core provides notification operations for Twitter/X.
// X.com uses REST API v2 for notifications (not GraphQL).
package core

import (
	"fmt"
	"strings"
)

// Notification represents a Twitter/X notification
type Notification struct {
	ID         string `json:"id"`
	Type       string `json:"type"` // like, retweet, follow, reply, mention, quote
	Message    string `json:"message"`
	Timestamp  string `json:"timestamp,omitempty"`
	UserID     string `json:"user_id,omitempty"`
	UserName   string `json:"user_name,omitempty"`
	UserHandle string `json:"user_handle,omitempty"`
	TweetID    string `json:"tweet_id,omitempty"`
	TweetText  string `json:"tweet_text,omitempty"`
	Icon       string `json:"icon,omitempty"`
}

// NotificationsResponse holds a page of notifications
type NotificationsResponse struct {
	Notifications []Notification `json:"notifications"`
	CursorTop     string         `json:"cursor_top,omitempty"`
	CursorBottom  string         `json:"cursor_bottom,omitempty"`
	HasMore       bool           `json:"has_more"`
}

// notificationV2Params returns the standard query parameters for the REST v2 notifications endpoint
func notificationV2Params(count int, cursor string) map[string]interface{} {
	params := map[string]interface{}{
		"include_profile_interstitial_type":    "1",
		"include_blocking":                     "1",
		"include_blocked_by":                   "1",
		"include_followed_by":                  "1",
		"include_want_retweets":                "1",
		"include_mute_edge":                    "1",
		"include_can_dm":                       "1",
		"include_can_media_tag":                "1",
		"include_ext_is_blue_verified":         "1",
		"include_ext_verified_type":            "1",
		"include_ext_profile_image_shape":      "1",
		"skip_status":                          "1",
		"cards_platform":                       "Web-12",
		"include_cards":                        "1",
		"include_ext_alt_text":                 "true",
		"include_ext_limited_action_results":   "true",
		"include_quote_count":                  "true",
		"include_reply_count":                  "1",
		"tweet_mode":                           "extended",
		"include_ext_views":                    "true",
		"include_entities":                     "true",
		"include_user_entities":                "true",
		"include_ext_media_color":              "true",
		"include_ext_media_availability":       "true",
		"include_ext_sensitive_media_warning":  "true",
		"include_ext_trusted_friends_metadata": "true",
		"send_error_codes":                     "true",
		"simple_quoted_tweet":                  "true",
		"count":                                fmt.Sprintf("%d", count),
		"ext":                                  "mediaStats,highlightedLabel,hasNftAvatar,voiceInfo,birdwatchPivot,superFollowMetadata,unmentionInfo,editControl",
	}
	if cursor != "" {
		params["cursor"] = cursor
	}
	return params
}

// GetNotifications fetches the user's notification timeline via REST API v2.
// X.com does not use GraphQL for notifications; it uses the /i/api/2/notifications/all.json endpoint.
func GetNotifications(client *XClient, count int, cursor string) (*NotificationsResponse, error) {
	params := notificationV2Params(count, cursor)
	data, err := client.RestGet(APIBase+"/2/notifications/all.json", params)
	if err != nil {
		return nil, err
	}
	return parseNotificationsV2(data), nil
}

// parseNotificationsV2 parses the REST v2 notifications response (URT format).
// Response structure: { "globalObjects": { "notifications": {...}, "tweets": {...}, "users": {...} }, "timeline": { "instructions": [...] } }
func parseNotificationsV2(data map[string]interface{}) *NotificationsResponse {
	response := &NotificationsResponse{
		Notifications: []Notification{},
	}

	// Extract global objects (notifications, tweets, users)
	globalObjects, _ := data["globalObjects"].(map[string]interface{})
	notifObjects, _ := getMapSafe(globalObjects, "notifications")
	tweetObjects, _ := getMapSafe(globalObjects, "tweets")
	userObjects, _ := getMapSafe(globalObjects, "users")

	// Extract timeline instructions
	timeline, _ := data["timeline"].(map[string]interface{})
	instructions, _ := timeline["instructions"].([]interface{})

	// Collect notification IDs in display order from timeline instructions
	var orderedNotifIDs []string
	for _, inst := range instructions {
		instMap, ok := inst.(map[string]interface{})
		if !ok {
			continue
		}

		// Handle addEntries
		if addEntries, ok := instMap["addEntries"].(map[string]interface{}); ok {
			entries, _ := addEntries["entries"].([]interface{})
			for _, entry := range entries {
				entryMap, ok := entry.(map[string]interface{})
				if !ok {
					continue
				}
				entryID := getString(entryMap, "entryId")

				// Handle cursors
				if strings.HasPrefix(entryID, "cursor-top-") {
					response.CursorTop = extractV2Cursor(entryMap)
					continue
				}
				if strings.HasPrefix(entryID, "cursor-bottom-") {
					response.CursorBottom = extractV2Cursor(entryMap)
					response.HasMore = response.CursorBottom != ""
					continue
				}

				// Extract notification ID from entry
				if strings.HasPrefix(entryID, "notification-") {
					notifID := strings.TrimPrefix(entryID, "notification-")
					orderedNotifIDs = append(orderedNotifIDs, notifID)
				}
			}
		}
	}

	// Build notifications from global objects in display order
	for _, notifID := range orderedNotifIDs {
		notifData, ok := notifObjects[notifID].(map[string]interface{})
		if !ok {
			continue
		}
		notification := parseNotificationV2Entry(notifID, notifData, tweetObjects, userObjects)
		if notification != nil {
			response.Notifications = append(response.Notifications, *notification)
		}
	}

	return response
}

// parseNotificationV2Entry parses a single notification from REST v2 global objects
func parseNotificationV2Entry(id string, notif map[string]interface{}, tweets, users map[string]interface{}) *Notification {
	message, _ := notif["message"].(map[string]interface{})
	messageText := getString(message, "text")

	timestampMs := getString(notif, "timestampMs")
	icon, _ := notif["icon"].(map[string]interface{})
	iconID := getString(icon, "id")

	notification := &Notification{
		ID:        id,
		Type:      classifyNotificationIcon(iconID),
		Message:   messageText,
		Timestamp: timestampMs,
		Icon:      iconID,
	}

	// Extract user and tweet info from template
	if template, ok := notif["template"].(map[string]interface{}); ok {
		if aggregate, ok := template["aggregateUserActionsV1"].(map[string]interface{}); ok {
			// Extract first "from" user (the one who performed the action)
			if fromUsers, ok := aggregate["fromUsers"].([]interface{}); ok && len(fromUsers) > 0 {
				userID := extractNestedID(fromUsers[0], "user")
				if userID != "" {
					notification.UserID = userID
					if userData, ok := users[userID].(map[string]interface{}); ok {
						notification.UserName = getString(userData, "name")
						notification.UserHandle = getString(userData, "screen_name")
					}
				}
			}

			// Extract target tweet
			if targetObjects, ok := aggregate["targetObjects"].([]interface{}); ok && len(targetObjects) > 0 {
				tweetID := extractNestedID(targetObjects[0], "tweet")
				if tweetID != "" {
					notification.TweetID = tweetID
					if tweetData, ok := tweets[tweetID].(map[string]interface{}); ok {
						notification.TweetText = getString(tweetData, "full_text")
					}
				}
			}
		}
	}

	if notification.Message == "" && notification.Type == "" {
		return nil
	}

	return notification
}

// extractNestedID extracts an ID from a nested object like {"user": {"id": "123"}} or {"tweet": {"id": "456"}}
func extractNestedID(item interface{}, key string) string {
	obj, ok := item.(map[string]interface{})
	if !ok {
		return ""
	}
	nested, ok := obj[key].(map[string]interface{})
	if !ok {
		// Try direct string (fallback)
		if s, ok := obj[key].(string); ok {
			return s
		}
		return ""
	}
	// ID can be a string or a number
	if s, ok := nested["id"].(string); ok {
		return s
	}
	// JSON numbers are float64 in Go
	if f, ok := nested["id"].(float64); ok {
		return fmt.Sprintf("%.0f", f)
	}
	return ""
}

// extractV2Cursor extracts cursor value from a REST v2 timeline entry
func extractV2Cursor(entry map[string]interface{}) string {
	content, ok := entry["content"].(map[string]interface{})
	if !ok {
		return ""
	}
	// REST v2 format: { "content": { "operation": { "cursor": { "value": "..." } } } }
	if op, ok := content["operation"].(map[string]interface{}); ok {
		if cursor, ok := op["cursor"].(map[string]interface{}); ok {
			return getString(cursor, "value")
		}
	}
	// Alternative: { "content": { "value": "..." } }
	return getString(content, "value")
}

// getMapSafe safely extracts a map[string]interface{} from a parent map
func getMapSafe(parent map[string]interface{}, key string) (map[string]interface{}, bool) {
	if parent == nil {
		return nil, false
	}
	v, ok := parent[key].(map[string]interface{})
	return v, ok
}

// classifyNotificationIcon maps icon IDs to notification types
func classifyNotificationIcon(iconID string) string {
	switch iconID {
	case "heart_icon":
		return "like"
	case "retweet_icon":
		return "retweet"
	case "person_icon":
		return "follow"
	case "reply_icon":
		return "reply"
	case "mention_icon", "at_icon":
		return "mention"
	case "quote_icon":
		return "quote"
	case "bell_icon":
		return "recommendation"
	case "bird_icon":
		return "news"
	case "community_icon":
		return "community"
	case "list_icon":
		return "list"
	case "space_icon", "microphone_icon":
		return "space"
	default:
		if iconID != "" {
			return iconID
		}
		return "notification"
	}
}
