package render

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/julian0xff/save/internal/db"
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true)
	mutedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
)

func ArticleHeader(a *db.Article) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(a.Title))
	b.WriteString("\n")

	var meta []string
	if a.Author != "" {
		meta = append(meta, "By: "+a.Author)
	}
	meta = append(meta, fmt.Sprintf("%d words, ~%d min read", a.WordCount, a.ReadTime))
	meta = append(meta, a.URL)
	b.WriteString(mutedStyle.Render(strings.Join(meta, " | ")))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render(strings.Repeat("─", 60)))
	b.WriteString("\n\n")
	return b.String()
}

func ArticleContent(content string) (string, error) {
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return content, nil
	}
	rendered, err := r.Render(content)
	if err != nil {
		return content, nil
	}
	return rendered, nil
}

func ArticleList(articles []db.Article) string {
	if len(articles) == 0 {
		return mutedStyle.Render("No articles saved yet. Use `save <url>` to add one.")
	}

	rows := make([][]string, len(articles))
	totalWords := 0
	for i, a := range articles {
		status := "○"
		if a.ReadAt != nil {
			status = "✓"
		}

		title := a.Title
		if len(title) > 50 {
			title = title[:47] + "..."
		}

		tags := a.Tags
		if len(tags) > 20 {
			tags = tags[:17] + "..."
		}

		rows[i] = []string{
			fmt.Sprintf("%d", a.ID),
			status,
			title,
			tags,
			RelativeTime(a.SavedAt),
			fmt.Sprintf("%d", a.WordCount),
		}
		totalWords += a.WordCount
	}

	t := table.New().
		Border(lipgloss.HiddenBorder()).
		Headers("#", "", "Title", "Tags", "Saved", "Words").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			if col == 1 { // status column
				return lipgloss.NewStyle().Foreground(lipgloss.Color("34"))
			}
			return lipgloss.NewStyle()
		}).
		Rows(rows...)

	var b strings.Builder
	b.WriteString(t.Render())
	b.WriteString("\n\n")

	totalReadTime := totalWords / 238
	b.WriteString(mutedStyle.Render(fmt.Sprintf("%d articles (%d words, ~%d min total)", len(articles), totalWords, totalReadTime)))
	b.WriteString("\n")

	return b.String()
}

func SearchResults(results []db.SearchResult) string {
	if len(results) == 0 {
		return mutedStyle.Render("No results found.")
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%d result(s):\n\n", len(results)))

	for _, r := range results {
		b.WriteString(fmt.Sprintf("  #%d  %s\n", r.ID, titleStyle.Render(r.Title)))
		if r.Tags != "" {
			b.WriteString(fmt.Sprintf("       Tags: %s\n", r.Tags))
		}
		b.WriteString(fmt.Sprintf("       %s\n", mutedStyle.Render(r.Snippet)))
		b.WriteString(fmt.Sprintf("       %s\n\n", mutedStyle.Render(RelativeTime(r.SavedAt))))
	}

	return b.String()
}

func RelativeTime(t time.Time) string {
	d := time.Since(t)

	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", h)
	case d < 48*time.Hour:
		return "yesterday"
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	case d < 30*24*time.Hour:
		w := int(d.Hours() / 24 / 7)
		if w == 1 {
			return "1w ago"
		}
		return fmt.Sprintf("%dw ago", w)
	default:
		return t.Format("Jan 2")
	}
}
