// Package cmd provides Twitter Spaces commands for xsh.
package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/spf13/cobra"
)

// spaceCmd represents the space parent command
var spaceCmd = &cobra.Command{
	Use:   "space",
	Short: "Twitter Spaces operations - view, search",
	Long: `View and search Twitter/X Spaces.

Examples:
  xsh space view 1AbCdEfGhIjKl      # View Space details
  xsh space search "AI"              # Search for Spaces`,
}

// spaceViewCmd views a Space's details
var spaceViewCmd = &cobra.Command{
	Use:   "view [space-id]",
	Short: "View a Space's details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		space, err := core.GetSpace(client, args[0])
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to fetch Space: %v", err)))
			os.Exit(core.ExitError)
		}

		output(space, func() {
			fmt.Println(display.FormatSpace(space))
		})
	},
}

// spaceSearchCmd searches for Spaces
var spaceSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for Twitter Spaces",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		count, _ := strconv.Atoi(cmd.Flag("count").Value.String())

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		response, err := core.SearchSpaces(client, args[0], count)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to search Spaces: %v", err)))
			os.Exit(core.ExitError)
		}

		output(response.Spaces, func() {
			fmt.Println(display.FormatSpaces(response.Spaces))
		})
	},
}

func init() {
	rootCmd.AddCommand(spaceCmd)
	spaceCmd.AddCommand(spaceViewCmd)
	spaceCmd.AddCommand(spaceSearchCmd)

	spaceSearchCmd.Flags().IntP("count", "n", 20, "Number of results")
}
