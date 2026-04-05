// Package core provides core functionality for the xsh Twitter/X client.
package core

// Base URLs
const (
	BaseURL     = "https://x.com"
	APIBase     = "https://x.com/i/api"
	GraphQLBase = APIBase + "/graphql"
)

// Bearer token (public, embedded in Twitter web app)
const BearerToken = "AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA"

// GraphQL operation IDs - extracted from X.com JS bundles (updated 2026-04-02)
// These are fallback values. The actual endpoints are fetched dynamically from X.com
// and cached in ~/.config/xsh/graphql_ops.json
var GraphQLEndpoints = map[string]string{
	// Read operations - Timeline
	"HomeTimeline":       "gXtpuBkna6SRLFFKaT2OTg/HomeTimeline",
	"HomeLatestTimeline": "JVzDMxTXbT9bRXSpUR16CQ/HomeLatestTimeline",

	// Read operations - Search
	"SearchTimeline": "OFvapAUD5xrCWn9xcD0A6A/SearchTimeline",

	// Read operations - Tweets
	"TweetDetail":           "1eAGnXrtvTBUePpQfTXZzA/TweetDetail",
	"TweetResultByRestId":   "tcA4FFMIjGSDv48Cu_FS5Q/TweetResultByRestId",
	"TweetResultsByRestIds": "M441-7OPnT7o_TzVwteU3Q/TweetResultsByRestIds",

	// Read operations - Users
	"UserByScreenName":   "pLsOiyHJ1eFwPJlNmLp4Bg/UserByScreenName",
	"UserByRestId":       "FJ17ptkJuQAZGWilcySi5w/UserByRestId",
	"UsersByRestIds":     "8OKmcyotfczJb44QTTu5tQ/UsersByRestIds",
	"UsersByScreenNames": "Ats5GnHiQxT-Nnzw09raMw/UsersByScreenNames",

	// Read operations - User Timelines
	"UserTweets":           "5M8UuGym7_VyIEggQIyjxQ/UserTweets",
	"UserTweetsAndReplies": "C3YpYjTsQZznJIdyy2JKuQ/UserTweetsAndReplies",
	"UserMedia":            "mWo2yKjZEaqK7_vKox_67Q/UserMedia",
	"Likes":                "dv5-II7_Bup_PHish7p6fw/Likes",

	// Read operations - Followers/Following
	"Followers":             "8sIMO3RbSCdvk2QzxcPpIg/Followers",
	"Following":             "lEJDj0bTio9-s0hSukCD9Q/Following",
	"FollowersYouKnow":      "fBi9FJP1haBdGoZuVfZVzQ/FollowersYouKnow",
	"BlueVerifiedFollowers": "ZH16zF8R8YAJAAfIGbef9A/BlueVerifiedFollowers",

	// Read operations - Bookmarks
	"Bookmarks":              "uKP9v_I31k0_VSBmlpq2Xg/Bookmarks",
	"BookmarkSearchTimeline": "vBCEp1KR36nwY3u2K-2nEA/BookmarkSearchTimeline",
	"BookmarkFoldersSlice":   "i78YDd0Tza-dV4SYs58kRg/BookmarkFoldersSlice",
	"BookmarkFolderTimeline": "hNY7X2xE2N7HVF6Qb_mu6w/BookmarkFolderTimeline",

	// Read operations - Lists
	"ListLatestTweetsTimeline":    "gNXkRRRV67cSRJkmpgGPnA/ListLatestTweetsTimeline",
	"ListsManagementPageTimeline": "qXWhKTaNeJianB02_JovCg/ListsManagementPageTimeline",
	"ListByRestId":                "bSE1lqLBnovM86uu4p4Iqg/ListByRestId",
	"ListMembers":                 "fqecRWCF4EcSAOs5yXh7Ig/ListMembers",
	"ListMemberships":             "cRsrj8HASXYzxaf90wMDPQ/ListMemberships",

	// Read operations - Jobs
	"JobSearchQueryScreenJobsQuery": "jVMK9qcOUB5xQQdSLr5ECg/JobSearchQueryScreenJobsQuery",
	"JobScreenQuery":                "8uZH_OBKTFNIMzTJaV5lbQ/JobScreenQuery",

	// Read operations - Explore/Trends
	"ExplorePage": "Z6s1tFEq4BveGOj0N80z8g/ExplorePage",
	"Trends":      "Z6s1tFEq4BveGOj0N80z8g/ExplorePage",

	// Write operations - Tweets (updated 2026-04-02 from dynamic cache)
	"CreateTweet":     "BLx8gngFhHE5eBgLBCT_0Q/CreateTweet",
	"CreateNoteTweet": "4e-YHiuiNDaITMxa29cerw/CreateNoteTweet",
	"DeleteTweet":     "nxpZCY2K-I6QoFHAHeojFQ/DeleteTweet",

	// Write operations - Likes
	"FavoriteTweet":   "lI07N6Otwv1PhnEgXILM7A/FavoriteTweet",
	"UnfavoriteTweet": "ZYKSe-w7KEslx3JhSIk5LA/UnfavoriteTweet",

	// Write operations - Retweets
	"CreateRetweet": "mbRO74GrOvSfRcJnlMapnQ/CreateRetweet",
	"DeleteRetweet": "ZyZigVsNiFO6v1dEks1eWg/DeleteRetweet",

	// Write operations - Bookmarks
	"CreateBookmark": "aoDbu3RHznuiSkQ9aNM67Q/CreateBookmark",
	"DeleteBookmark": "Wlmlj2-xzyS1GN3a6cj-mQ/DeleteBookmark",

	// Write operations - DMs
	"DMMessageDeleteMutation": "BJ6DtxA2llfjnRoRjaiIiw/DMMessageDeleteMutation",

	// Write operations - Scheduled Tweets
	"CreateScheduledTweet": "LCVzRQGxOaGnOnYH01NQXg/CreateScheduledTweet",
	"FetchScheduledTweets": "ITtjAzvlZni2wWXwf295Qg/FetchScheduledTweets",
	"DeleteScheduledTweet": "CTOVqej0JBXAZSwkp1US0g/DeleteScheduledTweet",

	// Write operations - Lists
	"CreateList":       "QXil-VE8uEJPfUKFiO36Bg/CreateList",
	"UpdateList":       "qE2QVWL84jqa6CmH-m-D3w/UpdateList",
	"DeleteList":       "UnN9Th1BDbeLjpgjGSpL3Q/DeleteList",
	"ListAddMember":    "nAi8BAjn1xQOyCH0hWZpPA/ListAddMember",
	"ListRemoveMember": "pGMiwtWRMx08r4XCYxai4Q/ListRemoveMember",
	"PinTimeline":      "t-vQkLuhUq-GvXLbRXXMFA/PinTimeline",
	"UnpinTimeline":    "agrJf0pu-b_3p53wUjZEFA/UnpinTimeline",

	// Write operations - Follow
	"FollowUser":   "iNZ4xPly3JveJqEshzJdLA/FollowUser",
	"UnfollowUser": "tTT1x7v4h9zLVzT1BA/UnfollowUser",

	// Other operations
	"Viewer":             "zWQLM9HIVahRSUvzUH4lDw/Viewer",
	"BlockedAccountsAll": "tnGJ8xjxfnv032LSOpKYWQ/BlockedAccountsAll",
	"MutedAccounts":      "CI5MBQ8KdnNx87ZXOfTsBg/MutedAccounts",

	// Read operations - Communities
	"CommunitiesMainPageTimeline": "ZYMHvNRSCYjfaxcK_XBIOQ/CommunitiesMainPageTimeline",
	"CommunityByRestId":           "nNjM4SaF_W1Hk0JQ8pOPQg/CommunityByRestId",
	"CommunityTweetsTimeline":     "aeJsV55JMPw15_-N4VCnPA/CommunityTweetsTimeline",
	"JoinCommunity":               "bTM1mhHSML-sKB9EjT0RRg/JoinCommunity",
	"LeaveCommunity":              "OoS6Nd__MZ0LHAiBbmc3gQ/LeaveCommunity",

	// Read operations - Spaces
	"AudioSpaceById":   "Uv5R_x2bBDPMVxfGjkzxtA/AudioSpaceById",
	"AudioSpaceSearch": "NTq79TuSz6fHj8lQaferJw/AudioSpaceSearch",

	// Notifications: uses REST API v2 (/i/api/2/notifications/all.json), not GraphQL

	// Read operations - Quotes
	"TweetQuotes": "kv6oBLVLyPi0JWVqrY4kJQ/TweetQuotes",
}

// DefaultFeatures for GraphQL requests
var DefaultFeatures = map[string]bool{
	"rweb_tipjar_consumption_enabled":                                         true,
	"responsive_web_graphql_exclude_directive_enabled":                        true,
	"verified_phone_label_enabled":                                            false,
	"creator_subscriptions_tweet_preview_api_enabled":                         true,
	"responsive_web_graphql_timeline_navigation_enabled":                      true,
	"responsive_web_graphql_skip_user_profile_image_extensions_enabled":       false,
	"communities_web_enable_tweet_community_results_fetch":                    true,
	"c9s_tweet_anatomy_moderator_badge_enabled":                               true,
	"articles_preview_enabled":                                                true,
	"responsive_web_edit_tweet_api_enabled":                                   true,
	"graphql_is_translatable_rweb_tweet_is_translatable_enabled":              true,
	"view_counts_everywhere_api_enabled":                                      true,
	"longform_notetweets_consumption_enabled":                                 true,
	"responsive_web_twitter_article_tweet_consumption_enabled":                true,
	"tweet_awards_web_tipping_enabled":                                        false,
	"creator_subscriptions_quote_tweet_preview_enabled":                       false,
	"freedom_of_speech_not_reach_fetch_enabled":                               true,
	"standardized_nudges_misinfo":                                             true,
	"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": true,
	"rweb_video_timestamps_enabled":                                           true,
	"longform_notetweets_rich_text_read_enabled":                              true,
	"longform_notetweets_inline_media_enabled":                                true,
	"responsive_web_enhance_cards_enabled":                                    true,
}

// DefaultFieldToggles for GraphQL requests
var DefaultFieldToggles = map[string]bool{
	"withArticlePlainText": false,
}

// Request defaults
const (
	DefaultCount     = 20
	MaxCount         = 100
	DefaultDelaySec  = 1.5
	MinWriteDelaySec = 1.5
	MaxWriteDelaySec = 4.0
)

// Exit codes
const (
	ExitSuccess   = 0
	ExitError     = 1
	ExitAuthError = 2
	ExitRateLimit = 3
)

// Config paths
const (
	ConfigDirName  = "xsh"
	ConfigFileName = "config.toml"
	AuthFileName   = "auth.json"
)

// UserAgentVersions contains Chrome version strings for User-Agent rotation
var UserAgentVersions = []string{
	"120.0.0.0",
	"123.0.0.0",
	"124.0.0.0",
	"126.0.0.0",
	"127.0.0.0",
	"131.0.0.0",
	"133.0.0.0",
}

// TLSFingerprintType represents different browser fingerprint types
type TLSFingerprintType string

const (
	Chrome120 TLSFingerprintType = "chrome120"
	Chrome123 TLSFingerprintType = "chrome123"
	Chrome124 TLSFingerprintType = "chrome124"
	Chrome126 TLSFingerprintType = "chrome126"
	Chrome127 TLSFingerprintType = "chrome127"
	Chrome131 TLSFingerprintType = "chrome131"
	Chrome133 TLSFingerprintType = "chrome133"
)

// ChromeVersions contains all supported Chrome versions for User-Agent rotation
var ChromeVersions = []TLSFingerprintType{
	Chrome131, Chrome133, Chrome127, Chrome126, Chrome124, Chrome123, Chrome120,
}

// chromeVersionStrings maps fingerprint types to version strings for User-Agent
var chromeVersionStrings = map[TLSFingerprintType]string{
	Chrome120: "120.0.0.0",
	Chrome123: "123.0.0.0",
	Chrome124: "124.0.0.0",
	Chrome126: "126.0.0.0",
	Chrome127: "127.0.0.0",
	Chrome131: "131.0.0.0",
	Chrome133: "133.0.0.0",
}
