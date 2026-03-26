package tests

import (
	"testing"

	"github.com/benoitpetit/xsh/models"
	"github.com/benoitpetit/xsh/utils"
)

// TestScoreTweet teste le calcul du score d'engagement
func TestScoreTweet(t *testing.T) {
	// Créer un tweet avec des métriques connues
	tweet := &models.Tweet{
		ID:   "123",
		Text: "Test tweet",
		Engagement: models.TweetEngagement{
			Likes:     100,
			Retweets:  50,
			Replies:   25,
			Bookmarks: 10,
			Views:     1000,
		},
	}

	cfg := utils.DefaultFilterConfig()
	score := utils.ScoreTweet(tweet, &cfg)

	// Calcul attendu:
	// likes: 100 * 1.0 = 100
	// retweets: 50 * 1.5 = 75
	// replies: 25 * 0.5 = 12.5
	// bookmarks: 10 * 2.0 = 20
	// views_log: log10(1001) * 0.3 ≈ 0.9
	// Total ≈ 208.4

	if score <= 0 {
		t.Errorf("ScoreTweet() = %f, want positive value", score)
	}

	// Vérifier que le score est raisonnable (entre 200 et 210)
	if score < 200 || score > 210 {
		t.Errorf("ScoreTweet() = %f, want between 200 and 210", score)
	}
}

// TestScoreTweetWithNilConfig teste le scoring avec config nil
func TestScoreTweetWithNilConfig(t *testing.T) {
	tweet := &models.Tweet{
		ID:   "123",
		Text: "Test tweet",
		Engagement: models.TweetEngagement{
			Likes: 100,
		},
	}

	// Test avec config nil (doit utiliser la config par défaut)
	score := utils.ScoreTweet(tweet, nil)

	if score <= 0 {
		t.Errorf("ScoreTweet() with nil config = %f, want positive value", score)
	}
}

// TestScoreTweetZeroViews teste le scoring avec 0 vues
func TestScoreTweetZeroViews(t *testing.T) {
	tweet := &models.Tweet{
		ID:   "123",
		Text: "Test tweet",
		Engagement: models.TweetEngagement{
			Likes: 100,
			Views: 0,
		},
	}

	cfg := utils.DefaultFilterConfig()
	score := utils.ScoreTweet(tweet, &cfg)

	// Avec 0 vues, views_log devrait être 0
	expectedScore := 100.0 // Juste les likes
	if score != expectedScore {
		t.Errorf("ScoreTweet() = %f, want %f", score, expectedScore)
	}
}

// TestFilterTweetsAllMode teste le filtrage en mode "all"
func TestFilterTweetsAllMode(t *testing.T) {
	tweets := []*models.Tweet{
		{ID: "1", Engagement: models.TweetEngagement{Likes: 100}},
		{ID: "2", Engagement: models.TweetEngagement{Likes: 50}},
		{ID: "3", Engagement: models.TweetEngagement{Likes: 200}},
	}

	result := utils.FilterTweets(tweets, "all", 0, 0, nil)

	if len(result) != 3 {
		t.Errorf("FilterTweets(all) returned %d tweets, want 3", len(result))
	}

	// Vérifier que les tweets sont triés par score (décroissant)
	if result[0].ID != "3" {
		t.Errorf("First tweet should be ID 3 (highest score), got %s", result[0].ID)
	}
	if result[2].ID != "2" {
		t.Errorf("Last tweet should be ID 2 (lowest score), got %s", result[2].ID)
	}
}

// TestFilterTweetsTopMode teste le filtrage en mode "top"
func TestFilterTweetsTopMode(t *testing.T) {
	tweets := []*models.Tweet{
		{ID: "1", Engagement: models.TweetEngagement{Likes: 100}},
		{ID: "2", Engagement: models.TweetEngagement{Likes: 50}},
		{ID: "3", Engagement: models.TweetEngagement{Likes: 200}},
		{ID: "4", Engagement: models.TweetEngagement{Likes: 150}},
		{ID: "5", Engagement: models.TweetEngagement{Likes: 25}},
	}

	result := utils.FilterTweets(tweets, "top", 0, 3, nil)

	if len(result) != 3 {
		t.Errorf("FilterTweets(top, 3) returned %d tweets, want 3", len(result))
	}

	// Vérifier que ce sont les 3 meilleurs
	if result[0].ID != "3" || result[1].ID != "4" || result[2].ID != "1" {
		t.Error("FilterTweets(top) did not return top 3 tweets in correct order")
	}
}

// TestFilterTweetsScoreMode teste le filtrage en mode "score"
func TestFilterTweetsScoreMode(t *testing.T) {
	tweets := []*models.Tweet{
		{ID: "1", Engagement: models.TweetEngagement{Likes: 100}},
		{ID: "2", Engagement: models.TweetEngagement{Likes: 50}},
		{ID: "3", Engagement: models.TweetEngagement{Likes: 200}},
	}

	// Filtrer avec un seuil de 75 (devrait garder 1 et 3)
	result := utils.FilterTweets(tweets, "score", 75, 0, nil)

	if len(result) != 2 {
		t.Errorf("FilterTweets(score, 75) returned %d tweets, want 2", len(result))
	}

	// Vérifier que les tweets avec score >= 75 sont présents
	hasID1 := false
	hasID3 := false
	for _, t := range result {
		if t.ID == "1" {
			hasID1 = true
		}
		if t.ID == "3" {
			hasID3 = true
		}
	}

	if !hasID1 {
		t.Error("FilterTweets(score) should include tweet with ID 1")
	}
	if !hasID3 {
		t.Error("FilterTweets(score) should include tweet with ID 3")
	}
}

// TestFilterTweetsEmptyList teste le filtrage avec une liste vide
func TestFilterTweetsEmptyList(t *testing.T) {
	var tweets []*models.Tweet

	result := utils.FilterTweets(tweets, "all", 0, 0, nil)

	if len(result) != 0 {
		t.Errorf("FilterTweets(empty) returned %d tweets, want 0", len(result))
	}
}

// TestFilterTweetsTopNLargerThanList teste top N plus grand que la liste
func TestFilterTweetsTopNLargerThanList(t *testing.T) {
	tweets := []*models.Tweet{
		{ID: "1", Engagement: models.TweetEngagement{Likes: 100}},
		{ID: "2", Engagement: models.TweetEngagement{Likes: 50}},
	}

	// Demander top 10 mais il n'y a que 2 tweets
	result := utils.FilterTweets(tweets, "top", 0, 10, nil)

	if len(result) != 2 {
		t.Errorf("FilterTweets(top, 10) returned %d tweets, want 2", len(result))
	}
}

// TestFilterTweetsCustomWeights teste le filtrage avec des pondérations personnalisées
func TestFilterTweetsCustomWeights(t *testing.T) {
	tweets := []*models.Tweet{
		{ID: "1", Engagement: models.TweetEngagement{Likes: 100, Retweets: 0}},
		{ID: "2", Engagement: models.TweetEngagement{Likes: 0, Retweets: 100}},
	}

	// Config avec pondération élevée pour les retweets
	cfg := &utils.FilterConfig{
		LikesWeight:    1.0,
		RetweetsWeight: 10.0, // Les retweets comptent 10x plus
		RepliesWeight:  0.5,
		BookmarksWeight: 2.0,
		ViewsLogWeight: 0.3,
	}

	result := utils.FilterTweets(tweets, "all", 0, 0, cfg)

	// Le tweet 2 devrait être premier car 100 retweets * 10 = 1000 > 100 likes * 1 = 100
	if result[0].ID != "2" {
		t.Errorf("First tweet should be ID 2 (high retweet weight), got %s", result[0].ID)
	}
}
