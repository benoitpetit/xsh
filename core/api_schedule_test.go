package core

import (
	"encoding/json"
	"testing"
)

func TestParseJSONIDPreservesExactRepresentations(t *testing.T) {
	tests := []struct {
		name string
		id   interface{}
		want string
	}{
		{name: "string ID", id: "1234567890123456789", want: "1234567890123456789"},
		{name: "JSON number", id: json.Number("1234567890123456789"), want: "1234567890123456789"},
		{name: "safe float", id: float64(1234), want: "1234"},
		{name: "unsafe float", id: float64(9007199254740992), want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseJSONID(tt.id); got != tt.want {
				t.Fatalf("parseJSONID(%v) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}
