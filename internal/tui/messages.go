package tui

import (
	"github.com/dulait/grit/internal/github"
	"github.com/dulait/grit/internal/llm"
)

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

type startActionMsg struct {
	kind        actionKind
	issueNumber int
}

type actionSuccessMsg struct {
	text string
}

type actionDoneMsg struct{}

type actionCancelledMsg struct{}

type navigateToCreateMsg struct{}

type issueGeneratedMsg struct {
	issue *llm.GeneratedIssue
}

type issueCreatedMsg struct {
	issue *github.Issue
}

type searchResultsMsg struct {
	issues     []github.Issue
	totalCount int
	page       int
}

type searchTickMsg struct {
	seq int
}

type navigateToEditMsg struct {
	issueNumber int
}

type issueUpdatedMsg struct {
	issue *github.Issue
}
