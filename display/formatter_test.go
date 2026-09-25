package display

import (
	"strings"
	"testing"

	"github.com/benoitpetit/xsh/models"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestFormatThreadMutesTreeSeparators(t *testing.T) {
	previousProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(previousProfile)

	tweets := []*models.Tweet{
		{ID: "root", Text: "Root", AuthorHandle: "author"},
		{ID: "reply-1", Text: "Reply one", AuthorHandle: "author", ReplyToID: "root"},
		{ID: "reply-2", Text: "Reply two", AuthorHandle: "author", ReplyToID: "root"},
	}
	output := FormatThread(tweets, "root")

	if !strings.Contains(output, StyleMuted.Render("├──")) {
		t.Fatalf("thread output does not mute branch separators: %q", output)
	}
	if !strings.Contains(output, StyleMuted.Render("   │  ")) {
		t.Fatalf("thread output does not mute vertical separators: %q", output)
	}
}
