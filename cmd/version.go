package cmd

import (
	"fmt"

	"github.com/benoitpetit/xsh/display"
	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags
var Version = "0.1.1"

// Commit and BuildDate are injected by Makefile/release builds.
var Commit = "unknown"
var BuildDate = "unknown"

// VersionInfo returns a stable, human-readable build identifier.
func VersionInfo() string {
	return fmt.Sprintf("%s (commit %s, built %s)", Version, Commit, BuildDate)
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of xsh",
	RunE: func(cmd *cobra.Command, args []string) error {
		output(map[string]string{
			"version":    Version,
			"commit":     Commit,
			"build_date": BuildDate,
		}, func() {
			fmt.Println(display.Title(fmt.Sprintf("xsh version %s", VersionInfo())))
		})

		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
