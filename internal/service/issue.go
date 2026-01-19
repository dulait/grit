package service

import (
	"context"
	"fmt"

	"github.com/dulait/grit/internal/config"
	"github.com/dulait/grit/internal/github"
	"github.com/dulait/grit/internal/llm"
)

type IssueInput struct {
	Prompt      string
	Title       string
	Description string
	Labels      []string
	Assignees   []string
}

type IssueService struct {
	github github.Client
	llm    llm.Client
	cfg    *config.Config
}

func NewIssueService(ghClient github.Client, llmClient llm.Client, cfg *config.Config) *IssueService {
	return &IssueService{
		github: ghClient,
		llm:    llmClient,
		cfg:    cfg,
	}
}

func (s *IssueService) GenerateIssue(ctx context.Context, input IssueInput, enhance bool) (*llm.GeneratedIssue, error) {
	if !enhance && input.Title != "" && input.Description != "" {
		return &llm.GeneratedIssue{
			Title:  input.Title,
			Body:   input.Description,
			Labels: input.Labels,
		}, nil
	}

	if s.llm == nil {
		return &llm.GeneratedIssue{
			Title:  input.Title,
			Body:   input.Description,
			Labels: input.Labels,
		}, nil
	}

	req := llm.IssueRequest{
		UserPrompt:      input.Prompt,
		TitleHint:       input.Title,
		DescriptionHint: input.Description,
		RepoContext:     fmt.Sprintf("%s/%s", s.cfg.Project.Owner, s.cfg.Project.Repo),
		IssuePrefix:     s.cfg.Project.IssuePrefix,
		AllowedLabels:   s.cfg.Project.Labels,
		GenerateTitle:   input.Title == "",
		GenerateBody:    true,
		SuggestLabels:   len(input.Labels) == 0 && len(s.cfg.Project.Labels) > 0,
	}

	issue, err := s.llm.GenerateIssue(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("generating issue: %w", err)
	}

	if input.Title != "" {
		issue.Title = input.Title
	}

	if len(input.Labels) > 0 {
		issue.Labels = input.Labels
	}

	return issue, nil
}

func (s *IssueService) CreateIssue(ctx context.Context, issue *llm.GeneratedIssue, assignees []string) (*github.Issue, error) {
	req := github.CreateIssueRequest{
		Title:     issue.Title,
		Body:      issue.Body,
		Labels:    issue.Labels,
		Assignees: assignees,
	}

	created, err := s.github.CreateIssue(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("creating issue on github: %w", err)
	}

	return created, nil
}

func (s *IssueService) CloseIssue(ctx context.Context, number int, comment string) (*github.Issue, error) {
	closed, err := s.github.CloseIssue(ctx, number, comment)
	if err != nil {
		return nil, fmt.Errorf("closing issue: %w", err)
	}
	return closed, nil
}

func (s *IssueService) AddComment(ctx context.Context, number int, userPrompt string) (*github.IssueComment, error) {
	issue, err := s.github.GetIssue(ctx, number)
	if err != nil {
		return nil, fmt.Errorf("fetching issue: %w", err)
	}

	issueContext := fmt.Sprintf("Title: %s\n\nBody:\n%s", issue.Title, issue.Body)

	commentBody, err := s.llm.GenerateComment(ctx, issueContext, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("generating comment: %w", err)
	}

	comment, err := s.github.AddComment(ctx, number, commentBody)
	if err != nil {
		return nil, fmt.Errorf("adding comment: %w", err)
	}

	return comment, nil
}

func (s *IssueService) AssignIssue(ctx context.Context, number int, assignees []string) (*github.Issue, error) {
	issue, err := s.github.AssignIssue(ctx, number, assignees)
	if err != nil {
		return nil, fmt.Errorf("assigning issue: %w", err)
	}
	return issue, nil
}

func (s *IssueService) LinkIssue(ctx context.Context, number, targetNumber int, linkType string) error {
	linkText := fmt.Sprintf("Related to #%d", targetNumber)
	switch linkType {
	case "blocks":
		linkText = fmt.Sprintf("Blocks #%d", targetNumber)
	case "blocked-by":
		linkText = fmt.Sprintf("Blocked by #%d", targetNumber)
	case "duplicates":
		linkText = fmt.Sprintf("Duplicates #%d", targetNumber)
	case "parent":
		linkText = fmt.Sprintf("Parent of #%d", targetNumber)
	case "child":
		linkText = fmt.Sprintf("Child of #%d", targetNumber)
	}

	_, err := s.github.AddComment(ctx, number, linkText)
	if err != nil {
		return fmt.Errorf("adding link comment: %w", err)
	}

	return nil
}

func (s *IssueService) CreateSubIssue(ctx context.Context, parentNumber int, generated *llm.GeneratedIssue, assignees []string) (*github.Issue, error) {
	body := fmt.Sprintf("Part of #%d\n\n---\n\n%s", parentNumber, generated.Body)

	req := github.CreateIssueRequest{
		Title:     generated.Title,
		Body:      body,
		Labels:    generated.Labels,
		Assignees: assignees,
	}

	created, err := s.github.CreateIssue(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("creating sub-issue: %w", err)
	}

	return created, nil
}
