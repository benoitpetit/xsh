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

func TestTopLevelCommandsUseIntentionalHelpGroups(t *testing.T) {
	want := map[string]string{
		"auth": groupStart, "accounts": groupStart, "switch": groupStart, "import": groupStart,
		"feed": groupExplore, "search": groupExplore, "user": groupExplore, "tweet": groupExplore,
		"compose": groupPublish, "schedule": groupPublish, "scheduled": groupPublish, "unschedule": groupPublish,
		"status": groupSystem, "doctor": groupSystem, "config": groupSystem, "endpoints": groupSystem,
	}

	for name, group := range want {
		command, _, err := rootCmd.Find([]string{name})
		if err != nil {
			t.Fatalf("failed to find %q: %v", name, err)
		}
		if command.GroupID != group {
			t.Fatalf("%s group = %q, want %q", name, command.GroupID, group)
		}
	}
}
