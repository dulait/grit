package tui

import "github.com/dulait/grit/internal/github"

type issuesLoadedMsg struct {
	issues []github.Issue
	page   int
}

type errMsg struct {
	err error
}

type statusMsg struct {
	text string
}
