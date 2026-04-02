// Package display provides rich terminal formatting for xsh.
package display

import (
	"github.com/charmbracelet/lipgloss"
)

// ─── Palette ─────────────────────────────────────────────────────────

var (
	// Primary accent — Twitter blue
	ColorPrimary = lipgloss.Color("#1DA1F2")
	// Success green
	ColorSuccess = lipgloss.Color("#00BA7C")
	// Error red
	ColorError = lipgloss.Color("#F4212E")
	// Warning yellow
	ColorWarning = lipgloss.Color("#FFAD1F")
	// Info cyan
	ColorInfo = lipgloss.Color("#00BCD4")
	// Muted gray
	ColorMuted = lipgloss.Color("#8899A6")
	// White text
	ColorText = lipgloss.Color("#FFFFFF")
	// Dark background for panels
	ColorPanelBg = lipgloss.Color("#151515")
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
			BorderStyle(lipgloss.RoundedBorder()).
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
