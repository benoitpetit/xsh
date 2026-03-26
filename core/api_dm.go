// Package core provides direct message operations for Twitter/X.
package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/benoitpetit/xsh/models"
	"github.com/google/uuid"
)

const (
	// DMAPIBase is the base URL for DM API
	DMAPIBase = "https://x.com/i/api/1.1/dm"
)

// GetDMInbox fetches the user's DM inbox
func GetDMInbox(client *XClient) ([]models.DMConversation, error) {
	conversations := []models.DMConversation{}
	url := fmt.Sprintf("%s/inbox_initial_state.json", DMAPIBase)

	params := map[string]string{
		"nsfw_filtering_enabled":                "false",
		"filter_low_quality":                    "false",
		"include_quality":                       "all",
		"include_profile_interstitial_type":     "1",
		"include_blocking":                      "1",
		"include_blocked_by":                    "1",
		"dm_secret_conversations_enabled":       "false",
		"krs_registration_enabled":              "true",
		"cards_platform":                        "Web-12",
		"include_cards":                         "1",
		"include_ext_alt_text":                  "true",
		"include_quote_count":                   "true",
		"include_reply_count":                   "1",
		"tweet_mode":                            "extended",
		"include_ext_collab_control":            "true",
		"include_ext_is_blue_verified":          "1",
		"include_ext_has_nft_avatar":            "1",
		"include_ext_vibe_tag":                  "1",
		"include_ext_sensitive_media_warning":   "true",
		"include_ext_media_color":               "true",
		"include_ext_media_availability":        "true",
		"include_ext_has_birdwatch_notes":       "1",
	}

	// Make the request
	result, err := client.restGetWithParams(url, params)
	if err != nil {
		return conversations, err
	}

	parsed := parseDMInbox(result)
	if parsed != nil {
		conversations = parsed
	}
	return conversations, nil
}

// parseDMInbox parses the inbox response
func parseDMInbox(data map[string]interface{}) []models.DMConversation {
	var conversations []models.DMConversation

	inboxState, ok := data["inbox_initial_state"].(map[string]interface{})
	if !ok {
		return conversations
	}

	conversationsRaw, ok := inboxState["conversations"].(map[string]interface{})
	if !ok {
		return conversations
	}

	entries, ok := inboxState["entries"].([]interface{})
	if !ok {
		return conversations
	}

	users, ok := inboxState["users"].(map[string]interface{})
	if !ok {
		return conversations
	}

	// Index latest messages per conversation
	latestMessages := make(map[string]map[string]interface{})
	for _, entry := range entries {
		entryMap, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}

		msg, ok := entryMap["message"].(map[string]interface{})
		if !ok {
			continue
		}

		msgData, ok := msg["message_data"].(map[string]interface{})
		if !ok {
			continue
		}

		convID, _ := msg["conversation_id"].(string)
		if convID == "" {
			convID, _ = entryMap["conversation_id"].(string)
		}

		if convID != "" && latestMessages[convID] == nil {
			latestMessages[convID] = map[string]interface{}{
				"text":      getString(msgData, "text"),
				"time":      getString(msgData, "time"),
				"sender_id": getString(msgData, "sender_id"),
			}
		}
	}

	// Build conversation list
	for convID, conv := range conversationsRaw {
		convMap, ok := conv.(map[string]interface{})
		if !ok {
			continue
		}

		convType, _ := convMap["type"].(string)
		isGroup := convType == "GROUP_DM"

		// Build participant list
		var participants []models.DMParticipant
		participantsRaw, ok := convMap["participants"].([]interface{})
		if ok {
			for _, p := range participantsRaw {
				pMap, ok := p.(map[string]interface{})
				if !ok {
					continue
				}

				userID, _ := pMap["user_id"].(string)
				if userID == "" {
					continue
				}

				userInfo, ok := users[userID].(map[string]interface{})
				if !ok {
					continue
				}

				participants = append(participants, models.DMParticipant{
					ID:     userID,
					Name:   getString(userInfo, "name"),
					Handle: getString(userInfo, "screen_name"),
				})
			}
		}

		// Get last message info
		lastMsg := latestMessages[convID]
		lastText := ""
		lastTime := ""
		if lastMsg != nil {
			lastText, _ = lastMsg["text"].(string)
			lastTime, _ = lastMsg["time"].(string)
		}

		// Calculate unread status
		readOnly, _ := convMap["read_only"].(bool)
		notifDisabled, _ := convMap["notifications_disabled"].(bool)
		lastRead, _ := convMap["last_read_event_id"].(string)
		sortEvent, _ := convMap["sort_event_id"].(string)
		unread := !readOnly && !notifDisabled && lastRead < sortEvent

		conversations = append(conversations, models.DMConversation{
			ID:              convID,
			Type:            "group",
			Participants:    participants,
			LastMessage:     lastText,
			LastMessageTime: lastTime,
			Unread:          unread,
		})

		if !isGroup {
			conversations[len(conversations)-1].Type = "one_to_one"
		}
	}

	return conversations
}

// SendDM sends a direct message to a user
func SendDM(client *XClient, userID, text string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/new2.json", DMAPIBase)

	body := map[string]interface{}{
		"recipient_ids":       userID,
		"request_id":          uuid.New().String(),
		"text":                text,
		"cards_platform":      "Web-12",
		"include_cards":       1,
		"include_quote_count": true,
		"dm_users":            false,
	}

	return client.RestPostWithOptions(url, nil, body, 30)
}

// DeleteDM deletes a DM message
func DeleteDM(client *XClient, messageID string) (map[string]interface{}, error) {
	// Uses hardcoded GraphQL query ID
	queryID := "BJ6DtxA2llfjnRoRjaiIiw"
	operationName := "DMMessageDeleteMutation"

	variables := map[string]interface{}{
		"messageId": messageID,
	}

	return client.GraphQLPostRaw(queryID, operationName, variables)
}

// restGetWithParams makes a GET request with query parameters
func (c *XClient) restGetWithParams(urlStr string, params map[string]string) (map[string]interface{}, error) {
	// Build query string
	first := true
	for k, v := range params {
		if first {
			urlStr += "?"
			first = false
		} else {
			urlStr += "&"
		}
		urlStr += fmt.Sprintf("%s=%s", k, v)
	}

	// Get credentials
	creds, err := c.getCredentials()
	if err != nil {
		return nil, err
	}

	// Build request
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}

	// Headers
	req.Header.Set("User-Agent", GetUserAgent())
	req.Header.Set("Authorization", "Bearer "+BearerToken)
	req.Header.Set("X-Csrf-Token", creds.Ct0)
	req.Header.Set("Accept", "application/json")

	// Cookies
	cookies := creds.GetSanitizedCookies()
	var cookieParts []string
	for k, v := range cookies {
		cookieParts = append(cookieParts, fmt.Sprintf("%s=%s", k, v))
	}
	if len(cookieParts) > 0 {
		req.Header.Set("Cookie", strings.Join(cookieParts, "; "))
	}

	// Execute
	httpClient, err := c.getHTTPClient()
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}
