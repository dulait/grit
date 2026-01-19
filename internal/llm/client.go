package llm

import "context"

// Client defines the interface for LLM-powered content generation.
type Client interface {
	GenerateIssue(ctx context.Context, req IssueRequest) (*GeneratedIssue, error)
	GenerateComment(ctx context.Context, issueContext, userPrompt string) (string, error)
}
