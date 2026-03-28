package tests

import (
	"testing"
)

// TestMain is the entry point for all tests
func TestMain(m *testing.M) {
	// Global test configuration if needed
	m.Run()
}

// TestSuite verifies that all packages are testable
func TestSuite(t *testing.T) {
	t.Run("Auth", func(t *testing.T) {
		t.Run("Credentials", TestAuthCredentials)
		t.Run("EnvAuth", TestEnvAuth)
		t.Run("StoredAuth", TestStoredAuth)
		t.Run("ListAccounts", TestListAccounts)
		t.Run("SetDefaultAccount", TestSetDefaultAccount)
		t.Run("RemoveAuth", TestRemoveAuth)
		t.Run("SanitizeCookieValue", TestSanitizeCookieValue)
		t.Run("CredentialsSanitization", TestAuthCredentialsSanitization)
	})

	t.Run("Config", func(t *testing.T) {
		t.Run("DefaultConfig", TestDefaultConfig)
		t.Run("LoadSave", TestConfigLoadSave)
		t.Run("FilterConfig", TestFilterConfig)
		t.Run("SetValues", TestConfigSetValues)
	})

	t.Run("Filter", func(t *testing.T) {
		t.Run("ScoreTweet", TestScoreTweet)
		t.Run("FilterTweetsAll", TestFilterTweetsAllMode)
		t.Run("FilterTweetsTop", TestFilterTweetsTopMode)
		t.Run("FilterTweetsScore", TestFilterTweetsScoreMode)
	})

	t.Run("Models", func(t *testing.T) {
		t.Run("TweetURL", TestTweetURL)
		t.Run("UserProfileURL", TestUserProfileURL)
		t.Run("TweetFromAPI", TestTweetFromAPIResult)
		t.Run("UserFromAPI", TestUserFromAPIResult)
	})

	t.Run("Display", func(t *testing.T) {
		t.Run("FormatConfig", TestFormatConfig)
		t.Run("FormatUserList", TestFormatUserList)
		t.Run("FormatUserListEmpty", TestFormatUserListEmpty)
		t.Run("FormatTweet", TestFormatTweet)
		t.Run("FormatTweetList", TestFormatTweetList)
		t.Run("PrintSuccess", TestPrintSuccess)
		t.Run("PrintError", TestPrintError)
		t.Run("PrintWarning", TestPrintWarning)
		t.Run("FormatUser", TestFormatUser)
		t.Run("RelativeTime", TestRelativeTime)
		t.Run("FormatNumber", TestFormatNumber)
	})

	t.Run("CLI", func(t *testing.T) {
		t.Run("ExitCodes", TestExitCodes)
		t.Run("Constants", TestDefaultConstants)
		t.Run("GraphQLEndpoints", TestGraphQLEndpointsCount)
		t.Run("TLSFingerprinting", TestTLSFingerprinting)
		t.Run("ConfigCommands", TestConfigCommands)
	})

	t.Run("MCP", func(t *testing.T) {
		t.Run("ToolSchema", TestMCPToolSchema)
		t.Run("ToolNames", TestMCPToolNames)
	})
}
