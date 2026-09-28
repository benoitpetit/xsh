package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/benoitpetit/xsh/models"
)

func TestSelectUserBatchInput(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		ids       []string
		wantByIDs bool
		wantCount int
		wantError bool
	}{
		{name: "handles", args: []string{"alice", "bob"}, wantCount: 2},
		{name: "ids", ids: []string{"10", "20"}, wantByIDs: true, wantCount: 2},
		{name: "empty", wantError: true},
		{name: "mixed", args: []string{"alice"}, ids: []string{"10"}, wantError: true},
		{name: "invalid ID", ids: []string{"alice"}, wantError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			byIDs, values, err := selectUserBatchInput(tt.args, tt.ids)
			if (err != nil) != tt.wantError {
				t.Fatalf("selectUserBatchInput() error = %v, wantError %v", err, tt.wantError)
			}
			if err == nil && (byIDs != tt.wantByIDs || len(values) != tt.wantCount) {
				t.Fatalf("selectUserBatchInput() = byIDs %v, %d values; want %v, %d", byIDs, len(values), tt.wantByIDs, tt.wantCount)
			}
		})
	}
}

func TestWriteArticleExportJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "article.json")
	article := map[string]interface{}{
		"result": map[string]interface{}{
			"title": "Test article",
			"content": map[string]interface{}{
				"content_state": map[string]interface{}{
					"blocks": []interface{}{map[string]interface{}{"text": "Hello article", "type": "unstyled"}},
				},
			},
		},
	}
	tweet := &models.Tweet{ID: "42", Text: "Attached tweet"}
	if err := writeArticleExport(article, tweet, path, "json"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var payload struct {
		Tweet   models.Tweet `json:"tweet"`
		Article struct {
			Title    string `json:"title"`
			Markdown string `json:"markdown"`
		} `json:"article"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("article export is not valid JSON: %v", err)
	}
	if payload.Tweet.ID != tweet.ID || payload.Article.Title != "Test article" || payload.Article.Markdown == "" {
		t.Fatalf("article export payload = %#v", payload)
	}
}

func TestWriteArticleExportRejectsUnknownFormat(t *testing.T) {
	if err := writeArticleExport(nil, nil, "article.bin", "xml"); err == nil {
		t.Fatal("unsupported format accepted")
	}
}

func TestValidateCommunityExploreCount(t *testing.T) {
	for _, count := range []int{1, 20, 100} {
		if err := validateCommunityExploreCount(count); err != nil {
			t.Errorf("validateCommunityExploreCount(%d) = %v", count, err)
		}
	}
	for _, count := range []int{0, -1, 101} {
		if err := validateCommunityExploreCount(count); err == nil {
			t.Errorf("validateCommunityExploreCount(%d) accepted invalid count", count)
		}
	}
}

func TestNewEndpointCommandsAreRegistered(t *testing.T) {
	if communityExploreCmd.Parent() != communityCmd {
		t.Fatal("community explore is not registered under community")
	}
	if tweetGetCmd.Parent() != tweetCmd {
		t.Fatal("tweet get is not registered under tweet")
	}
	if usersBatchCmd.Flags().Lookup("id") == nil {
		t.Fatal("users --id flag is not registered")
	}
	if communityExploreCmd.Flags().Lookup("cursor") == nil || communityExploreCmd.Flags().Lookup("count") == nil {
		t.Fatal("community explore pagination flags are not registered")
	}
}
