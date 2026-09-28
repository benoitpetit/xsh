package cmd

import (
	"errors"
	"testing"

	"github.com/benoitpetit/xsh/models"
)

func TestAuthWhoamiPayloadContainsViewerProfile(t *testing.T) {
	profile := &models.User{ID: "u1", Name: "Ben", Handle: "ben", FollowersCount: 12}
	viewer, err := resolveAuthWhoamiProfile(func() (*models.User, error) {
		return profile, nil
	})
	if err != nil {
		t.Fatalf("resolveAuthWhoamiProfile() error = %v", err)
	}
	payload := authWhoamiPayload("default", viewer)
	if payload["authenticated"] != true || payload["account"] != "default" || payload["profile"] != profile || viewer != profile {
		t.Fatalf("authWhoamiPayload() = %#v", payload)
	}
	if _, includesToken := payload["auth_token"]; includesToken {
		t.Fatal("whoami payload must not expose authentication cookies")
	}
}

func TestBuildAuthWhoamiPayloadReturnsViewerErrors(t *testing.T) {
	viewerErr := errors.New("viewer request failed")
	if _, err := resolveAuthWhoamiProfile(func() (*models.User, error) {
		return nil, viewerErr
	}); !errors.Is(err, viewerErr) {
		t.Fatalf("fetch error = %v, want wrapped viewer error", err)
	}

	if _, err := resolveAuthWhoamiProfile(func() (*models.User, error) {
		return nil, nil
	}); err == nil {
		t.Fatal("nil viewer was accepted")
	}
}
