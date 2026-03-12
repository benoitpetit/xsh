package models

import (
	"fmt"
	"time"
)

// User represents a Twitter/X user
type User struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Handle           string     `json:"handle"`
	Bio              string     `json:"bio"`
	Location         string     `json:"location"`
	Website          string     `json:"website"`
	Verified         bool       `json:"verified"`
	FollowersCount   int        `json:"followers_count"`
	FollowingCount   int        `json:"following_count"`
	TweetCount       int        `json:"tweet_count"`
	ListedCount      int        `json:"listed_count"`
	CreatedAt        *time.Time `json:"created_at,omitempty"`
	ProfileImageURL  string     `json:"profile_image_url"`
	ProfileBannerURL string     `json:"profile_banner_url"`
	PinnedTweetID    string     `json:"pinned_tweet_id,omitempty"`
}

// ProfileURL returns the full URL to the user's profile
func (u *User) ProfileURL() string {
	return fmt.Sprintf("https://x.com/%s", u.Handle)
}

// UserFromAPIResult parses a user from Twitter API GraphQL result
func UserFromAPIResult(result map[string]interface{}) *User {
	defer func() {
		if r := recover(); r != nil {
			// Handle parsing errors gracefully
		}
	}()

	legacy, _ := result["legacy"].(map[string]interface{})
	restID, _ := result["rest_id"].(string)

	if restID == "" {
		return nil
	}

	// Parse timestamp
	var createdAt *time.Time
	if rawDate, ok := legacy["created_at"].(string); ok && rawDate != "" {
		if t, err := time.Parse("Mon Jan 02 15:04:05 -0700 2006", rawDate); err == nil {
			createdAt = &t
		}
	}

	// Extract website from entities
	website := ""
	if entities, ok := legacy["entities"].(map[string]interface{}); ok {
		if urlEntity, ok := entities["url"].(map[string]interface{}); ok {
			if urls, ok := urlEntity["urls"].([]interface{}); ok && len(urls) > 0 {
				if urlObj, ok := urls[0].(map[string]interface{}); ok {
					website, _ = urlObj["expanded_url"].(string)
					if website == "" {
						website, _ = urlObj["url"].(string)
					}
				}
			}
		}
	}

	// Get pinned tweet
	var pinnedTweetID string
	if pinned, ok := legacy["pinned_tweet_ids_str"].([]interface{}); ok && len(pinned) > 0 {
		if id, ok := pinned[0].(string); ok {
			pinnedTweetID = id
		}
	}

	// Parse profile image URL (use larger size)
	profileImageURL, _ := legacy["profile_image_url_https"].(string)
	profileImageURL = replaceAll(profileImageURL, "_normal", "_400x400")

	// Get counts
	followersCount := getInt(legacy, "followers_count")
	followingCount := getInt(legacy, "friends_count")
	tweetCount := getInt(legacy, "statuses_count")
	listedCount := getInt(legacy, "listed_count")

	// Get verification status
	isBlueVerified, _ := result["is_blue_verified"].(bool)

	return &User{
		ID:               restID,
		Name:             getString(legacy, "name"),
		Handle:           getString(legacy, "screen_name"),
		Bio:              getString(legacy, "description"),
		Location:         getString(legacy, "location"),
		Website:          website,
		Verified:         isBlueVerified,
		FollowersCount:   followersCount,
		FollowingCount:   followingCount,
		TweetCount:       tweetCount,
		ListedCount:      listedCount,
		CreatedAt:        createdAt,
		ProfileImageURL:  profileImageURL,
		ProfileBannerURL: getString(legacy, "profile_banner_url"),
		PinnedTweetID:    pinnedTweetID,
	}
}

func replaceAll(s, old, new string) string {
	result := ""
	for {
		i := 0
		for i <= len(s)-len(old) {
			if s[i:i+len(old)] == old {
				result += new
				s = s[i+len(old):]
				break
			}
			i++
		}
		if i > len(s)-len(old) {
			result += s
			break
		}
	}
	return result
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}
