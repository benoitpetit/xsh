package core

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
)

// BodyLimitError reports a response body that exceeded its configured bound.
type BodyLimitError struct {
	Limit int64
	Size  int64
}

func (e *BodyLimitError) Error() string {
	return fmt.Sprintf("response body exceeds limit of %d bytes", e.Limit)
}

// ReadLimitedBody reads at most limit bytes and detects one byte over the cap.
func ReadLimitedBody(r io.Reader, limit int64) ([]byte, error) {
	if limit <= 0 {
		return nil, &BodyLimitError{Limit: limit}
	}
	data, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, &BodyLimitError{Limit: limit, Size: int64(len(data))}
	}
	return data, nil
}

// ResolveProxy resolves the xsh proxy precedence and validates its scheme.
func ResolveProxy(explicit string, environ func(string) string) (*url.URL, error) {
	if environ == nil {
		environ = os.Getenv
	}
	raw := strings.TrimSpace(explicit)
	if raw == "" {
		for _, key := range []string{
			"CLIX_PROXY", "clix_proxy",
			"X_PROXY", "x_proxy",
			"TWITTER_PROXY", "twitter_proxy",
			"HTTPS_PROXY", "https_proxy",
			"HTTP_PROXY", "http_proxy",
		} {
			if value := strings.TrimSpace(environ(key)); value != "" {
				raw = value
				break
			}
		}
	}
	if raw == "" {
		return nil, nil
	}
	proxyURL, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL %q: %w", raw, err)
	}
	scheme := strings.ToLower(proxyURL.Scheme)
	if (scheme != "http" && scheme != "https") || proxyURL.Host == "" {
		return nil, fmt.Errorf("unsupported or invalid proxy URL %q: use http:// or https://", raw)
	}
	proxyURL.Scheme = scheme
	return proxyURL, nil
}
