// Package cmd provides bookmark commands for xsh.
package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/spf13/cobra"
)

// bookmarksFoldersCmd lists bookmark folders
var bookmarksFoldersCmd = &cobra.Command{
	Use:   "bookmarks-folders",
	Short: "List bookmark folders",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		folders, err := core.GetBookmarkFolders(client)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitError)
		}

		output(folders, func() {
			fmt.Println(display.FormatBookmarkFolders(folders))
		})
	},
}

// bookmarksFolderCmd views tweets in a bookmark folder
var bookmarksFolderCmd = &cobra.Command{
	Use:   "bookmarks-folder <folder-id>",
	Short: "View tweets in a bookmark folder",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		folderID := args[0]
		count, _ := strconv.Atoi(cmd.Flag("count").Value.String())

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		response, err := core.GetBookmarkFolderTimeline(client, folderID, count, "")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitError)
		}

		output(response.Tweets, func() {
			fmt.Println(display.FormatTweets(response.Tweets))
		})
	},
}

func init() {
	rootCmd.AddCommand(bookmarksFoldersCmd)
	rootCmd.AddCommand(bookmarksFolderCmd)

	bookmarksFolderCmd.Flags().IntP("count", "n", 20, "Number of tweets to fetch")
}
