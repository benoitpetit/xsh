package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestCLIContractVersionJSONUsesStdoutOnly(t *testing.T) {
	oldArgs := rootCmd.Flags().Args()
	defer rootCmd.SetArgs(oldArgs)
	oldJSON, oldYAML, oldCompact := jsonOutput, yamlOutput, compactMode
	defer func() { jsonOutput, yamlOutput, compactMode = oldJSON, oldYAML, oldCompact }()
	rootCmd.SetArgs([]string{"version", "--json"})
	jsonOutput, yamlOutput, compactMode = false, false, false

	var stdout, stderr bytes.Buffer
	if err := ExecuteContext(context.Background(), strings.NewReader(""), &stdout, &stderr); err != nil {
		t.Fatalf("ExecuteContext() error = %v", err)
	}
	var payload map[string]string
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("stdout is not valid JSON: %v; output=%q", err, stdout.String())
	}
	if payload["version"] == "" || stderr.Len() != 0 {
		t.Fatalf("unexpected CLI contract output stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestCLIContractCancellationReturnsNoOutput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var stdout, stderr bytes.Buffer
	err := ExecuteContext(ctx, strings.NewReader(""), &stdout, &stderr)
	if err == nil || stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("canceled execution err=%v stdout=%q stderr=%q", err, stdout.String(), stderr.String())
	}
}
