// Package tests provides unit tests for article conversion utilities.
package tests

import (
	"strings"
	"testing"

	"github.com/benoitpetit/xsh/utils"
)

func TestArticleToMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		article  map[string]interface{}
		expected []string // Substrings expected in output
	}{
		{
			name: "empty article",
			article: map[string]interface{}{
				"result": map[string]interface{}{
					"content": map[string]interface{}{
						"content_state": map[string]interface{}{
							"blocks": []interface{}{},
						},
					},
				},
			},
			expected: []string{},
		},
		{
			name: "header one",
			article: map[string]interface{}{
				"result": map[string]interface{}{
					"content": map[string]interface{}{
						"content_state": map[string]interface{}{
							"blocks": []interface{}{
								map[string]interface{}{
									"type": "header-one",
									"text": "Test Header",
								},
							},
						},
					},
				},
			},
			expected: []string{"# Test Header"},
		},
		{
			name: "unordered list",
			article: map[string]interface{}{
				"result": map[string]interface{}{
					"content": map[string]interface{}{
						"content_state": map[string]interface{}{
							"blocks": []interface{}{
								map[string]interface{}{
									"type": "unordered-list-item",
									"text": "Item 1",
								},
								map[string]interface{}{
									"type": "unordered-list-item",
									"text": "Item 2",
								},
							},
						},
					},
				},
			},
			expected: []string{"- Item 1", "- Item 2"},
		},
		{
			name: "blockquote",
			article: map[string]interface{}{
				"result": map[string]interface{}{
					"content": map[string]interface{}{
						"content_state": map[string]interface{}{
							"blocks": []interface{}{
								map[string]interface{}{
									"type": "blockquote",
									"text": "A quote",
								},
							},
						},
					},
				},
			},
			expected: []string{"> A quote"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ArticleToMarkdown(tt.article)
			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("ArticleToMarkdown() missing expected content %q in:\n%s", expected, result)
				}
			}
		})
	}
}

func TestExtractArticleMetadata(t *testing.T) {
	article := map[string]interface{}{
		"result": map[string]interface{}{
			"title":            "Test Article",
			"lifecycle_state":  "published",
			"cover_media": map[string]interface{}{
				"media_info": map[string]interface{}{
					"original_img_url": "https://example.com/image.jpg",
				},
			},
		},
	}

	metadata := utils.ExtractArticleMetadata(article)

	if title, ok := metadata["title"].(string); !ok || title != "Test Article" {
		t.Errorf("ExtractArticleMetadata() Title = %v, want %v", metadata["title"], "Test Article")
	}

	if state, ok := metadata["lifecycle_state"].(string); !ok || state != "published" {
		t.Errorf("ExtractArticleMetadata() LifecycleState = %v, want %v", metadata["lifecycle_state"], "published")
	}

	if coverURL, ok := metadata["cover_image_url"].(string); !ok || coverURL != "https://example.com/image.jpg" {
		t.Errorf("ExtractArticleMetadata() CoverImageURL = %v, want %v", metadata["cover_image_url"], "https://example.com/image.jpg")
	}
}

func TestExtractArticleMetadata_NoCover(t *testing.T) {
	article := map[string]interface{}{
		"title":           "Simple Article",
		"lifecycle_state": "draft",
	}

	metadata := utils.ExtractArticleMetadata(article)

	if title, ok := metadata["title"].(string); !ok || title != "Simple Article" {
		t.Errorf("ExtractArticleMetadata() Title = %v, want %v", metadata["title"], "Simple Article")
	}

	if coverURL, ok := metadata["cover_image_url"].(string); ok && coverURL != "" {
		t.Errorf("ExtractArticleMetadata() CoverImageURL should be empty, got %v", coverURL)
	}
}
