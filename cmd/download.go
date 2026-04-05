// Package cmd provides download commands for xsh.
package cmd

import (
	"fmt"
	"os"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/spf13/cobra"
)

// downloadCmd downloads media from a tweet
var downloadCmd = &cobra.Command{
	Use:   "download <tweet-id>",
	Short: "Download media from a tweet",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tweetID := args[0]
		outputDir, _ := cmd.Flags().GetString("output-dir")

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		files, err := core.DownloadTweetMedia(client, tweetID, outputDir)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitError)
		}

		output(map[string]interface{}{
			"tweet_id": tweetID,
			"files":    files,
			"count":    len(files),
		}, func() {
			if len(files) == 0 {
				fmt.Println(display.Warning("No media found in tweet"))
			} else {
				fmt.Println(display.Success(fmt.Sprintf("Downloaded %d file(s)", len(files))))
				for _, f := range files {
					fmt.Println(display.Info(fmt.Sprintf("  %s", f)))
				}
			}
		})
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().StringP("output-dir", "o", ".", "Output directory")
}
