package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/benoitpetit/xsh/display"
)

var (
	countFile    string
	countPreview bool
	countWidth   int
)

// countCmd represents the count command
var countCmd = &cobra.Command{
	Use:   "count [text]",
	Short: "Count characters in a tweet",
	Long: `Count characters in a tweet and check if it fits Twitter's limit.

Supports both direct text input and file input.`,
	Example: `  xsh count "Hello world!"
  xsh count --file draft.txt
  echo "My tweet" | xsh count`,
	Run: func(cmd *cobra.Command, args []string) {
		var text string

		// Get text from various sources
		if countFile != "" {
			// From file
			content, err := os.ReadFile(countFile)
			if err != nil {
				fmt.Println(display.Error(fmt.Sprintf("Error reading file: %v", err)))
				os.Exit(1)
				return
			}
			text = string(content)
		} else if len(args) > 0 {
			// From arguments
			text = strings.Join(args, " ")
		} else {
			// From stdin
			scanner := bufio.NewScanner(os.Stdin)
			var lines []string
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
			text = strings.Join(lines, "\n")
		}

		if text == "" {
			fmt.Println(display.Error("No text provided. Use arguments, --file, or pipe text to stdin."))
			os.Exit(1)
			return
		}

		// Calculate metrics
		charCount := utf8.RuneCountInString(text)
		byteCount := len(text)
		wordCount := len(strings.Fields(text))
		lineCount := strings.Count(text, "\n") + 1

		// Twitter's limit
		limit := 280
		remaining := limit - charCount

		// Display stats
		fmt.Println(display.Title("📝 Tweet Character Count"))
		fmt.Println()

		// Main counter
		var status string
		if remaining < 0 {
			status = display.Error(fmt.Sprintf("%d / %d (%d over limit)", charCount, limit, -remaining))
		} else if remaining < 20 {
			status = display.Warning(fmt.Sprintf("%d / %d (%d remaining)", charCount, limit, remaining))
		} else {
			status = display.Success(fmt.Sprintf("%d / %d (%d remaining)", charCount, limit, remaining))
		}
		fmt.Println(display.KeyValue("Characters:", status))

		// Additional stats
		fmt.Println(display.KeyValue("Words:", fmt.Sprintf("%d", wordCount)))
		fmt.Println(display.KeyValue("Lines:", fmt.Sprintf("%d", lineCount)))
		fmt.Println(display.KeyValue("Bytes:", fmt.Sprintf("%d", byteCount)))

		// Preview mode
		if countPreview {
			fmt.Println()
			fmt.Println(display.Subtitle("Preview:"))
			fmt.Println()

			preview := formatPreview(text, countWidth)
			fmt.Println(preview)

			// Show URL preview if contains URLs
			if containsURL(text) {
				fmt.Println()
				fmt.Println(display.Info("Note: URLs will be shortened to 23 characters by Twitter"))
			}
		}

		// Exit with error if over limit
		if remaining < 0 {
			os.Exit(1)
		}
	},
}

func formatPreview(text string, width int) string {
	if width <= 0 {
		width = 60
	}

	var lines []string
	words := strings.Fields(text)
	currentLine := ""

	for _, word := range words {
		// Handle newlines in original text
		if strings.Contains(word, "\n") {
			parts := strings.Split(word, "\n")
			for i, part := range parts {
				if i > 0 {
					lines = append(lines, currentLine)
					currentLine = ""
				}
				if len(part) > 0 {
					if len(currentLine)+len(part)+1 <= width {
						if currentLine != "" {
							currentLine += " "
						}
						currentLine += part
					} else {
						if currentLine != "" {
							lines = append(lines, currentLine)
						}
						currentLine = part
					}
				}
			}
			continue
		}

		if len(currentLine)+len(word)+1 <= width {
			if currentLine != "" {
				currentLine += " "
			}
			currentLine += word
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return strings.Join(lines, "\n")
}

func containsURL(text string) bool {
	return strings.Contains(text, "http://") || strings.Contains(text, "https://") || strings.Contains(text, "www.")
}

func init() {
	rootCmd.AddCommand(countCmd)

	countCmd.Flags().StringVarP(&countFile, "file", "f", "", "Read text from file")
	countCmd.Flags().BoolVarP(&countPreview, "preview", "p", false, "Show formatted preview")
	countCmd.Flags().IntVarP(&countWidth, "width", "w", 60, "Preview width for wrapping")
}
