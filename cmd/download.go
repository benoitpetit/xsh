// Package cmd provides download commands for xsh.
package cmd

import (
	"fmt"
	"os"

	"github.com/benoitpetit/xsh/core"
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
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		files, err := core.DownloadTweetMedia(client, tweetID, outputDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		output(map[string]interface{}{
			"tweet_id": tweetID,
			"files":    files,
			"count":    len(files),
		}, func() {
			if len(files) == 0 {
				fmt.Println("No media found in tweet")
			} else {
				fmt.Printf("✓ Downloaded %d file(s)\n", len(files))
				for _, f := range files {
					fmt.Printf("  %s\n", f)
				}
			}
		})
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().StringP("output-dir", "o", ".", "Output directory")
}
