// Package models provides data models for xsh.
package models

// Trend represents a trending topic
type Trend struct {
	Name        string `json:"name"`
	Query       string `json:"query"`
	TweetVolume int    `json:"tweet_volume"`
	IsPromoted  bool   `json:"is_promoted"`
	Rank        int    `json:"rank"`
}
