package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type helpBinding struct {
	key  string
	desc string
}

var listHelpBindings = []helpBinding{
	{"j/k", "navigate up/down"},
	{"enter/l", "open issue"},
	{"n/p", "next/prev page"},
	{"r", "refresh"},
	{"1/2/3", "filter: open/closed/all"},
	{"?", "toggle help"},
	{"q", "quit"},
}

var detailHelpBindings = []helpBinding{
	{"j/k", "scroll up/down"},
	{"ctrl+u/d", "half page up/down"},
	{"o", "open in browser"},
	{"esc/h", "back to list"},
	{"?", "toggle help"},
	{"q", "quit"},
}

func renderHelp(width int, currentScreen screen) string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Render("Key Bindings")

	bindings := listHelpBindings
	if currentScreen == screenDetail {
		bindings = detailHelpBindings
	}

	var lines []string
	lines = append(lines, title)
	lines = append(lines, "")

	for _, b := range bindings {
		key := lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Width(12).Render(b.key)
		lines = append(lines, "  "+key+helpStyle.Render(b.desc))
	}

	lines = append(lines, "")
	lines = append(lines, helpStyle.Render("  Press ? to close"))

	content := strings.Join(lines, "\n")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Render(content)

	return lipgloss.Place(width, 0, lipgloss.Center, lipgloss.Top, box)
}
