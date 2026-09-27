package core

import (
	"fmt"
	"strings"
	"testing"
)

func FuzzCookieSanitizationAndDomain(f *testing.F) {
	for _, seed := range []string{"token", " x;bad\\value ", "", "é"} {
		f.Add(seed, "x.com")
	}
	f.Fuzz(func(t *testing.T, value, domain string) {
		sanitized := SanitizeCookieValue(value)
		if sanitized != SanitizeCookieValue(sanitized) {
			t.Fatalf("sanitization is not idempotent: %q", sanitized)
		}
		_ = isAllowedCookieDomain(strings.ToLower(strings.TrimSpace(domain)))
	})
}

func FuzzGraphQLErrorClassification(f *testing.F) {
	for _, seed := range []string{"Query not found", "operation not found", "temporary error", ""} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, message string) {
		result := map[string]interface{}{"errors": []interface{}{map[string]interface{}{"message": message}}}
		_ = isGraphQLEndpointNotFoundResponse(result)
		_ = isGraphQLEndpointNotFoundResponse(map[string]interface{}{"errors": message})
	})
}

func FuzzEndpointJavaScriptExtraction(f *testing.F) {
	for _, seed := range []string{
		`queryId:"abc-123",operationName:"HomeTimeline"`,
		`fetch("/i/api/graphql/id/UserTweets", {method:"GET"})`,
		`{"operationName":"SearchTimeline","queryId":"id"}`,
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, javascript string) {
		endpoints, features := extractOperationsFromJS(javascript)
		for operation, endpoint := range endpoints {
			if operation == "" || !strings.Contains(endpoint, "/") {
				t.Fatalf("invalid extracted endpoint %q=%q", operation, endpoint)
			}
		}
		for operation, values := range features {
			if operation == "" || values == nil {
				t.Fatalf("invalid feature mapping %q=%v", operation, values)
			}
		}
	})
}

func FuzzAPIErrorFormatting(f *testing.F) {
	f.Add("message", 404, "response")
	f.Fuzz(func(t *testing.T, message string, status int, response string) {
		err := (&APIError{Message: message, StatusCode: status, ResponseData: response}).Error()
		if err == "" {
			t.Fatal("APIError.Error() returned an empty string")
		}
		_ = fmt.Sprint(&StaleEndpointError{APIError: APIError{Message: message, StatusCode: status}})
	})
}
