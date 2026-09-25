package display

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestThemeUsesNeutralAccentPalette(t *testing.T) {
	colors := map[string]lipgloss.Color{
		"primary": ColorPrimary,
		"info":    ColorInfo,
		"muted":   ColorMuted,
	}

	for name, color := range colors {
		if string(color) == "#1DA1F2" || string(color) == "#00BCD4" {
			t.Fatalf("%s color still uses the vivid legacy accent %q", name, color)
		}
	}
}

func TestPanelUsesMinimalBorder(t *testing.T) {
	panel := Panel("content")
	if strings.Contains(panel, "╭") || !strings.Contains(panel, "┌") {
		t.Fatalf("Panel() should use a minimal normal border: %q", panel)
	}
}

func TestSimpleTableDoesNotUseVividHeaderBackground(t *testing.T) {
	previousProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(previousProfile)

	table := SimpleTable([]string{"Name"}, []TableRow{{"xsh"}})
	if strings.Contains(table, "48;2;29;161;242") {
		t.Fatalf("SimpleTable() still uses a vivid primary header background: %q", table)
	}
}

func TestEmptyMutedTextDoesNotEmitFormattingCodes(t *testing.T) {
	previousProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(previousProfile)

	if got := Muted(""); got != "" {
		t.Fatalf("Muted(\"\") = %q, want an empty string", got)
	}
}
