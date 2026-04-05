// Package core provides Twitter Spaces operations for Twitter/X.
package core

// Space represents a Twitter Space
type Space struct {
	ID                  string      `json:"id"`
	Title               string      `json:"title"`
	State               string      `json:"state"` // Running, Ended, Scheduled, NotStarted
	CreatedAt           string      `json:"created_at,omitempty"`
	ScheduledStart      string      `json:"scheduled_start,omitempty"`
	StartedAt           string      `json:"started_at,omitempty"`
	EndedAt             string      `json:"ended_at,omitempty"`
	HostIDs             []string    `json:"host_ids"`
	SpeakerIDs          []string    `json:"speaker_ids"`
	ParticipantCount    int         `json:"participant_count"`
	IsTicketed          bool        `json:"is_ticketed"`
	NarrowCastSpaceType int         `json:"narrow_cast_space_type"`
	Hosts               []SpaceUser `json:"hosts,omitempty"`
}

// SpaceUser represents a user in a Space context
type SpaceUser struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Handle string `json:"handle"`
}

// SpaceSearchResponse represents search results for Spaces
type SpaceSearchResponse struct {
	Spaces  []Space `json:"spaces"`
	HasMore bool    `json:"has_more"`
}

// GetSpace fetches a Space by ID
func GetSpace(client *XClient, spaceID string) (*Space, error) {
	variables := map[string]interface{}{
		"id":              spaceID,
		"isMetatagsQuery": false,
		"withReplays":     true,
		"withListeners":   true,
	}

	data, err := client.GraphQLGet("AudioSpaceById", variables)
	if err != nil {
		return nil, err
	}

	return parseSpace(data), nil
}

// SearchSpaces searches for Twitter Spaces
func SearchSpaces(client *XClient, query string, count int) (*SpaceSearchResponse, error) {
	variables := map[string]interface{}{
		"query": query,
		"count": count,
	}

	data, err := client.GraphQLGet("AudioSpaceSearch", variables)
	if err != nil {
		return nil, err
	}

	return parseSpaceSearch(data), nil
}

// parseSpace parses a Space from API response
func parseSpace(data map[string]interface{}) *Space {
	space := &Space{}

	dataMap, ok := data["data"].(map[string]interface{})
	if !ok {
		return space
	}

	audioSpace, ok := dataMap["audioSpace"].(map[string]interface{})
	if !ok {
		return space
	}

	metadata, ok := audioSpace["metadata"].(map[string]interface{})
	if !ok {
		return space
	}

	space.ID = getString(metadata, "rest_id")
	space.Title = getString(metadata, "title")
	space.State = getString(metadata, "state")
	space.CreatedAt = getString(metadata, "created_at")
	space.ScheduledStart = getString(metadata, "scheduled_start")
	space.StartedAt = getString(metadata, "started_at")
	space.EndedAt = getString(metadata, "ended_at")
	space.IsTicketed = false
	if ticketed, ok := metadata["is_space_available_for_replay"].(bool); ok {
		space.IsTicketed = ticketed
	}

	if count, ok := metadata["total_participating"].(float64); ok {
		space.ParticipantCount = int(count)
	}

	// Parse participants
	if participants, ok := audioSpace["participants"].(map[string]interface{}); ok {
		if admins, ok := participants["admins"].([]interface{}); ok {
			for _, admin := range admins {
				if u := parseSpaceUser(admin); u != nil {
					space.Hosts = append(space.Hosts, *u)
					space.HostIDs = append(space.HostIDs, u.ID)
				}
			}
		}
		if speakers, ok := participants["speakers"].([]interface{}); ok {
			for _, speaker := range speakers {
				if u := parseSpaceUser(speaker); u != nil {
					space.SpeakerIDs = append(space.SpeakerIDs, u.ID)
				}
			}
		}
	}

	return space
}

// parseSpaceUser parses a user from Space participant data
func parseSpaceUser(data interface{}) *SpaceUser {
	m, ok := data.(map[string]interface{})
	if !ok {
		return nil
	}

	userResults, ok := m["user_results"].(map[string]interface{})
	if !ok {
		return nil
	}

	result, ok := userResults["result"].(map[string]interface{})
	if !ok {
		return nil
	}

	legacy, ok := result["legacy"].(map[string]interface{})
	if !ok {
		return nil
	}

	return &SpaceUser{
		ID:     getString(result, "rest_id"),
		Name:   getString(legacy, "name"),
		Handle: getString(legacy, "screen_name"),
	}
}

// parseSpaceSearch parses Space search results
func parseSpaceSearch(data map[string]interface{}) *SpaceSearchResponse {
	response := &SpaceSearchResponse{
		Spaces: []Space{},
	}

	dataMap, ok := data["data"].(map[string]interface{})
	if !ok {
		return response
	}

	searchResult, ok := dataMap["search_by_raw_query"].(map[string]interface{})
	if !ok {
		return response
	}

	audioResults, ok := searchResult["audio_spaces_grouped_by_section"].(map[string]interface{})
	if !ok {
		return response
	}

	// Parse sections (live, scheduled, etc.)
	sections := []string{"live", "upcoming", "recorded"}
	for _, section := range sections {
		if items, ok := audioResults[section].(map[string]interface{}); ok {
			if spaces, ok := items["spaces"].([]interface{}); ok {
				for _, s := range spaces {
					sMap, ok := s.(map[string]interface{})
					if !ok {
						continue
					}
					space := Space{
						ID:    getString(sMap, "rest_id"),
						Title: getString(sMap, "title"),
						State: getString(sMap, "state"),
					}
					if count, ok := sMap["total_participating"].(float64); ok {
						space.ParticipantCount = int(count)
					}
					if space.ID != "" {
						response.Spaces = append(response.Spaces, space)
					}
				}
			}
		}
	}

	return response
}
