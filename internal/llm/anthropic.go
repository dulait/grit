package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AnthropicClient implements Client using the Anthropic API.
type AnthropicClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewAnthropicClient creates a new Anthropic API client.
func NewAnthropicClient(apiKey, model string) *AnthropicClient {
	return &AnthropicClient{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
	System    string             `json:"system,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *AnthropicClient) GenerateIssue(ctx context.Context, req IssueRequest) (*GeneratedIssue, error) {
	var titleInstruction string
	if req.TitleHint != "" {
		titleInstruction = fmt.Sprintf("Use this as the title (keep it concise, under 80 chars): %s", req.TitleHint)
	} else if req.GenerateTitle {
		titleInstruction = "Generate a clear, concise title (under 80 characters)"
	}

	var bodyInstruction string
	if req.DescriptionHint != "" {
		bodyInstruction = fmt.Sprintf(`Expand and structure the following into a well-formatted GitHub issue body:
"%s"

Include relevant sections such as:
- Description (what needs to be done)
- Acceptance Criteria (if applicable)
- Technical Notes (if applicable)

Use markdown formatting.`, req.DescriptionHint)
	} else if req.UserPrompt != "" {
		bodyInstruction = `Generate a well-structured GitHub issue body with:
- Description
- Acceptance Criteria (if applicable)
- Technical Notes (if applicable)

Use markdown formatting.`
	} else {
		bodyInstruction = "Generate a brief issue body based on the title."
	}

	labelInstruction := "Set labels to an empty array []."
	if req.SuggestLabels && len(req.AllowedLabels) > 0 {
		labelInstruction = fmt.Sprintf("Suggest labels ONLY from: %v. If none fit, use empty array.", req.AllowedLabels)
	}

	systemPrompt := fmt.Sprintf(`You are a GitHub issue generator that creates well-structured issues.

%s

%s

%s

Respond with ONLY valid JSON:
{
  "title": "Issue title",
  "body": "Markdown formatted body",
  "labels": []
}`, titleInstruction, bodyInstruction, labelInstruction)

	userMessage := req.UserPrompt
	if userMessage == "" && req.DescriptionHint != "" {
		userMessage = req.DescriptionHint
	}
	if userMessage == "" && req.TitleHint != "" {
		userMessage = req.TitleHint
	}
	if userMessage == "" {
		userMessage = "Generate a GitHub issue."
	}

	resp, err := c.call(ctx, systemPrompt, userMessage)
	if err != nil {
		return nil, err
	}

	var issue GeneratedIssue
	if err := json.Unmarshal([]byte(resp), &issue); err != nil {
		return nil, fmt.Errorf("parsing LLM response: %w\nraw response: %s", err, resp)
	}

	if req.IssuePrefix != "" && issue.Title != "" {
		issue.Title = req.IssuePrefix + issue.Title
	}

	if !req.SuggestLabels {
		issue.Labels = nil
	} else if len(req.AllowedLabels) > 0 {
		issue.Labels = filterLabels(issue.Labels, req.AllowedLabels)
	}

	return &issue, nil
}

func (c *AnthropicClient) GenerateComment(ctx context.Context, issueContext, userPrompt string) (string, error) {
	systemPrompt := `You are helping write a GitHub issue comment. Write a clear, professional comment based on the user's intent. Respond with ONLY the comment text, no JSON wrapping.`

	userMessage := fmt.Sprintf("Issue context:\n%s\n\nWrite a comment that: %s", issueContext, userPrompt)

	return c.call(ctx, systemPrompt, userMessage)
}

func (c *AnthropicClient) call(ctx context.Context, system, user string) (string, error) {
	reqBody := anthropicRequest{
		Model:     c.model,
		MaxTokens: 1024,
		System:    system,
		Messages: []anthropicMessage{
			{Role: "user", Content: user},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	var anthropicResp anthropicResponse
	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	if anthropicResp.Error != nil {
		return "", fmt.Errorf("anthropic api error: %s", anthropicResp.Error.Message)
	}

	if len(anthropicResp.Content) == 0 {
		return "", fmt.Errorf("empty response from anthropic")
	}

	return anthropicResp.Content[0].Text, nil
}
