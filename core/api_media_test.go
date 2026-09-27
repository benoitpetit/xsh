package core

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDownloadMediaRejectsOversizedBodyAndCleansPartialFile(t *testing.T) {
	output := filepath.Join(t.TempDir(), "media.bin")
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("0123456789")), Header: make(http.Header), Request: req}, nil
	})}
	if err := DownloadMediaContext(context.Background(), client, "https://example.test/media", output, 5); err == nil {
		t.Fatal("DownloadMediaContext() accepted oversized body")
	}
	if _, err := os.Stat(output); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("partial output still exists: %v", err)
	}
}

func TestDownloadMediaReportsHTTPStatus(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader("missing")), Header: make(http.Header), Request: req}, nil
	})}
	err := DownloadMediaContext(context.Background(), client, "https://example.test/media", filepath.Join(t.TempDir(), "media.bin"), 100)
	if err == nil || !strings.Contains(err.Error(), "HTTP 404") {
		t.Fatalf("DownloadMediaContext() error = %v, want HTTP status", err)
	}
}

func TestDownloadMediaHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, req.Context().Err()
	})}
	err := DownloadMediaContext(ctx, client, "https://example.test/media", filepath.Join(t.TempDir(), "media.bin"), 100)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("DownloadMediaContext() error = %v, want cancellation", err)
	}
}

func TestMakeUploadRequestWithTimeoutPropagatesDeadline(t *testing.T) {
	var deadline time.Time
	client := &XClient{
		credentials: &AuthCredentials{AuthToken: "token", Ct0: "ct0"},
		client: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			deadline, _ = req.Context().Deadline()
			return nil, context.DeadlineExceeded
		})},
	}
	_, err := makeUploadRequestWithTimeout(client, url.Values{"command": {"INIT"}}, 2)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("makeUploadRequestWithTimeout() error = %v, want deadline", err)
	}
	if deadline.IsZero() || time.Until(deadline) <= 0 || time.Until(deadline) > 2*time.Second {
		t.Fatalf("upload request deadline = %v, want about two seconds from now", deadline)
	}
}
