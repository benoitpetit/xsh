package utils

import (
	"math"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/benoitpetit/xsh/models"
)

// TwitterHandleRegex matches valid Twitter handles
// Rules: 1-15 chars, alphanumeric + underscore, no spaces, must start with letter or underscore
var twitterHandleRegex = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]{0,14}$`)

// ValidateTwitterHandle validates a Twitter handle
// Returns the cleaned handle and true if valid, false otherwise
func ValidateTwitterHandle(handle string) (string, bool) {
	// Remove @ prefix if present
	handle = strings.TrimPrefix(handle, "@")
	
	// Check basic constraints
	if handle == "" {
		return "", false
	}
	
	// Check length (Twitter allows 1-15 characters)
	if utf8.RuneCountInString(handle) > 15 {
		return "", false
	}
	
	// Check against regex
	if !twitterHandleRegex.MatchString(handle) {
		return "", false
	}
	
	return handle, true
}

// TweetIDRegex matches valid Twitter tweet IDs (numeric strings)
var TweetIDRegex = regexp.MustCompile(`^[0-9]+$`)

// ValidateTweetID validates a tweet ID
func ValidateTweetID(id string) bool {
	if id == "" {
		return false
	}
	// Tweet IDs are numeric and typically 19 digits
	if !TweetIDRegex.MatchString(id) {
		return false
	}
	// Check reasonable length (min 10, max 25 digits)
	if len(id) < 10 || len(id) > 25 {
		return false
	}
	return true
}

// ValidateTweetText validates tweet text length
// Twitter allows 280 characters for standard tweets, 4000 for Twitter Blue
func ValidateTweetText(text string, maxLength int) (string, bool) {
	if maxLength <= 0 {
		maxLength = 280 // Standard tweet limit
	}
	
	// Check if empty
	text = strings.TrimSpace(text)
	if text == "" {
		return "", false
	}
	
	// Check length (count runes for Unicode)
	if utf8.RuneCountInString(text) > maxLength {
		return "", false
	}
	
	// Remove null bytes and control characters except newlines
	var result strings.Builder
	for _, r := range text {
		if r == 0 {
			continue
		}
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			continue
		}
		result.WriteRune(r)
	}
	
	return result.String(), true
}

// SanitizeInput removes potentially dangerous characters from user input
func SanitizeInput(input string, maxLength int) string {
	if maxLength <= 0 {
		maxLength = 1000
	}
	
	// Trim whitespace
	input = strings.TrimSpace(input)
	
	// Limit length
	if len(input) > maxLength {
		input = input[:maxLength]
	}
	
	// Remove null bytes and control characters
	var result strings.Builder
	for _, r := range input {
		if r >= 32 || r == '\t' || r == '\n' || r == '\r' {
			result.WriteRune(r)
		}
	}
	
	return result.String()
}

// FilterConfig contains filter settings for tweet scoring
type FilterConfig struct {
	LikesWeight      float64
	RetweetsWeight   float64
	RepliesWeight    float64
	BookmarksWeight  float64
	ViewsLogWeight   float64
}

// DefaultFilterConfig returns the default filter configuration
func DefaultFilterConfig() *FilterConfig {
	return &FilterConfig{
		LikesWeight:     1.0,
		RetweetsWeight:  1.5,
		RepliesWeight:   0.5,
		BookmarksWeight: 2.0,
		ViewsLogWeight:  0.3,
	}
}

// ScoreTweet calculates engagement score for a tweet
func ScoreTweet(tweet *models.Tweet, config *FilterConfig) float64 {
	if config == nil {
		config = DefaultFilterConfig()
	}

	e := tweet.Engagement
	viewsLog := 0.0
	if e.Views > 0 {
		viewsLog = math.Log10(float64(e.Views) + 1)
	}

	return config.LikesWeight*float64(e.Likes) +
		config.RetweetsWeight*float64(e.Retweets) +
		config.RepliesWeight*float64(e.Replies) +
		config.BookmarksWeight*float64(e.Bookmarks) +
		config.ViewsLogWeight*viewsLog
}

// FilterTweets filters and sorts tweets by engagement score
// Modes:
//   - "all": Sort by score, no filtering
//   - "score": Keep tweets above threshold
//   - "top": Keep top N tweets
func FilterTweets(tweets []*models.Tweet, mode string, threshold float64, topN int, config *FilterConfig) []*models.Tweet {
	// Score all tweets
	type scoredTweet struct {
		tweet *models.Tweet
		score float64
	}

	scored := make([]scoredTweet, 0, len(tweets))
	for _, tweet := range tweets {
		scored = append(scored, scoredTweet{
			tweet: tweet,
			score: ScoreTweet(tweet, config),
		})
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Apply filter
	switch mode {
	case "score":
		filtered := make([]scoredTweet, 0)
		for _, s := range scored {
			if s.score >= threshold {
				filtered = append(filtered, s)
			}
		}
		scored = filtered
	case "top":
		if topN < len(scored) {
			scored = scored[:topN]
		}
	}

	// Extract tweets
	result := make([]*models.Tweet, 0, len(scored))
	for _, s := range scored {
		result = append(result, s.tweet)
	}

	return result
}
