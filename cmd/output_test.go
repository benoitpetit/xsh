package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/benoitpetit/xsh/models"
)

func TestEncodeJSONCompactUsesOneLine(t *testing.T) {
	var buffer bytes.Buffer
	if err := encodeJSON(&buffer, map[string]string{"status": "ok"}, true); err != nil {
		t.Fatalf("encodeJSON() error = %v", err)
	}
	if strings.Contains(buffer.String(), "\n\n") || strings.Contains(buffer.String(), "  ") {
		t.Fatalf("compact JSON contains indentation: %q", buffer.String())
	}
	if got, want := buffer.String(), `{"status":"ok"}`+"\n"; got != want {
		t.Fatalf("JSON = %q, want %q", got, want)
	}
}

func TestCompactPageReducesItemsAndKeepsPagination(t *testing.T) {
	compact := toCompact(models.Page[*models.Tweet]{
		Items:      []*models.Tweet{{ID: "1", Text: "hello", AuthorHandle: "ben"}},
		NextCursor: "next",
		HasMore:    true,
	})

	page, ok := compact.(models.Page[map[string]interface{}])
	if !ok || len(page.Items) != 1 {
		t.Fatalf("compact page = %#v, want one compact item", compact)
	}
	if page.Items[0]["id"] != "1" || page.NextCursor != "next" || !page.HasMore {
		t.Fatalf("compact page = %#v, pagination not preserved", page)
	}
}

func TestOutputPageIncludesPaginationMetadataInJSON(t *testing.T) {
	var stdout bytes.Buffer
	previousJSON, previousYAML, previousCompact := jsonOutput, yamlOutput, compactMode
	defer func() {
		jsonOutput, yamlOutput, compactMode = previousJSON, previousYAML, previousCompact
	}()
	jsonOutput, yamlOutput, compactMode = true, false, false

	var outputErr error
	withRuntime(&Runtime{
		Context: context.Background(),
		In:      strings.NewReader(""),
		Out:     &stdout,
		Err:     io.Discard,
		Mode:    JSONOutput,
	}, func() {
		outputErr = outputPage([]string{"item"}, "cursor-2", true, func() {
			t.Fatal("human output ran in JSON mode")
		})
	})
	if outputErr != nil {
		t.Fatalf("outputPage() error = %v", outputErr)
	}

	var payload struct {
		Items      []string `json:"items"`
		NextCursor string   `json:"next_cursor"`
		HasMore    bool     `json:"has_more"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("outputPage() emitted invalid JSON: %v", err)
	}
	if len(payload.Items) != 1 || payload.Items[0] != "item" || payload.NextCursor != "cursor-2" || !payload.HasMore {
		t.Fatalf("page output = %#v", payload)
	}
}

func TestStatusJSONFlagWritesJSONInHumanTerminalMode(t *testing.T) {
	var stdout bytes.Buffer
	previousJSON, previousYAML, previousCompact := jsonOutput, yamlOutput, compactMode
	defer func() {
		jsonOutput, yamlOutput, compactMode = previousJSON, previousYAML, previousCompact
	}()
	jsonOutput, yamlOutput, compactMode = false, false, false

	var outputErr error
	withRuntime(&Runtime{
		Context: context.Background(),
		In:      strings.NewReader(""),
		Out:     &stdout,
		Err:     io.Discard,
		Mode:    HumanOutput,
	}, func() {
		outputErr = outputSystemStatus(&systemStatus{Authenticated: true}, true)
	})
	if outputErr != nil {
		t.Fatalf("outputSystemStatus() error = %v", outputErr)
	}
	var payload systemStatus
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("status JSON output = %q: %v", stdout.String(), err)
	}
	if !payload.Authenticated {
		t.Fatalf("status JSON = %#v", payload)
	}
}
