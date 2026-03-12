package tests

import (
	"strings"
	"testing"

	"github.com/benoitpetit/xsh/display"
	"github.com/benoitpetit/xsh/models"
)

// TestFormatConfig teste l'affichage de la configuration
func TestFormatConfig(t *testing.T) {
	cv := &display.ConfigValues{
		DefaultCount:    50,
		DefaultAccount:  "test",
		Theme:           "dark",
		ShowEngagement:  true,
		ShowTimestamps:  false,
		MaxWidth:        120,
		Delay:           2.5,
		Proxy:           "http://proxy.example.com:8080",
		Timeout:         60,
		MaxRetries:      5,
		LikesWeight:     1.5,
		RetweetsWeight:  2.0,
		RepliesWeight:   0.8,
		BookmarksWeight: 2.5,
		ViewsLogWeight:  0.5,
	}

	result := display.FormatConfig(cv)

	if result == "" {
		t.Error("FormatConfig() returned empty string")
	}

	// Vérifier que les valeurs importantes sont présentes
	if !strings.Contains(result, "50") {
		t.Error("FormatConfig() should contain default_count value")
	}

	if !strings.Contains(result, "dark") {
		t.Error("FormatConfig() should contain theme value")
	}

	if !strings.Contains(result, "2.5") && !strings.Contains(result, "2.50") {
		t.Error("FormatConfig() should contain delay value")
	}
}

// TestFormatUserList teste l'affichage de la liste d'utilisateurs
func TestFormatUserList(t *testing.T) {
	users := []*models.User{
		{
			ID:             "1",
			Name:           "Test User 1",
			Handle:         "testuser1",
			Bio:            "Test bio 1",
			FollowersCount: 1000,
			FollowingCount: 500,
			Verified:       true,
		},
		{
			ID:             "2",
			Name:           "Test User 2",
			Handle:         "testuser2",
			Bio:            "Test bio 2",
			FollowersCount: 2000,
			FollowingCount: 1000,
			Verified:       false,
		},
	}

	result := display.FormatUserList(users)

	if result == "" {
		t.Error("FormatUserList() returned empty string")
	}

	// Vérifier que les handles sont présents
	if !strings.Contains(result, "testuser1") {
		t.Error("FormatUserList() should contain first user handle")
	}

	if !strings.Contains(result, "testuser2") {
		t.Error("FormatUserList() should contain second user handle")
	}
}

// TestFormatUserListEmpty teste l'affichage avec une liste vide
func TestFormatUserListEmpty(t *testing.T) {
	users := []*models.User{}

	result := display.FormatUserList(users)

	if result == "" {
		t.Error("FormatUserList() should return message for empty list")
	}

	if !strings.Contains(result, "No users") {
		t.Error("FormatUserList() should indicate no users found")
	}
}

// TestFormatTweet teste l'affichage simple d'un tweet
func TestFormatTweet(t *testing.T) {
	tweet := &models.Tweet{
		ID:             "123456",
		Text:           "This is a test tweet",
		AuthorName:     "Test User",
		AuthorHandle:   "testuser",
		AuthorVerified: true,
		Engagement: models.TweetEngagement{
			Likes:    100,
			Retweets: 50,
		},
	}

	result := display.FormatTweet(tweet, "", true)

	if result == "" {
		t.Error("FormatTweetSimple() returned empty string")
	}

	// Vérifier que le texte est présent
	if !strings.Contains(result, "This is a test tweet") {
		t.Error("FormatTweetSimple() should contain tweet text")
	}

	// Vérifier que le handle est présent
	if !strings.Contains(result, "testuser") {
		t.Error("FormatTweetSimple() should contain author handle")
	}

	// Vérifier que l'ID est présent
	if !strings.Contains(result, "123456") {
		t.Error("FormatTweetSimple() should contain tweet ID")
	}
}

// TestFormatTweetList teste l'affichage d'une liste de tweets en format simple
func TestFormatTweetList(t *testing.T) {
	tweets := []*models.Tweet{
		{
			ID:           "1",
			Text:         "First tweet",
			AuthorHandle: "user1",
		},
		{
			ID:           "2",
			Text:         "Second tweet",
			AuthorHandle: "user2",
		},
	}

	result := display.FormatTweetList(tweets)

	if result == "" {
		t.Error("FormatTweetListSimple() returned empty string")
	}

	// Vérifier que les deux tweets sont présents
	if !strings.Contains(result, "First tweet") {
		t.Error("FormatTweetListSimple() should contain first tweet")
	}

	if !strings.Contains(result, "Second tweet") {
		t.Error("FormatTweetListSimple() should contain second tweet")
	}

	// Vérifier que les IDs sont présents
	if !strings.Contains(result, "1") {
		t.Error("FormatTweetListSimple() should contain first tweet ID")
	}

	if !strings.Contains(result, "2") {
		t.Error("FormatTweetListSimple() should contain second tweet ID")
	}
}



// TestPrintSuccess teste l'affichage des messages de succès
func TestPrintSuccess(t *testing.T) {
	result := display.PrintSuccess("Test success")

	if result == "" {
		t.Error("PrintSuccess() returned empty string")
	}

	if !strings.Contains(result, "Test success") {
		t.Error("PrintSuccess() should contain message")
	}
}

// TestPrintError teste l'affichage des messages d'erreur
func TestPrintError(t *testing.T) {
	result := display.PrintError("Test error")

	if result == "" {
		t.Error("PrintError() returned empty string")
	}

	if !strings.Contains(result, "Test error") {
		t.Error("PrintError() should contain message")
	}
}

// TestPrintWarning teste l'affichage des messages d'avertissement
func TestPrintWarning(t *testing.T) {
	result := display.PrintWarning("Test warning")

	if result == "" {
		t.Error("PrintWarning() returned empty string")
	}

	if !strings.Contains(result, "Test warning") {
		t.Error("PrintWarning() should contain message")
	}
}

// TestFormatUser teste l'affichage d'un profil utilisateur
func TestFormatUser(t *testing.T) {
	user := &models.User{
		ID:             "123",
		Name:           "Test User",
		Handle:         "testuser",
		Bio:            "Test bio",
		FollowersCount: 10000,
		FollowingCount: 1000,
		TweetCount:     5000,
		Verified:       true,
		Location:       "Paris, France",
		Website:        "https://example.com",
	}

	result := display.FormatUser(user)

	if result == "" {
		t.Error("FormatUser() returned empty string")
	}

	// Vérifier que les informations sont présentes
	if !strings.Contains(result, "Test User") {
		t.Error("FormatUser() should contain user name")
	}

	if !strings.Contains(result, "testuser") {
		t.Error("FormatUser() should contain user handle")
	}

	if !strings.Contains(result, "Test bio") {
		t.Error("FormatUser() should contain bio")
	}
}

// TestRelativeTime teste le formatage du temps relatif
func TestRelativeTime(t *testing.T) {
	// Ce test vérifie simplement que la fonction existe et retourne une chaîne
	// Les tests complets dépendent du temps actuel
	now := display.RelativeTime(nil)
	if now != "" {
		t.Error("RelativeTime(nil) should return empty string")
	}
}

// TestFormatNumber teste le formatage des nombres
func TestFormatNumber(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{50, "50"},
		{1000, "1.0K"},
		{1500, "1.5K"},
		{1000000, "1.0M"},
		{2500000, "2.5M"},
	}

	for _, tt := range tests {
		result := display.FormatNumber(tt.input)
		if result != tt.expected {
			t.Errorf("FormatNumber(%d) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}
