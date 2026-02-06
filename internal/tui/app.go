package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	screenList screen = iota
	screenDetail
)

type app struct {
	deps     Dependencies
	screen   screen
	list     listModel
	detail   detailModel
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

	case navigateToDetailMsg:
		a.detail = newDetailModel(a.deps, msg.issueNumber)
		a.detail.width = a.width
		a.detail.height = a.height
		a.screen = screenDetail
		return a, a.detail.Init()

	case navigateToListMsg:
		a.screen = screenList
		return a, nil
	}

	if a.showHelp {
		return a, nil
	}

	switch a.screen {
	case screenList:
		var cmd tea.Cmd
		a.list, cmd = a.list.Update(msg)
		return a, cmd
	case screenDetail:
		var cmd tea.Cmd
		a.detail, cmd = a.detail.Update(msg)
		return a, cmd
	}

	return a, nil
}

func (a app) View() string {
	if a.showHelp {
		return renderHelp(a.width, a.screen)
	}

	switch a.screen {
	case screenList:
		return a.list.View()
	case screenDetail:
		return a.detail.View()
	}

	return ""
}

func Run(deps Dependencies) error {
	p := tea.NewProgram(newApp(deps), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
