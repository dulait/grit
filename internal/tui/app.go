package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	screenList screen = iota
)

type app struct {
	deps     Dependencies
	screen   screen
	list     listModel
	showHelp bool
	width    int
	height   int
}

func newApp(deps Dependencies) app {
	return app{
		deps:   deps,
		screen: screenList,
		list:   newListModel(deps),
	}
}

func (a app) Init() tea.Cmd {
	return a.list.Init()
}

func (a app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}

		if msg.String() == "?" {
			a.showHelp = !a.showHelp
			return a, nil
		}

		if a.showHelp {
			return a, nil
		}

		if msg.String() == "q" {
			return a, tea.Quit
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	}

	if a.showHelp {
		return a, nil
	}

	switch a.screen {
	case screenList:
		var cmd tea.Cmd
		a.list, cmd = a.list.Update(msg)
		return a, cmd
	}

	return a, nil
}

func (a app) View() string {
	if a.showHelp {
		return renderHelp(a.width)
	}

	switch a.screen {
	case screenList:
		return a.list.View()
	}

	return ""
}

func Run(deps Dependencies) error {
	p := tea.NewProgram(newApp(deps), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
