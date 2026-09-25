package core

import "testing"

func TestCreateNoteTweetBuildsExpectedVariables(t *testing.T) {
	client := &XClient{requestWithOperationHook: func(method, _ string, _ map[string]interface{}, jsonData map[string]interface{}, _ int, _, operation string) (map[string]interface{}, error) {
		if method != "POST" || operation != "CreateNoteTweet" {
			t.Fatalf("method=%q operation=%q", method, operation)
		}
		variables, ok := jsonData["variables"].(map[string]interface{})
		if !ok || variables["tweet_text"] != "long form content" || variables["dark_request"] != false {
			t.Fatalf("variables=%#v", variables)
		}
		media, ok := variables["media"].(map[string]interface{})
		if !ok || len(media["media_entities"].([]map[string]interface{})) != 0 {
			t.Fatalf("media=%#v", variables["media"])
		}
		return map[string]interface{}{
			"data": map[string]interface{}{
				"create_note_tweet": map[string]interface{}{
					"tweet_results": map[string]interface{}{
						"result": map[string]interface{}{"rest_id": "900"},
					},
				},
			},
		}, nil
	}}

	result, err := CreateNoteTweet(client, "long form content", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !isTweetCreated(result) {
		t.Fatalf("result was not recognized as created: %#v", result)
	}
}
