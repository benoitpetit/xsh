package models

import (
	"fmt"
	"strings"
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
	coreProfile, _ := result["core"].(map[string]interface{})
	profileBio, _ := result["profile_bio"].(map[string]interface{})
	locationProfile, _ := result["location"].(map[string]interface{})
	websiteProfile, _ := result["website"].(map[string]interface{})
	relationshipCounts, _ := result["relationship_counts"].(map[string]interface{})
	tweetCounts, _ := result["tweet_counts"].(map[string]interface{})
	verification, _ := result["verification"].(map[string]interface{})
	banner, _ := result["banner"].(map[string]interface{})
	restID, _ := result["rest_id"].(string)

	if restID == "" {
		return nil
	}

	// Parse timestamp
	var createdAt *time.Time
	rawDate := GetString(coreProfile, "created_at")
	if rawDate == "" {
		rawDate = GetString(legacy, "created_at")
	}
	if rawDate != "" {
		if t, err := time.Parse("Mon Jan 02 15:04:05 -0700 2006", rawDate); err == nil {
			createdAt = &t
		}
	}

	// Extract website from the current profile_bio entities or website object,
	// while retaining the legacy entities path for older responses.
	website := ""
	if entities, ok := profileBio["entities"].(map[string]interface{}); ok {
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
	if website == "" {
		website = GetString(websiteProfile, "url")
	}
	if website == "" {
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
	}

	// Get pinned tweet
	var pinnedTweetID string
	if pinned, ok := legacy["pinned_tweet_ids_str"].([]interface{}); ok && len(pinned) > 0 {
		if id, ok := pinned[0].(string); ok {
			pinnedTweetID = id
		}
	}

	// Parse profile image URL (use larger size)
	profileImageURL := ""
	if avatar, ok := result["avatar"].(map[string]interface{}); ok {
		profileImageURL = GetString(avatar, "image_url")
	}
	if profileImageURL == "" {
		profileImageURL = GetString(result, "profile_image_url_https")
	}
	if profileImageURL == "" {
		profileImageURL = GetString(legacy, "profile_image_url_https")
	}
	profileImageURL = replaceAll(profileImageURL, "_normal", "_400x400")

	name := GetString(coreProfile, "name")
	if name == "" {
		name = GetString(result, "name")
	}
	if name == "" {
		name = GetString(legacy, "name")
	}
	handle := GetString(coreProfile, "screen_name")
	if handle == "" {
		handle = GetString(coreProfile, "handle")
	}
	if handle == "" {
		handle = GetString(result, "screen_name")
	}
	if handle == "" {
		handle = GetString(result, "handle")
	}
	if handle == "" {
		handle = GetString(legacy, "screen_name")
	}
	profileBannerURL := GetString(banner, "image_url")
	if profileBannerURL == "" {
		profileBannerURL = GetString(legacy, "profile_banner_url")
	}
	if profileBannerURL == "" {
		profileBannerURL = GetString(result, "profile_banner_url")
	}

	// Get counts
	followersCount := getIntOrFallback(relationshipCounts, "followers", legacy, "followers_count")
	followingCount := getIntOrFallback(relationshipCounts, "following", legacy, "friends_count")
	tweetCount := getIntOrFallback(tweetCounts, "tweets", legacy, "statuses_count")
	listedCount := getInt(legacy, "listed_count")

	// Get verification status
	isBlueVerified, _ := result["is_blue_verified"].(bool)
	isBlueVerified = isBlueVerified || getBool(verification, "is_blue_verified") || getBool(verification, "verified")

	bio := GetString(profileBio, "description")
	if bio == "" {
		bio = GetString(legacy, "description")
	}
	location := GetString(locationProfile, "location")
	if location == "" {
		location = GetString(legacy, "location")
	}
	if website == "" {
		website = GetString(legacy, "url")
	}
	return &User{
		ID:               restID,
		Name:             name,
		Handle:           handle,
		Bio:              bio,
		Location:         location,
		Website:          website,
		Verified:         isBlueVerified,
		FollowersCount:   followersCount,
		FollowingCount:   followingCount,
		TweetCount:       tweetCount,
		ListedCount:      listedCount,
		CreatedAt:        createdAt,
		ProfileImageURL:  profileImageURL,
		ProfileBannerURL: profileBannerURL,
		PinnedTweetID:    pinnedTweetID,
	}
}

func getIntOrFallback(primary map[string]interface{}, primaryKey string, fallback map[string]interface{}, fallbackKey string) int {
	if value, ok := primary[primaryKey]; ok {
		if count, valid := integerValue(value); valid {
			return count
		}
	}
	return getInt(fallback, fallbackKey)
}

func integerValue(value interface{}) (int, bool) {
	switch value := value.(type) {
	case float64:
		return int(value), true
	case float32:
		return int(value), true
	case int:
		return value, true
	case int64:
		return int(value), true
	default:
		return 0, false
	}
}

func getBool(m map[string]interface{}, key string) bool {
	value, _ := m[key].(bool)
	return value
}

func replaceAll(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

func getInt(m map[string]interface{}, key string) int {
	if value, ok := integerValue(m[key]); ok {
		return value
	}
	return 0
}
