// Package frame provides consistent styling for terminal display.
// Uses background colors instead of borders to avoid emoji width issues.
package frame

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// Colors
var (
	ColorBlue   = lipgloss.Color("#1DA1F2")
	ColorGray   = lipgloss.Color("#8899A6")
	ColorCyan   = lipgloss.Color("#00BCD4")
	ColorGreen  = lipgloss.Color("#00BA7C")
	ColorRed    = lipgloss.Color("#F4212E")
	ColorYellow = lipgloss.Color("#FFAD1F")
	ColorWhite  = lipgloss.Color("#FFFFFF")
	
	// Background colors for cards
	BgDark    = lipgloss.Color("#1a1a1a")
	BgDarker  = lipgloss.Color("#0d0d0d")
)

// FixedWidth is the target width for cards
const FixedWidth = 80

// Card creates a card with top/bottom borders only (no side borders to avoid emoji issues)
// Uses a fixed width for consistency across all cards
func Card(content string, accentColor lipgloss.Color) string {
	lines := strings.Split(content, "\n")
	
	// Use fixed width for all cards
	maxWidth := FixedWidth
	
	accentStyle := lipgloss.NewStyle().Foreground(accentColor)
	
	var b strings.Builder
	
	// Top border with rounded corners: ╭──────╮
	topBorder := accentStyle.Render("╭" + strings.Repeat("─", maxWidth+2) + "╮")
	b.WriteString(topBorder)
	b.WriteString("\n")
	
	// Content without background - wrap long lines instead of truncating
	for _, line := range lines {
		lineWidth := runewidth.StringWidth(line)
		
		if lineWidth > maxWidth {
			// Word wrap: split long line into multiple lines
			wrappedLines := wrapLine(line, maxWidth)
			for _, wrappedLine := range wrappedLines {
				w := runewidth.StringWidth(wrappedLine)
				padding := maxWidth - w
				if padding < 0 {
					padding = 0
				}
				b.WriteString(accentStyle.Render("│ "))
				b.WriteString(wrappedLine + strings.Repeat(" ", padding))
				b.WriteString(accentStyle.Render(" │"))
				b.WriteString("\n")
			}
		} else {
			padding := maxWidth - lineWidth
			b.WriteString(accentStyle.Render("│ "))
			b.WriteString(line + strings.Repeat(" ", padding))
			b.WriteString(accentStyle.Render(" │"))
			b.WriteString("\n")
		}
	}
	
	// Bottom border with rounded corners: ╰──────╯
	bottomBorder := accentStyle.Render("╰" + strings.Repeat("─", maxWidth+2) + "╯")
	b.WriteString(bottomBorder)
	
	return b.String()
}

// CardBlue creates a blue-accented card
func CardBlue(content string) string {
	return Card(content, ColorBlue)
}

// CardCyan creates a cyan-accented card
func CardCyan(content string) string {
	return Card(content, ColorCyan)
}

// CardGray creates a gray-accented card
func CardGray(content string) string {
	return Card(content, ColorGray)
}

// SimpleSeparator creates a simple horizontal line
func SimpleSeparator() string {
	return lipgloss.NewStyle().
		Foreground(ColorGray).
		Render(strings.Repeat("─", FixedWidth+4))
}

// StringWidth returns the display width of a string
func StringWidth(s string) int {
	return runewidth.StringWidth(s)
}

// Truncate truncates a string to fit within maxWidth
func Truncate(s string, maxWidth int) string {
	if StringWidth(s) <= maxWidth {
		return s
	}
	return runewidth.Truncate(s, maxWidth-3, "...")
}

// wrapLine wraps a long line into multiple lines at word boundaries
func wrapLine(line string, maxWidth int) []string {
	var result []string
	words := strings.Fields(line)
	if len(words) == 0 {
		return []string{line}
	}
	
	currentLine := ""
	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word
		
		if StringWidth(testLine) <= maxWidth {
			currentLine = testLine
		} else {
			if currentLine != "" {
				result = append(result, currentLine)
			}
			// If single word is too long, force break it
			if StringWidth(word) > maxWidth {
				word = runewidth.Truncate(word, maxWidth-3, "...")
			}
			currentLine = word
		}
	}
	
	if currentLine != "" {
		result = append(result, currentLine)
	}
	
	return result
}
