package cmd

import (
	"strings"
	"testing"
)

func TestValidatePaginationArgsRequiresPositiveValues(t *testing.T) {
	for _, tc := range []struct {
		name       string
		count      int
		pages      int
		wantSubstr string
	}{
		{name: "zero count", count: 0, pages: 1, wantSubstr: "--count"},
		{name: "negative count", count: -1, pages: 1, wantSubstr: "--count"},
		{name: "zero pages", count: 1, pages: 0, wantSubstr: "--pages"},
		{name: "negative pages", count: 1, pages: -1, wantSubstr: "--pages"},
		{name: "product overflow", count: int(^uint(0) >> 1), pages: 2, wantSubstr: "too large"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validatePaginationArgs(tc.count, tc.pages)
			if err == nil {
				t.Fatal("validatePaginationArgs() returned nil")
			}
			if !strings.Contains(err.Error(), tc.wantSubstr) {
				t.Fatalf("error = %q, want it to contain %q", err, tc.wantSubstr)
			}
		})
	}
	if err := validatePaginationArgs(20, 3); err != nil {
		t.Fatalf("valid pagination arguments rejected: %v", err)
	}
}

func TestFeedAndSearchRejectInvalidPaginationBeforeClientCreation(t *testing.T) {
	oldFeedCount, oldFeedPages := feedCount, feedPages
	oldSearchCount, oldSearchPages := searchCount, searchPages
	t.Cleanup(func() {
		feedCount, feedPages = oldFeedCount, oldFeedPages
		searchCount, searchPages = oldSearchCount, oldSearchPages
	})

	feedCount, feedPages = 0, 1
	if err := feedCmd.RunE(feedCmd, nil); err == nil || !strings.Contains(err.Error(), "--count") {
		t.Fatalf("feed RunE error = %v, want a --count validation error", err)
	}

	searchCount, searchPages = 20, 0
	if err := searchCmd.RunE(searchCmd, []string{"golang"}); err == nil || !strings.Contains(err.Error(), "--pages") {
		t.Fatalf("search RunE error = %v, want a --pages validation error", err)
	}
}
