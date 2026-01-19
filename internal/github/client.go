package github

import "context"

// Client defines the interface for GitHub API operations.
type Client interface {
	CreateIssue(ctx context.Context, req CreateIssueRequest) (*Issue, error)
	CloseIssue(ctx context.Context, number int, comment string) (*Issue, error)
	GetIssue(ctx context.Context, number int) (*Issue, error)
	AddComment(ctx context.Context, number int, body string) (*IssueComment, error)
	AssignIssue(ctx context.Context, number int, assignees []string) (*Issue, error)
}
