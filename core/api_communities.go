// Package core provides community operations for Twitter/X.
package core

import (
	"github.com/benoitpetit/xsh/models"
)

// Community represents a Twitter/X community
type Community struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	MemberCount    int             `json:"member_count"`
	ModeratorCount int             `json:"moderator_count"`
	CreatedAt      string          `json:"created_at,omitempty"`
	Role           string          `json:"role,omitempty"` // Member, Moderator, Admin, NonMember
	IsNSFW         bool            `json:"is_nsfw"`
	Rules          []CommunityRule `json:"rules,omitempty"`
}

// CommunityRule represents a community rule
type CommunityRule struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GetCommunity fetches a community by ID
func GetCommunity(client *XClient, communityID string) (*Community, error) {
	variables := map[string]interface{}{
		"communityId":              communityID,
		"withDmMuting":             false,
		"withSafetyModeUserFields": false,
	}

	data, err := client.GraphQLGet("CommunityByRestId", variables)
	if err != nil {
		return nil, err
	}

	return parseCommunity(data), nil
}

// GetCommunityTimeline fetches tweets from a community
func GetCommunityTimeline(client *XClient, communityID string, count int, cursor string) (*models.TimelineResponse, error) {
	variables := map[string]interface{}{
		"communityId":   communityID,
		"count":         count,
		"withCommunity": true,
	}

	if cursor != "" {
		variables["cursor"] = cursor
	}

	data, err := client.GraphQLGet("CommunityTweetsTimeline", variables)
	if err != nil {
		return nil, err
	}

	return extractTweetsFromTimeline(data), nil
}

// JoinCommunity joins a community
func JoinCommunity(client *XClient, communityID string) (map[string]interface{}, error) {
	variables := map[string]interface{}{
		"communityId": communityID,
	}

	return client.GraphQLPost("JoinCommunity", variables)
}

// LeaveCommunity leaves a community
func LeaveCommunity(client *XClient, communityID string) (map[string]interface{}, error) {
	variables := map[string]interface{}{
		"communityId": communityID,
	}

	return client.GraphQLPost("LeaveCommunity", variables)
}

// parseCommunity parses community data from API response
func parseCommunity(data map[string]interface{}) *Community {
	community := &Community{}

	dataMap, ok := data["data"].(map[string]interface{})
	if !ok {
		return community
	}

	communityResult, ok := dataMap["communityResults"].(map[string]interface{})
	if !ok {
		return community
	}

	result, ok := communityResult["result"].(map[string]interface{})
	if !ok {
		return community
	}

	community.ID = getString(result, "id_str")
	if community.ID == "" {
		community.ID = getString(result, "rest_id")
	}
	community.Name = getString(result, "name")
	community.Description = getString(result, "description")
	community.IsNSFW = false
	if nsfw, ok := result["is_nsfw"].(bool); ok {
		community.IsNSFW = nsfw
	}

	// Member count
	if membersResult, ok := result["member_count"].(float64); ok {
		community.MemberCount = int(membersResult)
	}
	if modCount, ok := result["moderator_count"].(float64); ok {
		community.ModeratorCount = int(modCount)
	}

	// Role
	if role, ok := result["role"].(string); ok {
		community.Role = role
	}

	// Rules
	if rulesRaw, ok := result["rules"].([]interface{}); ok {
		for _, r := range rulesRaw {
			rMap, ok := r.(map[string]interface{})
			if !ok {
				continue
			}
			community.Rules = append(community.Rules, CommunityRule{
				Name:        getString(rMap, "name"),
				Description: getString(rMap, "description"),
			})
		}
	}

	return community
}
