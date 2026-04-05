// Package core provides scheduled tweet operations for Twitter/X.
package core

import (
	"fmt"
)

// ScheduledTweet represents a scheduled tweet
type ScheduledTweet struct {
	ID        string   `json:"id"`
	Text      string   `json:"text"`
	ExecuteAt int64    `json:"execute_at"` // Unix timestamp
	State     string   `json:"state"`
	MediaIDs  []string `json:"media_ids"`
}

// CreateScheduledTweet schedules a tweet for future posting
func CreateScheduledTweet(client *XClient, text string, executeAt int64, mediaIDs []string) (map[string]interface{}, error) {
	variables := map[string]interface{}{
		"post_tweet_request": map[string]interface{}{
			"auto_populate_reply_metadata": false,
			"status":                       text,
			"exclude_reply_user_ids":       []string{},
			"media_ids":                    mediaIDs,
		},
		"execute_at": executeAt,
	}

	return client.GraphQLPost("CreateScheduledTweet", variables)
}

// GetScheduledTweets lists all scheduled tweets
func GetScheduledTweets(client *XClient) ([]ScheduledTweet, error) {
	tweets := []ScheduledTweet{}
	variables := map[string]interface{}{
		"ascending": true,
	}

	data, err := client.GraphQLPost("FetchScheduledTweets", variables)
	if err != nil {
		return tweets, err
	}

	parsed := parseScheduledTweets(data)
	if parsed != nil {
		tweets = parsed
	}
	return tweets, nil
}

// parseScheduledTweets parses the scheduled tweets response
func parseScheduledTweets(data map[string]interface{}) []ScheduledTweet {
	var tweets []ScheduledTweet

	viewer, ok := data["data"].(map[string]interface{})
	if !ok {
		return tweets
	}

	viewer2, ok := viewer["viewer"].(map[string]interface{})
	if !ok {
		return tweets
	}

	scheduledList, ok := viewer2["scheduled_tweet_list"].([]interface{})
	if !ok {
		return tweets
	}

	for _, entry := range scheduledList {
		entryMap, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}

		schedulingInfo, ok := entryMap["scheduling_info"].(map[string]interface{})
		if !ok {
			continue
		}

		tweetInfo, ok := entryMap["tweet_create_request"].(map[string]interface{})
		if !ok {
			continue
		}

		// Extract ID
		id, _ := entryMap["rest_id"].(string)
		if id == "" {
			if idFloat, ok := entryMap["id"].(float64); ok {
				id = fmt.Sprintf("%.0f", idFloat)
			}
		}

		// Extract execute_at
		var executeAt int64
		if execAt, ok := schedulingInfo["execute_at"].(float64); ok {
			executeAt = int64(execAt)
		}

		// Extract media IDs
		var mediaIDs []string
		if media, ok := tweetInfo["media_ids"].([]interface{}); ok {
			for _, m := range media {
				if mStr, ok := m.(string); ok {
					mediaIDs = append(mediaIDs, mStr)
				}
			}
		}

		tweets = append(tweets, ScheduledTweet{
			ID:        id,
			Text:      getString(tweetInfo, "status"),
			ExecuteAt: executeAt,
			State:     getString(schedulingInfo, "state"),
			MediaIDs:  mediaIDs,
		})
	}

	return tweets
}

// DeleteScheduledTweet cancels a scheduled tweet
func DeleteScheduledTweet(client *XClient, scheduledTweetID string) (map[string]interface{}, error) {
	variables := map[string]interface{}{
		"scheduled_tweet_id": scheduledTweetID,
	}

	return client.GraphQLPost("DeleteScheduledTweet", variables)
}
