package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/benoitpetit/xsh/core"
)

func FuzzJSONRedaction(f *testing.F) {
	for _, seed := range []string{"short", "auth-token-value", "", "token with spaces"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, value string) {
		var output bytes.Buffer
		if err := encodeJSON(&output, &core.AuthCredentials{AuthToken: value, Ct0: value}, true); err != nil {
			t.Fatal(err)
		}
		if len(value) > 8 && strings.Contains(output.String(), `"auth_token":"`+value+`"`) {
			t.Fatalf("full credential leaked in %q", output.String())
		}
	})
}
