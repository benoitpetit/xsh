// Package core provides poll operations for Twitter/X.
package core

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// CreatePollCard creates a poll card via the Twitter Cards API and returns the card_uri
// to be used with CreateTweet. Choices must be 2-4 items, duration is in minutes (min 5, max 10080 = 7 days).
func CreatePollCard(client *XClient, choices []string, durationMinutes int) (string, error) {
	if len(choices) < 2 || len(choices) > 4 {
		return "", fmt.Errorf("polls require 2-4 choices, got %d", len(choices))
	}
	if durationMinutes < 5 {
		durationMinutes = 5
	}
	if durationMinutes > 10080 {
		durationMinutes = 10080
	}

	// Build the card_data JSON that the Cards API expects
	cardData := map[string]interface{}{
		"twitter:api:api:endpoint":      "1",
		"twitter:card":                  fmt.Sprintf("poll%dchoice_text_only", len(choices)),
		"twitter:long:duration_minutes": strconv.Itoa(durationMinutes),
	}
	for i, choice := range choices {
		cardData[fmt.Sprintf("twitter:string:choice%d_label", i+1)] = choice
	}

	cardJSON, err := json.Marshal(cardData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal card data: %w", err)
	}

	data := map[string]string{
		"card_data": string(cardJSON),
	}

	if Verbose {
		fmt.Printf("[DEBUG] Creating poll card with %d choices, duration %d min\n", len(choices), durationMinutes)
	}

	result, err := client.RestPost("https://caps.twitter.com/v2/cards/create.json", data)
	if err != nil {
		return "", fmt.Errorf("failed to create poll card: %w", err)
	}

	if Verbose {
		jsonData, _ := json.MarshalIndent(result, "", "  ")
		fmt.Printf("[DEBUG] CreatePollCard response:\n%s\n", string(jsonData))
	}

	// Extract card_uri from response
	cardURI := ""
	if uri, ok := result["card_uri"].(string); ok {
		cardURI = uri
	}

	if cardURI == "" {
		return "", fmt.Errorf("poll card creation did not return a card_uri")
	}

	return cardURI, nil
}
