package core

import "net/url"

// AddQueryParams appends encoded query parameters without discarding an
// existing query string.
func AddQueryParams(rawURL string, params url.Values) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	query := parsed.Query()
	for key, values := range params {
		for _, value := range values {
			query.Add(key, value)
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}
