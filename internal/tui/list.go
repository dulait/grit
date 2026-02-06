package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dulait/grit/internal/github"
)

type listModel struct {
	deps        Dependencies
	issues      []github.Issue
	cursor      int
	offset      int
	page        int
	perPage     int
	state       string
	loading     bool
	spinner     spinner.Model
	err         error
	width       int
	height      int
	searchQuery string
	searching   bool
	searchInput textinput.Model
	searchSeq   int
	totalCount  int
}

func newListModel(deps Dependencies) listModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

	ti := textinput.New()
	ti.Placeholder = "search issues..."
	ti.CharLimit = 128

	return listModel{
		deps:        deps,
		page:        1,
		perPage:     20,
		state:       "open",
		loading:     true,
		spinner:     s,
		searchInput: ti,
	}
}

func (m listModel) Init() tea.Cmd {
	return tea.Batch(m.loadIssues(), m.spinner.Tick)
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

func (m listModel) loadIssues() tea.Cmd {
	if m.searchQuery != "" {
		return m.searchIssues()
	}
	return m.fetchIssues()
}

func (m listModel) searchIssues() tea.Cmd {
	return func() tea.Msg {
		svc := m.deps.IssueServiceWithoutLLM()
		req := github.SearchIssuesRequest{
			Query:   m.searchQuery,
			State:   m.state,
			PerPage: m.perPage,
			Page:    m.page,
		}
		resp, err := svc.SearchIssues(context.Background(), req)
		if err != nil {
			return errMsg{err: err}
		}
		return searchResultsMsg{issues: resp.Items, totalCount: resp.TotalCount, page: m.page}
	}
}

func (m listModel) updateSearchInput(msg tea.Msg) (listModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.searching = false
			m.searchInput.Blur()
			return m, nil
		case "esc":
			m.searching = false
			m.searchInput.Blur()
			m.searchInput.SetValue("")
			if m.searchQuery != "" {
				m.searchQuery = ""
				m.totalCount = 0
				m.page = 1
				m.loading = true
				return m, tea.Batch(m.loadIssues(), m.spinner.Tick)
			}
			return m, nil
		}
	}

	prevValue := m.searchInput.Value()
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)

	if m.searchInput.Value() != prevValue {
		m.searchSeq++
		seq := m.searchSeq
		tickCmd := tea.Tick(300*time.Millisecond, func(t time.Time) tea.Msg {
			return searchTickMsg{seq: seq}
		})
		return m, tea.Batch(cmd, tickCmd)
	}

	return m, cmd
}

func (m listModel) hasNextPage() bool {
	if m.searchQuery != "" {
		return m.page*m.perPage < m.totalCount
	}
	return len(m.issues) == m.perPage
}

func (m listModel) Update(msg tea.Msg) (listModel, tea.Cmd) {
	if m.searching {
		if _, ok := msg.(tea.KeyMsg); ok {
			return m.updateSearchInput(msg)
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.adjustOffset()

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
		m.offset = 0
		m.err = nil

	case searchResultsMsg:
		m.issues = msg.issues
		m.page = msg.page
		m.totalCount = msg.totalCount
		m.loading = false
		m.cursor = 0
		m.offset = 0
		m.err = nil

	case searchTickMsg:
		if msg.seq != m.searchSeq {
			return m, nil
		}
		query := strings.TrimSpace(m.searchInput.Value())
		if query == "" {
			if m.searchQuery != "" {
				m.searchQuery = ""
				m.totalCount = 0
				m.page = 1
				m.loading = true
				return m, tea.Batch(m.loadIssues(), m.spinner.Tick)
			}
			return m, nil
		}
		m.searchQuery = query
		m.page = 1
		m.loading = true
		return m, tea.Batch(m.loadIssues(), m.spinner.Tick)

	case errMsg:
		m.err = msg.err
		m.loading = false

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		switch {
		case key.Matches(msg, listKeys.Search):
			m.searching = true
			m.searchInput.SetValue("")
			m.searchInput.Focus()
			return m, m.searchInput.Cursor.BlinkCmd()
		case msg.String() == "esc":
			if m.searchQuery != "" {
				m.searchQuery = ""
				m.searchInput.SetValue("")
				m.totalCount = 0
				m.page = 1
				m.loading = true
				return m, tea.Batch(m.loadIssues(), m.spinner.Tick)
			}
		case key.Matches(msg, listKeys.Down):
			if m.cursor < len(m.issues)-1 {
				m.cursor++
				m.adjustOffset()
			}
		case key.Matches(msg, listKeys.Up):
			if m.cursor > 0 {
				m.cursor--
				m.adjustOffset()
			}
		case key.Matches(msg, listKeys.Open):
			if len(m.issues) > 0 {
				issue := m.issues[m.cursor]
				return m, func() tea.Msg {
					return navigateToDetailMsg{issueNumber: issue.Number}
				}
			}
		case key.Matches(msg, listKeys.NextPage):
			if m.hasNextPage() {
				m.page++
				m.loading = true
				return m, tea.Batch(m.loadIssues(), m.spinner.Tick)
			}
		case key.Matches(msg, listKeys.PrevPage):
			if m.page > 1 {
				m.page--
				m.loading = true
				return m, tea.Batch(m.loadIssues(), m.spinner.Tick)
			}
		case key.Matches(msg, listKeys.Refresh):
			m.loading = true
			return m, tea.Batch(m.loadIssues(), m.spinner.Tick)
		case key.Matches(msg, listKeys.FilterOpen):
			m.state = "open"
			m.page = 1
			m.loading = true
			return m, tea.Batch(m.loadIssues(), m.spinner.Tick)
		case key.Matches(msg, listKeys.FilterClosed):
			m.state = "closed"
			m.page = 1
			m.loading = true
			return m, tea.Batch(m.loadIssues(), m.spinner.Tick)
		case key.Matches(msg, listKeys.FilterAll):
			m.state = "all"
			m.page = 1
			m.loading = true
			return m, tea.Batch(m.loadIssues(), m.spinner.Tick)
		case key.Matches(msg, listKeys.Create):
			return m, func() tea.Msg { return navigateToCreateMsg{} }
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

	if m.searching {
		b.WriteString(fmt.Sprintf("  / %s\n", m.searchInput.View()))
	} else if m.searchQuery != "" {
		b.WriteString(dimStyle.Render(fmt.Sprintf("  search: %s (%d results)", m.searchQuery, m.totalCount)))
		b.WriteString("\n")
	}

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
		if m.searchQuery != "" {
			b.WriteString(dimStyle.Render("  No matching issues found."))
		} else {
			b.WriteString(dimStyle.Render("  No issues found."))
		}
		b.WriteString("\n")
	} else {
		visible := m.visibleRows()
		end := m.offset + visible
		if end > len(m.issues) {
			end = len(m.issues)
		}

		if m.offset > 0 {
			b.WriteString(dimStyle.Render(fmt.Sprintf("  ↑ %d more above", m.offset)))
			b.WriteString("\n")
		}

		for i := m.offset; i < end; i++ {
			b.WriteString(m.renderIssueRow(i, m.issues[i]))
			b.WriteString("\n")
		}

		remaining := len(m.issues) - end
		if remaining > 0 {
			b.WriteString(dimStyle.Render(fmt.Sprintf("  ↓ %d more below", remaining)))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(renderStatusBar(repo, m.state, m.page, len(m.issues), m.width))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(m.helpText()))

	return b.String()
}

func (m listModel) helpText() string {
	if m.searching {
		return "  type to search · enter done · esc cancel"
	}
	if m.searchQuery != "" {
		return "  j/k navigate · enter open · n/p page · 1/2/3 filter · / new search · esc clear search · ? help · q quit"
	}
	return "  j/k navigate · enter open · c create · n/p page · 1/2/3 filter · / search · ? help · q quit"
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

func (m listModel) visibleRows() int {
	rows := m.height - 6
	if m.searching || m.searchQuery != "" {
		rows--
	}
	if rows < 1 {
		return 1
	}
	return rows
}

func (m *listModel) adjustOffset() {
	visible := m.visibleRows()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+visible {
		m.offset = m.cursor - visible + 1
	}
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
