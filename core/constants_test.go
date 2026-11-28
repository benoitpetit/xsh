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
	// Every operation that appeared in the same-day authenticated discovery
	// snapshot is pinned here, including aliases and operations whose ID did
	// not change. Operations absent from discovery keep their historical ID.
	want := map[string]string{
		"BlockedAccountsAll":     "exDAeZuprdrqzXJu4UiNUQ/BlockedAccountsAll",
		"BlueVerifiedFollowers":  "HP_mH6gi0qyrYpnRxVovew/BlueVerifiedFollowers",
		"BookmarkSearchTimeline": "pgYWLRBVcyTTpHZPbfstqA/BookmarkSearchTimeline",
		"CreateBookmark":         "aoDbu3RHznuiSkQ9aNM67Q/CreateBookmark",
		"CreateNoteTweet":        "seOZkIS8AysCGfPo6yYPVA/CreateNoteTweet",
		"CreateRetweet":          "mbRO74GrOvSfRcJnlMapnQ/CreateRetweet",
		"CreateTweet":            "l4hk8uVGvR-a8kzQ4nOhPA/CreateTweet",
		"DeleteBookmark":         "Wlmlj2-xzyS1GN3a6cj-mQ/DeleteBookmark",
		"DeleteRetweet":          "ZyZigVsNiFO6v1dEks1eWg/DeleteRetweet",
		"DeleteTweet":            "nxpZCY2K-I6QoFHAHeojFQ/DeleteTweet",
		"ExplorePage":            "CVExg1KY6Ec9BioRBbI8Jg/ExplorePage",
		"FavoriteTweet":          "lI07N6Otwv1PhnEgXILM7A/FavoriteTweet",
		"FollowersYouKnow":       "N-5YQGqcnNO164XwsevHCA/FollowersYouKnow",
		"Following":              "eDqlbOS46zda8jDT8dEQ_Q/Following",
		"HomeTimeline":           "KmJCeZ__ef1fNos7tKN_fg/HomeTimeline",
		"Likes":                  "ZjyvI1V9gajq5eclHCObzA/Likes",
		"MutedAccounts":          "1heSSJ80TWeDGtoR7-X-iA/MutedAccounts",
		"PinTimeline":            "qaz5JhPxqI3UqAgkTpsGyQ/PinTimeline",
		"SearchTimeline":         "vLHjqbDv7TPFMBkIIO1nOA/SearchTimeline",
		"TweetDetail":            "dQDdKtt882nl-Km98FJGSQ/TweetDetail",
		"TweetResultByRestId":    "U8lc5aomF5krJzZ9vJ9wxg/TweetResultByRestId",
		"TweetResultsByRestIds":  "yW_CmYWEKwaOpAEHpNSReg/TweetResultsByRestIds",
		"Trends":                 "CVExg1KY6Ec9BioRBbI8Jg/ExplorePage",
		"UnfavoriteTweet":        "ZYKSe-w7KEslx3JhSIk5LA/UnfavoriteTweet",
		"UnpinTimeline":          "u_Nxpj-y-PoTFHEM9saz2Q/UnpinTimeline",
		"UserByRestId":           "mxaYWwt52eAb6OAzYHMVzg/UserByRestId",
		"UserByScreenName":       "AMIBMjtxEEATh4z8V9GtRg/UserByScreenName",
		"UserMedia":              "iTSlWbsmD00dAE7Sk9xvIw/UserMedia",
		"UserRepostsTimeline":    "bWzOK-jDN4SYVcF6AQ22Qw/UserRepostsTimeline",
		"UsersByRestIds":         "ubfbnAM4qTlMTOYleYhXSg/UsersByRestIds",
		"UsersByScreenNames":     "cACg_kWQkzNQ74lcK5rksQ/UsersByScreenNames",
		"UserTweets":             "D6z-FJI2pKuwEeYYxVIGbg/UserTweets",
		"UserTweetsAndReplies":   "FmuxExUC21Xn_C2NwtX5sQ/UserTweetsAndReplies",
		"Viewer":                 "TpQxQtWMTMnIAMPlMGMqNw/Viewer",
	}
	for operation, endpoint := range want {
		if got := GraphQLEndpoints[operation]; got != endpoint {
			t.Errorf("%s fallback = %q, want %q", operation, got, endpoint)
		}
	}
}
