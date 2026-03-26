// Package models provides data models for xsh.
package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TweetMedia represents a media attachment on a tweet
type TweetMedia struct {
	Type       string `json:"type"` // photo, video, animated_gif
	URL        string `json:"url"`
	PreviewURL string `json:"preview_url,omitempty"`
	AltText    string `json:"alt_text,omitempty"`
}

// TweetEngagement represents engagement metrics for a tweet
type TweetEngagement struct {
	Likes     int `json:"likes"`
	Retweets  int `json:"retweets"`
	Replies   int `json:"replies"`
	Quotes    int `json:"quotes"`
	Bookmarks int `json:"bookmarks"`
	Views     int `json:"views"`
}

// Tweet represents a tweet/post
type Tweet struct {
	ID             string          `json:"id"`
	Text           string          `json:"text"`
	AuthorID       string          `json:"author_id"`
	AuthorName     string          `json:"author_name"`
	AuthorHandle   string          `json:"author_handle"`
	AuthorVerified bool            `json:"author_verified"`
	CreatedAt      *time.Time      `json:"created_at,omitempty"`
	Engagement     TweetEngagement `json:"engagement"`
	Media          []TweetMedia    `json:"media"`
	QuotedTweet    *Tweet          `json:"quoted_tweet,omitempty"`
	ReplyToID      string          `json:"reply_to_id,omitempty"`
	ReplyToHandle  string          `json:"reply_to_handle,omitempty"`
	ConversationID string          `json:"conversation_id,omitempty"`
	Language       string          `json:"language,omitempty"`
	Source         string          `json:"source,omitempty"`
	IsRetweet      bool            `json:"is_retweet"`
	RetweetedBy    string          `json:"retweeted_by,omitempty"`
}

// TweetURL returns the full URL to the tweet
func (t *Tweet) TweetURL() string {
	return fmt.Sprintf("https://x.com/%s/status/%s", t.AuthorHandle, t.ID)
}

// TimelineResponse represents the response from a timeline/search API call
type TimelineResponse struct {
	Tweets       []*Tweet `json:"tweets"`
	CursorTop    string   `json:"cursor_top,omitempty"`
	CursorBottom string   `json:"cursor_bottom,omitempty"`
	HasMore      bool     `json:"has_more"`
}

// GetString safely gets a string from a map
func GetString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// extractUserFromAnywhere deeply searches for user info in the data structure
func extractUserFromAnywhere(data map[string]interface{}) (handle, name, id string) {
	var findUser func(m map[string]interface{})
	
	findUser = func(m map[string]interface{}) {
		// If we already found everything, stop
		if handle != "" && name != "" {
			return
		}
		
		// Check if this map has user info
		if h := GetString(m, "screen_name"); h != "" && handle == "" {
			handle = h
		}
		if n := GetString(m, "name"); n != "" && name == "" {
			name = n
		}
		if i := GetString(m, "rest_id"); i != "" && id == "" {
			id = i
		}
		if i := GetString(m, "id_str"); i != "" && id == "" {
			id = i
		}
		
		// Recursively search in nested maps
		for _, v := range m {
			switch val := v.(type) {
			case map[string]interface{}:
				findUser(val)
			case []interface{}:
				for _, item := range val {
					if itemMap, ok := item.(map[string]interface{}); ok {
						findUser(itemMap)
					}
				}
			}
		}
	}
	
	findUser(data)
	return
}

// getNestedMap safely navigates nested maps
// FromAPIResult parses a tweet from Twitter API GraphQL result
func TweetFromAPIResult(result map[string]interface{}) *Tweet {
	defer func() {
		if r := recover(); r != nil {
			// Handle parsing errors gracefully
		}
	}()

	// Handle different result wrappers
	tweetData := result
	if tweet, ok := result["tweet"].(map[string]interface{}); ok {
		tweetData = tweet
	}

	legacy, _ := tweetData["legacy"].(map[string]interface{})

	restID, _ := tweetData["rest_id"].(string)
	if restID == "" {
		restID = GetString(legacy, "id_str")
	}

	if restID == "" {
		return nil
	}

	// Extract author info - try multiple paths for robustness
	var authorID, authorName, authorHandle string
	var authorVerified bool

	// Path 1: Standard core -> user_results -> result -> legacy
	if core, ok := tweetData["core"].(map[string]interface{}); ok {
		if userResults, ok := core["user_results"].(map[string]interface{}); ok {
			if userResult, ok := userResults["result"].(map[string]interface{}); ok {
				authorID, _ = userResult["rest_id"].(string)
				authorVerified, _ = userResult["is_blue_verified"].(bool)
				if legacyUser, ok := userResult["legacy"].(map[string]interface{}); ok {
					authorName = GetString(legacyUser, "name")
					authorHandle = GetString(legacyUser, "screen_name")
				}
			}
		}
	}

	// Path 2: Try user_results directly (without core wrapper)
	if authorHandle == "" {
		if userResults, ok := tweetData["user_results"].(map[string]interface{}); ok {
			if userResult, ok := userResults["result"].(map[string]interface{}); ok {
				if id, ok := userResult["rest_id"].(string); ok && authorID == "" {
					authorID = id
				}
				if verified, ok := userResult["is_blue_verified"].(bool); ok {
					authorVerified = verified
				}
				if legacyUser, ok := userResult["legacy"].(map[string]interface{}); ok {
					if name := GetString(legacyUser, "name"); name != "" {
						authorName = name
					}
					if handle := GetString(legacyUser, "screen_name"); handle != "" {
						authorHandle = handle
					}
				}
			}
		}
	}

	// Path 3: Try to extract from legacy user_id_str
	if authorHandle == "" {
		if userID := GetString(legacy, "user_id_str"); userID != "" {
			authorID = userID
		}
	}

	// Path 4: Sometimes user info is in result directly (not nested in core)
	if authorHandle == "" {
		if user, ok := tweetData["user"].(map[string]interface{}); ok {
			authorName = GetString(user, "name")
			authorHandle = GetString(user, "screen_name")
			if id, ok := user["id_str"].(string); ok {
				authorID = id
			}
		}
	}

	// Path 5: Try from result.legacy directly (some endpoints return user info in legacy)
	if authorHandle == "" {
		if name := GetString(legacy, "name"); name != "" {
			authorName = name
		}
		if handle := GetString(legacy, "screen_name"); handle != "" {
			authorHandle = handle
		}
	}

	// Path 6: Deep search for user_results anywhere in tweetData
	if authorHandle == "" {
		authorHandle, authorName, authorID = extractUserFromAnywhere(tweetData)
	}

	// Check for retweet
	var retweetedStatus map[string]interface{}
	if rts, ok := legacy["retweeted_status_result"].(map[string]interface{}); ok {
		if r, ok := rts["result"].(map[string]interface{}); ok {
			retweetedStatus = r
		}
	}
	isRetweet := retweetedStatus != nil

	if isRetweet && retweetedStatus != nil {
		// Parse the original tweet instead
		original := TweetFromAPIResult(retweetedStatus)
		if original != nil {
			original.IsRetweet = true
			// Use the authorHandle we extracted from the retweeting user
			if authorHandle != "" {
				original.RetweetedBy = authorHandle
			}
		}
		return original
	}

	// Parse engagement
	engagement := TweetEngagement{}
	if v, ok := legacy["favorite_count"].(float64); ok {
		engagement.Likes = int(v)
	}
	if v, ok := legacy["retweet_count"].(float64); ok {
		engagement.Retweets = int(v)
	}
	if v, ok := legacy["reply_count"].(float64); ok {
		engagement.Replies = int(v)
	}
	if v, ok := legacy["quote_count"].(float64); ok {
		engagement.Quotes = int(v)
	}
	if v, ok := legacy["bookmark_count"].(float64); ok {
		engagement.Bookmarks = int(v)
	}
	if views, ok := tweetData["views"].(map[string]interface{}); ok {
		if count, ok := views["count"].(string); ok {
			if n, err := strconv.Atoi(count); err == nil {
				engagement.Views = n
			}
		}
	}

	// Parse media
	var mediaList []TweetMedia
	if entities, ok := legacy["extended_entities"].(map[string]interface{}); ok {
		if media, ok := entities["media"].([]interface{}); ok {
			for _, m := range media {
				if mediaItem, ok := m.(map[string]interface{}); ok {
					mediaType, _ := mediaItem["type"].(string)
					if mediaType == "" {
						mediaType = "photo"
					}

					var url string
					if mediaType == "video" || mediaType == "animated_gif" {
						if videoInfo, ok := mediaItem["video_info"].(map[string]interface{}); ok {
							if variants, ok := videoInfo["variants"].([]interface{}); ok {
								var maxBitrate float64
								for _, v := range variants {
									if variant, ok := v.(map[string]interface{}); ok {
										contentType, _ := variant["content_type"].(string)
										if contentType == "video/mp4" {
											if bitrate, ok := variant["bitrate"].(float64); ok {
												if bitrate > maxBitrate {
													maxBitrate = bitrate
													url, _ = variant["url"].(string)
												}
											}
										}
									}
								}
							}
						}
					} else {
						url, _ = mediaItem["media_url_https"].(string)
					}

					previewURL, _ := mediaItem["media_url_https"].(string)
					altText, _ := mediaItem["ext_alt_text"].(string)

					mediaList = append(mediaList, TweetMedia{
						Type:       mediaType,
						URL:        url,
						PreviewURL: previewURL,
						AltText:    altText,
					})
				}
			}
		}
	}

	// Parse timestamp
	var createdAt *time.Time
	if rawDate, ok := legacy["created_at"].(string); ok && rawDate != "" {
		if t, err := time.Parse("Mon Jan 02 15:04:05 -0700 2006", rawDate); err == nil {
			createdAt = &t
		}
	}

	// Get full text
	text := GetString(legacy, "full_text")
	if text == "" {
		text = GetString(legacy, "text")
	}

	// Expand URLs in text (replace t.co with real URLs)
	if entities, ok := legacy["entities"].(map[string]interface{}); ok {
		if urls, ok := entities["urls"].([]interface{}); ok {
			for _, u := range urls {
				if urlObj, ok := u.(map[string]interface{}); ok {
					shortURL := GetString(urlObj, "url")
					expandedURL := GetString(urlObj, "expanded_url")
					if shortURL != "" && expandedURL != "" {
						text = strings.ReplaceAll(text, shortURL, expandedURL)
					}
				}
			}
		}
	}

	// Parse quoted tweet
	var quoted *Tweet
	if quotedStatus, ok := tweetData["quoted_status_result"].(map[string]interface{}); ok {
		if quotedResult, ok := quotedStatus["result"].(map[string]interface{}); ok {
			quoted = TweetFromAPIResult(quotedResult)
		}
	}

	return &Tweet{
		ID:             restID,
		Text:           text,
		AuthorID:       authorID,
		AuthorName:     authorName,
		AuthorHandle:   authorHandle,
		AuthorVerified: authorVerified,
		CreatedAt:      createdAt,
		Engagement:     engagement,
		Media:          mediaList,
		QuotedTweet:    quoted,
		ReplyToID:      GetString(legacy, "in_reply_to_status_id_str"),
		ReplyToHandle:  GetString(legacy, "in_reply_to_screen_name"),
		ConversationID: GetString(legacy, "conversation_id_str"),
		Language:       GetString(legacy, "lang"),
		Source:         GetString(tweetData, "source"),
		IsRetweet:      false,
	}
}
