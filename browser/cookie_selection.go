package browser

import (
	"fmt"
	"strings"

	"github.com/benoitpetit/xsh/core"
)

type cookieAccumulator struct {
	cookies map[string]string
	hosts   map[string]string
	errors  map[string]error
}

func newCookieAccumulator() *cookieAccumulator {
	return &cookieAccumulator{
		cookies: make(map[string]string),
		hosts:   make(map[string]string),
		errors:  make(map[string]error),
	}
}

func (a *cookieAccumulator) add(host, name, value string, decryptErr error) {
	sanitized := core.SanitizeCookieValue(value)
	if sanitized == "" && decryptErr == nil {
		return
	}

	existingHost, exists := a.hosts[name]
	if exists && !preferCookieHost(host, existingHost, sanitized, a.cookies[name]) {
		return
	}

	a.hosts[name] = host
	if sanitized != "" {
		a.cookies[name] = sanitized
		delete(a.errors, name)
	} else if decryptErr != nil && isEssentialCookie(name) {
		a.errors[name] = decryptErr
	}
}

func (a *cookieAccumulator) credentials(browserName string) (*core.AuthCredentials, error) {
	authToken := a.cookies["auth_token"]
	ct0 := a.cookies["ct0"]
	if authToken == "" || ct0 == "" {
		if a.errors["auth_token"] != nil || a.errors["ct0"] != nil {
			return nil, fmt.Errorf("could not decrypt %s authentication cookies; browser key storage is unavailable: auth_token=%v, ct0=%v", browserName, a.errors["auth_token"] != nil, a.errors["ct0"] != nil)
		}
		return nil, fmt.Errorf("auth_token or ct0 not found in %s cookies. Make sure you're logged into x.com in %s", browserName, browserName)
	}

	cookies := make(map[string]string, len(a.cookies))
	for name, value := range a.cookies {
		cookies[name] = value
	}
	return &core.AuthCredentials{AuthToken: authToken, Ct0: ct0, Cookies: cookies}, nil
}

func isEssentialCookie(name string) bool {
	return name == "auth_token" || name == "ct0"
}

func preferCookieHost(candidateHost, existingHost, candidateValue, existingValue string) bool {
	candidatePriority := cookieHostPriority(candidateHost)
	existingPriority := cookieHostPriority(existingHost)
	if candidatePriority != existingPriority {
		return candidatePriority > existingPriority
	}
	return existingValue == "" && candidateValue != ""
}

func cookieHostPriority(host string) int {
	normalized := strings.ToLower(strings.TrimSpace(host))
	switch normalized {
	case "x.com", "twitter.com":
		return 3
	case ".x.com", ".twitter.com":
		return 2
	default:
		return 1
	}
}
