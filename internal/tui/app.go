package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	screenList screen = iota
	screenDetail
	screenCreate
	screenEdit
)

type app struct {
	deps     Dependencies
	screen   screen
	list     listModel
	detail   detailModel
	create   createModel
	edit     editModel
	action   *actionModel
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

		if a.action != nil {
			return a.updateAction(msg)
		}

		if a.screen == screenCreate || a.screen == screenEdit {
			break
		}

		if a.screen == screenList && a.list.searching {
			break
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
		a.list = newListModel(a.deps)
		a.list.width = a.width
		a.list.height = a.height
		a.screen = screenList
		return a, a.list.Init()

	case navigateToEditMsg:
		a.edit = newEditModel(a.deps, msg.issueNumber, a.width, a.height)
		a.screen = screenEdit
		return a, a.edit.Init()

	case navigateToCreateMsg:
		a.create = newCreateModel(a.deps, a.width, a.height)
		a.screen = screenCreate
		return a, a.create.Init()

	case startActionMsg:
		action := newActionModel(a.deps, msg.kind, msg.issueNumber, a.width, a.height)
		a.action = &action
		return a, a.action.Init()

	case actionDoneMsg:
		a.action = nil
		a.detail = newDetailModel(a.deps, a.detail.issueNumber)
		a.detail.width = a.width
		a.detail.height = a.height
		return a, a.detail.Init()

	case actionCancelledMsg:
		a.action = nil
		return a, nil
	}

	if a.action != nil {
		return a.updateAction(msg)
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
	case screenCreate:
		var cmd tea.Cmd
		a.create, cmd = a.create.Update(msg)
		return a, cmd
	case screenEdit:
		var cmd tea.Cmd
		a.edit, cmd = a.edit.Update(msg)
		return a, cmd
	}

	return a, nil
}

func (a app) updateAction(msg tea.Msg) (tea.Model, tea.Cmd) {
	action, cmd := a.action.Update(msg)
	a.action = &action
	return a, cmd
}

func (a app) View() string {
	if a.action != nil {
		return a.action.View()
	}

	if a.showHelp {
		return renderHelp(a.width, a.screen)
	}

	switch a.screen {
	case screenList:
		return a.list.View()
	case screenDetail:
		return a.detail.View()
	case screenCreate:
		return a.create.View()
	case screenEdit:
		return a.edit.View()
	}

	return ""
}

func Run(deps Dependencies) error {
	p := tea.NewProgram(newApp(deps), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
