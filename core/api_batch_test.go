package core

import (
	"fmt"
	"testing"
)

func TestGetUsersByHandlesDeduplicatesCaseInsensitively(t *testing.T) {
	var requested []string
	client := &XClient{requestWithOperationHook: func(_, _ string, params, _ map[string]interface{}, _ int, _, operation string) (map[string]interface{}, error) {
		if operation != "UsersByScreenNames" {
			t.Fatalf("operation = %q, want UsersByScreenNames", operation)
		}
		variables := params["variables"].(map[string]interface{})
		requested = variables["screenNames"].([]string)
		return batchUsersResponse(requested...), nil
	}}

	users, err := GetUsersByHandles(client, []string{"Alice", "alice", "Bob"})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := fmt.Sprint(requested), "[Alice Bob]"; got != want {
		t.Fatalf("requested handles = %s, want %s", got, want)
	}
	if len(users) != 2 || users[0].Handle != "Alice" || users[1].Handle != "Bob" {
		t.Fatalf("users = %#v, want Alice and Bob in first-seen order", users)
	}
}

func TestGetUsersByHandlesFallsBackForPartialBatchResponse(t *testing.T) {
	var fallbackCalls int
	client := &XClient{requestWithOperationHook: func(_, _ string, params, _ map[string]interface{}, _ int, _, operation string) (map[string]interface{}, error) {
		switch operation {
		case "UsersByScreenNames":
			return batchUsersResponse("Alice"), nil
		case "UserByScreenName":
			fallbackCalls++
			variables := params["variables"].(map[string]interface{})
			return singleUserResponse(variables["screen_name"].(string)), nil
		default:
			t.Fatalf("unexpected operation %q", operation)
			return nil, nil
		}
	}}

	users, err := GetUsersByHandles(client, []string{"Alice", "Bob"})
	if err != nil {
		t.Fatal(err)
	}
	if fallbackCalls != 1 || len(users) != 2 || users[1].Handle != "Bob" {
		t.Fatalf("fallbackCalls=%d users=%#v, want one fallback for Bob", fallbackCalls, users)
	}
}

func TestGetUsersByHandlesChunksMoreThanOneHundredHandles(t *testing.T) {
	var batches [][]string
	client := &XClient{requestWithOperationHook: func(_, _ string, params, _ map[string]interface{}, _ int, _, operation string) (map[string]interface{}, error) {
		if operation != "UsersByScreenNames" {
			t.Fatalf("operation = %q, want UsersByScreenNames", operation)
		}
		variables := params["variables"].(map[string]interface{})
		handles := variables["screenNames"].([]string)
		batches = append(batches, handles)
		return batchUsersResponse(handles...), nil
	}}

	handles := make([]string, 101)
	for i := range handles {
		handles[i] = fmt.Sprintf("user-%d", i)
	}
	users, err := GetUsersByHandles(client, handles)
	if err != nil {
		t.Fatal(err)
	}
	if len(batches) != 2 || len(batches[0]) != 100 || len(batches[1]) != 1 {
		t.Fatalf("batches sizes = %d/%d, want 100/1", len(batches[0]), len(batches[1]))
	}
	if len(users) != len(handles) || users[100].Handle != "user-100" {
		t.Fatalf("users length/last handle = %d/%q, want 101/user-100", len(users), users[100].Handle)
	}
}

func TestGetUsersByIDsChunksMoreThanOneHundredIDs(t *testing.T) {
	var batches [][]string
	client := &XClient{requestWithOperationHook: func(_, _ string, params, _ map[string]interface{}, _ int, _, operation string) (map[string]interface{}, error) {
		if operation != "UsersByRestIds" {
			t.Fatalf("operation = %q, want UsersByRestIds", operation)
		}
		variables := params["variables"].(map[string]interface{})
		ids := variables["userIds"].([]string)
		batches = append(batches, ids)
		return batchUsersResponse(ids...), nil
	}}

	ids := make([]string, 205)
	for i := range ids {
		ids[i] = fmt.Sprintf("user-%d", i)
	}
	users, err := GetUsersByIDs(client, ids)
	if err != nil {
		t.Fatal(err)
	}
	if len(batches) != 3 || len(batches[0]) != 100 || len(batches[1]) != 100 || len(batches[2]) != 5 {
		t.Fatalf("batch sizes = %#v, want [100 100 5]", batchSizes(batches))
	}
	if len(users) != len(ids) || users[len(users)-1].Handle != "user-204" {
		t.Fatalf("users length/last handle = %d/%q, want 205/user-204", len(users), users[len(users)-1].Handle)
	}
}

func batchSizes(batches [][]string) []int {
	sizes := make([]int, len(batches))
	for i, batch := range batches {
		sizes[i] = len(batch)
	}
	return sizes
}

func batchUsersResponse(handles ...string) map[string]interface{} {
	results := make([]interface{}, 0, len(handles))
	for _, handle := range handles {
		results = append(results, map[string]interface{}{"result": userResult(handle)})
	}
	return map[string]interface{}{"data": map[string]interface{}{"users": results}}
}

func singleUserResponse(handle string) map[string]interface{} {
	return map[string]interface{}{"data": map[string]interface{}{"user": map[string]interface{}{"result": userResult(handle)}}}
}

func userResult(handle string) map[string]interface{} {
	return map[string]interface{}{
		"rest_id": handle + "-id",
		"legacy": map[string]interface{}{
			"name":        handle,
			"screen_name": handle,
		},
	}
}
