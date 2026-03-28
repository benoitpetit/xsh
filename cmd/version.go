package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags
var Version = "0.0.3"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of xsh",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("xsh version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
