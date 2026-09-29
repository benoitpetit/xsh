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
