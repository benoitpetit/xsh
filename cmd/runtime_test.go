package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/benoitpetit/xsh/core"
)

func TestEncodeJSONRedactsShortCredentials(t *testing.T) {
	var out bytes.Buffer
	creds := &core.AuthCredentials{AuthToken: "short", Ct0: "csrf", AccountName: "demo"}

	if err := encodeJSON(&out, creds, false); err != nil {
		t.Fatalf("encodeJSON() error = %v", err)
	}
	var payload map[string]string
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("encoded JSON is invalid: %v", err)
	}
	if payload["auth_token"] == "short" || payload["ct0"] == "csrf" {
		t.Fatalf("short credentials were exposed: %#v", payload)
	}
	if !strings.Contains(out.String(), "demo") {
		t.Fatalf("account name missing from output: %s", out.String())
	}
}

func TestValidateOutputFlagsRejectsConflicts(t *testing.T) {
	for _, tc := range []struct {
		name       string
		json, yaml bool
		compact    bool
	}{
		{name: "json yaml", json: true, yaml: true},
		{name: "json compact", json: true, compact: true},
		{name: "yaml compact", yaml: true, compact: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateOutputFlags(tc.json, tc.yaml, tc.compact); err == nil {
				t.Fatal("ValidateOutputFlags() accepted conflicting formats")
			}
		})
	}
}

func TestOutputUsesRuntimeStdoutAndStderr(t *testing.T) {
	var stdout, stderr bytes.Buffer
	runtime := &Runtime{Context: context.Background(), In: strings.NewReader(""), Out: &stdout, Err: &stderr, Mode: JSONOutput}
	withRuntime(runtime, func() {
		if err := outputJSON(map[string]string{"status": "ok"}); err != nil {
			t.Fatalf("outputJSON() error = %v", err)
		}
		fmtDiagnostic("diagnostic")
	})

	if got := strings.TrimSpace(stdout.String()); got != `{"status":"ok"}` && got != "{\n  \"status\": \"ok\"\n}" {
		t.Fatalf("stdout = %q, want JSON only", got)
	}
	if stderr.String() != "diagnostic\n" {
		t.Fatalf("stderr = %q, want diagnostic", stderr.String())
	}
}

func TestJSONModeFallsBackToMachineOutputWhenStatFails(t *testing.T) {
	withRuntime(&Runtime{Context: context.Background(), Out: statErrorWriter{}, Err: io.Discard}, func() {
		jsonOutput = false
		yamlOutput = false
		compactMode = false
		if !isJSONMode() {
			t.Fatal("isJSONMode() = false after stdout Stat failure, want machine mode")
		}
	})
}

func TestOutputReturnsWriterErrors(t *testing.T) {
	errBroken := errors.New("broken pipe")
	writer := errorWriter{err: errBroken}
	if err := encodeJSON(writer, map[string]string{"status": "ok"}, false); !errors.Is(err, errBroken) {
		t.Fatalf("encodeJSON() error = %v, want %v", err, errBroken)
	}
}

func TestExecuteContextReturnsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	oldArgs := rootCmd.Flags().Args()
	_ = oldArgs

	err := ExecuteContext(ctx, strings.NewReader(""), io.Discard, io.Discard)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ExecuteContext() error = %v, want context.Canceled", err)
	}
}

type errorWriter struct{ err error }

func (w errorWriter) Write([]byte) (int, error) { return 0, w.err }

type statErrorWriter struct{}

func (statErrorWriter) Write(p []byte) (int, error) { return len(p), nil }
func (statErrorWriter) Stat() (os.FileInfo, error)  { return nil, errors.New("stat failed") }
