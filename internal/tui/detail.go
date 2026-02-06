package tui

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dulait/grit/internal/github"
)

type detailModel struct {
	deps        Dependencies
	issueNumber int
	issue       *github.Issue
	loading     bool
	spinner     spinner.Model
	viewport    viewport.Model
	ready       bool
	err         error
	width       int
	height      int
}

func newDetailModel(deps Dependencies, issueNumber int) detailModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

	return detailModel{
		deps:        deps,
		issueNumber: issueNumber,
		loading:     true,
		spinner:     s,
	}
}

func (m detailModel) Init() tea.Cmd {
	return tea.Batch(m.fetchIssue(), m.spinner.Tick)
}

func (m detailModel) fetchIssue() tea.Cmd {
	number := m.issueNumber
	return func() tea.Msg {
		issue, err := m.deps.GitHubClient.GetIssue(context.Background(), number)
		if err != nil {
			return errMsg{err: err}
		}
		return issueDetailLoadedMsg{issue: issue}
	}
}

func (m detailModel) Update(msg tea.Msg) (detailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerHeight := 8
		footerHeight := 2
		if m.ready {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - headerHeight - footerHeight
		}

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case issueDetailLoadedMsg:
		m.issue = msg.issue
		m.loading = false
		m.err = nil
		m.viewport = viewport.New(m.width, m.height-10)
		m.viewport.SetContent(m.renderBody())
		m.ready = true

	case errMsg:
		m.err = msg.err
		m.loading = false

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		switch {
		case key.Matches(msg, detailKeys.Back):
			return m, func() tea.Msg { return navigateToListMsg{} }
		case key.Matches(msg, detailKeys.OpenBrowser):
			if m.issue != nil {
				openBrowser(m.issue.HTMLURL)
			}
			return m, nil
		case key.Matches(msg, detailKeys.HalfPageUp):
			m.viewport.HalfViewUp()
			return m, nil
		case key.Matches(msg, detailKeys.HalfPageDwn):
			m.viewport.HalfViewDown()
			return m, nil
		case key.Matches(msg, detailKeys.Close):
			if m.issue != nil {
				number := m.issueNumber
				return m, func() tea.Msg {
					return startActionMsg{kind: actionClose, issueNumber: number}
				}
			}
		case key.Matches(msg, detailKeys.Assign):
			if m.issue != nil {
				number := m.issueNumber
				return m, func() tea.Msg {
					return startActionMsg{kind: actionAssign, issueNumber: number}
				}
			}
		case key.Matches(msg, detailKeys.Comment):
			if m.issue != nil {
				number := m.issueNumber
				return m, func() tea.Msg {
					return startActionMsg{kind: actionComment, issueNumber: number}
				}
			}
		}

		if m.ready {
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m detailModel) View() string {
	var b strings.Builder

	header := headerStyle.Width(m.width).Render(fmt.Sprintf(" grit · Issue #%d", m.issueNumber))
	b.WriteString(header)
	b.WriteString("\n")

	if m.loading {
		b.WriteString(fmt.Sprintf("\n  %s Loading issue...\n", m.spinner.View()))
		return b.String()
	}

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(fmt.Sprintf("  Error: %v", m.err)))
		b.WriteString("\n")
		return b.String()
	}

	if m.issue == nil {
		return b.String()
	}

	b.WriteString("\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("  %s", m.issue.Title)))
	b.WriteString("\n\n")
	b.WriteString(m.renderMetadata())
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", m.width))
	b.WriteString("\n")

	if m.ready {
		b.WriteString(m.viewport.View())
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  j/k scroll · x close · a assign · m comment · o browser · esc/h back · ? help"))

	return b.String()
}

func (m detailModel) renderMetadata() string {
	issue := m.issue
	var parts []string

	var state string
	if issue.State == "open" {
		state = stateOpenStyle.Render("open")
	} else {
		state = stateClosedStyle.Render("closed")
	}
	parts = append(parts, fmt.Sprintf("  State: %s", state))

	if len(issue.Labels) > 0 {
		names := make([]string, len(issue.Labels))
		for i, l := range issue.Labels {
			names[i] = l.Name
		}
		parts = append(parts, fmt.Sprintf("  Labels: %s", labelStyle.Render(strings.Join(names, ", "))))
	}

	if len(issue.Assignees) > 0 {
		names := make([]string, len(issue.Assignees))
		for i, u := range issue.Assignees {
			names[i] = u.Login
		}
		parts = append(parts, fmt.Sprintf("  Assignees: %s", assigneeStyle.Render(strings.Join(names, ", "))))
	}

	parts = append(parts, fmt.Sprintf("  URL: %s", dimStyle.Render(issue.HTMLURL)))

	return strings.Join(parts, "\n")
}

func (m detailModel) renderBody() string {
	if m.issue.Body == "" {
		return dimStyle.Render("  No description provided.")
	}
	return "  " + strings.ReplaceAll(m.issue.Body, "\n", "\n  ")
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	_ = cmd.Start()
}
