// Package cmd provides list management commands for xsh.
package cmd

import (
	"fmt"
	"os"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/spf13/cobra"
)

var (
	listDescription string
	listPrivate     bool
)

// listsCmd represents the lists command
var listsCmd = &cobra.Command{
	Use:   "lists",
	Short: "Manage Twitter lists",
	Long:  `View, create, and manage your Twitter lists.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default: list user's lists
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		lists, err := core.GetUserLists(client)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitError)
		}

		output(lists, func() {
			fmt.Println(display.FormatLists(lists))
		})
	},
}

// listViewCmd views tweets from a list
var listViewCmd = &cobra.Command{
	Use:   "view <list-id>",
	Short: "View tweets from a list",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		listID := args[0]
		count, _ := cmd.Flags().GetInt("count")

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		response, err := core.GetListTweets(client, listID, count, "")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitError)
		}

		output(response.Tweets, func() {
			fmt.Println(display.FormatTweets(response.Tweets))
		})
	},
}

// listCreateCmd creates a new list
var listCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new list",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		result, err := core.CreateList(client, name, listDescription, listPrivate)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitError)
		}

		output(result, func() {
			fmt.Println(display.Success(fmt.Sprintf("Created list '%s'", name)))
		})
	},
}

// listDeleteCmd deletes a list
var listDeleteCmd = &cobra.Command{
	Use:   "delete <list-id>",
	Short: "Delete a list",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		listID := args[0]
		force, _ := cmd.Flags().GetBool("force")

		if !force {
			fmt.Printf("Delete list %s? [y/N] ", listID)
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

		_, err = core.DeleteList(client, listID)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitError)
		}

		output(map[string]string{
			"action":  "delete",
			"list_id": listID,
			"status":  "success",
		}, func() {
			fmt.Println(display.Success(fmt.Sprintf("Deleted list %s", listID)))
		})
	},
}

// listMembersCmd views members of a list
var listMembersCmd = &cobra.Command{
	Use:   "members <list-id>",
	Short: "View members of a list",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		listID := args[0]
		count, _ := cmd.Flags().GetInt("count")

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		users, _, err := core.GetListMembers(client, listID, count, "")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitError)
		}

		output(users, func() {
			fmt.Println(display.FormatUsers(users))
		})
	},
}

// listAddMemberCmd adds a member to a list
var listAddMemberCmd = &cobra.Command{
	Use:   "add-member <list-id> <handle>",
	Short: "Add a member to a list",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		listID := args[0]
		handle := args[1]

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

		_, err = core.AddListMember(client, listID, user.ID)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitError)
		}

		output(map[string]string{
			"action":  "add_member",
			"list_id": listID,
			"handle":  handle,
			"status":  "success",
		}, func() {
			fmt.Println(display.Success(fmt.Sprintf("Added @%s to list %s", handle, listID)))
		})
	},
}

// listRemoveMemberCmd removes a member from a list
var listRemoveMemberCmd = &cobra.Command{
	Use:   "remove-member <list-id> <handle>",
	Short: "Remove a member from a list",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		listID := args[0]
		handle := args[1]

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

		_, err = core.RemoveListMember(client, listID, user.ID)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitError)
		}

		output(map[string]string{
			"action":  "remove_member",
			"list_id": listID,
			"handle":  handle,
			"status":  "success",
		}, func() {
			fmt.Println(display.Success(fmt.Sprintf("Removed @%s from list %s", handle, listID)))
		})
	},
}

// listPinCmd pins a list
var listPinCmd = &cobra.Command{
	Use:   "pin <list-id>",
	Short: "Pin a list",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		listID := args[0]

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		_, err = core.PinList(client, listID)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitError)
		}

		output(map[string]string{
			"action":  "pin",
			"list_id": listID,
			"status":  "success",
		}, func() {
			fmt.Println(display.Success(fmt.Sprintf("Pinned list %s", listID)))
		})
	},
}

// listUnpinCmd unpins a list
var listUnpinCmd = &cobra.Command{
	Use:   "unpin <list-id>",
	Short: "Unpin a list",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		listID := args[0]

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		_, err = core.UnpinList(client, listID)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitError)
		}

		output(map[string]string{
			"action":  "unpin",
			"list_id": listID,
			"status":  "success",
		}, func() {
			fmt.Println(display.Success(fmt.Sprintf("Unpinned list %s", listID)))
		})
	},
}

func init() {
	rootCmd.AddCommand(listsCmd)

	// Add subcommands
	listsCmd.AddCommand(listViewCmd)
	listsCmd.AddCommand(listCreateCmd)
	listsCmd.AddCommand(listDeleteCmd)
	listsCmd.AddCommand(listMembersCmd)
	listsCmd.AddCommand(listAddMemberCmd)
	listsCmd.AddCommand(listRemoveMemberCmd)
	listsCmd.AddCommand(listPinCmd)
	listsCmd.AddCommand(listUnpinCmd)

	// Flags
	listViewCmd.Flags().IntP("count", "n", 20, "Number of tweets to fetch")
	listCreateCmd.Flags().StringVarP(&listDescription, "description", "d", "", "List description")
	listCreateCmd.Flags().BoolVar(&listPrivate, "private", false, "Make the list private")
	listDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation")
	listMembersCmd.Flags().IntP("count", "n", 20, "Number of members to fetch")
}




