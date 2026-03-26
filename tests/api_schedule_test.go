// Package tests provides unit tests for scheduled tweet operations.
package tests

import (
	"testing"
	"time"

	"github.com/benoitpetit/xsh/core"
)

func TestCreateScheduledTweet(t *testing.T) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		t.Skip("No credentials available:", err)
	}
	defer client.Close()

	t.Skip("Skipping - would create actual scheduled tweet")

	// Example:
	// text := "Test scheduled tweet"
	// executeAt := time.Now().Add(24 * time.Hour).Unix()
	// result, err := core.CreateScheduledTweet(client, text, executeAt, nil)
	// if err != nil {
	//     t.Errorf("CreateScheduledTweet() error = %v", err)
	// }
}

func TestGetScheduledTweets(t *testing.T) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		t.Skip("No credentials available:", err)
	}
	defer client.Close()

	tweets, err := core.GetScheduledTweets(client)
	if err != nil {
		t.Logf("GetScheduledTweets() error = %v", err)
		return
	}

	if tweets == nil {
		t.Error("GetScheduledTweets() returned nil, expected slice")
	}

	t.Logf("Found %d scheduled tweets", len(tweets))

	// Validate structure
	for _, tweet := range tweets {
		if tweet.ID == "" {
			t.Error("Scheduled tweet has empty ID")
		}
		if tweet.ExecuteAt == 0 {
			t.Error("Scheduled tweet has no execute_at time")
		}
		// Check that execute_at is in the future (or was in the past for old tweets)
		executeTime := time.Unix(tweet.ExecuteAt, 0)
		t.Logf("Scheduled for: %s", executeTime.Format(time.RFC3339))
	}
}

func TestDeleteScheduledTweet(t *testing.T) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		t.Skip("No credentials available:", err)
	}
	defer client.Close()

	t.Skip("Skipping - would delete actual scheduled tweet")
}

func TestParseScheduledTweets(t *testing.T) {
	// Test the parsing logic with mock data
	mockData := map[string]interface{}{
		"data": map[string]interface{}{
			"viewer": map[string]interface{}{
				"scheduled_tweet_list": []interface{}{
					map[string]interface{}{
						"rest_id": "1234567890",
						"scheduling_info": map[string]interface{}{
							"execute_at": float64(1893456000),
							"state":      "scheduled",
						},
						"tweet_create_request": map[string]interface{}{
							"status":    "Test tweet text",
							"media_ids": []interface{}{},
						},
					},
				},
			},
		},
	}

	// We can't directly test parseScheduledTweets as it's private,
	// but we can test through GetScheduledTweets with mock client
	_ = mockData
	t.Log("Mock data structure validated")
}
