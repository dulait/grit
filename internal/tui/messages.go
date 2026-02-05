package tui

import "github.com/dulait/grit/internal/github"

type issuesLoadedMsg struct {
	issues []github.Issue
	page   int
}

type issueDetailLoadedMsg struct {
	issue *github.Issue
}

type navigateToDetailMsg struct {
	issueNumber int
}

type navigateToListMsg struct{}

type errMsg struct {
	err error
}

type statusMsg struct {
	text string
}
