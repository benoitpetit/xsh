// Package utils provides helper utilities.
package utils

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/benoitpetit/xsh/models"
)

// FilterConfig holds configuration for tweet filtering and scoring
type FilterConfig struct {
	LikesWeight     float64
	RetweetsWeight  float64
	RepliesWeight   float64
	BookmarksWeight float64
	ViewsLogWeight  float64
	MinScore        float64
}

// DefaultFilterConfig returns default filter configuration
func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		LikesWeight:     1.0,
		RetweetsWeight:  1.5,
		RepliesWeight:   0.5,
		BookmarksWeight: 2.0,
		ViewsLogWeight:  0.3,
		MinScore:        0,
	}
}

// ScoreTweet calculates engagement score for a tweet
func ScoreTweet(tweet *models.Tweet, cfg *FilterConfig) float64 {
	if cfg == nil {
		default_ := DefaultFilterConfig()
		cfg = &default_
	}

	// Calculate log10 of views (add 1 to avoid log(0))
	viewsLog := math.Log10(float64(tweet.Engagement.Views) + 1)

	score := cfg.LikesWeight*float64(tweet.Engagement.Likes) +
		cfg.RetweetsWeight*float64(tweet.Engagement.Retweets) +
		cfg.RepliesWeight*float64(tweet.Engagement.Replies) +
		cfg.BookmarksWeight*float64(tweet.Engagement.Bookmarks) +
		cfg.ViewsLogWeight*viewsLog

	return score
}

// FilterMode defines filtering behavior
type FilterMode string

const (
	// FilterModeAll returns all tweets sorted by score
	FilterModeAll FilterMode = "all"
	// FilterModeTop returns top N tweets by score
	FilterModeTop FilterMode = "top"
	// FilterModeScore returns tweets above threshold
	FilterModeScore FilterMode = "score"
)

// FilterTweets filters and sorts tweets based on engagement score.
// mode is one of "all", "top", "score". minScore is used for score mode,
// limit is used for top mode. cfg may be nil to use defaults.
func FilterTweets(tweets []*models.Tweet, mode string, minScore float64, limit int, cfg *FilterConfig) []*models.Tweet {
	if len(tweets) == 0 {
		return tweets
	}

	if cfg == nil {
		default_ := DefaultFilterConfig()
		cfg = &default_
	}

	// Calculate scores and store with tweets
	type scoredTweet struct {
		tweet *models.Tweet
		score float64
	}

	scored := make([]scoredTweet, 0, len(tweets))
	for _, tweet := range tweets {
		score := ScoreTweet(tweet, cfg)
		scored = append(scored, scoredTweet{tweet, score})
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Apply filtering based on mode
	switch strings.ToLower(mode) {
	case string(FilterModeTop):
		if limit > 0 && limit < len(scored) {
			scored = scored[:limit]
		}
	case string(FilterModeScore):
		filtered := make([]scoredTweet, 0)
		for _, s := range scored {
			if s.score >= minScore {
				filtered = append(filtered, s)
			}
		}
		scored = filtered
	}

	// Extract tweets
	result := make([]*models.Tweet, 0, len(scored))
	for _, s := range scored {
		result = append(result, s.tweet)
	}

	return result
}

// SortTweetsByDate sorts tweets by creation date (newest first)
func SortTweetsByDate(tweets []*models.Tweet) {
	sort.Slice(tweets, func(i, j int) bool {
		if tweets[i].CreatedAt == nil || tweets[j].CreatedAt == nil {
			return tweets[i].CreatedAt != nil
		}
		return tweets[i].CreatedAt.After(*tweets[j].CreatedAt)
	})
}

// SortTweetsByEngagement sorts tweets by total engagement
func SortTweetsByEngagement(tweets []*models.Tweet) {
	sort.Slice(tweets, func(i, j int) bool {
		ei := tweets[i].Engagement.Likes + tweets[i].Engagement.Retweets + tweets[i].Engagement.Replies
		ej := tweets[j].Engagement.Likes + tweets[j].Engagement.Retweets + tweets[j].Engagement.Replies
		return ei > ej
	})
}

// GetTopTweets returns the top N tweets by engagement score
func GetTopTweets(tweets []*models.Tweet, n int) []*models.Tweet {
	return FilterTweets(tweets, string(FilterModeTop), 0, n, nil)
}

// GetTopTweetsByViews returns top tweets by views
func GetTopTweetsByViews(tweets []*models.Tweet, n int) []*models.Tweet {
	sorted := make([]*models.Tweet, len(tweets))
	copy(sorted, tweets)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Engagement.Views > sorted[j].Engagement.Views
	})

	if n > 0 && n < len(sorted) {
		return sorted[:n]
	}
	return sorted
}

// GetTopTweetsByBookmarks returns top tweets by bookmarks
func GetTopTweetsByBookmarks(tweets []*models.Tweet, n int) []*models.Tweet {
	sorted := make([]*models.Tweet, len(tweets))
	copy(sorted, tweets)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Engagement.Bookmarks > sorted[j].Engagement.Bookmarks
	})

	if n > 0 && n < len(sorted) {
		return sorted[:n]
	}
	return sorted
}

// FilterTweetsByTime filters tweets by time range
func FilterTweetsByTime(tweets []*models.Tweet, since, until time.Time) []*models.Tweet {
	var result []*models.Tweet
	for _, tweet := range tweets {
		if !tweet.CreatedAt.Before(since) && !tweet.CreatedAt.After(until) {
			result = append(result, tweet)
		}
	}
	return result
}

// FilterTweetsByAuthor filters tweets by author handle
func FilterTweetsByAuthor(tweets []*models.Tweet, handle string) []*models.Tweet {
	handle = SanitizeHandle(handle)
	var result []*models.Tweet
	for _, tweet := range tweets {
		if SanitizeHandle(tweet.AuthorHandle) == handle {
			result = append(result, tweet)
		}
	}
	return result
}

// FilterTweetsByMedia filters tweets by media presence
func FilterTweetsByMedia(tweets []*models.Tweet, hasMedia bool) []*models.Tweet {
	var result []*models.Tweet
	for _, tweet := range tweets {
		if (len(tweet.Media) > 0) == hasMedia {
			result = append(result, tweet)
		}
	}
	return result
}

// FilterTweetsByText filters tweets containing text (case-insensitive)
func FilterTweetsByText(tweets []*models.Tweet, query string) []*models.Tweet {
	if query == "" {
		return tweets
	}
	query = SanitizeInput(query)
	var result []*models.Tweet
	for _, tweet := range tweets {
		if strings.Contains(strings.ToLower(tweet.Text), strings.ToLower(query)) {
			result = append(result, tweet)
		}
	}
	return result
}

// CalculateStats calculates statistics for a slice of tweets
func CalculateStats(tweets []*models.Tweet) map[string]interface{} {
	if len(tweets) == 0 {
		return map[string]interface{}{
			"count":          0,
			"total_likes":    0,
			"total_retweets": 0,
			"total_replies":  0,
			"avg_engagement": 0.0,
		}
	}

	var totalLikes, totalRetweets, totalReplies, totalViews int
	for _, tweet := range tweets {
		totalLikes += tweet.Engagement.Likes
		totalRetweets += tweet.Engagement.Retweets
		totalReplies += tweet.Engagement.Replies
		totalViews += tweet.Engagement.Views
	}

	count := float64(len(tweets))
	return map[string]interface{}{
		"count":          len(tweets),
		"total_likes":    totalLikes,
		"total_retweets": totalRetweets,
		"total_replies":  totalReplies,
		"total_views":    totalViews,
		"avg_likes":      float64(totalLikes) / count,
		"avg_retweets":   float64(totalRetweets) / count,
		"avg_replies":    float64(totalReplies) / count,
		"avg_views":      float64(totalViews) / count,
	}
}

// CalculateEngagementRate calculates engagement rate for a tweet
func CalculateEngagementRate(tweet *models.Tweet) float64 {
	if tweet.Engagement.Views == 0 {
		return 0
	}
	total := tweet.Engagement.Likes + tweet.Engagement.Retweets + tweet.Engagement.Replies
	return float64(total) / float64(tweet.Engagement.Views) * 100
}

// CalculateAverageEngagementRate calculates average engagement rate for tweets
func CalculateAverageEngagementRate(tweets []*models.Tweet) float64 {
	if len(tweets) == 0 {
		return 0
	}
	var total float64
	for _, tweet := range tweets {
		total += CalculateEngagementRate(tweet)
	}
	return total / float64(len(tweets))
}
