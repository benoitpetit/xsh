package frame

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

func TestCardsUseNeutralAccentPalette(t *testing.T) {
	if string(ColorBlue) == "#1DA1F2" || string(ColorCyan) == "#00BCD4" {
		t.Fatal("card accents still use vivid legacy colors")
	}

	previousProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(previousProfile)

	card := CardBlue("content")
	if strings.Contains(card, "╭") || !strings.Contains(card, "┌") {
		t.Fatalf("CardBlue() should use a minimal normal border: %q", card)
	}
	if strings.Contains(card, "38;2;29;161;242") {
		t.Fatalf("CardBlue() still renders the vivid legacy accent: %q", card)
	}
}

func TestCardKeepsStyledContentAligned(t *testing.T) {
	previousProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(previousProfile)

	content := lipgloss.NewStyle().Foreground(ColorCyan).Render("@devbyben") +
		" " + lipgloss.NewStyle().Bold(true).Render(strings.Repeat("Architecte logiciel ", 5))
	card := Card(content, ColorCyan)
	lines := strings.Split(card, "\n")
	wantWidth := ansi.StringWidth(lines[0])

	for index, line := range lines {
		if got := ansi.StringWidth(line); got != wantWidth {
			t.Fatalf("card line %d width = %d, want %d: %q", index, got, wantWidth, line)
		}
	}
}
