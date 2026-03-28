package tests

import (
	"testing"
	"time"

	"github.com/benoitpetit/xsh/models"
)

// TestTweetURL tests tweet URL generation
func TestTweetURL(t *testing.T) {
	tweet := &models.Tweet{
		ID:           "123456789",
		AuthorHandle: "testuser",
	}

	expected := "https://x.com/testuser/status/123456789"
	if got := tweet.TweetURL(); got != expected {
		t.Errorf("TweetURL() = %v, want %v", got, expected)
	}
}

// TestUserProfileURL tests profile URL generation
func TestUserProfileURL(t *testing.T) {
	user := &models.User{
		ID:     "123",
		Handle: "testuser",
	}

	expected := "https://x.com/testuser"
	if got := user.ProfileURL(); got != expected {
		t.Errorf("ProfileURL() = %v, want %v", got, expected)
	}
}

// TestTweetFromAPIResult tests tweet parsing from API
func TestTweetFromAPIResult(t *testing.T) {
	// Simulate a simplified API response
	result := map[string]interface{}{
		"rest_id": "123456",
		"legacy": map[string]interface{}{
			"full_text":                 "Test tweet content",
			"created_at":                "Mon Jan 01 12:00:00 +0000 2024",
			"favorite_count":            100.0,
			"retweet_count":             50.0,
			"reply_count":               25.0,
			"quote_count":               10.0,
			"bookmark_count":            5.0,
			"in_reply_to_status_id_str": "",
			"in_reply_to_screen_name":   "",
			"conversation_id_str":       "123456",
			"lang":                      "en",
		},
		"core": map[string]interface{}{
			"user_results": map[string]interface{}{
				"result": map[string]interface{}{
					"rest_id":          "789",
					"is_blue_verified": true,
					"legacy": map[string]interface{}{
						"name":        "Test User",
						"screen_name": "testuser",
					},
				},
			},
		},
		"views": map[string]interface{}{
			"count": "1000",
		},
	}

	tweet := models.TweetFromAPIResult(result)

	if tweet == nil {
		t.Fatal("TweetFromAPIResult() returned nil")
	}

	if tweet.ID != "123456" {
		t.Errorf("ID = %v, want 123456", tweet.ID)
	}

	if tweet.Text != "Test tweet content" {
		t.Errorf("Text = %v, want 'Test tweet content'", tweet.Text)
	}

	if tweet.AuthorHandle != "testuser" {
		t.Errorf("AuthorHandle = %v, want 'testuser'", tweet.AuthorHandle)
	}

	if !tweet.AuthorVerified {
		t.Error("AuthorVerified should be true")
	}

	if tweet.Engagement.Likes != 100 {
		t.Errorf("Likes = %v, want 100", tweet.Engagement.Likes)
	}

	if tweet.Engagement.Retweets != 50 {
		t.Errorf("Retweets = %v, want 50", tweet.Engagement.Retweets)
	}
}

// TestTweetFromAPIResultNil tests parsing with nil data
func TestTweetFromAPIResultNil(t *testing.T) {
	// Test with empty map
	result := map[string]interface{}{}

	tweet := models.TweetFromAPIResult(result)

	if tweet != nil {
		t.Error("TweetFromAPIResult(empty) should return nil")
	}
}

// TestTweetFromAPIResultRetweet tests parsing a retweet
func TestTweetFromAPIResultRetweet(t *testing.T) {
	result := map[string]interface{}{
		"rest_id": "original123",
		"legacy": map[string]interface{}{
			"full_text":      "Original tweet",
			"favorite_count": 1000.0,
			"retweeted_status_result": map[string]interface{}{
				"result": map[string]interface{}{
					"rest_id": "original456",
					"legacy": map[string]interface{}{
						"full_text":      "This is the original",
						"favorite_count": 5000.0,
						"created_at":     "Mon Jan 01 10:00:00 +0000 2024",
					},
					"core": map[string]interface{}{
						"user_results": map[string]interface{}{
							"result": map[string]interface{}{
								"rest_id":          "original_user",
								"is_blue_verified": false,
								"legacy": map[string]interface{}{
									"name":        "Original Author",
									"screen_name": "original_author",
								},
							},
						},
					},
				},
			},
		},
		"core": map[string]interface{}{
			"user_results": map[string]interface{}{
				"result": map[string]interface{}{
					"rest_id": "retweeter",
					"legacy": map[string]interface{}{
						"name":        "Retweeter",
						"screen_name": "retweeter_user",
					},
				},
			},
		},
	}

	tweet := models.TweetFromAPIResult(result)

	if tweet == nil {
		t.Fatal("TweetFromAPIResult(retweet) returned nil")
	}

	// Should parse the original tweet, not the retweet wrapper
	if tweet.ID != "original456" {
		t.Errorf("ID = %v, want original456 (original tweet)", tweet.ID)
	}

	if !tweet.IsRetweet {
		t.Error("IsRetweet should be true")
	}

	if tweet.RetweetedBy != "retweeter_user" {
		t.Errorf("RetweetedBy = %v, want 'retweeter_user'", tweet.RetweetedBy)
	}
}

// TestUserFromAPIResult tests user parsing from API
func TestUserFromAPIResult(t *testing.T) {
	result := map[string]interface{}{
		"rest_id":          "123",
		"is_blue_verified": true,
		"legacy": map[string]interface{}{
			"name":                    "Test User",
			"screen_name":             "testuser",
			"description":             "Test bio",
			"location":                "Test Location",
			"followers_count":         1000.0,
			"friends_count":           500.0,
			"statuses_count":          10000.0,
			"listed_count":            50.0,
			"created_at":              "Mon Jan 01 00:00:00 +0000 2020",
			"profile_image_url_https": "https://example.com/normal.jpg",
			"profile_banner_url":      "https://example.com/banner.jpg",
			"pinned_tweet_ids_str":    []interface{}{"456"},
		},
	}

	user := models.UserFromAPIResult(result)

	if user == nil {
		t.Fatal("UserFromAPIResult() returned nil")
	}

	if user.ID != "123" {
		t.Errorf("ID = %v, want 123", user.ID)
	}

	if user.Name != "Test User" {
		t.Errorf("Name = %v, want 'Test User'", user.Name)
	}

	if user.Handle != "testuser" {
		t.Errorf("Handle = %v, want 'testuser'", user.Handle)
	}

	if user.Bio != "Test bio" {
		t.Errorf("Bio = %v, want 'Test bio'", user.Bio)
	}

	if user.Location != "Test Location" {
		t.Errorf("Location = %v, want 'Test Location'", user.Location)
	}

	if !user.Verified {
		t.Error("Verified should be true")
	}

	if user.FollowersCount != 1000 {
		t.Errorf("FollowersCount = %v, want 1000", user.FollowersCount)
	}

	if user.FollowingCount != 500 {
		t.Errorf("FollowingCount = %v, want 500", user.FollowingCount)
	}

	if user.PinnedTweetID != "456" {
		t.Errorf("PinnedTweetID = %v, want 456", user.PinnedTweetID)
	}
}

// TestUserFromAPIResultNil tests user parsing with nil data
func TestUserFromAPIResultNil(t *testing.T) {
	result := map[string]interface{}{}

	user := models.UserFromAPIResult(result)

	if user != nil {
		t.Error("UserFromAPIResult(empty) should return nil")
	}
}

// TestTimelineResponse tests TimelineResponse structure
func TestTimelineResponse(t *testing.T) {
	response := &models.TimelineResponse{
		Tweets: []*models.Tweet{
			{ID: "1"},
			{ID: "2"},
		},
		CursorTop:    "top_cursor",
		CursorBottom: "bottom_cursor",
		HasMore:      true,
	}

	if len(response.Tweets) != 2 {
		t.Errorf("len(Tweets) = %v, want 2", len(response.Tweets))
	}

	if response.CursorTop != "top_cursor" {
		t.Errorf("CursorTop = %v, want 'top_cursor'", response.CursorTop)
	}

	if response.CursorBottom != "bottom_cursor" {
		t.Errorf("CursorBottom = %v, want 'bottom_cursor'", response.CursorBottom)
	}

	if !response.HasMore {
		t.Error("HasMore should be true")
	}
}

// TestTweetMedia tests TweetMedia structure
func TestTweetMedia(t *testing.T) {
	media := models.TweetMedia{
		Type:       "photo",
		URL:        "https://example.com/image.jpg",
		PreviewURL: "https://example.com/preview.jpg",
		AltText:    "Test image",
	}

	if media.Type != "photo" {
		t.Errorf("Type = %v, want 'photo'", media.Type)
	}

	if media.URL != "https://example.com/image.jpg" {
		t.Errorf("URL = %v, want image URL", media.URL)
	}

	if media.PreviewURL != "https://example.com/preview.jpg" {
		t.Errorf("PreviewURL = %v, want preview URL", media.PreviewURL)
	}

	if media.AltText != "Test image" {
		t.Errorf("AltText = %v, want 'Test image'", media.AltText)
	}
}

// TestTweetEngagement tests TweetEngagement structure
func TestTweetEngagement(t *testing.T) {
	engagement := models.TweetEngagement{
		Likes:     100,
		Retweets:  50,
		Replies:   25,
		Quotes:    10,
		Bookmarks: 5,
		Views:     1000,
	}

	if engagement.Likes != 100 {
		t.Errorf("Likes = %v, want 100", engagement.Likes)
	}

	if engagement.Retweets != 50 {
		t.Errorf("Retweets = %v, want 50", engagement.Retweets)
	}

	if engagement.Replies != 25 {
		t.Errorf("Replies = %v, want 25", engagement.Replies)
	}

	if engagement.Quotes != 10 {
		t.Errorf("Quotes = %v, want 10", engagement.Quotes)
	}

	if engagement.Bookmarks != 5 {
		t.Errorf("Bookmarks = %v, want 5", engagement.Bookmarks)
	}

	if engagement.Views != 1000 {
		t.Errorf("Views = %v, want 1000", engagement.Views)
	}
}

// TestTweetWithTime tests a tweet with CreatedAt
func TestTweetWithTime(t *testing.T) {
	now := time.Now()
	tweet := &models.Tweet{
		ID:        "123",
		Text:      "Test",
		CreatedAt: &now,
	}

	if tweet.CreatedAt == nil {
		t.Error("CreatedAt should not be nil")
	}

	if tweet.ID != "123" {
		t.Errorf("ID = %v, want '123'", tweet.ID)
	}

	if tweet.Text != "Test" {
		t.Errorf("Text = %v, want 'Test'", tweet.Text)
	}

	if !tweet.CreatedAt.Equal(now) {
		t.Error("CreatedAt should match the set time")
	}
}

// TestQuotedTweet tests a tweet with quote
func TestQuotedTweet(t *testing.T) {
	quoted := &models.Tweet{
		ID:   "quoted123",
		Text: "Original tweet",
	}

	tweet := &models.Tweet{
		ID:          "123",
		Text:        "Tweet with quote",
		QuotedTweet: quoted,
	}

	if tweet.QuotedTweet == nil {
		t.Fatal("QuotedTweet should not be nil")
	}

	if tweet.ID != "123" {
		t.Errorf("ID = %v, want '123'", tweet.ID)
	}

	if tweet.Text != "Tweet with quote" {
		t.Errorf("Text = %v, want 'Tweet with quote'", tweet.Text)
	}

	if tweet.QuotedTweet.ID != "quoted123" {
		t.Errorf("QuotedTweet.ID = %v, want 'quoted123'", tweet.QuotedTweet.ID)
	}
}
