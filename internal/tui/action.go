package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type actionKind int

const (
	actionClose actionKind = iota
	actionAssign
	actionComment
)

type actionModel struct {
	kind        actionKind
	issueNumber int
	deps        Dependencies
	input       textinput.Model
	loading     bool
	spinner     spinner.Model
	err         error
	done        bool
	result      string
	width       int
	height      int
}

func newActionModel(deps Dependencies, kind actionKind, issueNumber, width, height int) actionModel {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	switch kind {
	case actionClose:
		ti.Placeholder = "closing comment (optional)"
	case actionAssign:
		ti.Placeholder = "usernames (comma-separated)"
	case actionComment:
		ti.Placeholder = "comment text"
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

	return actionModel{
		kind:        kind,
		issueNumber: issueNumber,
		deps:        deps,
		input:       ti,
		spinner:     s,
		width:       width,
		height:      height,
	}
}

func (m actionModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m actionModel) Update(msg tea.Msg) (actionModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		if m.done || m.err != nil {
			return m, func() tea.Msg { return actionDoneMsg{} }
		}

		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return actionCancelledMsg{} }
		case "enter":
			if m.kind == actionAssign && strings.TrimSpace(m.input.Value()) == "" {
				return m, nil
			}
			if m.kind == actionComment && strings.TrimSpace(m.input.Value()) == "" {
				return m, nil
			}
			m.loading = true
			return m, tea.Batch(m.execute(), m.spinner.Tick)
		}

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case actionSuccessMsg:
		m.done = true
		m.loading = false
		m.result = msg.text

	case errMsg:
		m.err = msg.err
		m.loading = false
	}

	if !m.loading && !m.done && m.err == nil {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m actionModel) execute() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		switch m.kind {
		case actionClose:
			_, err := m.deps.GitHubClient.CloseIssue(ctx, m.issueNumber, m.input.Value())
			if err != nil {
				return errMsg{err: err}
			}
			return actionSuccessMsg{text: fmt.Sprintf("Issue #%d closed", m.issueNumber)}

		case actionAssign:
			raw := strings.Split(m.input.Value(), ",")
			assignees := make([]string, 0, len(raw))
			for _, a := range raw {
				if trimmed := strings.TrimSpace(a); trimmed != "" {
					assignees = append(assignees, trimmed)
				}
			}
			_, err := m.deps.GitHubClient.AssignIssue(ctx, m.issueNumber, assignees)
			if err != nil {
				return errMsg{err: err}
			}
			return actionSuccessMsg{text: fmt.Sprintf("Issue #%d assigned to %s", m.issueNumber, strings.Join(assignees, ", "))}

		case actionComment:
			_, err := m.deps.GitHubClient.AddComment(ctx, m.issueNumber, m.input.Value())
			if err != nil {
				return errMsg{err: err}
			}
			return actionSuccessMsg{text: fmt.Sprintf("Comment added to issue #%d", m.issueNumber)}
		}
		return nil
	}
}

func (m actionModel) View() string {
	var lines []string

	title := m.actionTitle()
	lines = append(lines, modalTitleStyle.Render(title))
	lines = append(lines, "")

	if m.loading {
		lines = append(lines, fmt.Sprintf("%s Processing...", m.spinner.View()))
	} else if m.done {
		lines = append(lines, successStyle.Render(m.result))
		lines = append(lines, "")
		lines = append(lines, dimStyle.Render("Press any key to continue"))
	} else if m.err != nil {
		lines = append(lines, errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		lines = append(lines, "")
		lines = append(lines, dimStyle.Render("Press any key to continue"))
	} else {
		lines = append(lines, m.input.View())
		lines = append(lines, "")
		lines = append(lines, dimStyle.Render("enter submit · esc cancel"))
	}

	content := strings.Join(lines, "\n")
	modal := modalStyle.Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
}

func (m actionModel) actionTitle() string {
	switch m.kind {
	case actionClose:
		return fmt.Sprintf("Close Issue #%d", m.issueNumber)
	case actionAssign:
		return fmt.Sprintf("Assign Issue #%d", m.issueNumber)
	case actionComment:
		return fmt.Sprintf("Comment on Issue #%d", m.issueNumber)
	}
	return ""
}
