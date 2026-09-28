package cmd

import (
	"fmt"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/spf13/cobra"
)

var (
	communityExploreCount  int
	communityExploreCursor string
)

var communityExploreCmd = &cobra.Command{
	Use:   "explore",
	Short: "Discover communities",
	Long:  "List communities from X's discovery timeline. Use --cursor from structured output to fetch the next page.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateCommunityExploreCount(communityExploreCount); err != nil {
			return err
		}
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Community discovery requires authentication: %v", err)))
			abortCommand(core.ExitAuthError)
			return nil
		}
		defer client.Close()

		communities, nextCursor, err := core.GetCommunityDiscovery(client, communityExploreCount, communityExploreCursor)
		if err != nil {
			return fmt.Errorf("failed to discover communities: %w", err)
		}
		return outputPage(communities, nextCursor, nextCursor != "", func() {
			if len(communities) == 0 {
				fmt.Fprintln(runtimeOutput(), display.EmptyState("No communities found."))
				return
			}
			for i, community := range communities {
				if i > 0 {
					fmt.Fprintln(runtimeOutput())
				}
				fmt.Fprintln(runtimeOutput(), display.FormatCommunity(community))
			}
		})
	},
}

func validateCommunityExploreCount(count int) error {
	if count <= 0 || count > core.MaxCount {
		return fmt.Errorf("--count must be between 1 and %d", core.MaxCount)
	}
	return nil
}

func init() {
	communityCmd.AddCommand(communityExploreCmd)
	communityExploreCmd.Flags().IntVarP(&communityExploreCount, "count", "n", 20, "Number of communities to fetch (1-100)")
	communityExploreCmd.Flags().StringVar(&communityExploreCursor, "cursor", "", "Pagination cursor from the previous response")
}
