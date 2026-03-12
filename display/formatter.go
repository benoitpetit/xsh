// Package display provides rich terminal formatting for xsh.
package display

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/benoitpetit/xsh/display/frame"
	"github.com/benoitpetit/xsh/models"
)

// Style definitions
var (
	// ColorBlue is the Twitter blue color
	ColorBlue   = lipgloss.Color("#1DA1F2")
	// ColorGreen is the success green color
	ColorGreen  = lipgloss.Color("#00BA7C")
	// ColorRed is the error red color
	ColorRed    = lipgloss.Color("#F4212E")
	// ColorYellow is the warning yellow color
	ColorYellow = lipgloss.Color("#FFAD1F")
	// ColorGray is the muted gray color
	ColorGray   = lipgloss.Color("#8899A6")
	// ColorWhite is the white color
	ColorWhite  = lipgloss.Color("#FFFFFF")
	// ColorCyan is the cyan color
	ColorCyan   = lipgloss.Color("#00BCD4")

	colorBlue   = ColorBlue
	colorGreen  = ColorGreen
	colorRed    = ColorRed
	colorYellow = ColorYellow
	colorGray   = ColorGray
	colorWhite  = ColorWhite
	colorCyan   = ColorCyan

	successStyle = lipgloss.NewStyle().Foreground(colorGreen).Bold(true)
	errorStyle = lipgloss.NewStyle().Foreground(colorRed).Bold(true)
	warningStyle = lipgloss.NewStyle().Foreground(colorYellow).Bold(true)

	// StyleBold is bold text style
	StyleBold = lipgloss.NewStyle().Bold(true).Foreground(colorWhite)
	// StyleMuted is muted/gray text style
	StyleMuted = lipgloss.NewStyle().Foreground(colorGray)
)

func RelativeTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	now := time.Now()
	diff := now.Sub(*t)
	seconds := int(diff.Seconds())
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	} else if seconds < 3600 {
		return fmt.Sprintf("%dm", seconds/60)
	} else if seconds < 86400 {
		return fmt.Sprintf("%dh", seconds/3600)
	} else if seconds < 604800 {
		return fmt.Sprintf("%dd", seconds/86400)
	}
	return t.Format("Jan 2, 2006")
}

func FormatNumber(n int) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	} else if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}

func FormatUser(user *models.User) string {
	var lines []string

	// Header
	verified := ""
	if user.Verified {
		verified = " ✓"
	}
	lines = append(lines, fmt.Sprintf("%s%s @%s", user.Name, verified, user.Handle))
	lines = append(lines, "")

	// Bio
	if user.Bio != "" {
		lines = append(lines, user.Bio)
		lines = append(lines, "")
	}

	// Stats
	lines = append(lines, fmt.Sprintf("%s Following  ·  %s Followers  ·  %s Tweets",
		FormatNumber(user.FollowingCount),
		FormatNumber(user.FollowersCount),
		FormatNumber(user.TweetCount)))

	// Details
	var details []string
	if user.Location != "" {
		details = append(details, "📍 "+user.Location)
	}
	if user.Website != "" {
		details = append(details, "🔗 "+user.Website)
	}
	if user.CreatedAt != nil {
		details = append(details, "📅 Joined "+user.CreatedAt.Format("January 2006"))
	}
	if len(details) > 0 {
		lines = append(lines, "")
		lines = append(lines, strings.Join(details, "  ·  "))
	}

	content := strings.Join(lines, "\n")
	return frame.CardCyan(content)
}

func FormatUserList(users []*models.User) string {
	if len(users) == 0 {
		return lipgloss.NewStyle().Foreground(colorGray).Render("No users found.")
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(colorWhite).Background(colorBlue).Padding(0, 1)
	cellStyle := lipgloss.NewStyle().Padding(0, 1)
	altCellStyle := lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("#232323"))

	var b strings.Builder

	headers := []string{"Handle", "Name", "Followers", "Following", "Bio"}
	colWidths := []int{20, 20, 12, 12, 40}

	for i, h := range headers {
		b.WriteString(headerStyle.Width(colWidths[i]).Render(h))
	}
	b.WriteString("\n")

	for i, user := range users {
		style := cellStyle
		if i%2 == 1 {
			style = altCellStyle
		}

		verified := ""
		if user.Verified {
			verified = "✓ "
		}

		bio := user.Bio
		if len(bio) > 35 {
			bio = bio[:32] + "..."
		}
		if bio == "" {
			bio = "-"
		}

		b.WriteString(style.Width(colWidths[0]).Render("@" + user.Handle))
		b.WriteString(style.Width(colWidths[1]).Render(verified + user.Name))
		b.WriteString(style.Width(colWidths[2]).Align(lipgloss.Right).Render(FormatNumber(user.FollowersCount)))
		b.WriteString(style.Width(colWidths[3]).Align(lipgloss.Right).Render(FormatNumber(user.FollowingCount)))
		b.WriteString(style.Width(colWidths[4]).Render(bio))
		b.WriteString("\n")
	}

	return b.String()
}

type ConfigValues struct {
	DefaultCount    int
	DefaultAccount  string
	Theme           string
	ShowEngagement  bool
	ShowTimestamps  bool
	MaxWidth        int
	Delay           float64
	Proxy           string
	Timeout         int
	MaxRetries      int
	LikesWeight     float64
	RetweetsWeight  float64
	RepliesWeight   float64
	BookmarksWeight float64
	ViewsLogWeight  float64
}

func FormatConfig(cv *ConfigValues) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(colorCyan).MarginBottom(1)
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(colorBlue).MarginTop(1)
	keyStyle := lipgloss.NewStyle().Foreground(colorGray).Width(25)
	valueStyle := lipgloss.NewStyle().Foreground(colorWhite)
	boolTrueStyle := lipgloss.NewStyle().Foreground(colorGreen)
	boolFalseStyle := lipgloss.NewStyle().Foreground(colorRed)

	var b strings.Builder
	b.WriteString(titleStyle.Render("⚙️  Configuration"))
	b.WriteString("\n")

	b.WriteString(sectionStyle.Render("General"))
	b.WriteString("\n")
	b.WriteString(keyStyle.Render("default_count") + valueStyle.Render(fmt.Sprintf("%d", cv.DefaultCount)) + "\n")
	b.WriteString(keyStyle.Render("default_account") + valueStyle.Render(cv.DefaultAccount) + "\n")

	b.WriteString(sectionStyle.Render("Display"))
	b.WriteString("\n")
	b.WriteString(keyStyle.Render("theme") + valueStyle.Render(cv.Theme) + "\n")
	if cv.ShowEngagement {
		b.WriteString(keyStyle.Render("show_engagement") + boolTrueStyle.Render("true") + "\n")
	} else {
		b.WriteString(keyStyle.Render("show_engagement") + boolFalseStyle.Render("false") + "\n")
	}
	if cv.ShowTimestamps {
		b.WriteString(keyStyle.Render("show_timestamps") + boolTrueStyle.Render("true") + "\n")
	} else {
		b.WriteString(keyStyle.Render("show_timestamps") + boolFalseStyle.Render("false") + "\n")
	}
	b.WriteString(keyStyle.Render("max_width") + valueStyle.Render(fmt.Sprintf("%d", cv.MaxWidth)) + "\n")

	b.WriteString(sectionStyle.Render("Request"))
	b.WriteString("\n")
	b.WriteString(keyStyle.Render("delay") + valueStyle.Render(fmt.Sprintf("%.2fs", cv.Delay)) + "\n")
	proxy := cv.Proxy
	if proxy == "" {
		proxy = "(none)"
	}
	b.WriteString(keyStyle.Render("proxy") + valueStyle.Render(proxy) + "\n")
	b.WriteString(keyStyle.Render("timeout") + valueStyle.Render(fmt.Sprintf("%ds", cv.Timeout)) + "\n")
	b.WriteString(keyStyle.Render("max_retries") + valueStyle.Render(fmt.Sprintf("%d", cv.MaxRetries)) + "\n")

	b.WriteString(sectionStyle.Render("Filter Weights"))
	b.WriteString("\n")
	b.WriteString(keyStyle.Render("likes") + valueStyle.Render(fmt.Sprintf("%.1f", cv.LikesWeight)) + "\n")
	b.WriteString(keyStyle.Render("retweets") + valueStyle.Render(fmt.Sprintf("%.1f", cv.RetweetsWeight)) + "\n")
	b.WriteString(keyStyle.Render("replies") + valueStyle.Render(fmt.Sprintf("%.1f", cv.RepliesWeight)) + "\n")
	b.WriteString(keyStyle.Render("bookmarks") + valueStyle.Render(fmt.Sprintf("%.1f", cv.BookmarksWeight)) + "\n")
	b.WriteString(keyStyle.Render("views_log") + valueStyle.Render(fmt.Sprintf("%.1f", cv.ViewsLogWeight)) + "\n")

	return b.String()
}

func PrintSuccess(message string) string {
	return successStyle.Render("✓ " + message)
}

func PrintError(message string) string {
	return errorStyle.Render("✗ " + message)
}

func PrintWarning(message string) string {
	return warningStyle.Render("⚠ " + message)
}

func FormatThread(tweets []*models.Tweet, focalID string) string {
	if len(tweets) == 0 {
		return lipgloss.NewStyle().Foreground(colorGray).Render("No tweets in thread.")
	}

	var focal *models.Tweet
	for _, t := range tweets {
		if t.ID == focalID {
			focal = t
			break
		}
	}
	if focal == nil {
		focal = tweets[0]
	}

	repliesMap := make(map[string][]*models.Tweet)
	for _, t := range tweets {
		parent := t.ReplyToID
		repliesMap[parent] = append(repliesMap[parent], t)
	}

	const (
		vertical   = "│  "
		branch     = "├──"
		lastBranch = "└──"
		indent     = "   "
	)

	var b strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(colorWhite)
	b.WriteString(headerStyle.Render(fmt.Sprintf("Thread · %d tweets", len(tweets))))
	b.WriteString("\n\n")

	var buildTree func(tweet *models.Tweet, prefix string, isLast bool)
	buildTree = func(tweet *models.Tweet, prefix string, isLast bool) {
		branchChar := branch
		if isLast {
			branchChar = lastBranch
		}

		verified := ""
		if tweet.AuthorVerified {
			verified = " ✓"
		}

		timeStr := RelativeTime(tweet.CreatedAt)
		b.WriteString(fmt.Sprintf("%s@%s%s %s\n", prefix+branchChar, tweet.AuthorHandle, verified, timeStr))

		// Replace newlines with spaces to keep tree structure intact
		text := strings.ReplaceAll(tweet.Text, "\n", " ")
		text = strings.ReplaceAll(text, "\r", " ")
		// Collapse multiple spaces into one
		text = strings.Join(strings.Fields(text), " ")

		textPrefix := prefix
		if isLast {
			textPrefix += indent
		} else {
			textPrefix += vertical
		}

		// Wrap text at 70 chars with proper indentation
		wrappedLines := wrapText(text, 70)
		for _, line := range wrappedLines {
			b.WriteString(textPrefix + line + "\n")
		}

		e := tweet.Engagement
		if e.Likes > 0 || e.Retweets > 0 {
			stats := []string{}
			if e.Likes > 0 {
				stats = append(stats, fmt.Sprintf("❤️ %s", FormatNumber(e.Likes)))
			}
			if e.Retweets > 0 {
				stats = append(stats, fmt.Sprintf("🔁 %s", FormatNumber(e.Retweets)))
			}
			b.WriteString(textPrefix + strings.Join(stats, " ") + "\n")
		}

		// Tweet ID (styled in gray)
		idStyle := lipgloss.NewStyle().Foreground(colorGray)
		b.WriteString(textPrefix + idStyle.Render("🆔 "+tweet.ID) + "\n")

		b.WriteString(textPrefix + "\n")

		replies := repliesMap[tweet.ID]
		for i, reply := range replies {
			newPrefix := prefix
			if isLast {
				newPrefix += indent
			} else {
				newPrefix += vertical
			}
			buildTree(reply, newPrefix, i == len(replies)-1)
		}
	}

	buildTree(focal, "", true)

	return b.String()
}

// wrapText wraps text at maxWidth, breaking at word boundaries
func wrapText(text string, maxWidth int) []string {
	if len(text) <= maxWidth {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	var currentLine string

	for _, word := range words {
		if len(currentLine)+len(word)+1 > maxWidth {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		} else {
			if currentLine == "" {
				currentLine = word
			} else {
				currentLine += " " + word
			}
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// FormatTweet displays a tweet in simple tree-like style (no frame)
func FormatTweet(tweet *models.Tweet, prefix string, isLast bool) string {
	const (
		vertical   = "│  "
		branch     = "├──"
		lastBranch = "└──"
		indent     = "   "
	)

	var b strings.Builder

	verified := ""
	if tweet.AuthorVerified {
		verified = " ✓"
	}

	branchChar := branch
	if isLast {
		branchChar = lastBranch
	}

	timeStr := RelativeTime(tweet.CreatedAt)
	b.WriteString(fmt.Sprintf("%s%s@%s%s %s\n", prefix, branchChar, tweet.AuthorHandle, verified, timeStr))

	// Replace newlines with spaces and wrap text
	text := strings.ReplaceAll(tweet.Text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.Join(strings.Fields(text), " ")

	// Determine text prefix based on position
	textPrefix := prefix
	if isLast {
		textPrefix += indent
	} else {
		textPrefix += vertical
	}

	// Wrap text at 70 chars
	wrappedLines := wrapText(text, 70)
	for _, line := range wrappedLines {
		b.WriteString(textPrefix + line + "\n")
	}

	// Engagement stats
	e := tweet.Engagement
	if e.Likes > 0 || e.Retweets > 0 {
		stats := []string{}
		if e.Likes > 0 {
			stats = append(stats, fmt.Sprintf("❤️ %s", FormatNumber(e.Likes)))
		}
		if e.Retweets > 0 {
			stats = append(stats, fmt.Sprintf("🔁 %s", FormatNumber(e.Retweets)))
		}
		b.WriteString(textPrefix + strings.Join(stats, " ") + "\n")
	}

	// Tweet ID (styled in gray to be distinct but visible)
	idStyle := lipgloss.NewStyle().Foreground(colorGray)
	b.WriteString(textPrefix + idStyle.Render("🆔 "+tweet.ID) + "\n")

	return b.String()
}

// FormatTweetList displays a list of tweets in simple style (no frames)
func FormatTweetList(tweets []*models.Tweet) string {
	if len(tweets) == 0 {
		return lipgloss.NewStyle().Foreground(colorGray).Render("No tweets found.")
	}

	var b strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(colorWhite)
	b.WriteString(headerStyle.Render(fmt.Sprintf("Tweets · %d", len(tweets))))
	b.WriteString("\n\n")

	// Format each tweet
	for i, tweet := range tweets {
		isLast := i == len(tweets)-1
		b.WriteString(FormatTweet(tweet, "", isLast))
	}

	return b.String()
}

// FormatSingleTweet displays a single tweet in simple tree-like style
func FormatSingleTweet(tweet *models.Tweet) string {
	var b strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(colorWhite)
	b.WriteString(headerStyle.Render("Tweet"))
	b.WriteString("\n\n")

	verified := ""
	if tweet.AuthorVerified {
		verified = " ✓"
	}

	timeStr := RelativeTime(tweet.CreatedAt)
	b.WriteString(fmt.Sprintf("└──@%s%s %s\n", tweet.AuthorHandle, verified, timeStr))

	// Replace newlines with spaces and wrap text
	text := strings.ReplaceAll(tweet.Text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.Join(strings.Fields(text), " ")

	// Wrap text at 70 chars
	wrappedLines := wrapText(text, 70)
	for _, line := range wrappedLines {
		b.WriteString("   " + line + "\n")
	}

	// Engagement stats
	e := tweet.Engagement
	if e.Replies > 0 || e.Retweets > 0 || e.Likes > 0 || e.Bookmarks > 0 || e.Views > 0 {
		stats := []string{}
		if e.Replies > 0 {
			stats = append(stats, fmt.Sprintf("💬 %s", FormatNumber(e.Replies)))
		}
		if e.Retweets > 0 {
			stats = append(stats, fmt.Sprintf("🔁 %s", FormatNumber(e.Retweets)))
		}
		if e.Likes > 0 {
			stats = append(stats, fmt.Sprintf("❤️ %s", FormatNumber(e.Likes)))
		}
		if e.Bookmarks > 0 {
			stats = append(stats, fmt.Sprintf("🔖 %s", FormatNumber(e.Bookmarks)))
		}
		if e.Views > 0 {
			stats = append(stats, fmt.Sprintf("👁️ %s", FormatNumber(e.Views)))
		}
		b.WriteString("   " + strings.Join(stats, "  ") + "\n")
	}

	// Tweet ID (styled in gray)
	idStyle := lipgloss.NewStyle().Foreground(colorGray)
	b.WriteString("   " + idStyle.Render("🆔 "+tweet.ID) + "\n")

	return b.String()
}
