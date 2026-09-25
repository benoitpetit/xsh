package cmd

import "github.com/spf13/cobra"

const (
	groupStart      = "start"
	groupExplore    = "explore"
	groupPublish    = "publish"
	groupEngage     = "engage"
	groupOrganize   = "organize"
	groupMessages   = "messages"
	groupAutomation = "automation"
	groupSystem     = "system"
)

// Group top-level commands by user intent without changing their command
// paths. Existing scripts keep working while `xsh --help` becomes scannable.
func init() {
	ensureCommandGroups()
}

func ensureCommandGroups() {
	for _, group := range []*cobra.Group{
		{ID: groupStart, Title: "Getting Started:"},
		{ID: groupExplore, Title: "Discover & Read:"},
		{ID: groupPublish, Title: "Create & Schedule:"},
		{ID: groupEngage, Title: "Engage:"},
		{ID: groupOrganize, Title: "Organize & Export:"},
		{ID: groupMessages, Title: "Messages:"},
		{ID: groupAutomation, Title: "Automation & Data:"},
		{ID: groupSystem, Title: "System & Maintenance:"},
	} {
		if !rootCmd.ContainsGroup(group.ID) {
			rootCmd.AddGroup(group)
		}
	}
	// These commands are intentionally exposed both as top-level shortcuts and
	// under `social`, so the parent command needs the same group definition.
	if !socialCmd.ContainsGroup(groupEngage) {
		socialCmd.AddGroup(&cobra.Group{ID: groupEngage, Title: "Engage:"})
	}
	if !socialCmd.ContainsGroup(groupExplore) {
		socialCmd.AddGroup(&cobra.Group{ID: groupExplore, Title: "Discover & Read:"})
	}

	assignCommandGroups(map[string]string{
		"auth": groupStart, "accounts": groupStart, "switch": groupStart,
		"import": groupStart,
		"feed":   groupExplore, "search": groupExplore, "user": groupExplore,
		"tweet": groupExplore, "thread": groupExplore, "unroll": groupExplore,
		"quotes": groupExplore, "pinned": groupExplore, "trends": groupExplore,
		"space": groupExplore, "community": groupExplore, "jobs": groupExplore,
		"notifications": groupExplore,
		"compose":       groupPublish, "schedule": groupPublish,
		"scheduled": groupPublish, "unschedule": groupPublish,
		"follow": groupEngage, "unfollow": groupEngage, "social": groupEngage,
		"block": groupEngage, "unblock": groupEngage, "mute": groupEngage,
		"unmute":    groupEngage,
		"bookmarks": groupOrganize, "bookmarks-folder": groupOrganize,
		"bookmarks-folders": groupOrganize, "lists": groupOrganize,
		"download": groupOrganize, "gallery": groupOrganize, "export": groupOrganize,
		"dm":        groupMessages,
		"analytics": groupAutomation, "count": groupAutomation,
		"multi": groupAutomation, "stream": groupAutomation,
		"tweets": groupAutomation, "users": groupAutomation,
		"config": groupSystem, "endpoints": groupSystem,
		"status": groupSystem, "doctor": groupSystem,
		"auto-update": groupSystem, "ratelimit": groupSystem,
		"mcp": groupSystem, "version": groupSystem, "completion": groupSystem,
	})
	socialBlockedCmd.GroupID = groupExplore
	socialMutedCmd.GroupID = groupExplore
	rootCmd.SetHelpCommandGroupID(groupSystem)
	rootCmd.SetCompletionCommandGroupID(groupSystem)
}

func assignCommandGroups(groups map[string]string) {
	for _, command := range rootCmd.Commands() {
		if groupID, ok := groups[command.Name()]; ok {
			command.GroupID = groupID
		}
	}
}
