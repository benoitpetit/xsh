package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"
)

func TestUserFollowersErrorIsStructuredInJSONMode(t *testing.T) {
	oldJSON, oldYAML, oldCompact := jsonOutput, yamlOutput, compactMode
	defer func() { jsonOutput, yamlOutput, compactMode = oldJSON, oldYAML, oldCompact }()
	jsonOutput, yamlOutput, compactMode = true, false, false
	var stdout bytes.Buffer
	withRuntime(&Runtime{Out: &stdout, Err: io.Discard}, func() {
		showUserFollowersError("endpoint unavailable")
	})
	var payload map[string]string
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("followers error is not JSON: %v; output=%q", err, stdout.String())
	}
	if payload["error"] != "endpoint unavailable" {
		t.Fatalf("unexpected followers error: %#v", payload)
	}
}
