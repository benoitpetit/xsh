package browser

import (
	"errors"
	"testing"
)

var errCookieUnavailable = errors.New("cookie unavailable")

func TestCookieAccumulatorPrefersExactHost(t *testing.T) {
	acc := newCookieAccumulator()
	acc.add("profile.x.com", "ct0", "subdomain", nil)
	acc.add(".x.com", "ct0", "domain", nil)
	acc.add("x.com", "ct0", "exact", nil)
	acc.add("x.com", "auth_token", `"token"`, nil)

	creds, err := acc.credentials("Chrome")
	if err != nil {
		t.Fatalf("credentials() error = %v", err)
	}
	if creds.Ct0 != "exact" {
		t.Fatalf("Ct0 = %q, want exact-host value", creds.Ct0)
	}
	if creds.AuthToken != "token" {
		t.Fatalf("AuthToken = %q, want sanitized exact-host value", creds.AuthToken)
	}
}

func TestCookieAccumulatorRequiresEssentialCookies(t *testing.T) {
	acc := newCookieAccumulator()
	acc.add("x.com", "auth_token", "token", nil)

	if _, err := acc.credentials("Firefox"); err == nil {
		t.Fatal("credentials() error = nil, want missing ct0 error")
	}
}

func TestCookieAccumulatorIgnoresFailedNonEssentialCookie(t *testing.T) {
	acc := newCookieAccumulator()
	acc.add("x.com", "guest_id", "", errCookieUnavailable)
	acc.add("x.com", "auth_token", "token", nil)
	acc.add("x.com", "ct0", "csrf", nil)

	if _, err := acc.credentials("Chrome"); err != nil {
		t.Fatalf("credentials() error = %v, want non-essential failure ignored", err)
	}
}
