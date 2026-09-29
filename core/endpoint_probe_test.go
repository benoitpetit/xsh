package core

import (
	"context"
	"net/http"
	"reflect"
	"testing"
)

func TestEndpointProbeRequestUsesKnownSafeReadVariables(t *testing.T) {
	tests := []struct {
		operation string
		method    string
		variables map[string]interface{}
	}{
		{"Viewer", http.MethodGet, map[string]interface{}{}},
		{"ExplorePage", http.MethodGet, map[string]interface{}{}},
		{"AudioSpaceSearch", http.MethodGet, map[string]interface{}{"query": "x", "count": 1}},
	}
	for _, tt := range tests {
		t.Run(tt.operation, func(t *testing.T) {
			method, variables, ok := endpointProbeRequest(tt.operation)
			if !ok {
				t.Fatal("known read operation has no safe probe")
			}
			if method != tt.method {
				t.Fatalf("method = %q, want %q", method, tt.method)
			}
			if !reflect.DeepEqual(variables, tt.variables) {
				t.Fatalf("variables = %#v, want %#v", variables, tt.variables)
			}
		})
	}
}

func TestProbeGraphQLEndpointSendsSafeReadRequests(t *testing.T) {
	tests := []struct {
		operation string
		variables map[string]interface{}
	}{
		{"Viewer", map[string]interface{}{}},
		{"ExplorePage", map[string]interface{}{}},
		{"AudioSpaceSearch", map[string]interface{}{"query": "x", "count": 1}},
	}
	for _, tt := range tests {
		t.Run(tt.operation, func(t *testing.T) {
			client := &XClient{requestWithOperationHook: func(method, url string, params, body map[string]interface{}, attempts int, _ string, operation string) (map[string]interface{}, error) {
				if method != http.MethodGet || url != GraphQLBase+"/query/"+tt.operation || attempts != 1 || operation != tt.operation {
					t.Fatalf("unexpected probe request: %s %s attempts=%d operation=%s", method, url, attempts, operation)
				}
				if body != nil {
					t.Fatalf("read-only GET unexpectedly had a request body: %#v", body)
				}
				gotVariables, _ := params["variables"].(map[string]interface{})
				if !reflect.DeepEqual(gotVariables, tt.variables) {
					t.Fatalf("request variables = %#v, want %#v", gotVariables, tt.variables)
				}
				return map[string]interface{}{"data": map[string]interface{}{}}, nil
			}}
			probe := ProbeGraphQLEndpoint(context.Background(), client, tt.operation, "query/"+tt.operation, nil)
			if probe.State != "healthy" {
				t.Fatalf("probe state = %q, want healthy (%s)", probe.State, probe.Message)
			}
		})
	}
}

func TestProbeGraphQLEndpointClassifiesLiveResponses(t *testing.T) {
	tests := []struct {
		name   string
		result map[string]interface{}
		err    error
		want   string
	}{
		{"data", map[string]interface{}{"data": map[string]interface{}{"user": map[string]interface{}{}}}, nil, "healthy"},
		{"missing query", map[string]interface{}{"errors": []interface{}{map[string]interface{}{"message": "Query not found"}}}, nil, "obsolete"},
		{"expired session", map[string]interface{}{"errors": []interface{}{map[string]interface{}{"message": "Authentication: Not authenticated"}}}, nil, "auth_error"},
		{"application error", map[string]interface{}{"errors": []interface{}{map[string]interface{}{"message": "User not found"}}}, nil, "inconclusive"},
		{"404", nil, &StaleEndpointError{APIError: APIError{StatusCode: http.StatusNotFound}}, "obsolete"},
		{"403", nil, &APIError{StatusCode: http.StatusForbidden}, "auth_error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &XClient{requestWithOperationHook: func(method, url string, _, _ map[string]interface{}, attempts int, _, operation string) (map[string]interface{}, error) {
				if method != http.MethodGet || url != GraphQLBase+"/query/UserByScreenName" || attempts != 1 || operation != "UserByScreenName" {
					t.Fatalf("unexpected probe request: %s %s attempts=%d operation=%s", method, url, attempts, operation)
				}
				return tt.result, tt.err
			}}
			got := ProbeGraphQLEndpoint(context.Background(), client, "UserByScreenName", "query/UserByScreenName", nil)
			if got.State != tt.want {
				t.Fatalf("probe state = %q, want %q (%s)", got.State, tt.want, got.Message)
			}
		})
	}
}

func TestProbeGraphQLEndpointDoesNotSendUnknownOperation(t *testing.T) {
	client := &XClient{requestWithOperationHook: func(_, _ string, _, _ map[string]interface{}, _ int, _, _ string) (map[string]interface{}, error) {
		t.Fatal("unsupported operation was sent")
		return nil, nil
	}}
	got := ProbeGraphQLEndpoint(context.Background(), client, "CreateTweet", "query/CreateTweet", nil)
	if got.State != "unsupported" {
		t.Fatalf("probe state = %q, want unsupported", got.State)
	}
}
