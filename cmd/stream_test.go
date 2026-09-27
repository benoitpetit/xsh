package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/benoitpetit/xsh/models"
)

func TestRunWithWatchHonorsRuntimeCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	previous := watchInterval
	watchInterval = 1
	defer func() { watchInterval = previous }()

	var stderr bytes.Buffer
	var gotFetch int
	var gotErr error
	withRuntime(&Runtime{Context: ctx, Out: &bytes.Buffer{}, Err: &stderr}, func() {
		gotErr = runWithWatch(func() error {
			gotFetch++
			return nil
		})
	})

	if !errors.Is(gotErr, context.Canceled) {
		t.Fatalf("runWithWatch() error = %v, want context.Canceled", gotErr)
	}
	if gotFetch != 1 {
		t.Fatalf("fetch called %d times, want once", gotFetch)
	}
}

func TestEmitNewTweetsWritesNDJSONOnlyToEncoder(t *testing.T) {
	var stdout, stderr bytes.Buffer
	withRuntime(&Runtime{Context: context.Background(), Out: &stdout, Err: &stderr}, func() {
		tweets := []*models.Tweet{{ID: "1", Text: "hello", AuthorHandle: "ben"}}
		if err := emitNewTweets(tweets, map[string]bool{}, json.NewEncoder(&stdout)); err != nil {
			t.Fatalf("emitNewTweets() error = %v", err)
		}
	})

	line := strings.TrimSpace(stdout.String())
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(line), &payload); err != nil {
		t.Fatalf("stdout is not one JSON object: %v; output=%q", err, stdout.String())
	}
	if payload["id"] != "1" || payload["type"] != "tweet" {
		t.Fatalf("unexpected event: %#v", payload)
	}
	if !strings.Contains(stderr.String(), "+1 new items") {
		t.Fatalf("progress was not sent to stderr: %q", stderr.String())
	}
}

func TestStreamOnceRejectsUnknownSource(t *testing.T) {
	err := streamOnce(nil, "unknown", nil, map[string]bool{}, json.NewEncoder(&bytes.Buffer{}))
	if err == nil || !strings.Contains(err.Error(), "unknown source") {
		t.Fatalf("streamOnce() error = %v, want unknown source error", err)
	}
}
