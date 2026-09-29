package core

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// EndpointProbe reports what a real, read-only GraphQL request established.
// Only an actual data response establishes a healthy operation.
type EndpointProbe struct {
	State   string `json:"state"`
	Message string `json:"message"`
}

// NewEndpointProbeClient uses stored credentials only, so diagnostics cannot
// import browser cookies or change the selected account.
func NewEndpointProbeClient(account string) (*XClient, error) {
	var creds *AuthCredentials
	if account != "" {
		var err error
		creds, err = LoadStoredAuth(account)
		if err != nil {
			return nil, err
		}
	} else {
		creds = getDiscoveryCredentials()
	}
	if creds == nil || !creds.IsValid() {
		return nil, &AuthError{Message: "No stored authentication available for live endpoint checks"}
	}
	return NewXClientWithRequestConfig(creds, account, "", RequestConfig{Timeout: 10, MaxRetries: 1, MaxResponseBytes: defaultMaxResponseBytes})
}

// ProbeGraphQLEndpoint checks a selected operation ID without triggering the
// normal request path's automatic refresh or replaying a mutation.
func ProbeGraphQLEndpoint(ctx context.Context, client *XClient, operation, endpoint string, features map[string]bool) EndpointProbe {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return EndpointProbe{"unreachable", err.Error()}
	}
	parts := strings.Split(endpoint, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] != operation {
		return EndpointProbe{"inconclusive", "Invalid or mismatched operation ID"}
	}
	if client == nil {
		return EndpointProbe{"inconclusive", "No authenticated client available"}
	}
	method, variables, ok := endpointProbeRequest(operation)
	if !ok {
		return EndpointProbe{"unsupported", "No safe probe with known variables for this operation"}
	}

	// A copy keeps the monitor's client context and credentials untouched.
	probeClient := *client
	probeClient.client = nil
	defer probeClient.Close()
	probeClient.SetContext(ctx)
	probeClient.authRefreshAttempted = true
	probeClient.readDelayHook = func() {}
	params := map[string]interface{}{"variables": variables, "features": features, "fieldToggles": DefaultFieldToggles}
	var result map[string]interface{}
	var err error
	if method == http.MethodGet {
		result, err = probeClient.executeRequestWithOperation(method, GraphQLBase+"/"+endpoint, params, nil, 1, "", operation)
	} else {
		result, err = probeClient.executeRequestWithOperation(method, GraphQLBase+"/"+endpoint, nil,
			map[string]interface{}{"variables": variables, "features": features, "queryId": parts[0]}, 1, "", operation)
	}
	if err != nil {
		var apiErr *APIError
		var staleErr *StaleEndpointError
		switch {
		case errors.As(err, &staleErr), errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound:
			return EndpointProbe{"obsolete", "GraphQL operation returned HTTP 404"}
		case errors.As(err, &apiErr) && (apiErr.StatusCode == http.StatusUnauthorized || apiErr.StatusCode == http.StatusForbidden):
			return EndpointProbe{"auth_error", "Authentication was rejected"}
		default:
			var authErr *AuthError
			if errors.As(err, &authErr) {
				return EndpointProbe{"auth_error", "Authentication was rejected"}
			}
			return EndpointProbe{"inconclusive", fmt.Sprintf("GraphQL request failed: %v", err)}
		}
	}
	if isGraphQLEndpointNotFoundResponse(result) {
		return EndpointProbe{"obsolete", "GraphQL response reported query not found"}
	}
	if items, ok := result["errors"].([]interface{}); ok && len(items) > 0 {
		for _, item := range items {
			entry, _ := item.(map[string]interface{})
			message, _ := entry["message"].(string)
			if strings.Contains(strings.ToLower(message), "not authenticated") || strings.Contains(strings.ToLower(message), "unauthorized") {
				return EndpointProbe{"auth_error", "GraphQL response rejected authentication"}
			}
		}
		return EndpointProbe{"inconclusive", "GraphQL returned operation errors"}
	}
	if result["data"] == nil {
		return EndpointProbe{"inconclusive", "GraphQL returned no data"}
	}
	return EndpointProbe{"healthy", "GraphQL returned data"}
}

func endpointProbeRequest(operation string) (string, map[string]interface{}, bool) {
	switch operation {
	case "HomeTimeline":
		return http.MethodGet, map[string]interface{}{"count": 1, "includePromotedContent": false, "latestControlAvailable": true}, true
	case "HomeLatestTimeline":
		return http.MethodGet, map[string]interface{}{"count": 1, "includePromotedContent": false, "latestControlAvailable": true, "requestContext": "launch"}, true
	case "UserByScreenName":
		return http.MethodGet, map[string]interface{}{"screen_name": "x", "withSafetyModeUserFields": true}, true
	case "Viewer", "ExplorePage":
		return http.MethodGet, map[string]interface{}{}, true
	case "AudioSpaceSearch":
		return http.MethodGet, map[string]interface{}{"query": "x", "count": 1}, true
	case "SearchTimeline":
		return http.MethodPost, map[string]interface{}{"rawQuery": "x", "count": 1, "querySource": "typed_query", "product": "Top", "withGrokTranslatedBio": false}, true
	default:
		return "", nil, false
	}
}
