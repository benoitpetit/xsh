package cmd

import "testing"

func TestVersionInfoIncludesBuildMetadata(t *testing.T) {
	oldVersion, oldCommit, oldDate := Version, Commit, BuildDate
	t.Cleanup(func() { Version, Commit, BuildDate = oldVersion, oldCommit, oldDate })
	Version = "1.2.3"
	Commit = "abc1234"
	BuildDate = "2026-09-25"

	if got, want := VersionInfo(), "1.2.3 (commit abc1234, built 2026-09-25)"; got != want {
		t.Fatalf("VersionInfo() = %q, want %q", got, want)
	}
}
