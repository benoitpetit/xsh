// Package cmd provides DM commands for xsh.
package cmd

import (
	"fmt"
	"strings"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/spf13/cobra"
)

// dmCmd represents the dm command group
var dmCmd = &cobra.Command{
	Use:   "dm",
	Short: "Direct message commands",
	Long:  `Send and manage direct messages.`,
}

// dmInboxCmd views DM inbox
var dmInboxCmd = &cobra.Command{
	Use:   "inbox",
	Short: "View DM inbox",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			abortCommand(core.ExitAuthError)
		}
		defer client.Close()

		conversations, err := core.GetDMInbox(client)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			abortCommand(core.ExitError)
		}

		output(conversations, func() {
			fmt.Println(display.FormatDMInbox(conversations))
		})

		return nil
	},
}

// dmSendCmd sends a DM
var dmSendCmd = &cobra.Command{
	Use:   "send <handle> <message>",
	Short: "Send a direct message",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		handle := strings.TrimPrefix(args[0], "@")
		message := strings.Join(args[1:], " ")

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			abortCommand(core.ExitAuthError)
		}
		defer client.Close()

		// Get user by handle
		user, err := core.GetUserByHandle(client, handle)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error fetching user: %v", err)))
			abortCommand(core.ExitError)
		}
		if user == nil {
			fmt.Println(display.Error(fmt.Sprintf("User @%s not found", handle)))
			abortCommand(core.ExitError)
		}

		result, err := core.SendDM(client, user.ID, message)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			abortCommand(core.ExitError)
		}

		output(result, func() {
			fmt.Println(display.Success(fmt.Sprintf("DM sent to @%s", handle)))
		})

		return nil
	},
}

// dmDeleteCmd deletes a DM
var dmDeleteCmd = &cobra.Command{
	Use:   "delete <message-id>",
	Short: "Delete a DM message",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		messageID := args[0]
		force, _ := cmd.Flags().GetBool("force")

		if !force {
			fmt.Printf("Delete message %s? [y/N] ", messageID)
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println(display.Warning("Cancelled"))
				return nil
			}
		}

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			abortCommand(core.ExitAuthError)
		}
		defer client.Close()

		_, err = core.DeleteDM(client, messageID)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			abortCommand(core.ExitError)
		}

		output(map[string]string{
			"action":     "delete_dm",
			"message_id": messageID,
			"status":     "success",
		}, func() {
			fmt.Println(display.Success("Message deleted"))
		})

		return nil
	},
}

func init() {
	rootCmd.AddCommand(dmCmd)
	dmCmd.AddCommand(dmInboxCmd)
	dmCmd.AddCommand(dmSendCmd)
	dmCmd.AddCommand(dmDeleteCmd)

	dmDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation")
}
