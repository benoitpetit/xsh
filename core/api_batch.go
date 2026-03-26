// Package core provides batch operations for Twitter/X.
package core

import (
	"fmt"
	"os"
	"strings"

	"github.com/benoitpetit/xsh/models"
)

// GetTweetsByIDs fetches multiple tweets by their IDs in a single request
func GetTweetsByIDs(client *XClient, tweetIDs []string) ([]*models.Tweet, error) {
	if len(tweetIDs) == 0 {
		return []*models.Tweet{}, nil
	}

	// Limit to 100 IDs per request (Twitter's limit)
	if len(tweetIDs) > 100 {
		tweetIDs = tweetIDs[:100]
	}

	variables := map[string]interface{}{
		"tweetIds": tweetIDs,
		"includePromotedContent":               false,
		"withBirdwatchNotes":                   false,
		"withVoice":                            true,
		"withCommunity":                        true,
		"withQuickPromoteEligibilityTweetFields": false,
	}

	data, err := client.GraphQLGet("TweetResultsByRestIds", variables)
	if err != nil {
		return nil, err
	}

	var tweets []*models.Tweet

	// Parse results
	if result, ok := data["data"].(map[string]interface{}); ok {
		if tweetResults, ok := result["tweetResult"].([]interface{}); ok {
			for _, tr := range tweetResults {
				if trMap, ok := tr.(map[string]interface{}); ok {
					if result, ok := trMap["result"].(map[string]interface{}); ok {
						// Skip tombstone tweets
						if typeName, ok := result["__typename"].(string); ok && typeName == "TweetTombstone" {
							continue
						}
						// Handle TweetWithVisibilityResults wrapper
						if typeName, ok := result["__typename"].(string); ok && typeName == "TweetWithVisibilityResults" {
							if tweet, ok := result["tweet"].(map[string]interface{}); ok {
								result = tweet
							}
						}
						tweet := models.TweetFromAPIResult(result)
						if tweet != nil {
							tweets = append(tweets, tweet)
						}
					}
				}
			}
		}
	}

	return tweets, nil
}

// GetUsersByHandles fetches multiple users by their handles
func GetUsersByHandles(client *XClient, handles []string) ([]*models.User, error) {
	var users []*models.User

	// Process one by one since there's no batch endpoint for user lookup by screen name
	for _, handle := range handles {
		user, err := GetUserByHandle(client, handle)
		if err != nil {
			continue // Skip failed lookups
		}
		if user != nil {
			users = append(users, user)
		}
	}

	return users, nil
}

// GetUsersByIDs fetches multiple users by their IDs
func GetUsersByIDs(client *XClient, userIDs []string) ([]*models.User, error) {
	if len(userIDs) == 0 {
		return []*models.User{}, nil
	}

	// Limit to 100 IDs per request
	if len(userIDs) > 100 {
		userIDs = userIDs[:100]
	}

	variables := map[string]interface{}{
		"userIds": userIDs,
	}

	data, err := client.GraphQLGet("UsersByRestIds", variables)
	if err != nil {
		return nil, err
	}

	var users []*models.User

	// Parse results
	if result, ok := data["data"].(map[string]interface{}); ok {
		if usersResults, ok := result["users"].([]interface{}); ok {
			for _, ur := range usersResults {
				if urMap, ok := ur.(map[string]interface{}); ok {
					if result, ok := urMap["result"].(map[string]interface{}); ok {
						// Skip unavailable users
						if typeName, ok := result["__typename"].(string); ok && typeName == "UserUnavailable" {
							continue
						}
						user := models.UserFromAPIResult(result)
						if user != nil {
							users = append(users, user)
						}
					}
				}
			}
		}
	}

	return users, nil
}

// DownloadTweetMedia downloads all media from a tweet
func DownloadTweetMedia(client *XClient, tweetID, outputDir string) ([]string, error) {
	// Get tweet detail
	tweets, err := GetTweetDetail(client, tweetID, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get tweet: %w", err)
	}

	// Find the focal tweet
	var focal *models.Tweet
	for _, t := range tweets {
		if t.ID == tweetID {
			focal = t
			break
		}
	}

	if focal == nil {
		return nil, fmt.Errorf("tweet not found")
	}

	if len(focal.Media) == 0 {
		return nil, fmt.Errorf("no media found in tweet")
	}

	// Create output directory
	if err := ensureDir(outputDir); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	var downloaded []string
	for i, media := range focal.Media {
		mediaURL := media.URL
		if mediaURL == "" {
			continue
		}

		// For photos, request original quality
		if media.Type == "photo" && !contains(mediaURL, "?") {
			mediaURL = mediaURL + "?format=jpg&name=orig"
		}

		ext := getExtensionFromURL(media.URL)
		filename := fmt.Sprintf("%s_%d.%s", tweetID, i, ext)
		filepath := fmt.Sprintf("%s/%s", outputDir, filename)

		if err := DownloadMedia(mediaURL, filepath); err != nil {
			continue // Skip failed downloads
		}

		downloaded = append(downloaded, filepath)
	}

	return downloaded, nil
}

// getExtensionFromURL extracts file extension from URL
func getExtensionFromURL(url string) string {
	// Simple extension extraction
	parts := split(url, ".")
	if len(parts) > 1 {
		ext := parts[len(parts)-1]
		// Clean up query params
		queryParts := split(ext, "?")
		if len(queryParts) > 0 {
			ext = queryParts[0]
		}
		// Limit length
		if len(ext) > 4 {
			ext = ext[:4]
		}
		return ext
	}
	return "jpg"
}

// split splits a string by separator
func split(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
		}
	}
	result = append(result, s[start:])
	return result
}

// ensureDir ensures a directory exists
func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
