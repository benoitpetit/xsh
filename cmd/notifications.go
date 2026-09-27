// Package cmd provides notification commands for xsh.
package cmd

import (
	"fmt"
	"strconv"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/spf13/cobra"
)

// notificationsCmd fetches the notification timeline
var notificationsCmd = &cobra.Command{
	Use:   "notifications",
	Short: "View your notifications",
	Long: `View your Twitter/X notification timeline.

Examples:
  xsh notifications              # View recent notifications
  xsh notifications -n 50        # View last 50 notifications`,
	RunE: func(cmd *cobra.Command, args []string) error {
		count, _ := strconv.Atoi(cmd.Flag("count").Value.String())

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			abortCommand(core.ExitAuthError)
		}
		defer client.Close()

		runWithWatch(func() error {
			response, err := core.GetNotifications(client, count, "")
			if err != nil {
				return fmt.Errorf("failed to fetch notifications: %w", err)
			}

			output(response.Notifications, func() {
				fmt.Println(display.FormatNotifications(response.Notifications))
			})
			return nil
		})

		return nil
	},
}

func init() {
	rootCmd.AddCommand(notificationsCmd)
	notificationsCmd.Flags().IntP("count", "n", 20, "Number of notifications to fetch")
}
