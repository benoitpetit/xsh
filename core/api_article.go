// Package core provides article operations for Twitter/X.
package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/benoitpetit/xsh/models"
	"github.com/benoitpetit/xsh/utils"
)

// ArticleResult represents a Twitter/X article (long-form content)
type ArticleResult struct {
	ID        string                 `json:"id"`
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata"`
	Author    string                 `json:"author"`
	CreatedAt *time.Time             `json:"created_at,omitempty"`
}

// GetArticle fetches article data for a tweet that contains an article
func GetArticle(client *XClient, tweetID string) (map[string]interface{}, error) {
	// Use TweetDetail query to get full tweet data including article
	variables := map[string]interface{}{
		"focalTweetId": tweetID,
		"withArticle":  true,
		"withArticlePlainText": false,
		"withArticleRichText":  true,
	}

	data, err := client.GraphQLGet("TweetDetail", variables)
	if err != nil {
		return nil, err
	}

	// Navigate to find article_results
	return extractArticleFromTweetData(data), nil
}

// extractArticleFromTweetData extracts article data from tweet detail response
func extractArticleFromTweetData(data map[string]interface{}) map[string]interface{} {
	// Path: data.threaded_conversation_with_injections_v2.instructions[].entries[].content.itemContent.tweet_results.result.article_results
	
	threadedConv, ok := data["data"].(map[string]interface{})
	if !ok {
		return nil
	}

	instructions, ok := threadedConv["threaded_conversation_with_injections_v2"].(map[string]interface{})
	if !ok {
		return nil
	}

	instructionsArr, ok := instructions["instructions"].([]interface{})
	if !ok {
		return nil
	}

	for _, inst := range instructionsArr {
		instMap, ok := inst.(map[string]interface{})
		if !ok {
			continue
		}

		entries, ok := instMap["entries"].([]interface{})
		if !ok {
			continue
		}

		for _, entry := range entries {
			entryMap, ok := entry.(map[string]interface{})
			if !ok {
				continue
			}

			content, ok := entryMap["content"].(map[string]interface{})
			if !ok {
				continue
			}

			itemContent, ok := content["itemContent"].(map[string]interface{})
			if !ok {
				continue
			}

			tweetResults, ok := itemContent["tweet_results"].(map[string]interface{})
			if !ok {
				continue
			}

			result, ok := tweetResults["result"].(map[string]interface{})
			if !ok {
				continue
			}

			// Check for article_results
			if articleResults, ok := result["article_results"].(map[string]interface{}); ok {
				return articleResults
			}
		}
	}

	return nil
}

// ExportArticleToFile exports an article to a Markdown file
func ExportArticleToFile(articleData map[string]interface{}, tweet *models.Tweet, outputPath string) error {
	if articleData == nil {
		return fmt.Errorf("no article data to export")
	}

	// Extract metadata
	metadata := utils.ExtractArticleMetadata(articleData)
	title, _ := metadata["title"].(string)
	if title == "" {
		title = fmt.Sprintf("Article by @%s", tweet.AuthorHandle)
	}

	// Convert to markdown
	contentMD := utils.ArticleToMarkdown(articleData)

	// Build full markdown with frontmatter
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", title))
	sb.WriteString(fmt.Sprintf("*By @%s*\n\n", tweet.AuthorHandle))
	
	if tweet.CreatedAt != nil {
		sb.WriteString(fmt.Sprintf("*Published: %s*\n\n", tweet.CreatedAt.Format("2006-01-02 15:04")))
	}

	// Add cover image if present
	if coverURL, ok := metadata["cover_image_url"].(string); ok && coverURL != "" {
		sb.WriteString(fmt.Sprintf("![Cover](%s)\n\n", coverURL))
	}

	sb.WriteString(contentMD)

	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Write file
	if err := os.WriteFile(outputPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ArticleToJSON serializes article data to JSON string
func ArticleToJSON(articleData map[string]interface{}, tweet *models.Tweet) (string, error) {
	metadata := utils.ExtractArticleMetadata(articleData)
	contentMD := utils.ArticleToMarkdown(articleData)

	result := map[string]interface{}{
		"tweet": tweet,
		"article": map[string]interface{}{
			"title":         metadata["title"],
			"cover_image_url": metadata["cover_image_url"],
			"markdown":      contentMD,
		},
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}
