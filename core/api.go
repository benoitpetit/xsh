package core

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/benoitpetit/xsh/models"
)

// extractTweetsFromTimeline extracts tweets and cursors from a timeline-style API response
func extractTweetsFromTimeline(data map[string]interface{}) *models.TimelineResponse {
	response := &models.TimelineResponse{
		Tweets: []*models.Tweet{},
	}

	instructions := findInstructions(data)

	for _, instruction := range instructions {
		instType := getString(instruction, "type")

		var entries []interface{}
		switch instType {
		case "TimelineAddEntries":
			entries, _ = instruction["entries"].([]interface{})
		case "TimelineAddToModule":
			entries, _ = instruction["moduleItems"].([]interface{})
		}

		for _, entry := range entries {
			if entryMap, ok := entry.(map[string]interface{}); ok {
				entryID := getString(entryMap, "entryId")
				content, _ := entryMap["content"].(map[string]interface{})

				if len(entryID) >= 10 && entryID[:10] == "cursor-top" {
					response.CursorTop = extractCursor(content)
				} else if len(entryID) >= 14 && entryID[:14] == "cursor-bottom" {
					response.CursorBottom = extractCursor(content)
					response.HasMore = response.CursorBottom != ""
				} else if itemContent, ok := content["itemContent"].(map[string]interface{}); ok {
					tweet := parseTweetEntry(itemContent)
					if tweet != nil {
						response.Tweets = append(response.Tweets, tweet)
					}
				} else if entryType, ok := content["entryType"].(string); ok && entryType == "TimelineTimelineModule" {
					if items, ok := content["items"].([]interface{}); ok {
						for _, item := range items {
							if itemMap, ok := item.(map[string]interface{}); ok {
								if innerItem, ok := itemMap["item"].(map[string]interface{}); ok {
									if itemContent, ok := innerItem["itemContent"].(map[string]interface{}); ok {
										tweet := parseTweetEntry(itemContent)
										if tweet != nil {
											response.Tweets = append(response.Tweets, tweet)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return response
}

// findInstructions finds the instructions list in API response
func findInstructions(data map[string]interface{}) []map[string]interface{} {
	paths := [][]string{
		{"data", "home", "home_timeline_urt", "instructions"},
		{"data", "search_by_raw_query", "search_timeline", "timeline", "instructions"},
		{"data", "user", "result", "timeline_v2", "timeline", "instructions"},
		{"data", "user", "result", "timeline", "timeline", "instructions"},
		{"data", "bookmark_timeline_v2", "timeline", "instructions"},
		{"data", "bookmark_timeline", "timeline", "instructions"},
		{"data", "search_by_raw_query", "bookmarks_search_timeline", "timeline", "instructions"},
		{"data", "list", "tweets_timeline", "timeline", "instructions"},
		{"data", "threaded_conversation_with_injections_v2", "instructions"},
		{"data", "tweetResult", "result"},
	}

	for _, path := range paths {
		result := data
		for _, key := range path {
			if next, ok := result[key].(map[string]interface{}); ok {
				result = next
			} else if arr, ok := result[key].([]interface{}); ok {
				var instructions []map[string]interface{}
				for _, item := range arr {
					if m, ok := item.(map[string]interface{}); ok {
						instructions = append(instructions, m)
					}
				}
				return instructions
			} else {
				result = nil
				break
			}
		}
	}

	return nil
}

// extractCursor extracts cursor value from content
func extractCursor(content map[string]interface{}) string {
	if cursorType, ok := content["cursorType"].(string); ok && cursorType != "" {
		if value, ok := content["value"].(string); ok {
			return value
		}
	}
	// Try nested
	if itemContent, ok := content["itemContent"].(map[string]interface{}); ok {
		if value, ok := itemContent["value"].(string); ok {
			return value
		}
	}
	return ""
}

// parseTweetEntry parses a tweet from a timeline entry's itemContent
func parseTweetEntry(itemContent map[string]interface{}) *models.Tweet {
	if itemType, ok := itemContent["itemType"].(string); !ok || itemType != "TimelineTweet" {
		return nil
	}

	tweetResults, _ := itemContent["tweet_results"].(map[string]interface{})
	result, _ := tweetResults["result"].(map[string]interface{})

	// Handle tombstone tweets
	if typeName, ok := result["__typename"].(string); ok && typeName == "TweetTombstone" {
		return nil
	}

	// Handle TweetWithVisibilityResults wrapper
	if typeName, ok := result["__typename"].(string); ok && typeName == "TweetWithVisibilityResults" {
		if tweet, ok := result["tweet"].(map[string]interface{}); ok {
			result = tweet
		}
	}

	return models.TweetFromAPIResult(result)
}

// GetHomeTimeline fetches home timeline
func GetHomeTimeline(client *XClient, timelineType string, count int, cursor string) (*models.TimelineResponse, error) {
	operation := "HomeTimeline"
	if timelineType == "following" {
		operation = "HomeLatestTimeline"
	}

	variables := map[string]interface{}{
		"count":                  count,
		"includePromotedContent": false,
		"latestControlAvailable": true,
	}

	if timelineType == "following" {
		variables["requestContext"] = "launch"
	}

	if cursor != "" {
		variables["cursor"] = cursor
	}

	data, err := client.GraphQLGet(operation, variables)
	if err != nil {
		return nil, err
	}

	return extractTweetsFromTimeline(data), nil
}

// GetTweetDetail fetches a tweet and its conversation thread
func GetTweetDetail(client *XClient, tweetID string, count int) ([]*models.Tweet, error) {
	if count <= 0 {
		count = 20
	}
	if count > 100 {
		count = 100
	}

	variables := map[string]interface{}{
		"focalTweetId":                           tweetID,
		"with_rux_injections":                    false,
		"includePromotedContent":                 false,
		"withCommunity":                          true,
		"withQuickPromoteEligibilityTweetFields": false,
		"withBirdwatchNotes":                     true,
		"withVoice":                              true,
		"withV2Timeline":                         true,
		"count":                                  count,
	}

	data, err := client.GraphQLGet("TweetDetail", variables)
	if err != nil {
		return nil, err
	}

	var tweets []*models.Tweet

	if dataData, ok := data["data"].(map[string]interface{}); ok {
		if conversation, ok := dataData["threaded_conversation_with_injections_v2"].(map[string]interface{}); ok {
			if instructions, ok := conversation["instructions"].([]interface{}); ok {
				for _, inst := range instructions {
					if instruction, ok := inst.(map[string]interface{}); ok {
						if entries, ok := instruction["entries"].([]interface{}); ok {
							for _, entry := range entries {
								if entryMap, ok := entry.(map[string]interface{}); ok {
									if content, ok := entryMap["content"].(map[string]interface{}); ok {
										if itemContent, ok := content["itemContent"].(map[string]interface{}); ok {
											tweet := parseTweetEntry(itemContent)
											if tweet != nil {
												tweets = append(tweets, tweet)
											}
										} else if entryType, ok := content["entryType"].(string); ok && entryType == "TimelineTimelineModule" {
											if items, ok := content["items"].([]interface{}); ok {
												for _, item := range items {
													if itemMap, ok := item.(map[string]interface{}); ok {
														if innerItem, ok := itemMap["item"].(map[string]interface{}); ok {
															if itemContent, ok := innerItem["itemContent"].(map[string]interface{}); ok {
																tweet := parseTweetEntry(itemContent)
																if tweet != nil {
																	tweets = append(tweets, tweet)
																}
															}
														}
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return tweets, nil
}

// SearchTweets searches for tweets
func SearchTweets(client *XClient, query, searchType string, count int, cursor string) (*models.TimelineResponse, error) {
	variables := map[string]interface{}{
		"rawQuery":              query,
		"count":                 count,
		"querySource":           "typed_query",
		"product":               searchType,
		"withGrokTranslatedBio": false,
	}
	if cursor != "" {
		variables["cursor"] = cursor
	}

	// Build proper referer for search
	referer := BaseURL + "/search?q=" + url.QueryEscape(query) + "&src=typed_query"
	if searchType == "Latest" {
		referer += "&f=live"
	}

	// SearchTimeline requires POST (not GET)
	data, err := client.GraphQLPostWithReferer("SearchTimeline", variables, referer)
	if err != nil {
		return nil, err
	}

	return extractTweetsFromTimeline(data), nil
}

// GetUserByHandle fetches user profile by screen name
func GetUserByHandle(client *XClient, handle string) (*models.User, error) {
	variables := map[string]interface{}{
		"screen_name":              handle,
		"withSafetyModeUserFields": true,
	}

	data, err := client.GraphQLGet("UserByScreenName", variables)
	if err != nil {
		return nil, err
	}

	if dataData, ok := data["data"].(map[string]interface{}); ok {
		if user, ok := dataData["user"].(map[string]interface{}); ok {
			if result, ok := user["result"].(map[string]interface{}); ok {
				if typeName, ok := result["__typename"].(string); ok && typeName == "UserUnavailable" {
					return nil, nil
				}
				return models.UserFromAPIResult(result), nil
			}
		}
	}

	return nil, nil
}

// GetUserTweets fetches tweets from a user
func GetUserTweets(client *XClient, userID string, count int, cursor string, includeReplies bool) (*models.TimelineResponse, error) {
	operation := "UserTweets"
	if includeReplies {
		operation = "UserTweetsAndReplies"
	}

	variables := map[string]interface{}{
		"userId":                                 userID,
		"count":                                  count,
		"includePromotedContent":                 false,
		"withQuickPromoteEligibilityTweetFields": false,
		"withVoice":                              true,
		"withV2Timeline":                         true,
	}
	if cursor != "" {
		variables["cursor"] = cursor
	}

	data, err := client.GraphQLGet(operation, variables)
	if err != nil {
		return nil, err
	}

	return extractTweetsFromTimeline(data), nil
}

// GetUserLikes fetches likes from a user
func GetUserLikes(client *XClient, userID string, count int, cursor string) (*models.TimelineResponse, error) {
	variables := map[string]interface{}{
		"userId":                 userID,
		"count":                  count,
		"includePromotedContent": false,
	}
	if cursor != "" {
		variables["cursor"] = cursor
	}

	data, err := client.GraphQLGet("Likes", variables)
	if err != nil {
		return nil, err
	}

	return extractTweetsFromTimeline(data), nil
}

// GetFollowers fetches followers of a user
func GetFollowers(client *XClient, userID string, count int, cursor string) ([]*models.User, string, error) {
	variables := map[string]interface{}{
		"userId":                 userID,
		"count":                  count,
		"includePromotedContent": false,
	}
	if cursor != "" {
		variables["cursor"] = cursor
	}

	data, err := client.GraphQLGet("Followers", variables)
	if err != nil {
		return nil, "", err
	}

	return extractUsersFromTimeline(data)
}

// GetFollowing fetches users followed by a user
func GetFollowing(client *XClient, userID string, count int, cursor string) ([]*models.User, string, error) {
	variables := map[string]interface{}{
		"userId":                 userID,
		"count":                  count,
		"includePromotedContent": false,
	}
	if cursor != "" {
		variables["cursor"] = cursor
	}

	data, err := client.GraphQLGet("Following", variables)
	if err != nil {
		return nil, "", err
	}

	return extractUsersFromTimeline(data)
}

// extractUsersFromTimeline extracts users from a timeline response
func extractUsersFromTimeline(data map[string]interface{}) ([]*models.User, string, error) {
	var users []*models.User
	var nextCursor string

	instructions := findInstructions(data)

	for _, instruction := range instructions {
		if entries, ok := instruction["entries"].([]interface{}); ok {
			for _, entry := range entries {
				if entryMap, ok := entry.(map[string]interface{}); ok {
					entryID := getString(entryMap, "entryId")
					content, _ := entryMap["content"].(map[string]interface{})

					if len(entryID) > 14 && entryID[:14] == "cursor-bottom" {
						nextCursor = extractCursor(content)
					} else if itemContent, ok := content["itemContent"].(map[string]interface{}); ok {
						if userResults, ok := itemContent["user_results"].(map[string]interface{}); ok {
							if result, ok := userResults["result"].(map[string]interface{}); ok {
								user := models.UserFromAPIResult(result)
								if user != nil {
									users = append(users, user)
								}
							}
						}
					}
				}
			}
		}
	}

	return users, nextCursor, nil
}

// GetBookmarks fetches bookmarked tweets
func GetBookmarks(client *XClient, count int, cursor string) (*models.TimelineResponse, error) {
	variables := map[string]interface{}{
		"count":                  count,
		"includePromotedContent": false,
		"rawQuery":               "a OR e OR i OR o OR u OR t OR s OR n OR r OR l",
	}
	if cursor != "" {
		variables["cursor"] = cursor
	}

	data, err := client.GraphQLGet("BookmarkSearchTimeline", variables)
	if err != nil {
		return nil, err
	}

	return extractTweetsFromTimeline(data), nil
}

// GetListTweets fetches tweets from a list
func GetListTweets(client *XClient, listID string, count int, cursor string) (*models.TimelineResponse, error) {
	variables := map[string]interface{}{
		"listId": listID,
		"count":  count,
	}
	if cursor != "" {
		variables["cursor"] = cursor
	}

	data, err := client.GraphQLGet("ListLatestTweetsTimeline", variables)
	if err != nil {
		return nil, err
	}

	return extractTweetsFromTimeline(data), nil
}

// CreateTweet creates a new tweet
func CreateTweet(client *XClient, text, replyToID, quoteTweetURL string, mediaIDs []string, cardURI string) (map[string]interface{}, error) {
	variables := map[string]interface{}{
		"tweet_text":              text,
		"dark_request":            false,
		"semantic_annotation_ids": []interface{}{},
	}

	mediaEntities := make([]map[string]interface{}, 0, len(mediaIDs))
	for _, mid := range mediaIDs {
		mediaEntities = append(mediaEntities, map[string]interface{}{
			"media_id":     mid,
			"tagged_users": []interface{}{},
		})
	}
	variables["media"] = map[string]interface{}{
		"media_entities":     mediaEntities,
		"possibly_sensitive": false,
	}

	if replyToID != "" {
		variables["reply"] = map[string]interface{}{
			"in_reply_to_tweet_id":   replyToID,
			"exclude_reply_user_ids": []interface{}{},
		}
	}

	if quoteTweetURL != "" {
		variables["attachment_url"] = quoteTweetURL
	}

	if cardURI != "" {
		variables["card_uri"] = cardURI
	}

	if Verbose {
		fmt.Printf("[DEBUG] Creating tweet with text: %s\n", text)
	}

	result, err := client.GraphQLPost("CreateTweet", variables)
	if err != nil {
		return nil, err
	}

	if Verbose {
		jsonData, _ := json.MarshalIndent(result, "", "  ")
		fmt.Printf("[DEBUG] CreateTweet response:\n%s\n", string(jsonData))
	}

	// Check for explicit errors first (top-level errors array or error object)
	if errMsg := extractErrorMessage(result); errMsg != "" {
		return nil, fmt.Errorf("tweet creation failed: %s", errMsg)
	}

	// Verify the tweet was actually created (must have a valid tweet ID)
	// An empty create_tweet response is NOT a success — X.com always returns the
	// created tweet's rest_id on success. Empty means the request was silently
	// rejected (stale endpoint, rate-limit, policy, etc.).
	if !isTweetCreated(result) {
		if Verbose {
			jsonData, _ := json.MarshalIndent(result, "", "  ")
			fmt.Printf("[DEBUG] CreateTweet rejected (no tweet ID in response):\n%s\n", string(jsonData))
		}
		return nil, fmt.Errorf("tweet creation failed: API did not return a tweet ID (request may have been silently rejected)")
	}

	return result, nil
}

// isTweetCreated checks if the response contains a created tweet
func isTweetCreated(result map[string]interface{}) bool {
	if result == nil {
		return false
	}

	// Check multiple possible response structures
	if data, ok := result["data"].(map[string]interface{}); ok {
		// Check create_tweet endpoint
		if createTweet, ok := data["create_tweet"].(map[string]interface{}); ok {
			if tweetResults, ok := createTweet["tweet_results"].(map[string]interface{}); ok {
				if result, ok := tweetResults["result"].(map[string]interface{}); ok {
					// Check rest_id
					if restID, ok := result["rest_id"].(string); ok && restID != "" {
						return true
					}
					// Check legacy.id_str (sometimes rest_id is empty but id_str has the value)
					if legacy, ok := result["legacy"].(map[string]interface{}); ok {
						if idStr, ok := legacy["id_str"].(string); ok && idStr != "" {
							return true
						}
					}
				}
			}
		}
		// Check for create_tweeting (alternative spelling)
		if createTweeting, ok := data["create_tweeting"].(map[string]interface{}); ok {
			if tweet, ok := createTweeting["tweet"].(map[string]interface{}); ok {
				if id, ok := tweet["id"].(string); ok && id != "" {
					return true
				}
			}
		}
	}

	return false
}

// extractErrorMessage tries to extract an error message from API response
func extractErrorMessage(result map[string]interface{}) string {
	if result == nil {
		return ""
	}

	// Check for errors array
	if errors, ok := result["errors"].([]interface{}); ok && len(errors) > 0 {
		if firstErr, ok := errors[0].(map[string]interface{}); ok {
			if message, ok := firstErr["message"].(string); ok {
				return message
			}
		}
	}

	// Check for error object
	if errObj, ok := result["error"].(map[string]interface{}); ok {
		if message, ok := errObj["message"].(string); ok {
			return message
		}
	}

	return ""
}

// DeleteTweet deletes a tweet
func DeleteTweet(client *XClient, tweetID string) (map[string]interface{}, error) {
	return client.GraphQLPost("DeleteTweet", map[string]interface{}{
		"tweet_id": tweetID,
	})
}

// LikeTweet likes a tweet
func LikeTweet(client *XClient, tweetID string) (map[string]interface{}, error) {
	return client.GraphQLPost("FavoriteTweet", map[string]interface{}{
		"tweet_id": tweetID,
	})
}

// UnlikeTweet unlikes a tweet
func UnlikeTweet(client *XClient, tweetID string) (map[string]interface{}, error) {
	return client.GraphQLPost("UnfavoriteTweet", map[string]interface{}{
		"tweet_id": tweetID,
	})
}

// Retweet retweets a tweet
func Retweet(client *XClient, tweetID string) (map[string]interface{}, error) {
	return client.GraphQLPost("CreateRetweet", map[string]interface{}{
		"tweet_id": tweetID,
	})
}

// Unretweet undoes a retweet
func Unretweet(client *XClient, tweetID string) (map[string]interface{}, error) {
	return client.GraphQLPost("DeleteRetweet", map[string]interface{}{
		"source_tweet_id": tweetID,
	})
}

// BookmarkTweet bookmarks a tweet
func BookmarkTweet(client *XClient, tweetID string) (map[string]interface{}, error) {
	return client.GraphQLPost("CreateBookmark", map[string]interface{}{
		"tweet_id": tweetID,
	})
}

// UnbookmarkTweet removes a bookmark
func UnbookmarkTweet(client *XClient, tweetID string) (map[string]interface{}, error) {
	return client.GraphQLPost("DeleteBookmark", map[string]interface{}{
		"tweet_id": tweetID,
	})
}

// GetTrends fetches trending topics
func GetTrends(client *XClient, woeid int) ([]*models.Trend, error) {
	variables := map[string]interface{}{}
	if woeid != 0 {
		variables["woeid"] = woeid
	}

	data, err := client.GraphQLGet("ExplorePage", variables)
	if err != nil {
		return nil, err
	}

	return parseTrendsFromExploreResponse(data), nil
}

func parseTrendsFromExploreResponse(data map[string]interface{}) []*models.Trend {
	trends := make([]*models.Trend, 0)
	rank := 1

	var timeline map[string]interface{}
	if d, ok := data["data"].(map[string]interface{}); ok {
		if explorePage, ok := d["explore_page"].(map[string]interface{}); ok {
			if body, ok := explorePage["body"].(map[string]interface{}); ok {
				if initialTimeline, ok := body["initialTimeline"].(map[string]interface{}); ok {
					if t1, ok := initialTimeline["timeline"].(map[string]interface{}); ok {
						if t2, ok := t1["timeline"].(map[string]interface{}); ok {
							timeline = t2
						}
					}
				}
			}
		}
	}

	if timeline == nil {
		if t, ok := data["timeline"].(map[string]interface{}); ok {
			timeline = t
		}
	}

	if timeline == nil {
		return trends
	}

	instructions, _ := timeline["instructions"].([]interface{})
	for _, instructionRaw := range instructions {
		instruction, ok := instructionRaw.(map[string]interface{})
		if !ok {
			continue
		}
		entries, _ := instruction["entries"].([]interface{})
		for _, entryRaw := range entries {
			entry, ok := entryRaw.(map[string]interface{})
			if !ok {
				continue
			}
			content, _ := entry["content"].(map[string]interface{})
			items, _ := content["items"].([]interface{})
			for _, itemRaw := range items {
				item, ok := itemRaw.(map[string]interface{})
				if !ok {
					continue
				}
				trend := extractTrendFromModuleItem(item, rank)
				if trend != nil {
					trends = append(trends, trend)
					rank++
				}
			}
		}
	}

	return trends
}

func extractTrendFromModuleItem(item map[string]interface{}, rank int) *models.Trend {
	itemMap, _ := item["item"].(map[string]interface{})
	if itemMap == nil {
		return nil
	}

	itemContent, _ := itemMap["itemContent"].(map[string]interface{})
	if itemContent == nil {
		itemContent, _ = itemMap["content"].(map[string]interface{})
	}

	if itemContent == nil {
		return nil
	}

	return extractTrendFromContent(itemContent, rank)
}

func extractTrendFromContent(content map[string]interface{}, rank int) *models.Trend {
	if typeName, _ := content["__typename"].(string); typeName == "TimelineTrend" {
		name, _ := content["name"].(string)
		if name == "" {
			return nil
		}

		context := ""
		if socialContext, ok := content["social_context"].(map[string]interface{}); ok {
			context, _ = socialContext["text"].(string)
		}

		query := name
		if trendMeta, ok := content["trend_metadata"].(map[string]interface{}); ok {
			if urlMap, ok := trendMeta["url"].(map[string]interface{}); ok {
				if u, ok := urlMap["url"].(string); ok && u != "" {
					query = u
				}
			}
		}

		isPromoted := false
		if promoted, ok := content["promoted_content"].(map[string]interface{}); ok && promoted != nil {
			isPromoted = true
		}

		return &models.Trend{
			Name:        name,
			Query:       query,
			TweetVolume: parseTweetVolume(context),
			IsPromoted:  isPromoted,
			Rank:        rank,
		}
	}

	trend, _ := content["trend"].(map[string]interface{})
	if trend == nil {
		return nil
	}

	name, _ := trend["name"].(string)
	if name == "" {
		return nil
	}

	context := ""
	if trendContext, ok := content["trendContext"].(map[string]interface{}); ok {
		context, _ = trendContext["text"].(string)
	} else if socialContext, ok := content["socialContext"].(map[string]interface{}); ok {
		context, _ = socialContext["text"].(string)
	}

	metaDescription := ""
	if trendMetaData, ok := trend["trendMetadata"].(map[string]interface{}); ok {
		metaDescription, _ = trendMetaData["metaDescription"].(string)
	}
	if metaDescription == "" {
		metaDescription = context
	}

	query := name
	if urlValue, ok := trend["url"].(map[string]interface{}); ok {
		if u, ok := urlValue["url"].(string); ok && u != "" {
			query = u
		}
	} else if u, ok := trend["url"].(string); ok && u != "" {
		query = u
	}

	isPromoted := false
	if promoted, ok := trend["promoted_content"].(map[string]interface{}); ok && promoted != nil {
		isPromoted = true
	}

	return &models.Trend{
		Name:        name,
		Query:       query,
		TweetVolume: parseTweetVolume(metaDescription),
		IsPromoted:  isPromoted,
		Rank:        rank,
	}
}

var tweetVolumePattern = regexp.MustCompile(`([\d,.]+)\s*([KkMm])?\s*(posts|tweets)?`)

func parseTweetVolume(text string) int {
	if strings.TrimSpace(text) == "" {
		return 0
	}

	match := tweetVolumePattern.FindStringSubmatch(text)
	if len(match) < 2 {
		return 0
	}

	numStr := strings.ReplaceAll(match[1], ",", "")
	value, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0
	}

	if len(match) > 2 {
		suffix := strings.ToUpper(match[2])
		switch suffix {
		case "K":
			value *= 1000
		case "M":
			value *= 1000000
		}
	}

	return int(value)
}

func getString(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
