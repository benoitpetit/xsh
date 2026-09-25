package cmd

import (
	"bytes"
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
