package llm

import "context"

type Client interface {
	GenerateIssue(ctx context.Context, req IssueRequest) (*GeneratedIssue, error)
	GenerateComment(ctx context.Context, issueContext, userPrompt string) (string, error)
}
