// Package display provides rich terminal formatting for xsh.
package display

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display/frame"
	"github.com/benoitpetit/xsh/models"
	"github.com/benoitpetit/xsh/utils"
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

	verified := ""
	if user.Verified {
		verified = Primary(" ✓")
	}
	lines = append(lines, Bold(user.Name)+verified+Muted(" @"+user.Handle))
	lines = append(lines, "")

	if user.Bio != "" {
		lines = append(lines, user.Bio)
		lines = append(lines, "")
	}

	stats := fmt.Sprintf("%s Following  ·  %s Followers  ·  %s Tweets",
		Bold(FormatNumber(user.FollowingCount)),
		Bold(FormatNumber(user.FollowersCount)),
		Bold(FormatNumber(user.TweetCount)))
	lines = append(lines, stats)

	var details []string
	if user.Location != "" {
		details = append(details, Muted("📍")+" "+user.Location)
	}
	if user.Website != "" {
		details = append(details, Muted("🔗")+" "+user.Website)
	}
	if user.CreatedAt != nil {
		details = append(details, Muted("📅")+" Joined "+user.CreatedAt.Format("January 2006"))
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
		return EmptyState("No users found.")
	}

	headers := []string{"Handle", "Name", "Followers", "Following", "Bio"}
	var rows []TableRow
	for _, user := range users {
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
		rows = append(rows, TableRow{
			"@" + user.Handle,
			verified + user.Name,
			FormatNumber(user.FollowersCount),
			FormatNumber(user.FollowingCount),
			bio,
		})
	}

	return Subtitle(fmt.Sprintf("Users · %d", len(users))) + "\n\n" + SimpleTable(headers, rows)
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
	var lines []string

	lines = append(lines, Title("⚙️  Configuration"))
	lines = append(lines, Section("General"))
	lines = append(lines, KeyValue("default_count:", fmt.Sprintf("%d", cv.DefaultCount)))
	lines = append(lines, KeyValue("default_account:", cv.DefaultAccount))

	lines = append(lines, Section("Display"))
	lines = append(lines, KeyValue("theme:", cv.Theme))
	lines = append(lines, KeyValueBool("show_engagement:", cv.ShowEngagement))
	lines = append(lines, KeyValueBool("show_timestamps:", cv.ShowTimestamps))
	lines = append(lines, KeyValue("max_width:", fmt.Sprintf("%d", cv.MaxWidth)))

	lines = append(lines, Section("Request"))
	lines = append(lines, KeyValue("delay:", fmt.Sprintf("%.2fs", cv.Delay)))
	proxy := cv.Proxy
	if proxy == "" {
		proxy = "(none)"
	}
	lines = append(lines, KeyValue("proxy:", proxy))
	lines = append(lines, KeyValue("timeout:", fmt.Sprintf("%ds", cv.Timeout)))
	lines = append(lines, KeyValue("max_retries:", fmt.Sprintf("%d", cv.MaxRetries)))

	lines = append(lines, Section("Filter Weights"))
	lines = append(lines, KeyValue("likes:", fmt.Sprintf("%.1f", cv.LikesWeight)))
	lines = append(lines, KeyValue("retweets:", fmt.Sprintf("%.1f", cv.RetweetsWeight)))
	lines = append(lines, KeyValue("replies:", fmt.Sprintf("%.1f", cv.RepliesWeight)))
	lines = append(lines, KeyValue("bookmarks:", fmt.Sprintf("%.1f", cv.BookmarksWeight)))
	lines = append(lines, KeyValue("views_log:", fmt.Sprintf("%.1f", cv.ViewsLogWeight)))

	return strings.Join(lines, "\n")
}

func FormatThread(tweets []*models.Tweet, focalID string) string {
	if len(tweets) == 0 {
		return EmptyState("No tweets in thread.")
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

	b.WriteString(Title(fmt.Sprintf("Thread · %d tweets", len(tweets))))
	b.WriteString("\n\n")

	var buildTree func(tweet *models.Tweet, prefix string, isLast bool)
	buildTree = func(tweet *models.Tweet, prefix string, isLast bool) {
		branchChar := branch
		if isLast {
			branchChar = lastBranch
		}

		verified := ""
		if tweet.AuthorVerified {
			verified = Primary(" ✓")
		}

		timeStr := Muted(RelativeTime(tweet.CreatedAt))
		handleLine := fmt.Sprintf("%s%s@%s%s %s", prefix, branchChar, tweet.AuthorHandle, verified, timeStr)
		b.WriteString(handleLine + "\n")

		text := strings.ReplaceAll(tweet.Text, "\n", " ")
		text = strings.ReplaceAll(text, "\r", " ")
		text = strings.Join(strings.Fields(text), " ")

		textPrefix := prefix
		if isLast {
			textPrefix += indent
		} else {
			textPrefix += vertical
		}

		wrappedLines := wrapText(text, 70)
		for _, line := range wrappedLines {
			b.WriteString(textPrefix + line + "\n")
		}

		e := tweet.Engagement
		if e.Likes > 0 || e.Retweets > 0 {
			stats := []string{}
			if e.Likes > 0 {
				stats = append(stats, StyleError.Render(fmt.Sprintf("❤️ %s", FormatNumber(e.Likes))))
			}
			if e.Retweets > 0 {
				stats = append(stats, StyleInfo.Render(fmt.Sprintf("🔁 %s", FormatNumber(e.Retweets))))
			}
			b.WriteString(textPrefix + strings.Join(stats, " ") + "\n")
		}

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

func FormatTweet(tweet *models.Tweet, prefix string, isLast bool, forceVertical bool) string {
	const (
		vertical   = "│  "
		branch     = "├──"
		lastBranch = "└──"
		indent     = "   "
	)

	var b strings.Builder

	verified := ""
	if tweet.AuthorVerified {
		verified = Primary(" ✓")
	}

	branchChar := branch
	if isLast {
		branchChar = lastBranch
	}

	timeStr := Muted(RelativeTime(tweet.CreatedAt))
	b.WriteString(fmt.Sprintf("%s%s@%s%s %s\n", prefix, branchChar, tweet.AuthorHandle, verified, timeStr))

	text := strings.ReplaceAll(tweet.Text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.Join(strings.Fields(text), " ")

	textPrefix := prefix
	if isLast && !forceVertical {
		textPrefix += indent
	} else {
		textPrefix += vertical
	}

	wrappedLines := wrapText(text, 70)
	for _, line := range wrappedLines {
		b.WriteString(textPrefix + line + "\n")
	}

	e := tweet.Engagement
	if e.Likes > 0 || e.Retweets > 0 {
		stats := []string{}
		if e.Likes > 0 {
			stats = append(stats, StyleError.Render(fmt.Sprintf("❤️ %s", FormatNumber(e.Likes))))
		}
		if e.Retweets > 0 {
			stats = append(stats, StyleInfo.Render(fmt.Sprintf("🔁 %s", FormatNumber(e.Retweets))))
		}
		b.WriteString(textPrefix + strings.Join(stats, " ") + "\n")
	}

	idStyle := lipgloss.NewStyle().Foreground(colorGray)
	b.WriteString(textPrefix + idStyle.Render("🆔 "+tweet.ID) + "\n")

	return b.String()
}

func FormatTweetList(tweets []*models.Tweet) string {
	if len(tweets) == 0 {
		return EmptyState("No tweets found.")
	}

	var b strings.Builder

	b.WriteString(Title(fmt.Sprintf("Tweets · %d", len(tweets))))
	b.WriteString("\n\n")

	for i, tweet := range tweets {
		isLast := i == len(tweets)-1
		b.WriteString(FormatTweet(tweet, "", isLast, true))
	}

	return b.String()
}

func FormatSingleTweet(tweet *models.Tweet) string {
	var b strings.Builder

	b.WriteString(Title("Tweet"))
	b.WriteString("\n\n")

	verified := ""
	if tweet.AuthorVerified {
		verified = Primary(" ✓")
	}

	timeStr := Muted(RelativeTime(tweet.CreatedAt))
	b.WriteString(fmt.Sprintf("└──@%s%s %s\n", tweet.AuthorHandle, verified, timeStr))

	text := strings.ReplaceAll(tweet.Text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.Join(strings.Fields(text), " ")

	wrappedLines := wrapText(text, 70)
	for _, line := range wrappedLines {
		b.WriteString("   " + line + "\n")
	}

	e := tweet.Engagement
	if e.Replies > 0 || e.Retweets > 0 || e.Likes > 0 || e.Bookmarks > 0 || e.Views > 0 {
		stats := []string{}
		if e.Replies > 0 {
			stats = append(stats, StyleInfo.Render(fmt.Sprintf("💬 %s", FormatNumber(e.Replies))))
		}
		if e.Retweets > 0 {
			stats = append(stats, StyleInfo.Render(fmt.Sprintf("🔁 %s", FormatNumber(e.Retweets))))
		}
		if e.Likes > 0 {
			stats = append(stats, StyleError.Render(fmt.Sprintf("❤️ %s", FormatNumber(e.Likes))))
		}
		if e.Bookmarks > 0 {
			stats = append(stats, StyleWarning.Render(fmt.Sprintf("🔖 %s", FormatNumber(e.Bookmarks))))
		}
		if e.Views > 0 {
			stats = append(stats, StyleMuted.Render(fmt.Sprintf("👁️ %s", FormatNumber(e.Views))))
		}
		b.WriteString("   " + strings.Join(stats, "  ") + "\n")
	}

	idStyle := lipgloss.NewStyle().Foreground(colorGray)
	b.WriteString("   " + idStyle.Render("🆔 "+tweet.ID) + "\n")

	return b.String()
}

func FormatUsers(users []*models.User) string {
	return FormatUserList(users)
}

func FormatTweets(tweets []*models.Tweet) string {
	return FormatTweetList(tweets)
}

func FormatDMInbox(conversations []models.DMConversation) string {
	if len(conversations) == 0 {
		return EmptyState("No conversations found.")
	}

	var b strings.Builder

	b.WriteString(Title(fmt.Sprintf("DM Inbox · %d conversations", len(conversations))))
	b.WriteString("\n\n")

	for i, conv := range conversations {
		indicator := "  "
		if conv.Unread {
			indicator = StylePrimary.Render("● ")
		}

		var participantNames []string
		for _, p := range conv.Participants {
			participantNames = append(participantNames, Bold("@"+p.Handle))
		}
		participants := strings.Join(participantNames, ", ")

		lastMsg := conv.LastMessage
		if len(lastMsg) > 50 {
			lastMsg = lastMsg[:47] + "..."
		}

		line := fmt.Sprintf("%s%s: %s", indicator, participants, lastMsg)
		if conv.LastMessageTime != "" {
			line += Muted(fmt.Sprintf(" (%s)", conv.LastMessageTime))
		}

		b.WriteString(line)
		if i < len(conversations)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func FormatScheduledTweets(tweets []core.ScheduledTweet) string {
	if len(tweets) == 0 {
		return EmptyState("No scheduled tweets.")
	}

	var b strings.Builder

	b.WriteString(Title(fmt.Sprintf("Scheduled Tweets · %d", len(tweets))))
	b.WriteString("\n\n")

	for i, tweet := range tweets {
		scheduledTime := time.Unix(tweet.ExecuteAt, 0)
		timeStr := scheduledTime.Format("2006-01-02 15:04")

		text := tweet.Text
		if len(text) > 60 {
			text = text[:57] + "..."
		}

		status := tweet.State
		if status == "" {
			status = "scheduled"
		}

		badge := StatusBadge(status)
		b.WriteString(fmt.Sprintf("%s %s: %s %s", badge, timeStr, text, Muted("(ID: "+tweet.ID+")")))
		if i < len(tweets)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func FormatLists(lists []core.ListInfo) string {
	if len(lists) == 0 {
		return EmptyState("No lists found.")
	}

	var b strings.Builder

	b.WriteString(Title(fmt.Sprintf("Lists · %d", len(lists))))
	b.WriteString("\n\n")

	for i, list := range lists {
		mode := list.Mode
		if mode == "" {
			mode = "Public"
		}

		b.WriteString(fmt.Sprintf("%s %s", Bold(list.Name), StatusBadge(mode)))
		b.WriteString(Muted(fmt.Sprintf("  (%d members)  ID: %s", list.MemberCount, list.ID)))
		if list.Description != "" {
			b.WriteString(fmt.Sprintf("\n  %s", list.Description))
		}
		if i < len(lists)-1 {
			b.WriteString("\n\n")
		}
	}

	return b.String()
}

func FormatBookmarkFolders(folders []core.BookmarkFolder) string {
	if len(folders) == 0 {
		return EmptyState("No bookmark folders.")
	}

	var b strings.Builder

	b.WriteString(Title(fmt.Sprintf("Bookmark Folders · %d", len(folders))))
	b.WriteString("\n\n")

	for i, folder := range folders {
		b.WriteString(fmt.Sprintf("%s %s", Bold(folder.Name), Muted("(ID: "+folder.ID+")")))
		if i < len(folders)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func FormatJobs(jobs []interface{}) string {
	if len(jobs) == 0 {
		return EmptyState("No jobs found.")
	}

	var b strings.Builder

	b.WriteString(Title(fmt.Sprintf("Jobs · %d", len(jobs))))
	b.WriteString("\n\n")

	for i, jobInterface := range jobs {
		if job, ok := jobInterface.(models.Job); ok {
			b.WriteString(Bold(job.Title) + Muted(" at ") + Info(job.Company.Name))
			if job.Location != "" {
				b.WriteString(Muted(fmt.Sprintf(" (%s)", job.Location)))
			}
			if job.WorkplaceType != "" {
				b.WriteString(" " + StatusBadge(job.WorkplaceType))
			}
			b.WriteString(Muted(fmt.Sprintf("\n  ID: %s", job.ID)))
			if i < len(jobs)-1 {
				b.WriteString("\n\n")
			}
		}
	}

	return b.String()
}

func FormatJobDetail(job *models.Job) string {
	var lines []string

	lines = append(lines, Title(job.Title))
	lines = append(lines, Info(job.Company.Name))
	lines = append(lines, "")

	if job.Location != "" {
		lines = append(lines, Muted("📍")+" "+job.Location)
	}
	if job.WorkplaceType != "" {
		lines = append(lines, Muted("🏢")+" "+job.WorkplaceType)
	}
	if job.EmploymentType != "" {
		lines = append(lines, Muted("💼")+" "+job.EmploymentType)
	}
	if job.Salary != "" {
		lines = append(lines, Muted("💰")+" "+job.Salary)
	}

	lines = append(lines, "")
	lines = append(lines, KeyValue("ID:", job.ID))

	if job.Description != "" {
		lines = append(lines, "")
		lines = append(lines, Section("Description"))
		var descData map[string]interface{}
		if err := json.Unmarshal([]byte(job.Description), &descData); err == nil {
			markdown := utils.ArticleToMarkdown(map[string]interface{}{"result": map[string]interface{}{"content": map[string]interface{}{"content_state": descData}}})
			if markdown != "" {
				lines = append(lines, markdown)
			} else {
				lines = append(lines, job.Description)
			}
		} else {
			lines = append(lines, job.Description)
		}
	}

	if job.ApplyURL != "" {
		lines = append(lines, "")
		lines = append(lines, Primary("🔗 Apply: "+job.ApplyURL))
	}

	return strings.Join(lines, "\n")
}

func FormatArticle(title, author, contentMD string, engagement models.TweetEngagement) string {
	var b strings.Builder

	b.WriteString(Title("Article"))
	b.WriteString("\n\n")

	if title != "" {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorWhite).Width(80).Render(title))
		b.WriteString("\n\n")
	}

	b.WriteString(Muted(fmt.Sprintf("By @%s", author)))
	b.WriteString("\n\n")

	if engagement.Likes > 0 || engagement.Retweets > 0 || engagement.Replies > 0 {
		stats := []string{}
		if engagement.Replies > 0 {
			stats = append(stats, StyleInfo.Render(fmt.Sprintf("💬 %s", FormatNumber(engagement.Replies))))
		}
		if engagement.Retweets > 0 {
			stats = append(stats, StyleInfo.Render(fmt.Sprintf("🔁 %s", FormatNumber(engagement.Retweets))))
		}
		if engagement.Likes > 0 {
			stats = append(stats, StyleError.Render(fmt.Sprintf("❤️ %s", FormatNumber(engagement.Likes))))
		}
		if engagement.Bookmarks > 0 {
			stats = append(stats, StyleWarning.Render(fmt.Sprintf("🔖 %s", FormatNumber(engagement.Bookmarks))))
		}
		b.WriteString(strings.Join(stats, "  "))
		b.WriteString("\n\n")
	}

	lines := strings.Split(contentMD, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			b.WriteString("\n")
			continue
		}
		if strings.HasPrefix(line, "# ") {
			b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorWhite).Render(line[2:]))
			b.WriteString("\n")
		} else if strings.HasPrefix(line, "## ") {
			b.WriteString(lipgloss.NewStyle().Bold(true).Render(line[3:]))
			b.WriteString("\n")
		} else if strings.HasPrefix(line, "### ") {
			b.WriteString(lipgloss.NewStyle().Bold(true).Render(line[4:]))
			b.WriteString("\n")
		} else if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			b.WriteString("  • " + line[2:])
			b.WriteString("\n")
		} else if strings.HasPrefix(line, "> ") {
			quoteStyle := lipgloss.NewStyle().Foreground(colorGray).Italic(true)
			b.WriteString(quoteStyle.Render("  \"" + line[2:] + "\""))
			b.WriteString("\n")
		} else if strings.HasPrefix(line, "```") {
			if line == "```" {
				continue
			}
			b.WriteString(Code("  " + line))
			b.WriteString("\n")
		} else {
			wrapped := wrapText(line, 78)
			for _, w := range wrapped {
				b.WriteString(w)
				b.WriteString("\n")
			}
		}
	}

	return b.String()
}
