package models

import "testing"

func TestUserFromAPIResultReadsAvatarImageURL(t *testing.T) {
	result := map[string]interface{}{
		"rest_id": "123",
		"avatar": map[string]interface{}{
			"image_url": "https://pbs.twimg.com/profile_images/123/avatar_normal.jpg",
		},
	}

	user := UserFromAPIResult(result)
	if user == nil {
		t.Fatal("UserFromAPIResult() returned nil")
	}
	if got, want := user.ProfileImageURL, "https://pbs.twimg.com/profile_images/123/avatar_400x400.jpg"; got != want {
		t.Fatalf("ProfileImageURL = %q, want %q", got, want)
	}
}

func TestUserFromAPIResultReadsModernProfileShape(t *testing.T) {
	result := map[string]interface{}{
		"rest_id": "123",
		"core": map[string]interface{}{
			"name":        "Modern Name",
			"screen_name": "modern_handle",
			"created_at":  "Mon Jan 02 15:04:05 +0000 2006",
		},
		"profile_bio": map[string]interface{}{
			"description": "Modern bio",
			"entities": map[string]interface{}{
				"url": map[string]interface{}{
					"urls": []interface{}{map[string]interface{}{
						"expanded_url": "https://example.com/profile",
					}},
				},
			},
		},
		"location": map[string]interface{}{"location": "Paris"},
		"website":  map[string]interface{}{"url": "https://example.com/profile"},
		"relationship_counts": map[string]interface{}{
			"followers": 1234.0,
			"following": 567.0,
		},
		"tweet_counts": map[string]interface{}{"tweets": 890.0},
		"verification": map[string]interface{}{"is_blue_verified": true},
		"avatar":       map[string]interface{}{"image_url": "https://example.com/avatar_normal.jpg"},
		"banner":       map[string]interface{}{"image_url": "https://example.com/banner"},
		"legacy":       nil,
	}

	user := UserFromAPIResult(result)
	if user == nil {
		t.Fatal("UserFromAPIResult() returned nil")
	}
	if user.Name != "Modern Name" || user.Handle != "modern_handle" {
		t.Errorf("name/handle = %q/%q, want Modern Name/modern_handle", user.Name, user.Handle)
	}
	if user.Bio != "Modern bio" || user.Location != "Paris" || user.Website != "https://example.com/profile" {
		t.Errorf("profile text = bio %q, location %q, website %q", user.Bio, user.Location, user.Website)
	}
	if user.FollowersCount != 1234 || user.FollowingCount != 567 || user.TweetCount != 890 {
		t.Errorf("counts = followers %d, following %d, tweets %d; want 1234/567/890", user.FollowersCount, user.FollowingCount, user.TweetCount)
	}
	if !user.Verified {
		t.Error("Verified = false, want true from verification.is_blue_verified")
	}
	if user.ProfileImageURL != "https://example.com/avatar_400x400.jpg" || user.ProfileBannerURL != "https://example.com/banner" {
		t.Errorf("images = %q/%q", user.ProfileImageURL, user.ProfileBannerURL)
	}
	if user.CreatedAt == nil {
		t.Error("CreatedAt = nil, want value from core.created_at")
	}
}

func TestUserFromAPIResultPrefersModernIdentityAndAvatar(t *testing.T) {
	result := map[string]interface{}{
		"rest_id": "123",
		"core": map[string]interface{}{
			"name":        "Current Name",
			"screen_name": "current_handle",
		},
		"avatar": map[string]interface{}{
			"image_url": "https://example.com/current_avatar.jpg",
		},
		"legacy": map[string]interface{}{
			"name":                    "Old Name",
			"screen_name":             "old_handle",
			"profile_image_url_https": "https://example.com/old_avatar.jpg",
		},
	}

	user := UserFromAPIResult(result)
	if user == nil {
		t.Fatal("UserFromAPIResult() returned nil")
	}
	if user.Name != "Current Name" || user.Handle != "current_handle" {
		t.Errorf("identity = %q/%q, want Current Name/current_handle", user.Name, user.Handle)
	}
	if user.ProfileImageURL != "https://example.com/current_avatar.jpg" {
		t.Errorf("ProfileImageURL = %q, want modern avatar URL", user.ProfileImageURL)
	}
}
