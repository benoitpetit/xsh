// Package display provides standardized error messages for xsh.
package display

import (
	"fmt"
)

// ─── Standardized Error Messages ─────────────────────────────────────

// ErrorAuthFailed returns a standardized authentication error message.
func ErrorAuthFailed() string {
	return Error("Authentication failed. Please run 'xsh auth login' to authenticate.")
}

// ErrorAuthRequired returns a standardized message when auth is required.
func ErrorAuthRequired() string {
	return Error("Authentication required. Run 'xsh auth login' to authenticate.")
}

// ErrorInvalidTweetID returns a standardized invalid tweet ID error.
func ErrorInvalidTweetID(id string) string {
	return Error(fmt.Sprintf("Invalid tweet ID: %s", id))
}

// ErrorTweetNotFound returns a standardized tweet not found error.
func ErrorTweetNotFound(id string) string {
	return Error(fmt.Sprintf("Tweet %s not found", id))
}

// ErrorInvalidHandle returns a standardized invalid user handle error.
func ErrorInvalidHandle(handle string) string {
	return Error(fmt.Sprintf("Invalid user handle: %s", handle))
}

// ErrorUserNotFound returns a standardized user not found error.
func ErrorUserNotFound(handle string) string {
	return Error(fmt.Sprintf("User '%s' not found", handle))
}

// ErrorAPIFailed returns a standardized API failure error.
func ErrorAPIFailed(action string, err error) string {
	return Error(fmt.Sprintf("Failed to %s: %v", action, err))
}

// ErrorRateLimited returns a standardized rate limit error.
func ErrorRateLimited() string {
	return Error("Rate limited by Twitter/X. Please wait a moment and try again.")
}

// ErrorInvalidInput returns a standardized invalid input error.
func ErrorInvalidInput(input string) string {
	return Error(fmt.Sprintf("Invalid input: %s", input))
}

// ErrorEmptyInput returns a standardized empty input error.
func ErrorEmptyInput(field string) string {
	return Error(fmt.Sprintf("%s cannot be empty", field))
}

// ErrorFileNotFound returns a standardized file not found error.
func ErrorFileNotFound(path string) string {
	return Error(fmt.Sprintf("File not found: %s", path))
}

// ErrorPermissionDenied returns a standardized permission denied error.
func ErrorPermissionDenied(path string) string {
	return Error(fmt.Sprintf("Permission denied: %s", path))
}

// ErrorNetwork returns a standardized network error.
func ErrorNetwork(err error) string {
	return Error(fmt.Sprintf("Network error: %v", err))
}

// ErrorTimeout returns a standardized timeout error.
func ErrorTimeout() string {
	return Error("Request timed out. Please try again.")
}

// ErrorCancelled returns a standardized cancelled operation error.
func ErrorCancelled() string {
	return Error("Operation cancelled by user.")
}

// ErrorNotImplemented returns a standardized not implemented error.
func ErrorNotImplemented(feature string) string {
	return Error(fmt.Sprintf("Feature not implemented: %s", feature))
}

// ErrorConfigInvalid returns a standardized invalid configuration error.
func ErrorConfigInvalid(err error) string {
	return Error(fmt.Sprintf("Invalid configuration: %v", err))
}

// ErrorAccountNotFound returns a standardized account not found error.
func ErrorAccountNotFound(account string) string {
	return Error(fmt.Sprintf("Account '%s' not found", account))
}

// ─── Standardized Success Messages ───────────────────────────────────

// SuccessPosted returns a standardized post success message.
func SuccessPosted(item string) string {
	return Success(fmt.Sprintf("%s posted successfully", item))
}

// SuccessDeleted returns a standardized delete success message.
func SuccessDeleted(item string) string {
	return Success(fmt.Sprintf("%s deleted successfully", item))
}

// SuccessUpdated returns a standardized update success message.
func SuccessUpdated(item string) string {
	return Success(fmt.Sprintf("%s updated successfully", item))
}

// SuccessSaved returns a standardized save success message.
func SuccessSaved(item string) string {
	return Success(fmt.Sprintf("%s saved successfully", item))
}

// SuccessFollowed returns a standardized follow success message.
func SuccessFollowed(handle string) string {
	return Success(fmt.Sprintf("Now following @%s", handle))
}

// SuccessUnfollowed returns a standardized unfollow success message.
func SuccessUnfollowed(handle string) string {
	return Success(fmt.Sprintf("Unfollowed @%s", handle))
}

// SuccessBlocked returns a standardized block success message.
func SuccessBlocked(handle string) string {
	return Success(fmt.Sprintf("Blocked @%s", handle))
}

// SuccessUnblocked returns a standardized unblock success message.
func SuccessUnblocked(handle string) string {
	return Success(fmt.Sprintf("Unblocked @%s", handle))
}

// SuccessMuted returns a standardized mute success message.
func SuccessMuted(handle string) string {
	return Success(fmt.Sprintf("Muted @%s", handle))
}

// SuccessUnmuted returns a standardized unmute success message.
func SuccessUnmuted(handle string) string {
	return Success(fmt.Sprintf("Unmuted @%s", handle))
}

// SuccessLiked returns a standardized like success message.
func SuccessLiked(id string) string {
	return Success(fmt.Sprintf("Liked tweet %s", id))
}

// SuccessUnliked returns a standardized unlike success message.
func SuccessUnliked(id string) string {
	return Success(fmt.Sprintf("Unliked tweet %s", id))
}

// SuccessRetweeted returns a standardized retweet success message.
func SuccessRetweeted(id string) string {
	return Success(fmt.Sprintf("Retweeted %s", id))
}

// SuccessUnretweeted returns a standardized unretweet success message.
func SuccessUnretweeted(id string) string {
	return Success(fmt.Sprintf("Unretweeted %s", id))
}

// SuccessBookmarked returns a standardized bookmark success message.
func SuccessBookmarked(id string) string {
	return Success(fmt.Sprintf("Bookmarked tweet %s", id))
}

// SuccessUnbookmarked returns a standardized unbookmark success message.
func SuccessUnbookmarked(id string) string {
	return Success(fmt.Sprintf("Unbookmarked tweet %s", id))
}

// SuccessSwitched returns a standardized account switch success message.
func SuccessSwitched(account string) string {
	return Success(fmt.Sprintf("Switched to account '%s'", account))
}

// SuccessLoggedOut returns a standardized logout success message.
func SuccessLoggedOut(account string) string {
	return Success(fmt.Sprintf("Logged out from account '%s'", account))
}
