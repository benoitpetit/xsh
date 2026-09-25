// Package display provides rich terminal formatting for xsh.
package display

import (
	"github.com/charmbracelet/lipgloss"
)

// ─── Palette ─────────────────────────────────────────────────────────

var (
	// Neutral accents keep the interface readable without competing with content.
	ColorPrimary = lipgloss.Color("#C8D0D8")
	ColorSuccess = lipgloss.Color("#A7C3B1")
	ColorError   = lipgloss.Color("#C99EA3")
	ColorWarning = lipgloss.Color("#C8B98F")
	ColorInfo    = lipgloss.Color("#A9B7C5")
	ColorMuted   = lipgloss.Color("#7F8994")
	ColorText    = lipgloss.Color("#E7EBEF")
	ColorPanelBg = lipgloss.Color("#15191D")
)

// Legacy aliases for compatibility inside display package
var (
	colorBlue   = ColorPrimary
	colorGreen  = ColorSuccess
	colorRed    = ColorError
	colorYellow = ColorWarning
	colorGray   = ColorMuted
	colorWhite  = ColorText
	colorCyan   = ColorInfo
)

// ─── Base Styles ─────────────────────────────────────────────────────

var (
	// StyleBold is bold white text
	StyleBold = lipgloss.NewStyle().Bold(true).Foreground(ColorText)

	// StyleMuted is gray text
	StyleMuted = lipgloss.NewStyle().Foreground(ColorMuted)

	// StylePrimary is bold primary-colored text
	StylePrimary = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)

	// StyleSuccess is bold green text with a success icon
	StyleSuccess = lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess)

	// StyleError is bold red text with an error icon
	StyleError = lipgloss.NewStyle().Bold(true).Foreground(ColorError)

	// StyleWarning is bold yellow text with a warning icon
	StyleWarning = lipgloss.NewStyle().Bold(true).Foreground(ColorWarning)

	// StyleInfo is cyan text
	StyleInfo = lipgloss.NewStyle().Foreground(ColorInfo)

	// StyleCode is text styled for inline code
	StyleCode = lipgloss.NewStyle().Foreground(ColorInfo)

	// StylePanel creates a subtle bordered panel
	StylePanel = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(ColorMuted).
			Padding(0, 1)
)

// ─── Icons ───────────────────────────────────────────────────────────

const (
	IconSuccess = "✓"
	IconError   = "✗"
	IconWarning = "⚠"
	IconInfo    = "ℹ"
	IconBullet  = "•"
	IconArrow   = "→"
	IconStar    = "★"
	IconDash    = "—"
)
