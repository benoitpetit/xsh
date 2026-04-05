// Package cmd provides DM commands for xsh.
package cmd

import (
	"fmt"
	"os"
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
	Run: func(cmd *cobra.Command, args []string) {
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		conversations, err := core.GetDMInbox(client)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitError)
		}

		output(conversations, func() {
			fmt.Println(display.FormatDMInbox(conversations))
		})
	},
}

// dmSendCmd sends a DM
var dmSendCmd = &cobra.Command{
	Use:   "send <handle> <message>",
	Short: "Send a direct message",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		handle := strings.TrimPrefix(args[0], "@")
		message := strings.Join(args[1:], " ")

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		// Get user by handle
		user, err := core.GetUserByHandle(client, handle)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error fetching user: %v", err)))
			os.Exit(core.ExitError)
		}
		if user == nil {
			fmt.Println(display.Error(fmt.Sprintf("User @%s not found", handle)))
			os.Exit(core.ExitError)
		}

		result, err := core.SendDM(client, user.ID, message)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitError)
		}

		output(result, func() {
			fmt.Println(display.Success(fmt.Sprintf("DM sent to @%s", handle)))
		})
	},
}

// dmDeleteCmd deletes a DM
var dmDeleteCmd = &cobra.Command{
	Use:   "delete <message-id>",
	Short: "Delete a DM message",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		messageID := args[0]
		force, _ := cmd.Flags().GetBool("force")

		if !force {
			fmt.Printf("Delete message %s? [y/N] ", messageID)
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println(display.Warning("Cancelled"))
				return
			}
		}

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		_, err = core.DeleteDM(client, messageID)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitError)
		}

		output(map[string]string{
			"action":     "delete_dm",
			"message_id": messageID,
			"status":     "success",
		}, func() {
			fmt.Println(display.Success("Message deleted"))
		})
	},
}

func init() {
	rootCmd.AddCommand(dmCmd)
	dmCmd.AddCommand(dmInboxCmd)
	dmCmd.AddCommand(dmSendCmd)
	dmCmd.AddCommand(dmDeleteCmd)

	dmDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation")
}
