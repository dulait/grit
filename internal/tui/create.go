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
	"github.com/dulait/grit/internal/llm"
	"github.com/dulait/grit/internal/service"
)

type createStep int

const (
	stepInput createStep = iota
	stepGenerating
	stepReview
	stepCreating
	stepDone
)

const (
	fieldTitle = iota
	fieldPrompt
	fieldLabels
	fieldAssignees
	fieldCount
)

type createModel struct {
	deps       Dependencies
	step       createStep
	inputs     []textinput.Model
	focusIndex int
	generated  *llm.GeneratedIssue
	assignees  []string
	created    *github.Issue
	spinner    spinner.Model
	err        error
	width      int
	height     int
}

func newCreateModel(deps Dependencies, width, height int) createModel {
	inputs := make([]textinput.Model, fieldCount)

	inputs[fieldTitle] = newInput("Issue title", 120)
	inputs[fieldPrompt] = newInput("Prompt or description for LLM", 256)
	inputs[fieldLabels] = newInput("Labels (comma-separated, optional)", 120)
	inputs[fieldAssignees] = newInput("Assignees (comma-separated, optional)", 120)

	inputs[fieldTitle].Focus()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

	return createModel{
		deps:   deps,
		step:   stepInput,
		inputs: inputs,
		spinner: s,
		width:  width,
		height: height,
	}
}

func newInput(placeholder string, charLimit int) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = charLimit
	ti.Width = 60
	return ti
}

func (m createModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m createModel) Update(msg tea.Msg) (createModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch m.step {
		case stepInput:
			return m.updateInput(msg)
		case stepReview:
			return m.updateReview(msg)
		case stepDone:
			if msg.String() == "o" && m.created != nil {
				openBrowser(m.created.HTMLURL)
				return m, nil
			}
			return m, func() tea.Msg { return navigateToListMsg{} }
		}
		if m.err != nil {
			m.err = nil
			m.step = stepInput
			return m, nil
		}

	case spinner.TickMsg:
		if m.step == stepGenerating || m.step == stepCreating {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case issueGeneratedMsg:
		m.generated = msg.issue
		m.step = stepReview

	case issueCreatedMsg:
		m.created = msg.issue
		m.step = stepDone

	case errMsg:
		m.err = msg.err
		m.step = stepInput
	}

	if m.step == stepInput {
		return m.updateInputFields(msg)
	}

	return m, nil
}

func (m createModel) updateInput(msg tea.KeyMsg) (createModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m, func() tea.Msg { return navigateToListMsg{} }
	case "tab", "down":
		m.focusIndex = (m.focusIndex + 1) % fieldCount
		return m.syncFocus(), nil
	case "shift+tab", "up":
		m.focusIndex = (m.focusIndex - 1 + fieldCount) % fieldCount
		return m.syncFocus(), nil
	case "ctrl+g":
		if m.inputTitle() == "" && m.inputPrompt() == "" {
			return m, nil
		}
		m.step = stepGenerating
		return m, tea.Batch(m.generate(true), m.spinner.Tick)
	case "ctrl+s":
		if m.inputTitle() == "" {
			return m, nil
		}
		m.step = stepCreating
		m.assignees = parseCSVInput(m.inputs[fieldAssignees].Value())
		return m, tea.Batch(m.createDirect(), m.spinner.Tick)
	}

	return m.updateInputFields(msg)
}

func (m createModel) updateReview(msg tea.KeyMsg) (createModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.step = stepInput
		m.generated = nil
		return m, nil
	case "enter", "ctrl+s":
		m.step = stepCreating
		m.assignees = parseCSVInput(m.inputs[fieldAssignees].Value())
		return m, tea.Batch(m.createFromGenerated(), m.spinner.Tick)
	}
	return m, nil
}

func (m createModel) syncFocus() createModel {
	for i := range m.inputs {
		m.inputs[i].Blur()
	}
	m.inputs[m.focusIndex].Focus()
	return m
}

func (m createModel) updateInputFields(msg tea.Msg) (createModel, tea.Cmd) {
	var cmds []tea.Cmd
	for i := range m.inputs {
		var cmd tea.Cmd
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m createModel) inputTitle() string {
	return strings.TrimSpace(m.inputs[fieldTitle].Value())
}

func (m createModel) inputPrompt() string {
	return strings.TrimSpace(m.inputs[fieldPrompt].Value())
}

func (m createModel) buildInput() service.IssueInput {
	return service.IssueInput{
		Title:       m.inputTitle(),
		Prompt:      m.inputPrompt(),
		Description: m.inputPrompt(),
		Labels:      parseCSVInput(m.inputs[fieldLabels].Value()),
		Assignees:   parseCSVInput(m.inputs[fieldAssignees].Value()),
	}
}

func (m createModel) generate(enhance bool) tea.Cmd {
	input := m.buildInput()
	deps := m.deps
	return func() tea.Msg {
		svc := deps.IssueService()
		issue, err := svc.GenerateIssue(context.Background(), input, enhance)
		if err != nil {
			return errMsg{err: err}
		}
		return issueGeneratedMsg{issue: issue}
	}
}

func (m createModel) createDirect() tea.Cmd {
	input := m.buildInput()
	deps := m.deps
	assignees := m.assignees
	return func() tea.Msg {
		svc := deps.IssueService()
		generated, err := svc.GenerateIssue(context.Background(), input, false)
		if err != nil {
			return errMsg{err: err}
		}
		created, err := svc.CreateIssue(context.Background(), generated, assignees)
		if err != nil {
			return errMsg{err: err}
		}
		return issueCreatedMsg{issue: created}
	}
}

func (m createModel) createFromGenerated() tea.Cmd {
	generated := m.generated
	deps := m.deps
	assignees := m.assignees
	return func() tea.Msg {
		svc := deps.IssueService()
		created, err := svc.CreateIssue(context.Background(), generated, assignees)
		if err != nil {
			return errMsg{err: err}
		}
		return issueCreatedMsg{issue: created}
	}
}

func (m createModel) View() string {
	var b strings.Builder

	header := headerStyle.Width(m.width).Render(" grit · Create Issue")
	b.WriteString(header)
	b.WriteString("\n\n")

	switch m.step {
	case stepInput:
		b.WriteString(m.viewInput())
	case stepGenerating:
		b.WriteString(fmt.Sprintf("  %s Generating with LLM...\n", m.spinner.View()))
	case stepReview:
		b.WriteString(m.viewReview())
	case stepCreating:
		b.WriteString(fmt.Sprintf("  %s Creating issue...\n", m.spinner.View()))
	case stepDone:
		b.WriteString(m.viewDone())
	}

	if m.err != nil && m.step == stepInput {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(fmt.Sprintf("  Error: %v", m.err)))
		b.WriteString("\n")
	}

	return b.String()
}

func (m createModel) viewInput() string {
	var b strings.Builder

	labels := []string{"  Title:", "  Prompt:", "  Labels:", "  Assignees:"}
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
	b.WriteString(helpStyle.Render("  tab/shift+tab navigate · ctrl+g generate with LLM · ctrl+s submit · esc cancel"))

	return b.String()
}

func (m createModel) viewReview() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Review Generated Issue"))
	b.WriteString("\n\n")

	b.WriteString(dimStyle.Render("  Title:"))
	b.WriteString("\n")
	b.WriteString("  " + m.generated.Title)
	b.WriteString("\n\n")

	if len(m.generated.Labels) > 0 {
		b.WriteString(dimStyle.Render("  Labels: "))
		b.WriteString(labelStyle.Render(strings.Join(m.generated.Labels, ", ")))
		b.WriteString("\n\n")
	}

	b.WriteString(dimStyle.Render("  Body:"))
	b.WriteString("\n")
	body := m.generated.Body
	if len(body) > 500 {
		body = body[:497] + "..."
	}
	for _, line := range strings.Split(body, "\n") {
		b.WriteString("  " + line + "\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  enter/ctrl+s create · esc back to edit"))

	return b.String()
}

func (m createModel) viewDone() string {
	var b strings.Builder

	b.WriteString(successStyle.Render(fmt.Sprintf("  Issue #%d created", m.created.Number)))
	b.WriteString("\n\n")
	b.WriteString("  " + m.created.Title)
	b.WriteString("\n")
	b.WriteString("  " + dimStyle.Render(m.created.HTMLURL))
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("  o open in browser · any other key to return to list"))

	return b.String()
}

func parseCSVInput(s string) []string {
	raw := strings.Split(s, ",")
	var result []string
	for _, item := range raw {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
