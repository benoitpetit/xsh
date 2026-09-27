package core

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

func TestRunStartupCheckHonorsContextBeforeNetwork(t *testing.T) {
	t.Setenv("XSH_CONFIG_DIR", t.TempDir())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var diagnostics bytes.Buffer

	err := RunStartupCheck(ctx, &diagnostics)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RunStartupCheck() error = %v, want context.Canceled", err)
	}
	if diagnostics.Len() != 0 {
		t.Fatalf("startup diagnostics = %q, want empty output on cancellation", diagnostics.String())
	}
}

func TestStartupUpdatePromptUsesDiagnosticWriter(t *testing.T) {
	var diagnostics bytes.Buffer
	checker := &StartupChecker{}
	checker.ShowUpdatePrompt([]string{"SearchTimeline"}, &diagnostics)
	if diagnostics.Len() == 0 || !bytes.Contains(diagnostics.Bytes(), []byte("SearchTimeline")) {
		t.Fatalf("diagnostic output = %q, want endpoint warning", diagnostics.String())
	}
}
