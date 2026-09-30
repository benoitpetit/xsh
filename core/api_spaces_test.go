package core

import "testing"

func TestParseSpaceUserReadsModernUserResult(t *testing.T) {
	participant := map[string]interface{}{
		"user_results": map[string]interface{}{
			"result": map[string]interface{}{
				"rest_id": "user-42",
				"core": map[string]interface{}{
					"name":        "Modern Host",
					"screen_name": "modernhost",
				},
			},
		},
	}

	got := parseSpaceUser(participant)
	if got == nil {
		t.Fatal("parseSpaceUser() = nil, want parsed modern user")
	}
	if got.ID != "user-42" || got.Name != "Modern Host" || got.Handle != "modernhost" {
		t.Fatalf("parseSpaceUser() = %#v, want modern ID/name/handle", got)
	}
}

func TestParseSpaceKeepsReplayAvailabilitySeparateFromTicketing(t *testing.T) {
	data := map[string]interface{}{
		"data": map[string]interface{}{
			"audioSpace": map[string]interface{}{
				"metadata": map[string]interface{}{
					"rest_id":                       "space-1",
					"is_space_available_for_replay": true,
					"is_ticketed":                   false,
					"narrow_cast_space_type":        float64(2),
				},
			},
		},
	}

	got := parseSpace(data)
	if got.IsTicketed {
		t.Fatal("parseSpace() marked replay availability as ticketing")
	}
	if got.NarrowCastSpaceType != 2 {
		t.Fatalf("NarrowCastSpaceType = %d, want 2", got.NarrowCastSpaceType)
	}
}
