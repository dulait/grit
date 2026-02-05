package tui

import (
	"github.com/dulait/grit/internal/config"
	"github.com/dulait/grit/internal/github"
	"github.com/dulait/grit/internal/llm"
	"github.com/dulait/grit/internal/service"
)

type Dependencies struct {
	Config       *config.Config
	GitHubClient github.Client
	LLMClient    llm.Client
}

func (d Dependencies) IssueService() *service.IssueService {
	return service.NewIssueService(d.GitHubClient, d.LLMClient, d.Config)
}

func (d Dependencies) IssueServiceWithoutLLM() *service.IssueService {
	return service.NewIssueService(d.GitHubClient, nil, d.Config)
}
