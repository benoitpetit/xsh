// Package display provides rich terminal formatting for xsh.
package display

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display/frame"
	"github.com/benoitpetit/xsh/models"
	"github.com/benoitpetit/xsh/utils"
	"github.com/charmbracelet/lipgloss"
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

// FormatUnrolledThread formats a self-reply chain as a clean readable document.
// Only shows tweets by the thread author, concatenated with separators.
func FormatUnrolledThread(chain []*models.Tweet, author string) string {
	if len(chain) == 0 {
		return EmptyState("No thread content found.")
	}

	var b strings.Builder

	b.WriteString(Title(fmt.Sprintf("Thread by @%s · %d parts", author, len(chain))))
	b.WriteString("\n\n")

	for i, tweet := range chain {
		// Part header
		partHeader := Muted(fmt.Sprintf("[%d/%d]", i+1, len(chain)))
		timeStr := ""
		if tweet.CreatedAt != nil {
			timeStr = " " + Muted(RelativeTime(tweet.CreatedAt))
		}
		b.WriteString(partHeader + timeStr + "\n")

		// Tweet text — preserve newlines for readability
		text := strings.TrimSpace(tweet.Text)
		wrappedLines := wrapText(text, 76)
		for _, line := range wrappedLines {
			b.WriteString(line + "\n")
		}

		// Media indicators
		if len(tweet.Media) > 0 {
			for _, m := range tweet.Media {
				b.WriteString(Muted(fmt.Sprintf("  [%s: %s]", m.Type, m.URL)) + "\n")
			}
		}

		// Poll indicator
		if tweet.Poll != nil {
			b.WriteString(formatPoll(tweet.Poll, "  "))
		}

		if i < len(chain)-1 {
			b.WriteString("\n")
		}
	}

	// Summary
	b.WriteString("\n")
	b.WriteString(Muted(fmt.Sprintf("Thread by @%s · %d parts", author, len(chain))))
	if len(chain) > 0 && chain[0].CreatedAt != nil {
		b.WriteString(Muted(fmt.Sprintf(" · %s", chain[0].CreatedAt.Format("Jan 2, 2006"))))
	}
	b.WriteString("\n")

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

// formatPoll renders a poll with bar charts for each choice.
// prefix is prepended to each line (for tree indentation).
func formatPoll(poll *models.Poll, prefix string) string {
	if poll == nil || len(poll.Choices) == 0 {
		return ""
	}

	const barWidth = 20 // characters for the bar
	var b strings.Builder

	statusStyle := lipgloss.NewStyle().Foreground(colorGray)
	status := statusStyle.Render("[" + poll.Status + "]")
	b.WriteString(prefix + "📊 Poll " + status + "\n")

	total := poll.TotalVotes()
	for _, choice := range poll.Choices {
		pct := choice.Pct
		filled := int(pct / 100 * float64(barWidth))
		if filled > barWidth {
			filled = barWidth
		}
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

		pctStr := fmt.Sprintf("%4.1f%%", pct)
		votesStr := FormatNumber(choice.Votes)

		barStyle := lipgloss.NewStyle().Foreground(colorCyan)
		b.WriteString(fmt.Sprintf("%s  %s %s %s (%s)\n",
			prefix,
			barStyle.Render(bar),
			pctStr,
			choice.Label,
			votesStr,
		))
	}

	totalStr := Muted(fmt.Sprintf("  %s total votes", FormatNumber(total)))
	b.WriteString(prefix + totalStr + "\n")

	return b.String()
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

	// Display poll if present
	if tweet.Poll != nil {
		b.WriteString(formatPoll(tweet.Poll, textPrefix))
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

	// Display poll if present
	if tweet.Poll != nil {
		b.WriteString(formatPoll(tweet.Poll, "   "))
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

// FormatCommunity formats a community for display
func FormatCommunity(community *core.Community) string {
	if community == nil || community.ID == "" {
		return EmptyState("Community not found.")
	}

	var lines []string

	lines = append(lines, Bold(community.Name))
	lines = append(lines, Muted("Community ID: "+community.ID))
	lines = append(lines, "")

	if community.Description != "" {
		lines = append(lines, community.Description)
		lines = append(lines, "")
	}

	stats := fmt.Sprintf("%s Members  ·  %s Moderators",
		Bold(FormatNumber(community.MemberCount)),
		Bold(FormatNumber(community.ModeratorCount)))
	lines = append(lines, stats)

	if community.Role != "" {
		lines = append(lines, Muted("Your role: ")+community.Role)
	}

	if community.IsNSFW {
		lines = append(lines, StyleWarning.Render("NSFW"))
	}

	if len(community.Rules) > 0 {
		lines = append(lines, "")
		lines = append(lines, Section("Rules"))
		for i, rule := range community.Rules {
			lines = append(lines, Numbered(i+1, Bold(rule.Name)))
			if rule.Description != "" {
				lines = append(lines, "   "+Muted(rule.Description))
			}
		}
	}

	content := strings.Join(lines, "\n")
	return frame.CardCyan(content)
}

// FormatSpace formats a Twitter Space for display
func FormatSpace(space *core.Space) string {
	if space == nil || space.ID == "" {
		return EmptyState("Space not found.")
	}

	var lines []string

	// State badge
	stateBadge := StatusBadge(space.State)
	lines = append(lines, stateBadge+" "+Bold(space.Title))
	lines = append(lines, Muted("Space ID: "+space.ID))
	lines = append(lines, "")

	// Participants
	lines = append(lines, fmt.Sprintf("%s Participants", Bold(FormatNumber(space.ParticipantCount))))

	// Hosts
	if len(space.Hosts) > 0 {
		var hostNames []string
		for _, h := range space.Hosts {
			hostNames = append(hostNames, "@"+h.Handle)
		}
		lines = append(lines, Muted("Hosts: ")+strings.Join(hostNames, ", "))
	}

	// Timing
	if space.ScheduledStart != "" {
		lines = append(lines, Muted("Scheduled: ")+space.ScheduledStart)
	}
	if space.StartedAt != "" {
		lines = append(lines, Muted("Started: ")+space.StartedAt)
	}
	if space.EndedAt != "" {
		lines = append(lines, Muted("Ended: ")+space.EndedAt)
	}

	content := strings.Join(lines, "\n")
	return frame.CardCyan(content)
}

// FormatSpaces formats a list of Twitter Spaces
func FormatSpaces(spaces []core.Space) string {
	if len(spaces) == 0 {
		return EmptyState("No Spaces found.")
	}

	headers := []string{"ID", "Title", "State", "Participants"}
	var rows []TableRow
	for _, s := range spaces {
		title := s.Title
		if len(title) > 40 {
			title = title[:37] + "..."
		}
		rows = append(rows, TableRow{
			s.ID,
			title,
			s.State,
			FormatNumber(s.ParticipantCount),
		})
	}

	return Subtitle(fmt.Sprintf("Spaces · %d", len(spaces))) + "\n\n" + SimpleTable(headers, rows)
}

// FormatNotifications formats notifications for display
func FormatNotifications(notifications []core.Notification) string {
	if len(notifications) == 0 {
		return EmptyState("No notifications.")
	}

	var lines []string
	lines = append(lines, Subtitle(fmt.Sprintf("Notifications · %d", len(notifications))))
	lines = append(lines, "")

	for _, n := range notifications {
		icon := notificationIcon(n.Type)
		line := icon + " "

		if n.UserHandle != "" {
			line += Primary("@"+n.UserHandle) + " "
		}

		if n.Message != "" {
			line += n.Message
		} else {
			line += n.Type
		}

		if n.TweetText != "" {
			text := n.TweetText
			if len(text) > 60 {
				text = text[:57] + "..."
			}
			line += "\n   " + Muted(text)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// FormatAnalytics renders engagement analytics for terminal display
func FormatAnalytics(stats *models.TweetAnalytics, handle string) string {
	var lines []string

	lines = append(lines, Subtitle(fmt.Sprintf("Analytics · @%s · %d tweets", handle, stats.TotalTweets)))
	lines = append(lines, "")

	// ── Overview ──
	lines = append(lines, Section("Overview"))
	lines = append(lines, KeyValue("Total views", FormatNumber(stats.TotalViews)))
	lines = append(lines, KeyValue("Total likes", FormatNumber(stats.TotalLikes)))
	lines = append(lines, KeyValue("Total retweets", FormatNumber(stats.TotalRetweets)))
	lines = append(lines, KeyValue("Total replies", FormatNumber(stats.TotalReplies)))
	lines = append(lines, KeyValue("Total bookmarks", FormatNumber(stats.TotalBookmarks)))
	lines = append(lines, "")

	// ── Averages ──
	lines = append(lines, Section("Averages (per tweet)"))
	lines = append(lines, KeyValue("Avg views", fmt.Sprintf("%.0f", stats.AvgViews)))
	lines = append(lines, KeyValue("Avg likes", fmt.Sprintf("%.1f", stats.AvgLikes)))
	lines = append(lines, KeyValue("Avg retweets", fmt.Sprintf("%.1f", stats.AvgRetweets)))
	lines = append(lines, KeyValue("Avg replies", fmt.Sprintf("%.1f", stats.AvgReplies)))
	lines = append(lines, "")

	// ── Engagement Rate ──
	rateStr := fmt.Sprintf("%.2f%%", stats.EngagementRate)
	rateStyle := lipgloss.NewStyle().Bold(true)
	if stats.EngagementRate >= 5 {
		rateStyle = rateStyle.Foreground(ColorSuccess)
	} else if stats.EngagementRate >= 2 {
		rateStyle = rateStyle.Foreground(ColorWarning)
	} else {
		rateStyle = rateStyle.Foreground(ColorError)
	}
	lines = append(lines, KeyValue("Engagement rate", rateStyle.Render(rateStr)))
	lines = append(lines, Muted("  (likes + retweets + replies) / views"))
	lines = append(lines, "")

	// ── Media Breakdown ──
	if len(stats.MediaBreakdown) > 0 {
		lines = append(lines, Section("Content Type Breakdown"))
		for mediaType, count := range stats.MediaBreakdown {
			pct := float64(count) / float64(stats.TotalTweets) * 100
			lines = append(lines, KeyValue(mediaType, fmt.Sprintf("%d (%.0f%%)", count, pct)))
		}
		lines = append(lines, "")
	}

	// ── Top Tweets by Likes ──
	if len(stats.TopByLikes) > 0 {
		lines = append(lines, Section("Top by Likes"))
		for i, t := range stats.TopByLikes {
			text := t.Text
			if len(text) > 60 {
				text = text[:57] + "..."
			}
			lines = append(lines, fmt.Sprintf("  %s %s  %s  %s",
				Muted(fmt.Sprintf("%d.", i+1)),
				text,
				lipgloss.NewStyle().Foreground(ColorError).Render(fmt.Sprintf("♥ %s", FormatNumber(t.Engagement.Likes))),
				Muted(fmt.Sprintf("👁 %s", FormatNumber(t.Engagement.Views))),
			))
		}
		lines = append(lines, "")
	}

	// ── Top Tweets by Views ──
	if len(stats.TopByViews) > 0 {
		lines = append(lines, Section("Top by Views"))
		for i, t := range stats.TopByViews {
			text := t.Text
			if len(text) > 60 {
				text = text[:57] + "..."
			}
			lines = append(lines, fmt.Sprintf("  %s %s  %s  %s",
				Muted(fmt.Sprintf("%d.", i+1)),
				text,
				Muted(fmt.Sprintf("👁 %s", FormatNumber(t.Engagement.Views))),
				lipgloss.NewStyle().Foreground(ColorError).Render(fmt.Sprintf("♥ %s", FormatNumber(t.Engagement.Likes))),
			))
		}
		lines = append(lines, "")
	}

	// ── Top Tweets by Retweets ──
	if len(stats.TopByRetweets) > 0 {
		lines = append(lines, Section("Top by Retweets"))
		for i, t := range stats.TopByRetweets {
			text := t.Text
			if len(text) > 60 {
				text = text[:57] + "..."
			}
			lines = append(lines, fmt.Sprintf("  %s %s  %s  %s",
				Muted(fmt.Sprintf("%d.", i+1)),
				text,
				lipgloss.NewStyle().Foreground(ColorSuccess).Render(fmt.Sprintf("↻ %s", FormatNumber(t.Engagement.Retweets))),
				Muted(fmt.Sprintf("👁 %s", FormatNumber(t.Engagement.Views))),
			))
		}
	}

	return strings.Join(lines, "\n")
}

// FormatRateLimits renders rate limit information for terminal display
func FormatRateLimits(limits []*core.RateLimitInfo) string {
	var lines []string
	lines = append(lines, Subtitle(fmt.Sprintf("Rate Limits · %d endpoints tracked", len(limits))))
	lines = append(lines, "")

	headers := []string{"Endpoint", "Remaining", "Limit", "Used %", "Resets In"}
	var rows []TableRow

	for _, rl := range limits {
		usage := rl.UsagePercent()
		usageStr := fmt.Sprintf("%.0f%%", usage)

		// Color the usage percentage
		usageStyled := usageStr
		if usage >= 90 {
			usageStyled = lipgloss.NewStyle().Foreground(ColorError).Bold(true).Render(usageStr)
		} else if usage >= 70 {
			usageStyled = lipgloss.NewStyle().Foreground(ColorWarning).Render(usageStr)
		} else {
			usageStyled = lipgloss.NewStyle().Foreground(ColorSuccess).Render(usageStr)
		}

		resetStr := "—"
		secs := rl.SecondsUntilReset()
		if secs > 0 {
			if secs >= 60 {
				resetStr = fmt.Sprintf("%dm %ds", secs/60, secs%60)
			} else {
				resetStr = fmt.Sprintf("%ds", secs)
			}
		} else if rl.Limit > 0 {
			resetStr = "now"
		}

		rows = append(rows, TableRow{
			rl.Endpoint,
			fmt.Sprintf("%d", rl.Remaining),
			fmt.Sprintf("%d", rl.Limit),
			usageStyled,
			resetStr,
		})
	}

	lines = append(lines, SimpleTable(headers, rows))
	return strings.Join(lines, "\n")
}

// notificationIcon returns an icon for a notification type
func notificationIcon(notifType string) string {
	switch notifType {
	case "like":
		return lipgloss.NewStyle().Foreground(ColorError).Render("♥")
	case "retweet":
		return lipgloss.NewStyle().Foreground(ColorSuccess).Render("↻")
	case "follow":
		return lipgloss.NewStyle().Foreground(ColorPrimary).Render("👤")
	case "reply":
		return lipgloss.NewStyle().Foreground(ColorInfo).Render("💬")
	case "mention":
		return lipgloss.NewStyle().Foreground(ColorWarning).Render("@")
	case "quote":
		return lipgloss.NewStyle().Foreground(ColorInfo).Render("❝")
	default:
		return lipgloss.NewStyle().Foreground(ColorMuted).Render("•")
	}
}
