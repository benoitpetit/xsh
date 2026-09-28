package cmd

import "testing"

func TestCredentialHintHandlesShortValues(t *testing.T) {
	if got := credentialHint(""); got != "[redacted]" {
		t.Fatalf("credentialHint(empty) = %q, want %q", got, "[redacted]")
	}
	if got := credentialHint("ab"); got != "[redacted]" {
		t.Fatalf("credentialHint(short) = %q, want %q", got, "[redacted]")
	}
	if got := credentialHint("12345678abcdefgh"); got != "12345678..." {
		t.Fatalf("credentialHint(long) = %q, want %q", got, "12345678...")
	}
}
