package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func renderStatusBar(repo string, state string, page int, count int, width int) string {
	left := fmt.Sprintf(" %s · %s", repo, state)
	right := fmt.Sprintf("Page %d · %d issues ", page, count)

	gap := width - lipgloss.Width(left) - lipgloss.Width(right) - statusBarStyle.GetHorizontalPadding()
	if gap < 0 {
		gap = 0
	}

	bar := statusBarStyle.Width(width).Render(
		left + lipgloss.NewStyle().Width(gap).Render("") + right,
	)
	return bar
}
