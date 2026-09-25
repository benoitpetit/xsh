package core

import (
	"encoding/json"
	"os"
	"testing"
)

func readFixture(t *testing.T, name string) map[string]interface{} {
	t.Helper()
	data, err := os.ReadFile("../testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func TestGetUserMediaParsesTimeline(t *testing.T) {
	fixture := readFixture(t, "user_media.json")
	client := &XClient{requestWithOperationHook: func(_, _ string, _, _ map[string]interface{}, _ int, _, operation string) (map[string]interface{}, error) {
		if operation != "UserMedia" {
			t.Fatalf("operation = %q, want UserMedia", operation)
		}
		return fixture, nil
	}}

	response, err := GetUserMedia(client, "u1", 10, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Tweets) != 1 || response.Tweets[0].ID != "100" || response.CursorBottom != "next-media" {
		t.Fatalf("response = %#v", response)
	}
}

func TestGetTweetResultByRestIDParsesTweet(t *testing.T) {
	fixture := readFixture(t, "tweet_result.json")
	client := &XClient{requestWithOperationHook: func(_, _ string, _, _ map[string]interface{}, _ int, _, operation string) (map[string]interface{}, error) {
		if operation != "TweetResultByRestId" {
			t.Fatalf("operation = %q, want TweetResultByRestId", operation)
		}
		return fixture, nil
	}}

	tweet, err := GetTweetByID(client, "200")
	if err != nil {
		t.Fatal(err)
	}
	if tweet == nil || tweet.ID != "200" || tweet.Text != "direct tweet" {
		t.Fatalf("tweet = %#v", tweet)
	}
}

func TestGetQuoteTweetsParsesTimeline(t *testing.T) {
	fixture := readFixture(t, "quote_tweets.json")
	var operation string
	client := &XClient{requestWithOperationHook: func(_, _ string, _, jsonData map[string]interface{}, _ int, _, op string) (map[string]interface{}, error) {
		operation = op
		variables, ok := jsonData["variables"].(map[string]interface{})
		if !ok || variables["rawQuery"] != "quoted_tweet_id:100" {
			t.Fatalf("variables = %#v", variables)
		}
		return fixture, nil
	}}

	response, err := GetQuoteTweets(client, "100", 10, "")
	if err != nil {
		t.Fatal(err)
	}
	if operation != "SearchTimeline" || len(response.Tweets) != 1 || response.Tweets[0].ID != "300" {
		t.Fatalf("operation=%q response=%#v", operation, response)
	}
}

func TestGetUsersByHandlesUsesScreenNameBatchWhenAvailable(t *testing.T) {
	fixture := readFixture(t, "users_by_screen_names.json")
	var operation string
	client := &XClient{requestWithOperationHook: func(_, _ string, params, _ map[string]interface{}, _ int, _, op string) (map[string]interface{}, error) {
		operation = op
		variables, ok := params["variables"].(map[string]interface{})
		if !ok {
			t.Fatalf("params = %#v", params)
		}
		handles, ok := variables["screenNames"].([]string)
		if !ok || len(handles) != 1 || handles[0] != "batchuser" {
			t.Fatalf("variables = %#v", variables)
		}
		return fixture, nil
	}}

	users, err := GetUsersByHandles(client, []string{"batchuser"})
	if err != nil {
		t.Fatal(err)
	}
	if operation != "UsersByScreenNames" || len(users) != 1 || users[0].Handle != "batchuser" {
		t.Fatalf("operation=%q users=%#v", operation, users)
	}
}
