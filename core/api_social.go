// Package core provides social action operations (follow, block, mute) for Twitter/X.
package core

import (
	"fmt"
)

const (
	// APIBaseV1 is the base URL for Twitter REST API v1.1
	APIBaseV1 = "https://x.com/i/api/1.1"
)

// FollowUser follows a user by their user ID
func FollowUser(client *XClient, userID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/friendships/create.json", APIBaseV1)
	data := map[string]string{
		"user_id":                         userID,
		"include_profile_interstitial_type": "1",
	}
	return client.RestPost(url, data)
}

// UnfollowUser unfollows a user by their user ID
func UnfollowUser(client *XClient, userID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/friendships/destroy.json", APIBaseV1)
	data := map[string]string{
		"user_id":                         userID,
		"include_profile_interstitial_type": "1",
	}
	return client.RestPost(url, data)
}

// BlockUser blocks a user by their user ID
func BlockUser(client *XClient, userID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/blocks/create.json", APIBaseV1)
	data := map[string]string{
		"user_id": userID,
	}
	return client.RestPost(url, data)
}

// UnblockUser unblocks a user by their user ID
func UnblockUser(client *XClient, userID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/blocks/destroy.json", APIBaseV1)
	data := map[string]string{
		"user_id": userID,
	}
	return client.RestPost(url, data)
}

// MuteUser mutes a user by their user ID
func MuteUser(client *XClient, userID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/mutes/users/create.json", APIBaseV1)
	data := map[string]string{
		"user_id": userID,
	}
	return client.RestPost(url, data)
}

// UnmuteUser unmutes a user by their user ID
func UnmuteUser(client *XClient, userID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/mutes/users/destroy.json", APIBaseV1)
	data := map[string]string{
		"user_id": userID,
	}
	return client.RestPost(url, data)
}
