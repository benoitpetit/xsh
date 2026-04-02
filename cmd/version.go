package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/benoitpetit/xsh/display"
)

// Version is set at build time via ldflags
var Version = "0.0.3"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of xsh",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(display.Title(fmt.Sprintf("xsh version %s", Version)))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
