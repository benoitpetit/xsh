package cmd

import "testing"

func TestTopLevelCommandsAreGroupedForDiscoverability(t *testing.T) {
	if !rootCmd.ContainsGroup(groupEngage) {
		t.Fatal("engage help group was not registered")
	}
	for _, name := range []string{"auth", "feed", "tweet", "dm", "config", "endpoints", "mcp"} {
		command, _, err := rootCmd.Find([]string{name})
		if err != nil {
			t.Fatalf("failed to find %q: %v", name, err)
		}
		if command.GroupID == "" {
			t.Fatalf("command %q has no help group", name)
		}
	}
}
