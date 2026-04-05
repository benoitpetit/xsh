// Package models provides analytics computation for tweet engagement.
package models

import "sort"

// TweetAnalytics holds aggregated engagement metrics
type TweetAnalytics struct {
	TotalTweets    int            `json:"total_tweets"`
	TotalViews     int            `json:"total_views"`
	TotalLikes     int            `json:"total_likes"`
	TotalRetweets  int            `json:"total_retweets"`
	TotalReplies   int            `json:"total_replies"`
	TotalBookmarks int            `json:"total_bookmarks"`
	AvgViews       float64        `json:"avg_views"`
	AvgLikes       float64        `json:"avg_likes"`
	AvgRetweets    float64        `json:"avg_retweets"`
	AvgReplies     float64        `json:"avg_replies"`
	EngagementRate float64        `json:"engagement_rate_pct"` // (likes+retweets+replies)/views * 100
	TopByLikes     []*Tweet       `json:"top_by_likes"`
	TopByViews     []*Tweet       `json:"top_by_views"`
	TopByRetweets  []*Tweet       `json:"top_by_retweets"`
	MediaBreakdown map[string]int `json:"media_breakdown"` // text, photo, video, etc.
}

// ComputeAnalytics aggregates engagement metrics from a set of tweets
func ComputeAnalytics(tweets []*Tweet) *TweetAnalytics {
	stats := &TweetAnalytics{
		TotalTweets:    len(tweets),
		MediaBreakdown: make(map[string]int),
	}

	for _, t := range tweets {
		e := t.Engagement
		stats.TotalViews += e.Views
		stats.TotalLikes += e.Likes
		stats.TotalRetweets += e.Retweets
		stats.TotalReplies += e.Replies
		stats.TotalBookmarks += e.Bookmarks

		if len(t.Media) == 0 {
			stats.MediaBreakdown["text"]++
		} else {
			for _, m := range t.Media {
				stats.MediaBreakdown[m.Type]++
			}
		}
	}

	n := float64(len(tweets))
	stats.AvgViews = float64(stats.TotalViews) / n
	stats.AvgLikes = float64(stats.TotalLikes) / n
	stats.AvgRetweets = float64(stats.TotalRetweets) / n
	stats.AvgReplies = float64(stats.TotalReplies) / n

	if stats.TotalViews > 0 {
		totalEngagement := float64(stats.TotalLikes + stats.TotalRetweets + stats.TotalReplies)
		stats.EngagementRate = totalEngagement / float64(stats.TotalViews) * 100
	}

	// Top 3 by likes
	topN := 3

	byLikes := make([]*Tweet, len(tweets))
	copy(byLikes, tweets)
	sort.Slice(byLikes, func(i, j int) bool {
		return byLikes[i].Engagement.Likes > byLikes[j].Engagement.Likes
	})
	if len(byLikes) < topN {
		topN = len(byLikes)
	}
	stats.TopByLikes = byLikes[:topN]

	// Top 3 by views
	byViews := make([]*Tweet, len(tweets))
	copy(byViews, tweets)
	sort.Slice(byViews, func(i, j int) bool {
		return byViews[i].Engagement.Views > byViews[j].Engagement.Views
	})
	topN = 3
	if len(byViews) < topN {
		topN = len(byViews)
	}
	stats.TopByViews = byViews[:topN]

	// Top 3 by retweets
	byRetweets := make([]*Tweet, len(tweets))
	copy(byRetweets, tweets)
	sort.Slice(byRetweets, func(i, j int) bool {
		return byRetweets[i].Engagement.Retweets > byRetweets[j].Engagement.Retweets
	})
	topN = 3
	if len(byRetweets) < topN {
		topN = len(byRetweets)
	}
	stats.TopByRetweets = byRetweets[:topN]

	return stats
}
