package models

import (
	"encoding/json"
	"testing"
)

func TestPageMarshalsPaginationMetadata(t *testing.T) {
	page := Page[string]{
		Items:      []string{"one"},
		NextCursor: "cursor-2",
		HasMore:    true,
	}

	data, err := json.Marshal(page)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	got := string(data)
	want := `{"items":["one"],"next_cursor":"cursor-2","has_more":true}`
	if got != want {
		t.Fatalf("json = %s, want %s", got, want)
	}
}

func TestPageOmitsEmptyCursor(t *testing.T) {
	data, err := json.Marshal(Page[string]{Items: []string{}, HasMore: false})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if got, want := string(data), `{"items":[],"has_more":false}`; got != want {
		t.Fatalf("json = %s, want %s", got, want)
	}
}
