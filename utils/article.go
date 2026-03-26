// Package utils provides helper utilities.
package utils

import (
	"fmt"
	"strings"
)

// ArticleToMarkdown converts a Twitter Article (Draft.js content state) to Markdown
func ArticleToMarkdown(articleData map[string]interface{}) string {
	if articleData == nil {
		return ""
	}

	result, ok := articleData["result"].(map[string]interface{})
	if !ok {
		return ""
	}

	content, ok := result["content"].(map[string]interface{})
	if !ok {
		return ""
	}

	contentState, ok := content["content_state"].(map[string]interface{})
	if !ok {
		return ""
	}

	blocks, ok := contentState["blocks"].([]interface{})
	if !ok {
		return ""
	}

	entityMap := normalizeEntityMap(contentState["entityMap"])

	var lines []string
	for _, block := range blocks {
		blockMap, ok := block.(map[string]interface{})
		if !ok {
			continue
		}

		blockType, _ := blockMap["type"].(string)
		text, _ := blockMap["text"].(string)

		// Handle entity ranges for inline links
		if entityRanges, ok := blockMap["entityRanges"].([]interface{}); ok && len(entityRanges) > 0 {
			text = applyEntityRanges(text, entityRanges, entityMap)
		}

		// Handle inline styles
		if inlineStyleRanges, ok := blockMap["inlineStyleRanges"].([]interface{}); ok && len(inlineStyleRanges) > 0 {
			text = applyInlineStyles(text, inlineStyleRanges)
		}

		switch blockType {
		case "header-one":
			lines = append(lines, "# "+text)
		case "header-two":
			lines = append(lines, "## "+text)
		case "header-three":
			lines = append(lines, "### "+text)
		case "blockquote":
			lines = append(lines, "> "+text)
		case "unordered-list-item":
			lines = append(lines, "- "+text)
		case "ordered-list-item":
			// Find the index in the blocks for numbering
			// For simplicity, we just use 1. here; proper numbering would require tracking
			lines = append(lines, "1. "+text)
		case "code-block":
			lines = append(lines, "```", text, "```")
		case "atomic":
			// Handle atomic blocks (images, videos, etc.)
			atomicContent := handleAtomicBlock(blockMap, entityMap)
			if atomicContent != "" {
				lines = append(lines, atomicContent)
			}
		case "unstyled":
			if text != "" {
				lines = append(lines, text)
			}
		default:
			if text != "" {
				lines = append(lines, text)
			}
		}
	}

	return strings.Join(lines, "\n\n")
}

// ExtractArticleMetadata extracts metadata from article data and returns it as a map.
func ExtractArticleMetadata(articleData map[string]interface{}) map[string]interface{} {
	if articleData == nil {
		return nil
	}

	metadata := make(map[string]interface{})

	// Look for data nested under "result", or directly in articleData
	source := articleData
	if result, ok := articleData["result"].(map[string]interface{}); ok {
		source = result
	}

	// Extract title
	if title, ok := source["title"].(string); ok {
		metadata["title"] = title
	}

	// Extract lifecycle state
	if state, ok := source["lifecycle_state"].(string); ok {
		metadata["lifecycle_state"] = state
	}

	// Extract cover image URL from cover_media.media_info.original_img_url
	if cover, ok := source["cover_media"].(map[string]interface{}); ok {
		if mediaInfo, ok := cover["media_info"].(map[string]interface{}); ok {
			if url, ok := mediaInfo["original_img_url"].(string); ok {
				metadata["cover_image_url"] = url
			}
		}
	}

	return metadata
}

// ArticleMetadata holds article metadata
type ArticleMetadata struct {
	Title         string
	State         string
	CoverImageURL string
}

// normalizeEntityMap normalizes the entity map for easier lookup
func normalizeEntityMap(entityMap interface{}) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})

	if entityMap == nil {
		return result
	}

	switch em := entityMap.(type) {
	case map[string]interface{}:
		for key, value := range em {
			if vm, ok := value.(map[string]interface{}); ok {
				result[key] = vm
			}
		}
	case map[interface{}]interface{}:
		for key, value := range em {
			if keyStr, ok := key.(string); ok {
				if vm, ok := value.(map[string]interface{}); ok {
					result[keyStr] = vm
				}
			}
		}
	}

	return result
}

// applyEntityRanges applies entity ranges (links, mentions) to text
func applyEntityRanges(text string, ranges []interface{}, entityMap map[string]map[string]interface{}) string {
	if len(ranges) == 0 {
		return text
	}

	var result strings.Builder
	lastOffset := 0

	for _, r := range ranges {
		rangeMap, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		offset := 0
		length := 0
		entityKey := ""

		if o, ok := rangeMap["offset"].(float64); ok {
			offset = int(o)
		}
		if l, ok := rangeMap["length"].(float64); ok {
			length = int(l)
		}
		if ek, ok := rangeMap["key"].(float64); ok {
			entityKey = fmt.Sprintf("%d", int(ek))
		} else if ekStr, ok := rangeMap["key"].(string); ok {
			entityKey = ekStr
		}

		// Add text before entity
		if offset > lastOffset {
			result.WriteString(text[lastOffset:offset])
		}

		// Get entity text
		endOffset := offset + length
		if endOffset > len(text) {
			endOffset = len(text)
		}
		entityText := text[offset:endOffset]

		// Apply entity transformation
		if entity, ok := entityMap[entityKey]; ok {
			entityType, _ := entity["type"].(string)
			entityData, _ := entity["data"].(map[string]interface{})

			switch entityType {
			case "LINK":
				if url, ok := entityData["url"].(string); ok {
					result.WriteString(fmt.Sprintf("[%s](%s)", entityText, url))
				} else {
					result.WriteString(entityText)
				}
			case "MENTION":
				if handle, ok := entityData["handle"].(string); ok {
					result.WriteString(fmt.Sprintf("[@%s](https://x.com/%s)", handle, handle))
				} else {
					result.WriteString(entityText)
				}
			default:
				result.WriteString(entityText)
			}
		} else {
			result.WriteString(entityText)
		}

		lastOffset = endOffset
	}

	// Add remaining text
	if lastOffset < len(text) {
		result.WriteString(text[lastOffset:])
	}

	return result.String()
}

// applyInlineStyles applies inline styles (bold, italic, code) to text
func applyInlineStyles(text string, ranges []interface{}) string {
	if len(ranges) == 0 {
		return text
	}

	// Sort ranges by offset
	var sortedRanges []struct {
		offset int
		length int
		style  string
	}

	for _, r := range ranges {
		rangeMap, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		offset := 0
		length := 0
		style := ""

		if o, ok := rangeMap["offset"].(float64); ok {
			offset = int(o)
		}
		if l, ok := rangeMap["length"].(float64); ok {
			length = int(l)
		}
		if s, ok := rangeMap["style"].(string); ok {
			style = s
		}

		sortedRanges = append(sortedRanges, struct {
			offset int
			length int
			style  string
		}{offset, length, style})
	}

	// Simple approach: process styles one at a time
	// For a complete solution, we would need to handle overlapping styles
	result := text
	offset := 0

	for _, r := range sortedRanges {
		if r.offset+r.length > len(result) {
			continue
		}

		before := result[:r.offset+offset]
		textPart := result[r.offset+offset : r.offset+offset+r.length]
		after := result[r.offset+offset+r.length:]

		var wrapped string
		switch r.style {
		case "BOLD":
			wrapped = "**" + textPart + "**"
		case "ITALIC":
			wrapped = "*" + textPart + "*"
		case "CODE":
			wrapped = "`" + textPart + "`"
		default:
			wrapped = textPart
		}

		result = before + wrapped + after
		offset += len(wrapped) - len(textPart)
	}

	return result
}

// handleAtomicBlock handles atomic blocks (images, videos, embeds)
func handleAtomicBlock(block map[string]interface{}, entityMap map[string]map[string]interface{}) string {
	// Check if this block references an entity
	if entityRanges, ok := block["entityRanges"].([]interface{}); ok && len(entityRanges) > 0 {
		for _, r := range entityRanges {
			rangeMap, ok := r.(map[string]interface{})
			if !ok {
				continue
			}

			var entityKey string
			if ek, ok := rangeMap["key"].(float64); ok {
				entityKey = fmt.Sprintf("%d", int(ek))
			} else if ekStr, ok := rangeMap["key"].(string); ok {
				entityKey = ekStr
			}

			if entityData, ok := entityMap[entityKey]; ok {
				entityType, _ := entityData["type"].(string)
				data, _ := entityData["data"].(map[string]interface{})

				switch entityType {
				case "IMAGE":
					if src, ok := data["src"].(string); ok {
						return fmt.Sprintf("![%s](%s)", data["alt"], src)
					}
				case "EMBED":
					if src, ok := data["src"].(string); ok {
						return fmt.Sprintf("[%s](%s)", "Embedded content", src)
					}
				}
			}
		}
	}
	return ""
}
