package core

import (
	"strings"
	"testing"
)

// Static fallbacks are only used when discovery is unavailable, so a stale
// query ID would silently break commands. Keep them shaped like real
// GraphQL operation paths; Trends intentionally points at ExplorePage.
func TestGraphQLEndpointFallbacksMatchOperationNames(t *testing.T) {
	aliases := map[string]string{"Trends": "ExplorePage"}
	for operation, endpoint := range GraphQLEndpoints {
		parts := strings.Split(endpoint, "/")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			t.Errorf("%s fallback %q is not a query-id/operation pair", operation, endpoint)
			continue
		}
		want := aliases[operation]
		if want == "" {
			want = operation
		}
		if parts[1] != want {
			t.Errorf("%s fallback %q points at operation %q, want %q", operation, endpoint, parts[1], want)
		}
	}
}

func TestGraphQLEndpointFallbacksUseCurrentOperationIDs(t *testing.T) {
	want := map[string]string{
		"HomeTimeline":         "KmJCeZ__ef1fNos7tKN_fg/HomeTimeline",
		"UserRepostsTimeline":  "bWzOK-jDN4SYVcF6AQ22Qw/UserRepostsTimeline",
		"CreateTweet":          "l4hk8uVGvR-a8kzQ4nOhPA/CreateTweet",
		"UserTweets":           "D6z-FJI2pKuwEeYYxVIGbg/UserTweets",
		"UserTweetsAndReplies": "FmuxExUC21Xn_C2NwtX5sQ/UserTweetsAndReplies",
		"UserMedia":            "iTSlWbsmD00dAE7Sk9xvIw/UserMedia",
		"Likes":                "ZjyvI1V9gajq5eclHCObzA/Likes",
		"TweetDetail":          "dQDdKtt882nl-Km98FJGSQ/TweetDetail",
		"SearchTimeline":       "vLHjqbDv7TPFMBkIIO1nOA/SearchTimeline",
		"UserByScreenName":     "AMIBMjtxEEATh4z8V9GtRg/UserByScreenName",
	}
	for operation, endpoint := range want {
		if got := GraphQLEndpoints[operation]; got != endpoint {
			t.Errorf("%s fallback = %q, want %q", operation, got, endpoint)
		}
	}
}
