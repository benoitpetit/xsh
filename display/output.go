package display

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// ─── Simple Message Helpers ──────────────────────────────────────────

// Success returns a styled success message.
func Success(message string) string {
	return StyleSuccess.Render(IconSuccess + " " + message)
}

// Error returns a styled error message.
func Error(message string) string {
	return StyleError.Render(IconError + " " + message)
}

// Warning returns a styled warning message.
func Warning(message string) string {
	return StyleWarning.Render(IconWarning + " " + message)
}

// Info returns a styled info message.
func Info(message string) string {
	return StyleInfo.Render(IconInfo + " " + message)
}

// Muted returns muted gray text.
func Muted(text string) string {
	return StyleMuted.Render(text)
}

// Primary returns primary-accent bold text.
func Primary(text string) string {
	return StylePrimary.Render(text)
}

// Bold returns bold white text.
func Bold(text string) string {
	return StyleBold.Render(text)
}

// Code returns code-styled text.
func Code(text string) string {
	return StyleCode.Render(text)
}

// ─── Legacy aliases for backwards compatibility ──────────────────────

// PrintSuccess is an alias for Success (returns string).
func PrintSuccess(message string) string { return Success(message) }

// PrintError is an alias for Error (returns string).
func PrintError(message string) string { return Error(message) }

// PrintWarning is an alias for Warning (returns string).
func PrintWarning(message string) string { return Warning(message) }

// ─── Layout Helpers ──────────────────────────────────────────────────

// Title returns a large bold title.
func Title(text string) string {
	s := lipgloss.NewStyle().Bold(true).Foreground(ColorText).MarginBottom(1)
	return s.Render(text)
}

// Subtitle returns a medium bold subtitle with primary color.
func Subtitle(text string) string {
	s := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).MarginTop(1).MarginBottom(1)
	return s.Render(text)
}

// Section returns a section header.
func Section(text string) string {
	s := lipgloss.NewStyle().Bold(true).Foreground(ColorInfo).MarginTop(1).MarginBottom(1)
	return s.Render(text)
}

// Separator returns a horizontal separator line.
func Separator(width int) string {
	if width <= 0 {
		width = 60
	}
	return lipgloss.NewStyle().
		Foreground(ColorMuted).
		Render(strings.Repeat(IconDash, width))
}

// EmptyState returns a muted empty-state message.
func EmptyState(message string) string {
	return StyleMuted.Render(message)
}

// ─── List Helpers ────────────────────────────────────────────────────

// Bullet returns a bulleted list item.
func Bullet(text string) string {
	return fmt.Sprintf("  %s %s", StyleMuted.Render(IconBullet), text)
}

// Numbered returns a numbered list item.
func Numbered(index int, text string) string {
	num := StyleMuted.Render(fmt.Sprintf("%2d.", index))
	return fmt.Sprintf("%s %s", num, text)
}

// ArrowItem returns an arrow-prefixed list item.
func ArrowItem(text string) string {
	return fmt.Sprintf("  %s %s", StylePrimary.Render(IconArrow), text)
}

// ─── Key-Value Helpers ───────────────────────────────────────────────

// KeyValue renders a key/value pair with aligned key width.
func KeyValue(key string, value string) string {
	keyStyled := lipgloss.NewStyle().Foreground(ColorMuted).Width(20).Render(key)
	valueStyled := lipgloss.NewStyle().Foreground(ColorText).Render(value)
	return keyStyled + " " + valueStyled
}

// KeyValueCustom renders a key/value with custom key width.
func KeyValueCustom(key string, keyWidth int, value string) string {
	keyStyled := lipgloss.NewStyle().Foreground(ColorMuted).Width(keyWidth).Render(key)
	valueStyled := lipgloss.NewStyle().Foreground(ColorText).Render(value)
	return keyStyled + " " + valueStyled
}

// KeyValueBool renders a boolean value in green/red.
func KeyValueBool(key string, value bool) string {
	valStr := "true"
	valStyle := lipgloss.NewStyle().Foreground(ColorSuccess)
	if !value {
		valStr = "false"
		valStyle = lipgloss.NewStyle().Foreground(ColorError)
	}
	return KeyValue(key, valStyle.Render(valStr))
}

// StatusBadge returns a small badge for a status string.
func StatusBadge(status string) string {
	var c lipgloss.Color
	switch strings.ToLower(status) {
	case "ok", "healthy", "pass", "success", "active", "running":
		c = ColorSuccess
	case "fail", "error", "unhealthy", "critical", "expired":
		c = ColorError
	case "warn", "warning", "degraded", "static", "pending":
		c = ColorWarning
	default:
		c = ColorMuted
	}
	return lipgloss.NewStyle().
		Foreground(c).
		Bold(true).
		Render(strings.ToUpper(status))
}

// ─── Table Helper ────────────────────────────────────────────────────

// TableRow holds data for one table row.
type TableRow []string

// SimpleTable renders a minimal text table with headers and alternating row backgrounds.
func SimpleTable(headers []string, rows []TableRow) string {
	if len(headers) == 0 || len(rows) == 0 {
		return ""
	}

	// Compute column widths
	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = runewidth.StringWidth(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) {
				w := runewidth.StringWidth(cell)
				if w > colWidths[i] {
					colWidths[i] = w
				}
			}
		}
	}

	// Add some padding
	for i := range colWidths {
		colWidths[i] += 2
	}

	var b strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorText).Background(ColorPrimary).Padding(0, 1)
	cellStyle := lipgloss.NewStyle().Padding(0, 1)
	altCellStyle := lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("#1e1e1e"))

	// Headers
	for i, h := range headers {
		b.WriteString(headerStyle.Width(colWidths[i]).Render(h))
	}
	b.WriteString("\n")

	// Rows
	for i, row := range rows {
		style := cellStyle
		if i%2 == 1 {
			style = altCellStyle
		}
		for j, cell := range row {
			if j < len(colWidths) {
				b.WriteString(style.Width(colWidths[j]).Render(cell))
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

// ─── Panel / Box Helpers ─────────────────────────────────────────────

// Panel returns a string wrapped in a subtle rounded panel.
func Panel(content string) string {
	return StylePanel.Render(content)
}

// InfoBox returns content wrapped in a cyan-accent panel.
func InfoBox(content string) string {
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorInfo).
		Padding(0, 1)
	return style.Render(content)
}

// SuccessBox returns content wrapped in a green-accent panel.
func SuccessBox(content string) string {
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorSuccess).
		Padding(0, 1)
	return style.Render(content)
}

// ErrorBox returns content wrapped in a red-accent panel.
func ErrorBox(content string) string {
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorError).
		Padding(0, 1)
	return style.Render(content)
}

// WarningBox returns content wrapped in a yellow-accent panel.
func WarningBox(content string) string {
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorWarning).
		Padding(0, 1)
	return style.Render(content)
}

// ─── Progress / Action Helpers ───────────────────────────────────────

// Action returns a message indicating an ongoing action.
func Action(action, target string) string {
	return fmt.Sprintf("%s %s %s",
		StyleMuted.Render("..."),
		StyleInfo.Render(action),
		StyleMuted.Render(target),
	)
}

// Done returns a completion message for an action.
func Done(action, target string) string {
	return Success(fmt.Sprintf("%s %s", action, target))
}

// ─── Inline Helpers ──────────────────────────────────────────────────

// Join joins multiple styled strings with a separator.
func Join(sep string, parts ...string) string {
	return strings.Join(parts, sep)
}

// Indent indents every line of text.
func Indent(text string, spaces int) string {
	prefix := strings.Repeat(" ", spaces)
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}
