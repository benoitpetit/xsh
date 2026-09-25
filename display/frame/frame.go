// Package frame provides consistent styling for terminal display.
// Uses background colors instead of borders to avoid emoji width issues.
package frame

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// Colors
var (
	ColorBlue   = lipgloss.Color("#C8D0D8")
	ColorGray   = lipgloss.Color("#7F8994")
	ColorCyan   = lipgloss.Color("#A9B7C5")
	ColorGreen  = lipgloss.Color("#A7C3B1")
	ColorRed    = lipgloss.Color("#C99EA3")
	ColorYellow = lipgloss.Color("#C8B98F")
	ColorWhite  = lipgloss.Color("#E7EBEF")

	// Background colors for cards
	BgDark   = lipgloss.Color("#171B20")
	BgDarker = lipgloss.Color("#101317")
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

	// Minimal top border: ┌──────┐
	topBorder := accentStyle.Render("┌" + strings.Repeat("─", maxWidth+2) + "┐")
	b.WriteString(topBorder)
	b.WriteString("\n")

	// Content without background - wrap long lines instead of truncating
	for _, line := range lines {
		lineWidth := ansi.StringWidth(line)

		if lineWidth > maxWidth {
			// Word wrap: split long line into multiple lines
			wrappedLines := wrapLine(line, maxWidth)
			for _, wrappedLine := range wrappedLines {
				w := ansi.StringWidth(wrappedLine)
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

	// Minimal bottom border: └──────┘
	bottomBorder := accentStyle.Render("└" + strings.Repeat("─", maxWidth+2) + "┘")
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
	return ansi.StringWidth(s)
}

// Truncate truncates a string to fit within maxWidth
func Truncate(s string, maxWidth int) string {
	if StringWidth(s) <= maxWidth {
		return s
	}
	return ansi.Truncate(s, maxWidth-3, "...")
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
				word = ansi.Truncate(word, maxWidth-3, "...")
			}
			currentLine = word
		}
	}

	if currentLine != "" {
		result = append(result, currentLine)
	}

	return result
}
