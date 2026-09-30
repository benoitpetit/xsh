package core

import (
	"encoding/json"
	"testing"
)

func TestExtractNestedIDPreservesExactIDs(t *testing.T) {
	obj := map[string]interface{}{
		"user": map[string]interface{}{"id": json.Number("9007199254740993")},
	}
	if got := extractNestedID(obj, "user"); got != "9007199254740993" {
		t.Fatalf("extractNestedID() = %q, want exact ID", got)
	}
}

func TestExtractNestedIDRejectsRoundedFloatID(t *testing.T) {
	obj := map[string]interface{}{
		"tweet": map[string]interface{}{"id": float64(9007199254740992)},
	}
	if got := extractNestedID(obj, "tweet"); got != "" {
		t.Fatalf("extractNestedID() = %q, want empty for unsafe float ID", got)
	}
}
