package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dulait/grit/internal/github"
)

type listModel struct {
	deps    Dependencies
	issues  []github.Issue
	cursor  int
	page    int
	perPage int
	state   string
	loading bool
	spinner spinner.Model
	err     error
	width   int
	height  int
}

func newListModel(deps Dependencies) listModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

	return listModel{
		deps:    deps,
		page:    1,
		perPage: 20,
		state:   "open",
		loading: true,
		spinner: s,
	}
}

func (m listModel) Init() tea.Cmd {
	return tea.Batch(m.fetchIssues(), m.spinner.Tick)
}

func (m listModel) fetchIssues() tea.Cmd {
	return func() tea.Msg {
		svc := m.deps.IssueServiceWithoutLLM()
		req := github.ListIssuesRequest{
			State:   m.state,
			PerPage: m.perPage,
			Page:    m.page,
		}
		issues, err := svc.ListIssues(context.Background(), req)
		if err != nil {
			return errMsg{err: err}
		}
		return issuesLoadedMsg{issues: issues, page: m.page}
	}
}

func (m listModel) Update(msg tea.Msg) (listModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case issuesLoadedMsg:
		m.issues = msg.issues
		m.page = msg.page
		m.loading = false
		m.cursor = 0
		m.err = nil

	case errMsg:
		m.err = msg.err
		m.loading = false

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		switch {
		case key.Matches(msg, listKeys.Down):
			if m.cursor < len(m.issues)-1 {
				m.cursor++
			}
		case key.Matches(msg, listKeys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, listKeys.NextPage):
			if len(m.issues) == m.perPage {
				m.page++
				m.loading = true
				return m, tea.Batch(m.fetchIssues(), m.spinner.Tick)
			}
		case key.Matches(msg, listKeys.PrevPage):
			if m.page > 1 {
				m.page--
				m.loading = true
				return m, tea.Batch(m.fetchIssues(), m.spinner.Tick)
			}
		case key.Matches(msg, listKeys.Refresh):
			m.loading = true
			return m, tea.Batch(m.fetchIssues(), m.spinner.Tick)
		case key.Matches(msg, listKeys.FilterOpen):
			m.state = "open"
			m.page = 1
			m.loading = true
			return m, tea.Batch(m.fetchIssues(), m.spinner.Tick)
		case key.Matches(msg, listKeys.FilterClosed):
			m.state = "closed"
			m.page = 1
			m.loading = true
			return m, tea.Batch(m.fetchIssues(), m.spinner.Tick)
		case key.Matches(msg, listKeys.FilterAll):
			m.state = "all"
			m.page = 1
			m.loading = true
			return m, tea.Batch(m.fetchIssues(), m.spinner.Tick)
		}
	}

	return m, nil
}

func (m listModel) View() string {
	var b strings.Builder

	repo := fmt.Sprintf("%s/%s", m.deps.Config.Project.Owner, m.deps.Config.Project.Repo)
	header := headerStyle.Width(m.width).Render(fmt.Sprintf(" grit · %s", repo))
	b.WriteString(header)
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(fmt.Sprintf("  %s Loading issues...\n", m.spinner.View()))
		return b.String()
	}

	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("  Error: %v", m.err)))
		b.WriteString("\n")
		return b.String()
	}

	if len(m.issues) == 0 {
		b.WriteString(dimStyle.Render("  No issues found."))
		b.WriteString("\n")
	} else {
		for i, issue := range m.issues {
			b.WriteString(m.renderIssueRow(i, issue))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(renderStatusBar(repo, m.state, m.page, len(m.issues), m.width))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  j/k navigate · enter open · n/p page · 1/2/3 filter · ? help · q quit"))

	return b.String()
}

func (m listModel) renderIssueRow(index int, issue github.Issue) string {
	number := fmt.Sprintf("#%-4d", issue.Number)

	maxTitle := m.width - 30
	if maxTitle < 20 {
		maxTitle = 20
	}
	title := truncateStr(issue.Title, maxTitle)

	var state string
	if issue.State == "open" {
		state = stateOpenStyle.Render("open")
	} else {
		state = stateClosedStyle.Render("closed")
	}

	var labels string
	if len(issue.Labels) > 0 {
		names := make([]string, len(issue.Labels))
		for i, l := range issue.Labels {
			names[i] = l.Name
		}
		labels = labelStyle.Render(strings.Join(names, ","))
	}

	var assignees string
	if len(issue.Assignees) > 0 {
		names := make([]string, len(issue.Assignees))
		for i, u := range issue.Assignees {
			names[i] = u.Login
		}
		assignees = assigneeStyle.Render(strings.Join(names, ","))
	}

	parts := []string{number, title, state}
	if labels != "" {
		parts = append(parts, labels)
	}
	if assignees != "" {
		parts = append(parts, assignees)
	}
	row := "  " + strings.Join(parts, "  ")

	if index == m.cursor {
		return selectedStyle.Width(m.width).Render(row)
	}
	return normalStyle.Render(row)
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
