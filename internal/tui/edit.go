package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dulait/grit/internal/github"
	"github.com/dulait/grit/internal/service"
)

type editStep int

const (
	editLoading editStep = iota
	editInput
	editSaving
	editDone
)

const (
	editFieldTitle = iota
	editFieldBody
	editFieldLabels
	editFieldAssignees
	editFieldState
	editFieldCount
)

type editModel struct {
	deps        Dependencies
	issueNumber int
	original    *github.Issue
	step        editStep
	inputs      []textinput.Model
	focusIndex  int
	spinner     spinner.Model
	updated     *github.Issue
	err         error
	width       int
	height      int
}

func newEditModel(deps Dependencies, issueNumber, width, height int) editModel {
	inputs := make([]textinput.Model, editFieldCount)

	inputs[editFieldTitle] = newEditInput("Title", 120)
	inputs[editFieldBody] = newEditInput("Body", 512)
	inputs[editFieldLabels] = newEditInput("Labels (comma-separated)", 120)
	inputs[editFieldAssignees] = newEditInput("Assignees (comma-separated)", 120)
	inputs[editFieldState] = newEditInput("State (open/closed)", 10)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

	return editModel{
		deps:        deps,
		issueNumber: issueNumber,
		step:        editLoading,
		inputs:      inputs,
		spinner:     s,
		width:       width,
		height:      height,
	}
}

func newEditInput(placeholder string, charLimit int) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = charLimit
	ti.Width = 60
	return ti
}

func (m editModel) Init() tea.Cmd {
	return tea.Batch(m.fetchIssue(), m.spinner.Tick)
}

func (m editModel) fetchIssue() tea.Cmd {
	number := m.issueNumber
	deps := m.deps
	return func() tea.Msg {
		issue, err := deps.GitHubClient.GetIssue(context.Background(), number)
		if err != nil {
			return errMsg{err: err}
		}
		return issueDetailLoadedMsg{issue: issue}
	}
}

func (m editModel) Update(msg tea.Msg) (editModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch m.step {
		case editInput:
			return m.updateInput(msg)
		case editDone:
			if msg.String() == "o" && m.updated != nil {
				openBrowser(m.updated.HTMLURL)
				return m, nil
			}
			return m, func() tea.Msg {
				return navigateToDetailMsg{issueNumber: m.issueNumber}
			}
		}
		if m.err != nil {
			m.err = nil
			m.step = editInput
			return m, nil
		}

	case spinner.TickMsg:
		if m.step == editLoading || m.step == editSaving {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case issueDetailLoadedMsg:
		m.original = msg.issue
		m.populateInputs()
		m.step = editInput
		m.inputs[editFieldTitle].Focus()
		return m, textinput.Blink

	case issueUpdatedMsg:
		m.updated = msg.issue
		m.step = editDone

	case errMsg:
		m.err = msg.err
		if m.step == editLoading {
			m.step = editLoading
		} else {
			m.step = editInput
		}
	}

	if m.step == editInput {
		return m.updateInputFields(msg)
	}

	return m, nil
}

func (m *editModel) populateInputs() {
	m.inputs[editFieldTitle].SetValue(m.original.Title)
	m.inputs[editFieldBody].SetValue(m.original.Body)

	labelNames := make([]string, len(m.original.Labels))
	for i, l := range m.original.Labels {
		labelNames[i] = l.Name
	}
	m.inputs[editFieldLabels].SetValue(strings.Join(labelNames, ", "))

	assigneeNames := make([]string, len(m.original.Assignees))
	for i, u := range m.original.Assignees {
		assigneeNames[i] = u.Login
	}
	m.inputs[editFieldAssignees].SetValue(strings.Join(assigneeNames, ", "))

	m.inputs[editFieldState].SetValue(m.original.State)
}

func (m editModel) updateInput(msg tea.KeyMsg) (editModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m, func() tea.Msg {
			return navigateToDetailMsg{issueNumber: m.issueNumber}
		}
	case "tab", "down":
		m.focusIndex = (m.focusIndex + 1) % editFieldCount
		return m.syncFocus(), nil
	case "shift+tab", "up":
		m.focusIndex = (m.focusIndex - 1 + editFieldCount) % editFieldCount
		return m.syncFocus(), nil
	case "ctrl+s":
		m.step = editSaving
		return m, tea.Batch(m.save(), m.spinner.Tick)
	}

	return m.updateInputFields(msg)
}

func (m editModel) syncFocus() editModel {
	for i := range m.inputs {
		m.inputs[i].Blur()
	}
	m.inputs[m.focusIndex].Focus()
	return m
}

func (m editModel) updateInputFields(msg tea.Msg) (editModel, tea.Cmd) {
	var cmds []tea.Cmd
	for i := range m.inputs {
		var cmd tea.Cmd
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m editModel) save() tea.Cmd {
	original := m.original
	deps := m.deps
	number := m.issueNumber

	input := service.EditIssueInput{}

	title := strings.TrimSpace(m.inputs[editFieldTitle].Value())
	if title != original.Title {
		input.Title = &title
	}

	body := strings.TrimSpace(m.inputs[editFieldBody].Value())
	if body != original.Body {
		input.Body = &body
	}

	state := strings.TrimSpace(m.inputs[editFieldState].Value())
	if state != original.State && (state == "open" || state == "closed") {
		input.State = &state
	}

	labelNames := make([]string, len(original.Labels))
	for i, l := range original.Labels {
		labelNames[i] = l.Name
	}
	newLabels := parseCSVInput(m.inputs[editFieldLabels].Value())
	if !slicesEqual(newLabels, labelNames) {
		input.Labels = newLabels
		input.SetLabels = true
	}

	assigneeNames := make([]string, len(original.Assignees))
	for i, u := range original.Assignees {
		assigneeNames[i] = u.Login
	}
	newAssignees := parseCSVInput(m.inputs[editFieldAssignees].Value())
	if !slicesEqual(newAssignees, assigneeNames) {
		input.Assignees = newAssignees
		input.SetAssignees = true
	}

	return func() tea.Msg {
		svc := deps.IssueServiceWithoutLLM()
		issue, err := svc.EditIssue(context.Background(), number, input)
		if err != nil {
			return errMsg{err: err}
		}
		return issueUpdatedMsg{issue: issue}
	}
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (m editModel) View() string {
	var b strings.Builder

	header := headerStyle.Width(m.width).Render(fmt.Sprintf(" grit · Edit Issue #%d", m.issueNumber))
	b.WriteString(header)
	b.WriteString("\n\n")

	switch m.step {
	case editLoading:
		b.WriteString(fmt.Sprintf("  %s Loading issue...\n", m.spinner.View()))
	case editInput:
		b.WriteString(m.viewInput())
	case editSaving:
		b.WriteString(fmt.Sprintf("  %s Saving changes...\n", m.spinner.View()))
	case editDone:
		b.WriteString(m.viewDone())
	}

	if m.err != nil && m.step == editInput {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(fmt.Sprintf("  Error: %v", m.err)))
		b.WriteString("\n")
	}

	return b.String()
}

func (m editModel) viewInput() string {
	var b strings.Builder

	labels := []string{"  Title:", "  Body:", "  Labels:", "  Assignees:", "  State:"}
	for i, label := range labels {
		style := dimStyle
		if i == m.focusIndex {
			style = titleStyle
		}
		b.WriteString(style.Render(label))
		b.WriteString("\n")
		b.WriteString("  " + m.inputs[i].View())
		b.WriteString("\n\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  tab/shift+tab navigate · ctrl+s save · esc cancel"))

	return b.String()
}

func (m editModel) viewDone() string {
	var b strings.Builder

	b.WriteString(successStyle.Render(fmt.Sprintf("  Issue #%d updated", m.updated.Number)))
	b.WriteString("\n\n")
	b.WriteString("  " + m.updated.Title)
	b.WriteString("\n")
	b.WriteString("  " + dimStyle.Render(m.updated.HTMLURL))
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("  o open in browser · any other key to return to detail"))

	return b.String()
}
