package core

import "testing"

func TestIsGraphQLUnprocessableError(t *testing.T) {
	if !isGraphQLUnprocessableError(&APIError{StatusCode: 422, Message: "unprocessable"}) {
		t.Fatalf("expected 422 APIError to be classified as unprocessable")
	}

	if isGraphQLUnprocessableError(&APIError{StatusCode: 404, Message: "not found"}) {
		t.Fatalf("did not expect 404 APIError to be classified as unprocessable")
	}

	if isGraphQLUnprocessableError(nil) {
		t.Fatalf("did not expect nil error to be classified as unprocessable")
	}
}

func TestIsGraphQLEndpointNotFoundResponse(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]interface{}
		want  bool
	}{
		{
			name: "query not found",
			input: map[string]interface{}{
				"errors": []interface{}{
					map[string]interface{}{"message": "Query not found"},
				},
			},
			want: true,
		},
		{
			name: "operation not found",
			input: map[string]interface{}{
				"errors": []interface{}{
					map[string]interface{}{"message": "Operation not found in persisted query"},
				},
			},
			want: true,
		},
		{
			name: "non stale error",
			input: map[string]interface{}{
				"errors": []interface{}{
					map[string]interface{}{"message": "Authorization denied"},
				},
			},
			want: false,
		},
		{
			name:  "no errors",
			input: map[string]interface{}{"data": map[string]interface{}{}},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isGraphQLEndpointNotFoundResponse(tt.input)
			if got != tt.want {
				t.Fatalf("isGraphQLEndpointNotFoundResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}
