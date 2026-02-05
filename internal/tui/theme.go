package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	selectedStyle    = lipgloss.NewStyle().Background(lipgloss.Color("237")).Bold(true)
	normalStyle      = lipgloss.NewStyle()
	stateOpenStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	stateClosedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	labelStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	assigneeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("117"))
	errorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	statusBarStyle   = lipgloss.NewStyle().Background(lipgloss.Color("236")).Padding(0, 1)
	helpStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	headerStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")).Background(lipgloss.Color("62")).Padding(0, 1)
	dimStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)
