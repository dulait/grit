package tui

import "github.com/charmbracelet/bubbles/key"

type listKeyMap struct {
	Up           key.Binding
	Down         key.Binding
	Open         key.Binding
	Create       key.Binding
	NextPage     key.Binding
	PrevPage     key.Binding
	Refresh      key.Binding
	FilterOpen   key.Binding
	FilterClosed key.Binding
	FilterAll    key.Binding
	Help         key.Binding
	Quit         key.Binding
}

var listKeys = listKeyMap{
	Up:           key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "up")),
	Down:         key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "down")),
	Open:         key.NewBinding(key.WithKeys("enter", "l"), key.WithHelp("enter/l", "open")),
	Create:       key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "create issue")),
	NextPage:     key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "next page")),
	PrevPage:     key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "prev page")),
	Refresh:      key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
	FilterOpen:   key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "open")),
	FilterClosed: key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "closed")),
	FilterAll:    key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "all")),
	Help:         key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Quit:         key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
}

type detailKeyMap struct {
	Up          key.Binding
	Down        key.Binding
	HalfPageUp  key.Binding
	HalfPageDwn key.Binding
	Back        key.Binding
	OpenBrowser key.Binding
	Close       key.Binding
	Assign      key.Binding
	Comment     key.Binding
	Help        key.Binding
	Quit        key.Binding
}

var detailKeys = detailKeyMap{
	Up:          key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "scroll up")),
	Down:        key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "scroll down")),
	HalfPageUp:  key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl+u", "half page up")),
	HalfPageDwn: key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "half page down")),
	Back:        key.NewBinding(key.WithKeys("esc", "h", "backspace"), key.WithHelp("esc/h", "back")),
	OpenBrowser: key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open in browser")),
	Close:       key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "close issue")),
	Assign:      key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "assign")),
	Comment:     key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "comment")),
	Help:        key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Quit:        key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
}
