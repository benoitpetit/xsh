// Package cmd provides media gallery commands for xsh.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/benoitpetit/xsh/models"
	"github.com/benoitpetit/xsh/utils"
	"github.com/spf13/cobra"
)

var (
	galleryCount     int
	galleryOutputDir string
	galleryPhotos    bool
	galleryVideos    bool
)

// galleryCmd batch-downloads media from a user's tweets or a search query
var galleryCmd = &cobra.Command{
	Use:   "gallery [handle|query]",
	Short: "Batch download media from a user or search",
	Long: `Download all photos and/or videos from a user's recent tweets
or from a search query into a local directory.

Examples:
  xsh gallery @elonmusk                     # Download media from user
  xsh gallery @nasa -n 50 -o ./nasa-media   # 50 tweets, custom dir
  xsh gallery "cats" --search               # Download from search results
  xsh gallery @user --photos-only           # Photos only
  xsh gallery @user --videos-only           # Videos only`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		isSearch, _ := cmd.Flags().GetBool("search")

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			os.Exit(core.ExitAuthError)
			return
		}
		defer client.Close()

		var mediaFiles []galleryResult
		var label string

		if isSearch {
			label = fmt.Sprintf("search:%s", args[0])
			mediaFiles, err = galleryFromSearch(client, args[0])
		} else {
			handle, valid := utils.ValidateTwitterHandle(args[0])
			if !valid {
				fmt.Println(display.Error(fmt.Sprintf("Invalid handle: %s (use --search for search queries)", args[0])))
				os.Exit(core.ExitError)
				return
			}
			label = "@" + handle
			mediaFiles, err = galleryFromUser(client, handle)
		}

		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to fetch tweets: %v", err)))
			os.Exit(core.ExitError)
			return
		}

		if len(mediaFiles) == 0 {
			fmt.Println(display.Muted("No media found"))
			return
		}

		// Create output directory
		outDir := galleryOutputDir
		if outDir == "" {
			outDir = "xsh-gallery"
		}
		if err := os.MkdirAll(outDir, 0755); err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to create directory: %v", err)))
			os.Exit(core.ExitError)
			return
		}

		// Download media
		var downloaded []string
		var skipped int
		for i, m := range mediaFiles {
			fmt.Printf("\r%s Downloading %d/%d...", display.Muted("..."), i+1, len(mediaFiles))

			ext := getMediaExtension(m.URL, m.MediaType)
			filename := fmt.Sprintf("%s_%d.%s", m.TweetID, m.Index, ext)
			outPath := filepath.Join(outDir, filename)

			// Skip if already exists
			if _, statErr := os.Stat(outPath); statErr == nil {
				skipped++
				downloaded = append(downloaded, outPath)
				continue
			}

			mediaURL := m.URL
			if m.MediaType == "photo" && !containsQuery(mediaURL) {
				mediaURL = mediaURL + "?format=jpg&name=orig"
			}

			if dlErr := core.DownloadMedia(mediaURL, outPath); dlErr != nil {
				continue
			}
			downloaded = append(downloaded, outPath)
		}
		fmt.Println() // clear progress line

		output(map[string]interface{}{
			"source":     label,
			"directory":  outDir,
			"downloaded": len(downloaded),
			"skipped":    skipped,
			"total":      len(mediaFiles),
			"files":      downloaded,
		}, func() {
			fmt.Println(display.Success(fmt.Sprintf("Gallery %s: %d files downloaded to %s", label, len(downloaded), outDir)))
			if skipped > 0 {
				fmt.Println(display.Muted(fmt.Sprintf("  (%d already existed, skipped)", skipped)))
			}

			// Summary by type
			photos, videos := 0, 0
			for _, m := range mediaFiles {
				if m.MediaType == "photo" {
					photos++
				} else {
					videos++
				}
			}
			if photos > 0 {
				fmt.Println(display.Info(fmt.Sprintf("  Photos: %d", photos)))
			}
			if videos > 0 {
				fmt.Println(display.Info(fmt.Sprintf("  Videos: %d", videos)))
			}
		})
	},
}

type galleryResult struct {
	TweetID   string
	URL       string
	MediaType string // "photo", "video", "animated_gif"
	Index     int
}

func galleryFromUser(client *core.XClient, handle string) ([]galleryResult, error) {
	user, err := core.GetUserByHandle(client, handle)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user @%s not found", handle)
	}

	response, err := core.GetUserTweets(client, user.ID, galleryCount, "", false)
	if err != nil {
		return nil, err
	}

	return extractMedia(response.Tweets), nil
}

func galleryFromSearch(client *core.XClient, query string) ([]galleryResult, error) {
	response, err := core.SearchTweets(client, query, "Latest", galleryCount, "")
	if err != nil {
		return nil, err
	}

	return extractMedia(response.Tweets), nil
}

func extractMedia(tweets []*models.Tweet) []galleryResult {
	var results []galleryResult
	for _, t := range tweets {
		for i, m := range t.Media {
			if m.URL == "" {
				continue
			}
			// Filter by type if requested
			if galleryPhotos && m.Type != "photo" {
				continue
			}
			if galleryVideos && m.Type == "photo" {
				continue // videos-only: skip photos
			}
			results = append(results, galleryResult{
				TweetID:   t.ID,
				URL:       m.URL,
				MediaType: m.Type,
				Index:     i,
			})
		}
	}
	return results
}

func getMediaExtension(url, mediaType string) string {
	if mediaType == "video" || mediaType == "animated_gif" {
		return "mp4"
	}
	// Try to extract from URL
	ext := filepath.Ext(url)
	if ext != "" {
		ext = ext[1:] // remove leading dot
		// Clean query params
		for i, c := range ext {
			if c == '?' {
				ext = ext[:i]
				break
			}
		}
		if len(ext) > 0 && len(ext) <= 4 {
			return ext
		}
	}
	return "jpg"
}

func containsQuery(url string) bool {
	for _, c := range url {
		if c == '?' {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(galleryCmd)
	galleryCmd.Flags().IntVarP(&galleryCount, "count", "n", 20, "Number of tweets to scan for media")
	galleryCmd.Flags().StringVarP(&galleryOutputDir, "output-dir", "o", "xsh-gallery", "Output directory")
	galleryCmd.Flags().BoolVar(&galleryPhotos, "photos-only", false, "Download only photos")
	galleryCmd.Flags().BoolVar(&galleryVideos, "videos-only", false, "Download only videos")
	galleryCmd.Flags().Bool("search", false, "Treat argument as search query instead of handle")
}
