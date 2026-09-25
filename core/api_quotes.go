// Package core provides quote tweet operations for Twitter/X.
package core

import (
	"net/url"

	"github.com/benoitpetit/xsh/models"
)

// GetQuoteTweets fetches tweets that quote a specific tweet
func GetQuoteTweets(client *XClient, tweetID string, count int, cursor string) (*models.TimelineResponse, error) {
	variables := map[string]interface{}{
		"rawQuery":    "quoted_tweet_id:" + tweetID,
		"count":       count,
		"querySource": "tdqt",
		"product":     "Top",
	}

	if cursor != "" {
		variables["cursor"] = cursor
	}

	data, err := client.GraphQLPostWithReferer("SearchTimeline", variables, BaseURL+"/search?q="+url.QueryEscape("quoted_tweet_id:"+tweetID))
	if err != nil {
		return nil, err
	}

	return extractTweetsFromTimeline(data), nil
}
