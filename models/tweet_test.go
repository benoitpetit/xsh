package models

import "testing"

func TestTweetFromAPIResultReadsModernAuthorFields(t *testing.T) {
	result := map[string]interface{}{
		"rest_id": "tweet-123",
		"legacy":  map[string]interface{}{"full_text": "A test post"},
		"core": map[string]interface{}{
			"user_results": map[string]interface{}{
				"result": map[string]interface{}{
					"rest_id": "user-456",
					"core": map[string]interface{}{
						"name":        "Modern Author",
						"screen_name": "modern_author",
					},
					"verification": map[string]interface{}{"is_blue_verified": true},
					"legacy":       nil,
				},
			},
		},
	}

	tweet := TweetFromAPIResult(result)
	if tweet == nil {
		t.Fatal("TweetFromAPIResult() returned nil")
	}
	if tweet.AuthorID != "user-456" || tweet.AuthorName != "Modern Author" || tweet.AuthorHandle != "modern_author" {
		t.Errorf("author = %q/%q/%q, want user-456/Modern Author/modern_author", tweet.AuthorID, tweet.AuthorName, tweet.AuthorHandle)
	}
	if !tweet.AuthorVerified {
		t.Error("AuthorVerified = false, want true from verification.is_blue_verified")
	}
}
