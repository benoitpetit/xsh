package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
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
				fmt.Printf("Error reading file: %v\n", err)
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
			fmt.Println("No text provided. Use arguments, --file, or pipe text to stdin.")
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

		// Styles
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#1DA1F2"))
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#8899A6"))
		valueStyle := lipgloss.NewStyle().Bold(true)
		warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAD1F"))
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F4212E"))
		successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00BA7C"))

		// Display stats
		fmt.Println(titleStyle.Render("📝 Tweet Character Count"))
		fmt.Println()

		// Main counter
		var status string
		if remaining < 0 {
			status = errorStyle.Render(fmt.Sprintf("%d / %d (%d over limit)", charCount, limit, -remaining))
		} else if remaining < 20 {
			status = warningStyle.Render(fmt.Sprintf("%d / %d (%d remaining)", charCount, limit, remaining))
		} else {
			status = successStyle.Render(fmt.Sprintf("%d / %d (%d remaining)", charCount, limit, remaining))
		}
		fmt.Printf("%s %s\n", labelStyle.Render("Characters:"), status)

		// Additional stats
		fmt.Printf("%s %s\n", labelStyle.Render("Words:      "), valueStyle.Render(fmt.Sprintf("%d", wordCount)))
		fmt.Printf("%s %s\n", labelStyle.Render("Lines:      "), valueStyle.Render(fmt.Sprintf("%d", lineCount)))
		fmt.Printf("%s %s\n", labelStyle.Render("Bytes:      "), valueStyle.Render(fmt.Sprintf("%d", byteCount)))

		// Preview mode
		if countPreview {
			fmt.Println()
			fmt.Println(titleStyle.Render("Preview:"))
			fmt.Println()

			preview := formatPreview(text, countWidth)
			fmt.Println(preview)

			// Show URL preview if contains URLs
			if containsURL(text) {
				fmt.Println()
				fmt.Println(labelStyle.Render("Note: URLs will be shortened to 23 characters by Twitter"))
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
